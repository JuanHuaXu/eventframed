import assert from 'node:assert/strict';
import {solvePositive} from './residual-logit.mjs';

export const criticDimension=10,criticLambda=.01;
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
export function criticFeatures(original,disjoint,origin){
  assert.equal(original.Clock,160);assert.equal(disjoint.Clock,160);
  assert(original.Pool.includes(origin)&&disjoint.Pool.includes(origin));
  const a=original.Values.find(v=>v.Origin===origin),b=disjoint.Values.find(v=>v.Origin===origin);assert(a&&b);
  const p=a.Mass[1];assert(p>0&&p<1);assert(Math.abs(p-b.Mass[1])<=1e-10);
  assert(original.Origins.length>0);assert.deepEqual(original.Origins,disjoint.Origins);
  const risk=v=>mean(v.Base.map(p=>{assert(p>0&&p<1);return 4*p*(1-p);}));
  const x=[1,(160-origin)/8,4*p*(1-p),(-p*Math.log(p)-(1-p)*Math.log1p(-p))/Math.log(2),4*a.Gain,4*b.Gain,risk(a),risk(b),
    mean(original.Origins.map(j=>Number(j>=128))),mean(original.Origins.map(j=>(160-j)/176))];
  assert(x.length===criticDimension&&x.every(v=>Number.isFinite(v)&&v>=-1e-12&&v<=1+1e-12));return x;
}
function checkRow(r){assert(r.x.length===criticDimension&&r.x[0]===1&&r.x.every(Number.isFinite));assert(Number.isFinite(r.y)&&Math.abs(r.y)<=1);assert(Number.isFinite(r.weight)&&r.weight>0);}
export function criticTransform(model,x){assert(x.length===criticDimension&&x[0]===1&&x.every(Number.isFinite));return x.map((v,i)=>i?(v-model.center[i])/model.scale[i]:1);}
export function fitCritic(rows){
  assert(rows.length>0);rows.forEach(checkRow);const weight=rows.reduce((s,r)=>s+r.weight,0);
  const center=Array(criticDimension).fill(0),scale=Array(criticDimension).fill(1);
  for(let i=1;i<criticDimension;i++){
    center[i]=rows.reduce((s,r)=>s+r.weight*r.x[i],0)/weight;
    const variance=rows.reduce((s,r)=>s+r.weight*(r.x[i]-center[i])**2,0)/weight;
    scale[i]=variance>1e-12?Math.sqrt(variance):1;
  }
  const model={center,scale,beta:[],lambda:criticLambda,rows:rows.length,weight};
  const h=Array.from({length:criticDimension},()=>Array(criticDimension).fill(0)),g=Array(criticDimension).fill(0);
  for(const r of rows){const x=criticTransform(model,r.x),w=r.weight/weight;for(let i=0;i<criticDimension;i++){g[i]+=w*x[i]*r.y;for(let j=0;j<criticDimension;j++)h[i][j]+=w*x[i]*x[j];}}
  for(let i=1;i<criticDimension;i++)h[i][i]+=criticLambda;
  model.beta=solvePositive(h,g);
  model.normalResidual=Math.max(...g.map((v,i)=>Math.abs(h[i].reduce((s,x,j)=>s+x*model.beta[j],0)-v)));
  assert(model.normalResidual<1e-10);return model;
}
export function predictCritic(model,x){const z=criticTransform(model,x);assert(model.beta.length===criticDimension&&model.beta.every(Number.isFinite));return z.reduce((s,v,i)=>s+v*model.beta[i],0);}
export function chooseCritic(model,candidates){
  if(candidates.length===0)return {forced:-1,gated:-1,maximum:0,scores:[]};
  assert(new Set(candidates.map(c=>c.origin)).size===candidates.length);
  const scores=candidates.map(c=>predictCritic(model,c.x)),maximum=Math.max(...scores);
  const forced=Math.min(...candidates.filter((_,i)=>maximum-scores[i]<=1e-10).map(c=>c.origin));
  return {forced,gated:maximum>0?forced:-1,maximum,scores};
}
