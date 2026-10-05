import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const arms=[0,1,2,3,10,11],clocks=[128,160,192],windows=[32,64,128];
const clip=p=>Math.max(1e-12,Math.min(1-1e-12,p));
function terms(s,arm){const p=clip(s.P[arm]),b=clip(s.P[0]);return[Math.log(p/b),Math.log((1-p)/(1-b))];}
function evidence(steps,t,width,arm){
  const out={available:0,late:0,missing:0,actual:0,expected:0,fullExpected:0};
  for(let j=Math.max(0,t-width);j<t;j++){
    const s=steps[j],[a,b]=terms(s,arm),e=s.Q*a+(1-s.Q)*b;out.fullExpected+=e;
    if(s.Missing){out.missing++;continue;}
    if(j+s.Delay>t){out.late++;continue;}
    out.available++;out.actual+=s.Y?a:b;out.expected+=e;
  }
  assert.equal(out.available+out.late+out.missing,Math.min(t,width));return out;
}
const brier=(p,q)=>(p-q)**2+q*(1-q);
function choice(steps,t){
  const scores=arms.map(i=>steps.slice(t,t+64).reduce((v,s)=>v+brier(s.P[i],s.Q),0)/64);
  const index=scores.indexOf(Math.min(...scores)),arm=arms[index];
  const logGain=steps.slice(t,t+64).reduce((v,s)=>{const[a,b]=terms(s,arm);return v+s.Q*a+(1-s.Q)*b;},0)/64;
  return{arm,gain:scores[0]-scores[index],logGain};
}
// Independent short-product identity, avoiding long-product underflow.
let tests=0;
for(let seed=0;seed<32;seed++){
  const steps=Array.from({length:6},(_,j)=>({P:[.2+(seed%5)/10,.15+j/10],Y:(seed+j)%2,Q:.3,Delay:j%3,Missing:j%4===0}));
  for(let t=1;t<=6;t++){
    let ratio=1;for(let j=0;j<t;j++){const s=steps[j];if(s.Missing||j+s.Delay>t)continue;ratio*=s.Y?s.P[1]/s.P[0]:(1-s.P[1])/(1-s.P[0]);}
    assert(Math.abs(Math.log(ratio)-evidence(steps,t,6,1).actual)<1e-12);tests++;
  }
}
const input=process.argv[2],stream=fs.createReadStream(input),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));
const records=[];let header,asOf=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line);assert.equal(r.Steps.length,256);const probes=[];
  for(const t of clocks){
    const selected=choice(r.Steps,t);
    for(const width of windows){
      const e=evidence(r.Steps,t,width,selected.arm);probes.push({t,width,...selected,...e});
      if(r.Index===0){
        const poisoned=r.Steps.map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
        assert.equal(evidence(poisoned,t,width,selected.arm).actual,e.actual);
        const future=r.Steps.map((s,j)=>({...s,Y:j>=t?!s.Y:s.Y}));assert.deepEqual(choice(future,t),selected);asOf++;
      }
    }
  }
  records.push({phase:r.Phase,case:r.Case,schedule:r.Schedule,index:r.Index,probes});
}
assert.equal(records.length,2688);assert.equal(asOf,756);
const mean=a=>a.length?a.reduce((a,b)=>a+b,0)/a.length:null,groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(const t of clocks)for(const width of windows){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  const ps=rs.map(r=>r.probes.find(p=>p.t===t&&p.width===width)),op=ps.filter(p=>p.gain>=.005);
  groups.push({phase,case:c,schedule,t,width,armCounts:arms.map(i=>ps.filter(p=>p.arm===i).length),means:Object.fromEntries(['gain','logGain','available','late','missing','actual','expected','fullExpected'].map(k=>[k,mean(ps.map(p=>p[k]))])),opportunities:op.length,opMeans:Object.fromEntries(['gain','available','actual','expected','fullExpected'].map(k=>[k,mean(op.map(p=>p[k]))])),opWrongExpected:op.filter(p=>p.expected<=0).length,opWrongActual:op.filter(p=>p.actual<=0).length,opStrongActual:op.filter(p=>p.actual>Math.log(19)).length,opFutureLogMismatch:op.filter(p=>p.logGain<=0).length});
}
const paths=[import.meta.filename,'docs/experiments/mmm-evidence-availability-v120-protocol.md'];
console.log(JSON.stringify({scope:'Consumed hindsight evidence diagnostic, not online confidence or acquisition validation',artifactSHA256:hash.digest('hex'),hashes:Object.fromEntries(paths.map(p=>[p,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')])),tests,asOf,arms,clocks,windows,records,groups}));
