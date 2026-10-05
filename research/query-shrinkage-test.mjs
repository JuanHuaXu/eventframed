import assert from 'node:assert/strict';
import {fitQueryShrinkage,shrinkQueryMass,queryMixtureScore,chooseShrinkage} from './query-shrinkage.mjs';
const loss=(rows,a)=>rows.reduce((s,r)=>s+r.weight*(shrinkQueryMass(r.p,a)-r.y)**2,0);
const fixtures=[
  [{p:.5,y:1,weight:1}],
  [{p:1,y:0,weight:1}],
  [{p:.6,y:1,weight:1}],
  [{p:.9,y:.7,weight:1},{p:.1,y:.3,weight:1}],
  [{p:.9,y:1,weight:.2},{p:.1,y:1,weight:.8},{p:.7,y:0,weight:1}],
];
for(const rows of fixtures){const before=structuredClone(rows),m=fitQueryShrinkage(rows);for(let i=0;i<=1000;i++)assert(loss(rows,m.alpha)<=loss(rows,i/1000)+1e-14);assert.deepEqual(rows,before);}
assert.equal(fitQueryShrinkage(fixtures[0]).alpha,0);assert.equal(fitQueryShrinkage(fixtures[1]).alpha,0);assert.equal(fitQueryShrinkage(fixtures[2]).alpha,1);assert(Math.abs(fitQueryShrinkage(fixtures[3]).alpha-.5)<1e-14);
for(const p of [0,.1,.5,.9,1]){
  const c=[[.1,.7],[.9,.2]],m=c[0].map((x,j)=>(1-p)*x+p*c[1][j]),direct=c[0].reduce((s,x,j)=>s+((1-p)*(x-m[j])**2+p*(c[1][j]-m[j])**2)/2,0);
  assert(Math.abs(queryMixtureScore(c,p)-direct)<1e-14);assert.equal(shrinkQueryMass(p,1),p);assert.equal(shrinkQueryMass(p,0),.5);
}
const values=[{Origin:152,Mass:[.2,.8],Conditional:[[.2,.5],[.8,.5]]},{Origin:153,Mass:[.2,.8],Conditional:[[.2,.5],[.8,.5]]}],before=structuredClone(values);
assert.equal(chooseShrinkage(values,1).origin,152);assert.deepEqual(chooseShrinkage(values.toReversed(),1).scores,chooseShrinkage(values,1).scores);assert.deepEqual(values,before);
for(const v of values)for(const key of ['Q','ActualY','Teacher','future'])Object.defineProperty(v,key,{get(){throw Error(key);}});
assert.equal(chooseShrinkage(values,.5).origin,152);assert.equal(chooseShrinkage([],1).origin,-1);
assert.throws(()=>fitQueryShrinkage([]));assert.throws(()=>fitQueryShrinkage([{p:NaN,y:1,weight:1}]));assert.throws(()=>shrinkQueryMass(.5,2));assert.throws(()=>queryMixtureScore([[.2],[.1,.2]],.5));
console.log('PASS: convex fit grid, boundary/degenerate cases, mixture variance, raw/neutral maps, ownership, ties and oracle traps');
