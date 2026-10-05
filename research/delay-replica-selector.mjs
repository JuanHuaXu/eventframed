import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const prior=[.7,.1,.1,.1],eta=.5;
const dot=(w,p)=>w.reduce((s,x,j)=>s+x*p[j],0);
const update=(w,p,y)=>{
  const v=w.map((x,j)=>x*Math.exp(-eta*(p[j]-y)**2));const z=v.reduce((a,b)=>a+b,0);
  return v.map(x=>x/z);
};
class Replicas {
  constructor(){this.slots=[];this.issued=0;this.received=0;this.expired=0;}
  issue(origin,experts){
    assert.equal(origin,this.issued);assert.ok(experts.length===4&&experts.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
    let slot=this.slots.findIndex(s=>s.pending===null);
    if(slot<0){assert.ok(this.slots.length<33);slot=this.slots.length;this.slots.push({weights:[...prior],pending:null});}
    const s=this.slots[slot];s.pending={origin,experts:[...experts]};this.issued++;
    return {slot,origin,p:dot(s.weights,experts)};
  }
  feedback(token,y){
    assert.ok(y===0||y===1);
    const s=this.slots[token.slot];
    if(!s?.pending||s.pending.origin!==token.origin)return false;
    s.weights=update(s.weights,s.pending.experts,y);s.pending=null;this.received++;return true;
  }
  expire(clock){for(const s of this.slots)if(s.pending&&clock-s.pending.origin>=32){s.pending=null;this.expired++;}}
}
// Lifecycle falsifiers: immutable advice, one-use identity, and expired-slot reuse.
{
 const r=new Replicas(),p=[0,1,.5,.5],token=r.issue(0,p);p[0]=1;
 assert.ok(r.feedback(token,0));assert.ok(!r.feedback(token,0));
 assert.ok(r.slots[0].weights[0]>.7);
 const old=r.issue(1,[.5,.5,.5,.5]);r.expire(33);
 const fresh=r.issue(2,[.5,.5,.5,.5]);assert.equal(old.slot,fresh.slot);
 assert.ok(!r.feedback(old,1));assert.ok(r.feedback(fresh,1));
}

function replay(frames){
 const pool=new Replicas(),tokens=[],arrivalPred=[],replicaPred=[];
 let arrivalWeights=[...prior],observedReplica=0,observedBase=0,received=0;
 for(let clock=0;clock<frames.length+32;clock++){
  if(clock<frames.length){
   const advice=frames[clock].Predictions[0].experts;
   tokens.push(pool.issue(clock,advice));replicaPred.push(tokens[clock].p);
   arrivalPred.push(dot(arrivalWeights,advice));
  }
  for(let origin=Math.max(0,clock-31);origin<=Math.min(clock,frames.length-1);origin++){
   const f=frames[origin];
   if(!f.Missing&&f.Arrival===clock){
    const y=Number(f.Y),advice=f.Predictions[0].experts;
    assert.ok(pool.feedback(tokens[origin],y));arrivalWeights=update(arrivalWeights,advice,y);received++;
    observedReplica+=(replicaPred[origin]-y)**2;observedBase+=(advice[0]-y)**2;
   }
  }
  pool.expire(clock);
 }
 assert.equal(pool.received,received);assert.equal(pool.received+pool.expired,frames.length);
 // Each replica sees sequential observed rounds, with no update on censored rounds.
 const bound=pool.slots.length*Math.log(1/prior[0])/eta;
 assert.ok(observedReplica-observedBase<=bound+1e-10);
 const scores={};
 for(const [name,start] of [['full',0],['post',256]]){
  const fslice=frames.slice(start),n=fslice.length;assert.ok(n>0);
  scores[name]={};
  for(const arm of ['original','arrival','replica'])scores[name][arm]=fslice.reduce((s,f,i)=>{
   const p=arm==='original'?f.Predictions[0].p:arm==='arrival'?arrivalPred[i+start]:replicaPred[i+start];
   return s+(p-Number(f.Y))**2;
  },0)/n;
 }
 return {copies:pool.slots.length,received,expired:pool.expired,observedExcess:observedReplica-observedBase,observedBound:bound,scores};
}

const [input,output]=process.argv.slice(2),raw=fs.readFileSync(input);
const hash=crypto.createHash('sha256').update(raw).digest('hex');
assert.equal(hash,'1533a049c2146cc7839799e9a394691d96c3fdf4a1a8e8d1c504db86053bf8e6');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
assert.equal(header.Consumed,true);assert.equal(rows.length,192);
const results=[];
for(const r of rows)for(const schedule of ['Immediate','Delayed']){
 const frames=r[schedule].Frames;assert.equal(frames.length,512);
 for(let i=0;i<512;i++)assert.ok(frames[i].Arrival>=i&&frames[i].Arrival<i+32);
 const result=replay(frames);
 if(schedule==='Immediate')assert.ok(Math.abs(result.scores.full.replica-result.scores.full.arrival)<1e-12);
 results.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,...result});
}
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const summaries=[];
for(const phase of ['design','confirmation'])for(const name of [...new Set(results.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
 const a=results.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
 summaries.push({phase,case:name,schedule,n:a.length,post:Object.fromEntries(['original','arrival','replica'].map(arm=>[arm,mean(a.map(r=>r.scores.post[arm]))]))});
}
const result={sourceHash:hash,prior,eta,summaries,results,limits:'Consumed fixed-advice diagnostic. Original acquisition and expert tape held fixed. No fresh confirmation or deployed posterior claim.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(summaries.filter(r=>r.phase==='confirmation')));
