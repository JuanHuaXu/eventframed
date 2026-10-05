import assert from 'node:assert/strict';
import {expandQuery,fitQueryRidge,fitQueryFactorial,chooseQueryFactorial} from './query-error-factorial.mjs';
import {fitCritic} from './query-critic.mjs';
import {fitCenteredCritic,chooseCenteredCritic} from './query-centered-critic.mjs';
const pools=Array.from({length:20},(_,i)=>Array.from({length:3},(_,j)=>({origin:152+j,x:[1,...Array.from({length:9},(_,k)=>Math.sin(i+j*(k+1)))],y:.1*Math.cos(i+j),weight:1/3})));
assert.deepEqual(fitQueryRidge(pools.flat()),fitCritic(pools.flat()));
const base=fitCenteredCritic(pools),linear=fitQueryFactorial(pools,false);assert.deepEqual(base.context,linear.context);assert.deepEqual(base.rank,linear.rank);
for(const p of pools)assert.deepEqual(chooseQueryFactorial(linear,p),chooseCenteredCritic(base,p));
for(const d of [10,20]){const x=[1,...Array.from({length:d-1},(_,i)=>i/20)],q=expandQuery(x,true);assert.equal(q.length,d===10?55:210);let k=d;for(let i=1;i<d;i++)for(let j=i;j<d;j++)assert.equal(q[k++],x[i]*x[j]);}
const before=JSON.stringify(pools),quadratic=fitQueryFactorial(pools,true);assert.equal(before,JSON.stringify(pools));assert.deepEqual(fitQueryFactorial(pools,true),quadratic);
const trapped=pools[0].map(r=>{const c={origin:r.origin,x:r.x};for(const key of ['y','Q','Case','Phase','actual','weight'])Object.defineProperty(c,key,{get(){throw Error('target leaked');}});return c;});assert.deepEqual(chooseQueryFactorial(quadratic,trapped),chooseQueryFactorial(quadratic,pools[0]));
// Separate elimination reference checks the expanded 55-column normal equations.
const rows=pools.flat().map(r=>({...r,x:expandQuery(r.x,true)})),m=fitQueryRidge(rows),d=m.beta.length,h=Array.from({length:d},()=>Array(d+1).fill(0));
for(const r of rows){const x=r.x.map((v,i)=>i?(v-m.center[i])/m.scale[i]:1);for(let i=0;i<d;i++){for(let j=0;j<d;j++)h[i][j]+=r.weight*x[i]*x[j]/m.weight;h[i][d]+=r.weight*x[i]*r.y/m.weight;}}
for(let i=1;i<d;i++)h[i][i]+=.01;
for(let i=0;i<d;i++){let p=i;for(let j=i+1;j<d;j++)if(Math.abs(h[j][i])>Math.abs(h[p][i]))p=j;[h[i],h[p]]=[h[p],h[i]];const v=h[i][i];for(let j=i;j<=d;j++)h[i][j]/=v;for(let k=0;k<d;k++)if(k!==i){const a=h[k][i];for(let j=i;j<=d;j++)h[k][j]-=a*h[i][j];}}
for(let i=0;i<d;i++)assert(Math.abs(m.beta[i]-h[i][d])<1e-10);
assert.throws(()=>expandQuery([1,2],true));assert.throws(()=>fitQueryRidge([]));
console.log('PASS: exact base control, basis dimensions, independent expanded solver, inference traps, ownership and replay');
