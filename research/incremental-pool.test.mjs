import assert from 'node:assert/strict';
import {incrementalPoolWeights} from './incremental-pool.mjs';
import {poolWeights} from './strong-brier-pool.mjs';
let forecasts=0;
for(let bits=0;bits<64;bits++)for(const delayMode of [0,1,2,3])for(const sharing of [false,true]) {
  const rows=Array.from({length:16},(_,j)=>({p:[.1+j*.02,.8-j*.01,.3],y:!!(bits&(1<<(j%6))),delay:delayMode===0?0:delayMode===1?j%5:delayMode===2?30:15-j,missing:j%7===1}));
  const prior=[.5,.3,.2],fast=incrementalPoolWeights(rows,prior,sharing),slow=poolWeights(rows,prior,sharing);
  assert.deepEqual(fast.weights,slow);forecasts+=rows.length;
  for(const t of [0,5,15]) {
    const poison=rows.map((r,j)=>({...r,y:j>=t||r.missing||j+r.delay>t?!r.y:r.y}));
    assert.deepEqual(fast.weights.slice(0,t+1),incrementalPoolWeights(poison,prior,sharing).weights.slice(0,t+1));
  }
  if(delayMode===2)assert.equal(fast.transitions,16);
  if(delayMode===0)assert(fast.transitions<=31);
}
const none=incrementalPoolWeights([],[.5,.5]);assert.deepEqual(none,{weights:[],transitions:0});
assert.throws(()=>incrementalPoolWeights([{p:[.1,.2],y:0,delay:-1,missing:false}],[.5,.5]));
console.log(JSON.stringify({forecasts,tests:'bit-exact reference, immediate/delayed/missing/simultaneous arrivals, frozen issued outputs, as-of poisoning',limitations:'Batch finite-horizon kernel, no persistence/concurrent daemon lifecycle tested.'}));
