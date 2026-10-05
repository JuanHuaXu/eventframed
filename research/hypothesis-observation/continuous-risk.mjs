import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {canonicalHistory} from './count-planning.mjs';
import {noiseEnvelope,multiplyAffine} from './noise-envelope.mjs';
import {addTiePolicies} from './tie-policy.mjs';
import {addRiskBudgetPolicy} from './risk-budget-policy.mjs';
import {pathWeights} from './exact-regime.mjs';
import {restrict,bounds} from './bernstein-risk.mjs';
const dir='research/hypothesis-observation/',frozenRaw=readFileSync(dir+'risk-budget-evaluation.json'),frozen=JSON.parse(frozenRaw);
for(const [p,h]of Object.entries(frozen.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const model=JSON.parse(readFileSync(dir+'count-planning-exact.json'));
assert.equal(createHash('sha256').update(readFileSync(dir+'count-planning-exact.json')).digest('hex'),frozen.inputSHA256);
assert.deepEqual(addTiePolicies(model),frozen.tieStats);assert.deepEqual(addRiskBudgetPolicy(model),frozen.budgetStats);
const policies=['risk_budget','random','entropy','tie_entropy'];
const coefficients=Array.from({length:16},(_,mask)=>({mask,scores:Object.fromEntries(policies.map(p=>[p,{finalBrier:Array(11).fill(0),areaBrier:Array(11).fill(0)}]))}));
let histories=0;
for(const root of model.roots){
  const paths=pathWeights(root,policies);
  root.states.forEach((s,i)=>{
    const step=s.counts.reduce((a,n)=>a+n,0),weights=Object.fromEntries(policies.map(p=>[p,paths[p][step].get(i)||0]));
    if(!Object.values(weights).some(Boolean))return;histories++;
    const losses=Array.from({length:4},(_,y)=>s.forecast.reduce((v,p,z)=>v+(p-Number(y===z))**2,0));
    for(const m of noiseEnvelope(canonicalHistory(root.pattern,s.counts),undefined,.05,.35)){
      const classes=m.classCoefficients.map(c=>{while(c.length<11)c=multiplyAffine(c,1,1);return c;});
      const risk=Array.from({length:11},(_,k)=>classes.reduce((v,c,y)=>v+c[k]*losses[y],0));
      for(const p of policies){const w=weights[p];if(!w)continue;
        const target=coefficients[m.mask].scores[p][step===6?'finalBrier':'areaBrier'],factor=step===6?w:w/6;
        risk.forEach((v,k)=>target[k]+=factor*v);
      }
    }
  });
}
const ranges=[['inside',.1,.3],['below',.05,.1],['above',.3,.35]],envelopes=[];
for(const row of coefficients)for(const control of policies.slice(1))for(const metric of ['finalBrier','areaBrier']){
  const gain=row.scores[control][metric].map((v,k)=>v-row.scores.risk_budget[metric][k]);
  for(const [range,a,b]of ranges)envelopes.push({mask:row.mask,control,metric,range,...bounds(restrict(gain,(a-.05)/.3,(b-.05)/.3))});
}
const files=['continuous-risk.mjs','bernstein-risk.mjs','CONTINUOUS_RISK_PROTOCOL.md','noise-envelope.mjs','count-planning.mjs'];
console.log(JSON.stringify({scope:'Floating-point Bernstein population-risk envelopes; not outward-rounded certification',frozenSHA256:createHash('sha256').update(frozenRaw).digest('hex'),hashes:Object.fromEntries(files.map(f=>[f,createHash('sha256').update(readFileSync(dir+f)).digest('hex')])),histories,coefficients,envelopes},null,2));
