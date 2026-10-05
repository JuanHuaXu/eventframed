import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {switchWeights} from './switch-distribution.mjs';
const arms=[0,1,2,3,10,11],priors=[Array(6).fill(1/6),[.95,.01,.01,.01,.01,.01]];
function predict(steps,t,prior,switching) {
  const rows=steps.slice(0,t).map((s,j)=>({p:arms.map(i=>s.P[i]),y:!s.Missing&&j+s.Delay<=t?s.Y:null}));
  const {weights}=switchWeights(rows,prior,switching);
  assert(Math.abs(weights.reduce((a,b)=>a+b,0)-1)<1e-11);
  return weights.reduce((v,w,i)=>v+w*steps[t].P[arms[i]],0);
}
const input=process.argv[2],stream=fs.createReadStream(input),digest=crypto.createHash('sha256');stream.on('data',b=>digest.update(b));
let header,checks=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})) {
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line);assert.equal(r.Steps.length,256);
  const scores=Array.from({length:5},()=>[0,0]),accuracy=scores.map(()=>[0,0]),logLoss=scores.map(()=>[0,0]);
  for(let t=0;t<256;t++) {
    const s=r.Steps[t],ps=priors.flatMap(prior=>[true,false].map(sw=>predict(r.Steps,t,prior,sw))).concat(s.P[12]);
    for(let arm=0;arm<5;arm++) {
      const p=ps[arm],q=s.Q,v=Math.max(1e-12,Math.min(1-1e-12,p));assert(p>=0&&p<=1);
      for(const segment of t>=192?[0,1]:[0]) {
        const n=segment===0?256:64;
        scores[arm][segment]+=((p-q)**2+q*(1-q))/n;
        accuracy[arm][segment]+=(p>=.5?q:1-q)/n;
        logLoss[arm][segment]+=(-q*Math.log(v)-(1-q)*Math.log1p(-v))/n;
      }
    }
    if(r.Index===0&&[0,32,160,255].includes(t)) {
      const altered=r.Steps.map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
      for(let arm=0;arm<4;arm++){assert.equal(predict(altered,t,priors[Math.floor(arm/2)],arm%2===0),ps[arm]);checks++;}
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores,accuracy,logLoss});
}
assert.equal(records.length,2688);assert.equal(checks,1344);
const mean=a=>a.reduce((a,b)=>a+b,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++) {
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  const meta={phase,case:c,schedule,segment};
  groups.push({...meta,brier:[0,1,2,3,4].map(a=>mean(rs.map(r=>r.scores[a][segment])))});
  for(const arm of [0,2])for(const control of [arm+1,4]) {
    const gain=interval(rs.map(r=>r.scores[control][segment]-r.scores[arm][segment]));
    gates.push({...meta,arm,control,type:'nonharm',...gain,pass:gain.lower>=-.01});
    if(segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,arm,control,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});
  }
}
const candidates=[0,2].map(arm=>{const gs=gates.filter(g=>g.arm===arm);return{arm,nonharm:gs.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:336,gains:gs.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:64,status:gs.every(g=>g.pass)?'PASS':'FAIL'};});
const paths=[import.meta.filename,'research/switch-distribution.mjs','research/switch-distribution-test.mjs','docs/experiments/mmm-switch-distribution-v120-protocol.md'];
console.log(JSON.stringify({scope:'Consumed fixed-forecast switch mixture; not fresh validation',artifactSHA256:digest.digest('hex'),hashes:Object.fromEntries(paths.map(p=>[p,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')])),arms,priors,checks,candidates,records,groups,gates}));
