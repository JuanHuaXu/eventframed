import assert from 'node:assert/strict';
import {fitCritic,predictCritic,criticDimension} from './query-critic.mjs';
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
function center(candidates){
  assert(candidates.length>0);
  const averages=Array.from({length:criticDimension},(_,j)=>mean(candidates.map(c=>c.x[j])));
  return candidates.map(c=>({origin:c.origin,x:c.x.map((v,j)=>j?v-averages[j]:1)}));
}
export function fitCenteredCritic(pools){
  assert(pools.length>0&&pools.every(p=>p.length>0));
  const centered=[];
  for(const pool of pools){
    for(const r of pool)assert(Math.abs(r.weight-1/pool.length)<1e-12);
    const y=mean(pool.map(r=>r.y)),xs=center(pool);
    // Centered gains can span [-2,2]; reversible scaling preserves the ridge
    // solution while respecting the existing fitter's [-1,1] target contract.
    for(let i=0;i<pool.length;i++)centered.push({...xs[i],y:(pool[i].y-y)/2,weight:pool[i].weight});
  }
  return {context:fitCritic(pools.flat()),rank:fitCritic(centered)};
}
export function chooseCenteredCritic(model,candidates){
  if(!candidates.length)return {forced:-1,gated:-1,maximum:0,scores:[],poolMean:0};
  assert(new Set(candidates.map(c=>c.origin)).size===candidates.length);
  const poolMean=mean(candidates.map(c=>predictCritic(model.context,c.x)));
  const relative=center(candidates).map(c=>2*predictCritic(model.rank,c.x)),offset=mean(relative);
  const scores=relative.map(v=>poolMean+v-offset),maximum=Math.max(...scores);
  assert(Math.abs(mean(scores)-poolMean)<1e-12);
  const forced=Math.min(...candidates.filter((_,i)=>maximum-scores[i]<=1e-10).map(c=>c.origin));
  return {forced,gated:maximum>0?forced:-1,maximum,scores,poolMean};
}
