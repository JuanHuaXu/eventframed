import assert from 'node:assert/strict';
import {strongPool,poolWeights} from './strong-brier-pool.mjs';
import {strongBrier,brierWeights} from './strong-brier.mjs';
let checks=0,maxPathError=0,maxPotentialDefect=0;
const prior=[.5,.3,.2];
for(let bits=0;bits<16;bits++) {
  const rows=Array.from({length:4},(_,j)=>({p:[.1+j*.08,.85-j*.02,.4+j*.04],y:!!(bits&(1<<j)),delay:j%2,missing:j===1}));
  const ws=poolWeights(rows,prior);
  for(let t=0;t<4;t++) {
    const mass=prior.map(()=>0);let total=0;
    for(let path=0;path<3**(t+1);path++) {
      let code=path,prev=-1,prob=1,state;
      for(let j=0;j<=t;j++) {
        state=code%3;code=Math.floor(code/3);
        prob*=j===0?prior[state]:(1/(j+1))*prior[state]+(1-1/(j+1))*Number(state===prev);
        const r=rows[j];if(j<t&&!r.missing&&j+r.delay<=t)prob*=Math.exp(-2*(r.p[state]-Number(r.y))**2);
        prev=state;
      }
      total+=prob;mass[state]+=prob;
    }
    for(let k=0;k<3;k++){maxPathError=Math.max(maxPathError,Math.abs(ws[t][k]-mass[k]/total));assert(maxPathError<1e-12);checks++;}
    const p=strongPool(rows[t].p,ws[t]);
    for(const y of [0,1]){const mix=ws[t].reduce((s,w,k)=>s+w*Math.exp(-2*(rows[t].p[k]-y)**2),0),d=2*(p-y)**2+Math.log(mix);maxPotentialDefect=Math.max(maxPotentialDefect,d);assert(d<1e-12);checks++;}
  }
  const poison=rows.map((r,j)=>({...r,y:j>=2||r.missing||j+r.delay>2?!r.y:r.y}));assert.deepEqual(ws.slice(0,3),poolWeights(poison,prior).slice(0,3));
  const immediate=rows.map(r=>({...r,delay:0,missing:false})),wi=poolWeights(immediate,prior,false);let loss=0,experts=[0,0,0];
  immediate.forEach((r,j)=>{loss+=(strongPool(r.p,wi[j])-Number(r.y))**2;experts=experts.map((l,k)=>l+(r.p[k]-Number(r.y))**2);for(let k=0;k<3;k++){assert(loss-experts[k]<=Math.log(1/prior[k])/2+1e-12);checks++;}});
  const two=rows.map(r=>({b:r.p[0],c:r.p[1],...r,p:r.p.slice(0,2)})),w2=poolWeights(two,[.5,.5]),ref=brierWeights(two);
  two.forEach((r,j)=>{assert(Math.abs(w2[j][1]-ref[j])<1e-12);assert(Math.abs(strongPool(r.p,w2[j])-strongBrier(r.b,r.c,ref[j]))<1e-12);checks++;});
}
for(const ps of [[0,1,.5],[.1,.2,.3],[1,1,1]]) {
  const p=strongPool(ps,prior),permuted=strongPool([ps[2],ps[0],ps[1]],[prior[2],prior[0],prior[1]]);
  assert(Math.abs(p-permuted)<1e-12);assert(Math.abs(p-(1-strongPool(ps.map(x=>1-x),prior)))<1e-12);checks+=2;
}
assert.throws(()=>strongPool([0,1],[.5,.8]));assert.throws(()=>strongPool([NaN,1],[.5,.5]));
console.log(JSON.stringify({checks,maxPathError,maxPotentialDefect,limits:'Finite tests, no delayed/sharing cumulative theorem.'}));
