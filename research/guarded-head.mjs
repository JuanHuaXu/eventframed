import {delayedShare} from './delayed-fixed-share.mjs';
import {runBudget} from './spike-budget-core.mjs';
import {pointwiseGuard} from './pointwise-brier-guard.mjs';

export function guardedHead(rows){
  const weights=delayedShare(rows),point=rows.map((r,i)=>pointwiseGuard(r.b,r.c,weights[i]).p),local=[];
  for(let c=0;c<rows.length;c+=32)local.push(...runBudget(rows.slice(c,c+32),weights.slice(c,c+32)).map(f=>f.p));
  return{point,local};
}
