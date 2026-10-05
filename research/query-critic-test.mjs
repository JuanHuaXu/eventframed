import assert from 'node:assert/strict';
import {criticFeatures,fitCritic,criticTransform,predictCritic,chooseCritic,criticDimension} from './query-critic.mjs';

const rows=Array.from({length:80},(_,i)=>({x:[1,...Array.from({length:9},(_,j)=>Math.sin((i+1)*(j+1)))],y:.1*Math.cos(i/3),weight:1+(i%3)}));
const before=JSON.stringify(rows),model=fitCritic(rows);assert.equal(JSON.stringify(rows),before);
// Independent pivoted elimination, not the production Cholesky routine.
const h=Array.from({length:10},()=>Array(11).fill(0));
for(const r of rows){const x=criticTransform(model,r.x);for(let i=0;i<10;i++){for(let j=0;j<10;j++)h[i][j]+=r.weight*x[i]*x[j]/model.weight;h[i][10]+=r.weight*x[i]*r.y/model.weight;}}
for(let i=1;i<10;i++)h[i][i]+=.01;
for(let i=0;i<10;i++){let p=i;for(let j=i+1;j<10;j++)if(Math.abs(h[j][i])>Math.abs(h[p][i]))p=j;[h[i],h[p]]=[h[p],h[i]];const pivot=h[i][i];for(let j=i;j<=10;j++)h[i][j]/=pivot;for(let k=0;k<10;k++)if(k!==i){const m=h[k][i];for(let j=i;j<=10;j++)h[k][j]-=m*h[i][j];}}
for(let i=0;i<10;i++)assert(Math.abs(model.beta[i]-h[i][10])<1e-12);
const flat=fitCritic([-.1,.2,.5].map(y=>({x:[1,...Array(9).fill(0)],y,weight:1})));
assert(Math.abs(predictCritic(flat,[1,...Array(9).fill(0)])-.2)<1e-12);assert(flat.beta.slice(1).every(x=>Math.abs(x)<1e-12));
const d={Clock:160,Pool:[152],Origins:[-16,128,159],Values:[{Origin:152,Mass:[.4,.6],Gain:.01,Base:Array(8).fill(.5)}]};
for(const key of ['Phase','Case','Index','Seeds','Teacher','Q','Y','Selected','Costs','AtPublication','Predictions'])Object.defineProperty(d,key,{get(){throw Error('forbidden '+key);}});
const x=criticFeatures(d,d,152);assert.equal(x.length,criticDimension);
const tied=chooseCritic(flat,[{origin:159,x},{origin:152,x}]);assert.equal(tied.forced,152);assert.equal(tied.gated,152);
const trapped={origin:152,x};for(const key of ['y','weight','q','actual','counter','value'])Object.defineProperty(trapped,key,{get(){throw Error('inference read target '+key);}});
assert.equal(chooseCritic(flat,[trapped]).forced,152);
const negative={...flat,beta:[-.1,...Array(9).fill(0)]};assert.equal(chooseCritic(negative,[{origin:152,x}]).gated,-1);assert.equal(chooseCritic(negative,[]).forced,-1);
assert.throws(()=>fitCritic([]));assert.throws(()=>fitCritic([{x:Array(10).fill(0),y:0,weight:1}]));assert.throws(()=>predictCritic(model,[1,NaN]));
assert.deepEqual(fitCritic(rows),model);
console.log('PASS: independent pivoted solver, degenerate data, ownership/replay, oracle-field traps, ties, abstention and invalid inputs');
