import assert from 'node:assert/strict';
import {createContextRouter,contextMaskPrior} from './context-expert-router.mjs';

// Enumerate EVERY latent expert assignment, rather than using the component's
// per-cell factorization. Two bits and two experts give26 assignments overall.
function reference(samples,x,ps,contextual){
  const pi=[.6,.4],mp=contextMaskPrior(2,contextual);let z=0,num=0;
  for(let m=0;m<mp.length;m++){
    const keys=Array.from({length:4},(_,v)=>v).filter(v=>(v&m)===v);
    for(let assignment=0;assignment<2**keys.length;assignment++){
      let w=mp[m];for(let j=0;j<keys.length;j++)w*=pi[(assignment>>j)&1];
      const expert=bits=>(assignment>>keys.indexOf(bits&m))&1;
      for(const s of samples){const p=s.ps[expert(s.x)];w*=s.y?p:1-p;}
      z+=w;num+=w*ps[expert(x)];
    }
  }
  return {p:num/z,logZ:Math.log(z)};
}
let comparisons=0,maxProbabilityError=0,maxEvidenceError=0;
contextMaskPrior(2).forEach((p,j)=>assert(Math.abs(p-[.95,.02,.02,.01][j])<1e-14));
for(let outcomes=0;outcomes<64;outcomes++)for(let availability=0;availability<64;availability++)for(const contextual of [false,true]){
  const model=createContextRouter(2,[.6,.4],contextual,16),samples=[];
  for(let j=0;j<6;j++){const x=[0,1,2,3,0,3][j],ps=[.15+.05*j,.85-.03*j],y=Boolean(outcomes&(1<<j));model.issue(x,ps);if(availability&(1<<j))samples.push({id:j,x,ps,y});}
  for(const s of samples.toReversed())model.deliver(s.id,s.y);
  for(let x=0;x<4;x++){
    const actual=model.predict(x,[.25,.8]),expected=reference(samples,x,[.25,.8],contextual),pe=Math.abs(actual.p-expected.p),ze=Math.abs(model.inspect().logEvidence-expected.logZ);
    assert(pe<1e-12&&ze<1e-12);maxProbabilityError=Math.max(maxProbabilityError,pe);maxEvidenceError=Math.max(maxEvidenceError,ze);comparisons++;
  }
}
const model=createContextRouter(2,[.6,.4]),ps=[.2,.8],first=model.issue(1,ps);ps[0]=.99;first.weights[0]=99;
model.deliver(0,true);const state=model.inspect();assert.equal(model.deliver(0,true),false);assert.deepEqual(model.inspect(),state);
assert.throws(()=>model.deliver(0,false));assert.deepEqual(model.inspect(),state);
assert.throws(()=>model.issue(4,[.2,.8]));assert.deepEqual(model.inspect(),state);
model.issue(2,[.1,.9]);model.expireBefore(2);assert.throws(()=>model.deliver(1,true));
const detached=model.inspect();detached.maskWeights[0]=99;assert.notEqual(model.inspect().maskWeights[0],99);
// Extreme contradictory evidence must revive a formerly underflowed mass.
const revival=createContextRouter(0,[.5,.5],false);
for(let t=0;t<256;t++){revival.issue(0,[1e-9,1-1e-9]);revival.deliver(t,t<128);}
assert(Math.abs(revival.predict(0,[.2,.8]).p-.5)<1e-5);
assert.throws(()=>revival.issue(0,[.2,.8]));
assert.throws(()=>createContextRouter(10,[.5,.5]));assert.throws(()=>createContextRouter(2,[0,1]));
const edge=createContextRouter(9,[.95,.05]);assert.equal(edge.predict(511,[1-Number.EPSILON,1-Number.EPSILON]).p,1-Number.EPSILON);
console.log(JSON.stringify({comparisons,maxProbabilityError,maxEvidenceError,duplicateConflictExpiryOwnership:true,contradictionRevival:true}));
