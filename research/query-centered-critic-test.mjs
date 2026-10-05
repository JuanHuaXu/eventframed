import assert from 'node:assert/strict';
import {fitCritic,predictCritic} from './query-critic.mjs';
import {fitCenteredCritic,chooseCenteredCritic} from './query-centered-critic.mjs';
const pools=Array.from({length:32},(_,i)=>Array.from({length:4},(_,j)=>({origin:152+j,x:[1,...Array.from({length:9},(_,k)=>Math.sin((i+1)*(j+1)*(k+1)))],y:.2*Math.sin(i+j),weight:.25})));
const before=JSON.stringify(pools),m=fitCenteredCritic(pools);assert.equal(before,JSON.stringify(pools));assert.deepEqual(m.context,fitCritic(pools.flat()));
const shifted=pools.map((p,i)=>p.map(r=>({...r,x:r.x.map((v,j)=>j?v+Math.sin(i+j):1),y:r.y+.1*Math.sin(i)})));
const m2=fitCenteredCritic(shifted);for(let i=0;i<10;i++)assert(Math.abs(m.rank.beta[i]-m2.rank.beta[i])<1e-10);
for(const pool of pools){const decision=chooseCenteredCritic(m,pool),mean=pool.reduce((s,c)=>s+predictCritic(m.context,c.x),0)/pool.length;assert(Math.abs(mean-decision.poolMean)<1e-12);assert(Math.abs(decision.scores.reduce((s,x)=>s+x,0)/pool.length-mean)<1e-12);}
const trapped=pools[0].map(r=>{const c={origin:r.origin,x:r.x};for(const key of ['y','weight','Q','Case','Phase','value','actual'])Object.defineProperty(c,key,{get(){throw Error('forbidden '+key);}});return c;});
assert.deepEqual(chooseCenteredCritic(m,trapped),chooseCenteredCritic(m,pools[0]));
assert.deepEqual(fitCenteredCritic(pools),m);assert.equal(chooseCenteredCritic(m,[]).forced,-1);
const constant=fitCenteredCritic([[{origin:152,x:[1,...Array(9).fill(.5)],y:-.1,weight:1}]]);
assert.equal(chooseCenteredCritic(constant,[{origin:152,x:[1,...Array(9).fill(.5)]}]).gated,-1);
assert.throws(()=>fitCenteredCritic([]));assert.throws(()=>fitCenteredCritic([[]]));
const extreme=fitCenteredCritic([[...[-1,1,1].map((y,i)=>({origin:152+i,x:[1,i,...Array(8).fill(0)],y,weight:1/3}))]]);
assert(extreme.rank.beta.every(Number.isFinite));
console.log('PASS: centered shift invariance, original context fit, preserved mean, inference traps, replay, degenerate pools and invalid inputs');
