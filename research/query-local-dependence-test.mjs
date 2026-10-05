import assert from 'node:assert/strict';
import {observedInputs,dependenceCell,fitLocalDependence} from './query-local-dependence.mjs';
const source={Initial:Array.from({length:16},()=>({Bits:3})),Steps:Array.from({length:256},()=>({X:7,Missing:false,Delay:0}))};
const observed=observedInputs([-16,-1,159],source);assert.deepEqual([...observed],[3,7]);
assert.equal(dependenceCell(.1,3,observed),1);assert.equal(dependenceCell(.5,3,observed),3);assert.equal(dependenceCell(.1,8,observed),0);assert.equal(dependenceCell(.5,8,observed),2);
assert.deepEqual(observedInputs([-16,-15,-1,159],source),observed);
for(const key of ['Y','Q','teacher'])for(const step of [...source.Steps,...source.Initial])Object.defineProperty(step,key,{get(){throw Error(key);}});
source.Steps[160]=new Proxy({},{get(){throw Error('future');}});assert.deepEqual(observedInputs([-16,-1,159],source),observed);
assert.throws(()=>observedInputs([160],source));assert.throws(()=>observedInputs([1,1],source));source.Steps[159].Delay=2;assert.throws(()=>observedInputs([159],source));
const rows=[];
for(let cell=0;cell<4;cell++)for(const qy of [0,1])for(const y of [0,1]){
  const base=.5,delta=qy?.3:-.3,q=base+(cell/3)*delta;
  rows.push({cell,base,delta,y,weight:.5*(y?q:1-q)});
}
const copy=structuredClone(rows),fit=fitLocalDependence(rows);fit.forEach((v,i)=>assert(Math.abs(v.lambda-i/3)<1e-10));assert.deepEqual(rows,copy);
assert(fitLocalDependence([]).every(v=>v.lambda===1&&v.rows===0));assert.throws(()=>fitLocalDependence([{cell:4}]));
console.log('PASS: four-cell known optima; as-of support, duplicate inputs, oracle/future traps, ownership, empty-cell fallback and invalid inputs');
