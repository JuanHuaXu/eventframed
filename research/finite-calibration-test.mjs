import assert from 'node:assert/strict';
import {fitFiniteCalibration} from './finite-calibration.mjs';

// Multiplicative small-model enumeration does not use the log-domain fitter.
const grid=[-2,-1,0,1,2],normalizer=grid.reduce((s,x)=>s+Math.exp(-x*x/2),0);
const atoms=grid.flatMap(a=>grid.map(b=>[a,b,.5*Math.exp(-(a*a+b*b)/2)/normalizer**2+(a===0&&b===0?.5:0)]));
let checks=0,maxError=0;
for(let mask=0;mask<256;mask++){
  const rows=[];
  for(let t=0;t<=8;t++){
    const fit=fitFiniteCalibration(rows);
    const masses=atoms.map(([a,b,prior])=>rows.reduce((m,r)=>{const u=Math.log(r.p/(1-r.p)),p=1/(1+Math.exp(-(a+(1+b)*u)));return m*(r.y?p:1-p);},prior));
    const sum=masses.reduce((a,b)=>a+b,0);
    for(const p of [.1,.5,.9]){
      const u=Math.log(p/(1-p)),expected=atoms.reduce((v,[a,b],i)=>v+masses[i]/sum/(1+Math.exp(-(a+(1+b)*u))),0);
      const error=Math.abs(expected-fit.predict(p));maxError=Math.max(maxError,error);assert(error<1e-12);checks++;
    }
    assert(Math.abs(Math.log(sum)-fit.logEvidence)<1e-12);
    if(t<8)rows.push({p:[.1,.5,.9][t%3],y:Boolean(mask&(1<<t))});
  }
}
const prior=fitFiniteCalibration([]);
assert(Math.abs(prior.predict(.5)-.5)<1e-15);
assert(prior.predict(.9)<.9);assert(prior.predict(.1)>.1);
for(const p of [0,1]){const f=fitFiniteCalibration(Array.from({length:64},(_,i)=>({p,y:i%2===0})));assert(f.predict(p)>0&&f.predict(p)<1&&Number.isFinite(f.logEvidence));}
const fit=fitFiniteCalibration([{p:.7,y:true}]),before=fit.predict(.7);fit.weights.fill(0);assert.equal(fit.predict(.7),before);
assert.throws(()=>fitFiniteCalibration([{p:NaN,y:true}]));assert.throws(()=>fitFiniteCalibration([{p:.5,y:1}]));
assert.throws(()=>fitFiniteCalibration(Array(65).fill({p:.5,y:true})));
console.log(JSON.stringify({checks,maxError,priorOnlyIsNotIdentity:true,ownership:true,extremes:true}));
