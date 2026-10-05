import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {addTiePolicies} from './tie-policy.mjs';
import {addRiskBudgetPolicy} from './risk-budget-policy.mjs';
import {pathWeights} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'risk-budget-evaluation.json'),frozen=JSON.parse(raw);
for(const [p,h]of Object.entries(frozen.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const modelRaw=readFileSync(dir+'count-planning-exact.json'),model=JSON.parse(modelRaw);
assert.equal(createHash('sha256').update(modelRaw).digest('hex'),frozen.inputSHA256);
assert.deepEqual(addTiePolicies(model),frozen.tieStats);assert.deepEqual(addRiskBudgetPolicy(model),frozen.budgetStats);
const policies=['risk_budget','tie_entropy'];
const counts=Array.from({length:6},(_,step)=>({step,allDifferences:0,reachableDifferences:0,reachableWithoutContradiction:0}));
const masks=Array.from({length:16},(_,mask)=>({mask,unequalLayers:Array(7).fill(0),differingReachableActions:Array(6).fill(0)}));
const counterexamples=[];let layerChecks=0,integerChecks=0;
for(const root of model.roots){
  const paths=pathWeights(root,policies),index=new Map(root.states.map((s,i)=>[s.counts.join(','),i]));
  root.states.forEach((s,i)=>{
    const step=s.counts.reduce((a,n)=>a+n,0);if(step===6||s.actions.risk_budget===s.actions.tie_entropy)return;
    counts[step].allDifferences++;
    if(!policies.some(p=>paths[p][step].has(i)))return;
    counts[step].reachableDifferences++;
    const contradiction=s.counts[1-(root.pattern&1)]>0;
    if(!contradiction){counts[step].reachableWithoutContradiction++;if(step<5)counterexamples.push({pattern:root.pattern,counts:s.counts,actions:s.actions});}
  });
  for(const record of masks){
    const layers=Object.fromEntries(policies.map(p=>[p,new Map([[index.get('0,0,0,0,0,0,0,0'),1]])]));
    for(let step=0;step<=6;step++){
      const keys=new Set(policies.flatMap(p=>[...layers[p].keys()]));
      let equal=true;
      for(const i of keys){
        for(const p of policies){assert(Number.isSafeInteger(layers[p].get(i)||0));integerChecks++;}
        if((layers.risk_budget.get(i)||0)!==(layers.tie_entropy.get(i)||0))equal=false;
        if(step<6&&root.states[i].actions.risk_budget!==root.states[i].actions.tie_entropy)record.differingReachableActions[step]++;
      }
      layerChecks++;if(!equal)record.unequalLayers[step]++;
      if(step===6)break;
      for(const p of policies){
        const next=new Map();
        for(const [i,w]of layers[p]){
          const s=root.states[i],a=s.actions[p];
          const outcomes=record.mask&(1<<a)?[(root.pattern>>a)&1]:[0,1];
          for(const y of outcomes){const c=s.counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);next.set(j,(next.get(j)||0)+w);}
        }
        layers[p]=next;
      }
    }
  }
}
const files=['copy-trigger-audit.mjs','COPY_TRIGGER_PROTOCOL.md','exact-regime.mjs'];
console.log(JSON.stringify({scope:'Frozen policy support/trigger audit; integer occupancy comparison, not universal impossibility',frozenSHA256:createHash('sha256').update(raw).digest('hex'),hashes:Object.fromEntries(files.map(f=>[f,createHash('sha256').update(readFileSync(dir+f)).digest('hex')])),counts,counterexamples,masks,layerChecks,integerChecks},null,2));
