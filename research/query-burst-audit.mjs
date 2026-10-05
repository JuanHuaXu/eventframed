import assert from 'node:assert/strict';
import crypto from 'node:crypto';

const prior=[.95,.05/3,.05/3,.05/3];
const transition=(weights,ps,y)=>{
 let next=weights.map((w,k)=>w*(y===undefined?1:(y?ps[k]:1-ps[k])));
 const z=next.reduce((a,b)=>a+b,0);assert(z>0);
 next=next.map((v,k)=>.999*v/z+.001*prior[k]);
 return next;
};
const entropy=p=>{p=Math.max(1e-12,Math.min(1-1e-12,p));return -p*Math.log(p)-(1-p)*Math.log1p(-p);};

// Independent clock replay: derive availability from deliveries, never from
// the learner's recorded trigger bit. Q is used only after the policy audit.
export function auditBurstRecord(r, raw) {
 assert(!r.Error, r.Error);
 assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[raw.Phase,raw.Case,raw.Index,raw.Schedule]);
 assert.equal(r.Results.length,7);
 const metrics=[], diagnostics=[];
 for(let arm=0;arm<7;arm++) {
  const result=r.Results[arm],queries=result.Queries??[],paid=new Map(),known=new Set(),mixerKnown=new Set(),arrivals=[],empty=[];
  const byClock=new Map(queries.map(q=>[q.Clock,q]));assert.equal(byClock.size,queries.length);
  assert.equal(result.Predictions.length,256);
  let until=-1,spent=0,triggerClocks=0,paidTriggers=0,late=0,burstQueries=0,fallbackQueries=0;
  let base=0,checkpoint=[...prior];
  const counts=Array(8).fill(0),clocks=[];
  for(let t=0;t<256;t++) {
   if(t%32===0)spent=0;
   let surprise=false;
   for(let j=0;j<=t;j++) {
    const s=raw.Steps[j],paidNow=paid.has(j)&&paid.get(j)<=t;
    if(!known.has(j)&&((!s.Missing&&j+s.Delay<=t)||paidNow)) {
     known.add(j);
     if(j>=t-32)mixerKnown.add(j);
     const p=result.Predictions[j][4],unusual=(s.Y?p:1-p)<.2;
     surprise ||= unusual;
     if(arm>0)arrivals.push({Clock:t,Origin:j,Paid:paidNow,Surprise:unusual,Mixer:j>=t-32});
     if(unusual&&paidNow)paidTriggers++;
     if(j<t-32)late++;
    }
   }
   // Direct probability recursion, independent of Go's log-space journal.
   // After final admission at age32, prefix emissions can be checkpointed.
   const step=(w,j)=>transition(w,result.Predictions[j],mixerKnown.has(j)?raw.Steps[j].Y:undefined);
   while(base<=t-32){checkpoint=step(checkpoint,base);base++;}
   let weights=[...checkpoint];for(let j=base;j<t;j++)weights=step(weights,j);
   const predicted=weights.reduce((v,w,k)=>v+w*result.Predictions[t][k],0);
   assert(Math.abs(predicted-result.Predictions[t][4])<1e-10,`mixer arm${arm} clock${t}`);
   weights=step(weights,t);
   if(arm===0)continue;
   if(surprise){until=t+3;triggerClocks++;}
   const pool=[];for(let j=Math.max(0,t-32);j<t;j++)if(!known.has(j)&&!paid.has(j))pool.push(j);
   const block=Math.floor(t/32),budget=block===0?3:4,end=block===7?248:block*32+31;
   const periodic=t>0&&t%8===0;
   const due=arm<4?periodic:(t>=8&&t<=248&&spent<budget&&(t<=until||end-t+1<=budget-spent));
   if(pool.length===0&&(arm>=4||periodic))empty.push(t);
   const q=byClock.get(t);
   assert.equal(Boolean(q),due&&pool.length>0,`schedule arm${arm} clock${t}`);
   if(q) {
    if(arm>=4){if(t<=until)burstQueries++;else fallbackQueries++;}
    assert.equal(q.Reveal,t+1);assert(pool.includes(q.Origin));
    if(arm===1||arm===4) {
     const hash=crypto.createHash('sha256').update(`${r.Phase}:${r.Case}:${r.Index}:${r.Schedule}:${t}`).digest();
     const selected=Number((BigInt(hash.readUInt32BE(0))*BigInt(pool.length))>>32n);
     assert.equal(q.Origin,pool[selected]);
    } else {
     const scores=pool.map(j=>{
      const ps=result.Predictions[j];const m=weights.reduce((v,w,k)=>v+w*ps[k],0);
      return entropy(m)-((arm===3||arm===6)?weights.reduce((v,w,k)=>v+w*entropy(ps[k]),0):0);
     });
     assert(scores[pool.indexOf(q.Origin)]>=Math.max(...scores)-1e-10,'nonmaximal query utility');
    }
    paid.set(q.Origin,q.Reveal);spent++;counts[block]++;clocks.push(t);
   }
  }
  if(arm===0)assert.equal(queries.length,0);
  assert.deepEqual(result.Arrivals??[],arrivals);
  assert.deepEqual(result.EmptyQueryClocks??[],empty);
  if(arm>=4)counts.forEach((n,b)=>assert(n<=(b===0?3:4)));
  if(r.Schedule===0)assert.equal(queries.length,0);
  assert.equal(result.Fits.length,8);
  for(let f=0;f<8;f++) {
   const fit=result.Fits[f];assert.equal(fit.Clock,f*32);
   const eligible=[];
   for(let j=-16;j<fit.Clock;j++) {
    if(j<0||(!raw.Steps[j].Missing&&j+raw.Steps[j].Delay<=fit.Clock)||(paid.has(j)&&paid.get(j)<=fit.Clock))eligible.push(j);
   }
   for(let w=0;w<2;w++)assert.deepEqual(fit.Origins[w],eligible.slice(-(w===0?64:32)));
  }
  const brier=Array.from({length:5},()=>[0,0]);
  result.Predictions.forEach((ps,t)=>{
   assert.equal(ps.length,5);
   ps.forEach((p,a)=>{
    assert(Number.isFinite(p)&&p>0&&p<1);
    if(arm===0)assert(Math.abs(p-raw.Steps[t].P[[0,1,2,3,12][a]])<1e-12);
    if(r.Schedule===0)assert.equal(p,r.Results[0].Predictions[t][a]);
    const q=raw.Steps[t].Q,loss=(p-q)**2+q*(1-q);
    brier[a][0]+=loss/256;if(t>=192)brier[a][1]+=loss/64;
   });
  });
  let added=0,advanced=0,alreadyNatural=0,noFit=0,waitSum=0;
  for(const q of queries){
   const fit=result.Fits.find(f=>f.Clock>=q.Reveal&&f.Origins[0].includes(q.Origin));
   if(!fit){noFit++;continue;}
   waitSum+=fit.Clock-q.Reveal;
   const s=raw.Steps[q.Origin];if(s.Missing)added++;else if(q.Origin+s.Delay>fit.Clock)advanced++;else alreadyNatural++;
  }
  assert.equal(added+advanced+alreadyNatural+noFit,queries.length);
  metrics.push(brier);diagnostics.push({queries:queries.length,counts,clocks,triggerClocks,paidTriggers,lateMixerOmissions:late,emptyClocks:empty.length,burstQueries,fallbackQueries,training64:{added,advanced,alreadyNatural,noFit,waitSum}});
 }
 return {phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,brier:metrics,diagnostics};
}
