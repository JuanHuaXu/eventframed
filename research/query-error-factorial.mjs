import assert from 'node:assert/strict';
import {solvePositive} from './residual-logit.mjs';
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
export function expandQuery(x,quadratic){
  assert([10,20].includes(x.length)&&x[0]===1&&x.every(Number.isFinite));assert.equal(typeof quadratic,'boolean');
  const out=x.slice();if(quadratic)for(let i=1;i<x.length;i++)for(let j=i;j<x.length;j++)out.push(x[i]*x[j]);return out;
}
function transform(m,x){assert(x.length===m.beta.length&&x[0]===1&&x.every(Number.isFinite));return x.map((v,i)=>i?(v-m.center[i])/m.scale[i]:1);}
export function fitQueryRidge(rows){
  assert(rows.length>0);const d=rows[0].x.length;assert([10,20,55,210].includes(d));
  for(const r of rows)assert(r.x.length===d&&r.x[0]===1&&r.x.every(Number.isFinite)&&Number.isFinite(r.y)&&Math.abs(r.y)<=1&&Number.isFinite(r.weight)&&r.weight>0);
  const weight=rows.reduce((s,r)=>s+r.weight,0),center=Array(d).fill(0),scale=Array(d).fill(1);
  for(let i=1;i<d;i++){center[i]=rows.reduce((s,r)=>s+r.weight*r.x[i],0)/weight;const variance=rows.reduce((s,r)=>s+r.weight*(r.x[i]-center[i])**2,0)/weight;scale[i]=variance>1e-12?Math.sqrt(variance):1;}
  const m={center,scale,beta:Array(d).fill(0),lambda:.01,rows:rows.length,weight};
  const h=Array.from({length:d},()=>Array(d).fill(0)),g=Array(d).fill(0);
  for(const r of rows){const x=transform(m,r.x),w=r.weight/weight;for(let i=0;i<d;i++){g[i]+=w*x[i]*r.y;for(let j=0;j<d;j++)h[i][j]+=w*x[i]*x[j];}}
  for(let i=1;i<d;i++)h[i][i]+=.01;m.beta=solvePositive(h,g);
  m.normalResidual=Math.max(...g.map((v,i)=>Math.abs(h[i].reduce((s,x,j)=>s+x*m.beta[j],0)-v)));assert(m.normalResidual<1e-10);return m;
}
const predict=(m,x)=>transform(m,x).reduce((s,v,i)=>s+v*m.beta[i],0);
function centered(pool){const d=pool[0].x.length,av=Array.from({length:d},(_,j)=>mean(pool.map(r=>r.x[j])));return pool.map(r=>({origin:r.origin,x:r.x.map((v,j)=>j?v-av[j]:1)}));}
export function fitQueryFactorial(pools,quadratic){
  assert(pools.length>0&&pools.every(p=>p.length>0));
  const expanded=pools.map(p=>p.map(r=>({...r,x:expandQuery(r.x,quadratic)}))),rank=[];
  for(const pool of expanded){const ys=mean(pool.map(r=>r.y)),xs=centered(pool);pool.forEach((r,i)=>{assert(Math.abs(r.weight-1/pool.length)<1e-12);rank.push({...xs[i],y:(r.y-ys)/2,weight:r.weight});});}
  return {context:fitQueryRidge(expanded.flat()),rank:fitQueryRidge(rank),quadratic};
}
export function chooseQueryFactorial(model,candidates){
  if(!candidates.length)return {forced:-1,gated:-1,maximum:0,scores:[],poolMean:0};
  assert(new Set(candidates.map(c=>c.origin)).size===candidates.length);
  const expanded=candidates.map(c=>({origin:c.origin,x:expandQuery(c.x,model.quadratic)})),poolMean=mean(expanded.map(c=>predict(model.context,c.x)));
  const relative=centered(expanded).map(c=>2*predict(model.rank,c.x)),offset=mean(relative),scores=relative.map(v=>poolMean+v-offset),maximum=Math.max(...scores);
  assert(Math.abs(mean(scores)-poolMean)<1e-12);
  const forced=Math.min(...candidates.filter((_,i)=>maximum-scores[i]<=1e-10).map(c=>c.origin));return {forced,gated:maximum>0?forced:-1,maximum,scores,poolMean};
}
