import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {addRiskBudgetPolicy} from './risk-budget-policy.mjs';
import {addRegimeSafePolicy} from './regime-safe-policy.mjs';
import {addTiePolicies} from './tie-policy.mjs';
import {actualJoint,pathWeights} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',input=readFileSync(dir+'count-planning-exact.json'),d=JSON.parse(input);
for(const[p,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const frozenRaw=readFileSync(dir+'risk-budget-evaluation.json'),frozen=JSON.parse(frozenRaw);
for(const[p,h]of Object.entries(frozen.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const tieStats=addTiePolicies(d),safeStats=addRegimeSafePolicy(d),budgetStats=addRiskBudgetPolicy(d);
assert.deepEqual(tieStats,frozen.tieStats);assert.deepEqual(safeStats,frozen.safeStats);assert.deepEqual(budgetStats,frozen.budgetStats);
const policies=Object.keys(d.totals),prepared=d.roots.map(root=>({root,paths:pathWeights(root,policies)}));
const results=[],comparisons=[];let normalizationChecks=0;
for(const noise of [.05,.125,.175,.225,.275,.35])for(let mask=0;mask<16;mask++){
  const scores=Object.fromEntries(policies.map(p=>[p,{finalBrier:0,areaBrier:0,postSum:0,accuracy:0,confidentWrong:0,mass:Array(7).fill(0)}]));
  for(const{root,paths}of prepared)root.states.forEach((s,i)=>{
    const step=s.counts.reduce((a,n)=>a+n,0),joint=actualJoint(root.pattern,s.counts,noise,mask),mass=joint.reduce((a,v)=>a+v,0);
    if(mass===0)return;
    const square=s.forecast.reduce((a,p)=>a+p*p,0);
    const loss=mass*(1+square)-2*joint.reduce((a,v,y)=>a+v*s.forecast[y],0);
    const chosen=s.forecast.indexOf(Math.max(...s.forecast)),correct=joint[chosen];
    for(const policy of policies){
      const w=paths[policy][step].get(i)||0;if(!w)continue;const out=scores[policy];
      out.mass[step]+=w*mass;if(step<6)out.areaBrier+=w*loss/6;if(step>0)out.postSum+=w*loss;
      if(step===6){out.finalBrier+=w*loss;out.accuracy+=w*correct;if(Math.max(...s.forecast)>=.9)out.confidentWrong+=w*(mass-correct);}
    }
  });
  for(const out of Object.values(scores))for(const mass of out.mass){assert(Math.abs(mass-1)<1e-11);normalizationChecks++;}
  for(const candidate of ['risk_budget'])for(const control of ['random','entropy','tie_entropy']){
    const finalGain=scores[control].finalBrier-scores[candidate].finalBrier,areaGain=scores[control].areaBrier-scores[candidate].areaBrier;
    comparisons.push({noise,mask,candidate,control,finalGain,areaGain,finalNonharm:finalGain>=-.01,areaPositive:mask===15?null:areaGain>0});
  }
  results.push({noise,mask,scores});
}
const paths=['exact-regime.mjs','tie-policy.mjs','regime-safe-policy.mjs','risk-budget-policy.mjs','risk-budget-offgrid.mjs','RISK_BUDGET_OFFGRID_PROTOCOL.md'].map(p=>dir+p);
console.log(JSON.stringify({scope:'Frozen budget policy in96 off-grid regimes; interpolation and extrapolation; no semantic or continuous-noise validation',inputSHA256:createHash('sha256').update(input).digest('hex'),hashes:Object.fromEntries(paths.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),frozenSHA256:createHash('sha256').update(frozenRaw).digest('hex'),tieStats,safeStats,budgetStats,normalizationChecks,results,comparisons},null,2));
