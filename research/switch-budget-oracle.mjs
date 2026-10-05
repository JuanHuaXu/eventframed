import assert from 'node:assert/strict';

// Diagnostic hindsight optimum, never a deployable forecast. Costs may include
// hidden teacher-law information. At most s switches, arbitrary initial expert.
export function switchBudgetCosts(costs,maxSwitches){
 assert(costs.length>0&&Number.isInteger(maxSwitches)&&maxSwitches>=0);
 const k=costs[0].length;assert(k>0);for(const row of costs)assert(row.length===k&&row.every(x=>Number.isFinite(x)&&x>=0));
 const cap=Math.min(maxSwitches,costs.length-1);
 let dp=Array.from({length:cap+1},()=>costs[0].slice());
 for(let t=1;t<costs.length;t++){
  const next=Array.from({length:cap+1},()=>Array(k));
  for(let s=0;s<=cap;s++)for(let i=0;i<k;i++){
   let best=dp[s][i];if(s>0)for(let j=0;j<k;j++)if(j!==i)best=Math.min(best,dp[s-1][j]);
   next[s][i]=best+costs[t][i];
  }
  dp=next;
 }
 return dp.map(row=>Math.min(...row));
}
