import assert from 'node:assert/strict';
import {createIntervalSurrogate,coveringIntervals} from './interval-surrogate.mjs';

const near=(a,b)=>assert(Math.abs(a-b)<1e-11,`${a} != ${b}`);
// Direct reference retains absolute G for inactive AND active learners and
// multiplies unnormalized pair masses instead of accumulating R/V in log space.
function reference(T,prior,intervals){
  const rates=[.5,.25,.125,.0625],K=prior.length;
  const states=intervals.map(()=>({mass:rates.flatMap(()=>prior.map(p=>p/4)),G:0}));
  let t=0;
  return probabilities=>{
    const ids=intervals.flatMap(([a,b],i)=>a<=t&&t<=b?[i]:[]),mq=ids.map(i=>Math.exp(-states[i].G)),mz=mq.reduce((s,x)=>s+x,0),q=mq.map(x=>x/mz);
    const distributions=ids.map(i=>{const m=states[i].mass,z=m.reduce((s,x)=>s+x,0);return m.map(x=>x/z);});
    const w=Array(K).fill(0);
    ids.forEach((id,b)=>distributions[b].forEach((v,j)=>{w[j%K]+=q[b]*v*rates[Math.floor(j/K)];}));
    const z=w.reduce((s,x)=>s+x,0);w.forEach((x,j)=>{w[j]=x/z;});t++;
    const shortcut=Array(K).fill(0);
    ids.forEach((id,b)=>states[id].mass.forEach((v,j)=>{shortcut[j%K]+=q[b]*v*rates[Math.floor(j/K)];}));
    const shortcutZ=shortcut.reduce((s,x)=>s+x,0);shortcut.forEach((x,k)=>{shortcut[k]=x/shortcutZ;});
    return {weights:w,shortcut,observe:y=>{
      const losses=probabilities.map(p=>(p-Number(y))**2),avg=losses.reduce((s,l,k)=>s+w[k]*l,0),r=losses.map(l=>avg-l);
      const factors=rates.flatMap(eta=>r.map(v=>Math.exp(eta*v-eta*eta*v*v)));
      const zs=distributions.map(P=>P.reduce((s,p,j)=>s+p*factors[j],0)),ghat=-Math.log(zs.reduce((s,z,j)=>s+q[j]*z,0));
      states.forEach((state,id)=>{const b=ids.indexOf(id);state.G+=b<0?ghat:-Math.log(zs[b]);if(b>=0)state.mass.forEach((v,j)=>{state.mass[j]=v*factors[j];});});
    }};
  };
}
let referencePredictions=0,maxShortcutDifference=0;
for(let bits=0;bits<256;bits++)for(const intervals of [[[0,7]],coveringIntervals(8)]){
  const pi=[.8,.1,.1],model=createIntervalSurrogate(8,pi,intervals),ref=reference(8,pi,intervals);
  for(let t=0;t<8;t++){
    const ps=[.1+.1*(t%3),.6,.9-.1*(t%2)],y=Boolean(bits&(1<<t)),a=model.issue(ps),b=ref(ps);
    a.weights.forEach((w,k)=>{near(w,b.weights[k]);maxShortcutDifference=Math.max(maxShortcutDifference,Math.abs(w-b.shortcut[k]));});const result=model.deliver(a.id,y);assert(result.ghat>=-1e-12);b.observe(y);referencePredictions++;
  }
}
// Same issued evidence batch must commute up to floating-point summation.
const a=createIntervalSurrogate(8,[.5,.5]),b=createIntervalSurrogate(8,[.5,.5]);
for(let t=0;t<4;t++){a.issue([.2,.8]);b.issue([.2,.8]);}
a.deliver(0,true);a.deliver(2,false);b.deliver(2,false);b.deliver(0,true);
near(a.issue([.3,.7]).p,b.issue([.3,.7]).p);
const before=a.snapshot();assert(a.deliver(0,true).duplicate);assert.deepEqual(a.snapshot(),before);
assert.throws(()=>a.deliver(0,false));assert.deepEqual(a.snapshot(),before);
assert.throws(()=>a.deliver(7,true));assert.deepEqual(a.snapshot(),before);
const detached=a.snapshot();detached.states[0].R[0]=999;assert.deepEqual(a.snapshot(),before);
a.expireBefore(2);assert.throws(()=>a.deliver(1,true));
assert.throws(()=>createIntervalSurrogate(257,[.5,.5]));
assert.throws(()=>createIntervalSurrogate(8,[.5,.5],[[0,3]]));
assert.throws(()=>createIntervalSurrogate(8,[.5,.5],[[0,7],[0,7]]));
const own=[.5,.5],model=createIntervalSurrogate(2,own);own[0]=99;
const probabilities=[.2,.8],issued=model.issue(probabilities);probabilities[0]=99;issued.weights[0]=99;model.deliver(issued.id,true);
assert(model.issue([.2,.8]).p>=.2);
assert(maxShortcutDifference>1e-6,'normalization negative control must be nondegenerate');
console.log(JSON.stringify({referencePredictions,fullOutcomes:256,intervalFamilies:2,maxShortcutDifference,lateBatchCommutativity:true,duplicateConflictExpiryOwnership:true}));
