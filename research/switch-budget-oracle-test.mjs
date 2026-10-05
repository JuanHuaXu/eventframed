import assert from 'node:assert/strict';
import {switchBudgetCosts} from './switch-budget-oracle.mjs';
let checks=0;
for(let seed=0;seed<32;seed++){
 const costs=Array.from({length:5},(_,t)=>Array.from({length:3},(_,i)=>((seed*13+t*7+i*11+t*i*3)%29)/29));
 const reference=Array(5).fill(Infinity);
 for(let path=0;path<3**5;path++){
  let code=path,previous=-1,switches=0,total=0;
  for(let t=0;t<5;t++){const i=code%3;code=Math.floor(code/3);if(t&&i!==previous)switches++;previous=i;total+=costs[t][i]}
  for(let s=switches;s<5;s++)reference[s]=Math.min(reference[s],total);
 }
 const got=switchBudgetCosts(costs,4);for(let s=0;s<5;s++){assert(Math.abs(got[s]-reference[s])<1e-12);checks++}
 assert(Math.abs(got[4]-costs.reduce((n,row)=>n+Math.min(...row),0))<1e-12);
}
assert.deepEqual(switchBudgetCosts([[1],[2],[3]],8),[6,6,6]);
assert.throws(()=>switchBudgetCosts([[NaN]],1));
console.log(JSON.stringify({exhaustivePaths:32*3**5,checks,status:'PASS'}));
