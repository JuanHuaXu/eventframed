import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {addRiskBudgetPolicy} from './risk-budget-policy.mjs';
import {addRegimeSafePolicy} from './regime-safe-policy.mjs';
import {addTiePolicies,tieChoice} from './tie-policy.mjs';
import {actualJoint} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'risk-budget-offgrid.json'),d=JSON.parse(raw),model=JSON.parse(readFileSync(dir+'count-planning-exact.json'));
assert(raw.equals(execFileSync(process.execPath,[dir+'risk-budget-offgrid.mjs'],{maxBuffer:64*1024*1024})));
const compileStats=addTiePolicies(model);assert.deepEqual(compileStats,d.tieStats);
assert.deepEqual(addRegimeSafePolicy(model),d.safeStats);
assert.deepEqual(addRiskBudgetPolicy(model,0),d.safeStats);
let zeroParity=0;
for(const root of model.roots)for(const state of root.states)if(state.actions){
  assert.equal(state.actions.risk_budget,state.actions.regime_safe);zeroParity++;
}
for(const budget of [-1,NaN,Infinity])assert.throws(()=>addRiskBudgetPolicy({},budget));
assert.deepEqual(addRiskBudgetPolicy(model),d.budgetStats);
const changedByDepth=Array(6).fill(0);
for(const root of model.roots)for(const state of root.states)if(state.actions&&state.actions.risk_budget!==state.actions.tie_entropy){
  changedByDepth[state.counts.reduce((a,n)=>a+n,0)]++;
}
const frozenRaw=readFileSync(dir+'risk-budget-evaluation.json');
assert.equal(createHash('sha256').update(frozenRaw).digest('hex'),d.frozenSHA256);
const frozen=JSON.parse(frozenRaw);let frozenHashChecks=0;
for(const [path,hash]of Object.entries(frozen.hashes)){
  assert.equal(createHash('sha256').update(readFileSync(path)).digest('hex'),hash);frozenHashChecks++;
}
assert.equal(tieChoice([1+5e-13,1,2,3]),0);
assert.equal(tieChoice([1+2e-12,1,2,3]),1);
assert.equal(tieChoice([3,2,1,1]),2);
assert.throws(()=>tieChoice([NaN,1,2,3]));
const types=[0,1,2,7];let likelihoodChecks=0,maxLikelihoodError=0;
for(const root of model.roots)for(const s of root.states.filter((_,i)=>i%137===0))for(const noise of [.05,.125,.175,.225,.275,.35])for(const mask of [0,3,7,15]){
  const actual=actualJoint(root.pattern,s.counts,noise,mask),direct=[0,0,0,0];
  for(let h=0;h<16;h++){
    let weight=1/16;
    for(let j=0;j<4;j++){
      const t=types[j],truth=t===7?((h%2)^Math.floor(h/4)%2):Math.floor(h/2**t)%2,error=t===2?.01:noise,p=truth?1-error:error,initial=(root.pattern>>j)&1;
      weight*=initial?p:1-p;
      for(let y=0;y<2;y++)for(let k=0;k<s.counts[2*j+y];k++)weight*=mask&(1<<j)?Number(y===initial):y?p:1-p;
    }
    direct[h%4]+=weight;
  }
  actual.forEach((a,i)=>{const e=Math.abs(a-direct[i]);assert(e<1e-12);maxLikelihoodError=Math.max(maxLikelihoodError,e);likelihoodChecks++;});
}
let valueChecks=0,maxError=0;
for(const[noise,mask]of [[.05,0],[.35,3],[.275,7],[.225,15],[.175,5],[.125,10]]){
  const reference=d.results.find(r=>r.noise===noise&&r.mask===mask);
  const totals=Object.fromEntries(Object.keys(reference.scores).map(p=>[p,{preSum:0,postSum:0,finalBrier:0,accuracy:0,confidentWrong:0}]));
  for(const root of model.roots){
    const states=new Map(root.states.map(s=>[s.counts.join(','),s])),world=new Map();
    const state=key=>{
      if(!world.has(key)){
        const s=states.get(key),joint=actualJoint(root.pattern,s.counts,noise,mask),mass=joint.reduce((a,v)=>a+v,0);
        const loss=mass?joint.reduce((a,v,y)=>a+v*s.forecast.reduce((sum,p,i)=>sum+(p-Number(i===y))**2,0),0)/mass:0;
        const chosen=s.forecast.indexOf(Math.max(...s.forecast)),accuracy=mass?joint[chosen]/mass:0;
        world.set(key,{s,mass,loss,accuracy,confidentWrong:Math.max(...s.forecast)>=.9?1-accuracy:0});
      }
      return world.get(key);
    };
    for(const policy of Object.keys(totals)){
      const memo=new Map();
      const solve=key=>{
        if(memo.has(key))return memo.get(key);
        const w=state(key);assert(w.mass>0);
        const step=w.s.counts.reduce((s,n)=>s+n,0);
        if(step===6){const out={preSum:0,postSum:0,finalBrier:w.loss,accuracy:w.accuracy,confidentWrong:w.confidentWrong};memo.set(key,out);return out;}
        const out={preSum:w.loss,postSum:0,finalBrier:0,accuracy:0,confidentWrong:0};
        for(let a=0;a<4;a++){
          const aw=policy==='random'?.25:Number(w.s.actions[policy]===a);if(!aw)continue;
          for(let y=0;y<2;y++){
            const c=w.s.counts.slice();c[2*a+y]++;const childKey=c.join(','),child=state(childKey);if(!child.mass)continue;
            const next=solve(childKey),probability=aw*child.mass/w.mass;
            out.preSum+=probability*next.preSum;out.postSum+=probability*(child.loss+next.postSum);
            for(const k of ['finalBrier','accuracy','confidentWrong'])out[k]+=probability*next[k];
          }
        }
        memo.set(key,out);return out;
      };
      const key='0,0,0,0,0,0,0,0',value=solve(key),mass=state(key).mass;
      for(const k of Object.keys(totals[policy]))totals[policy][k]+=mass*value[k];
    }
  }
  for(const[policy,out]of Object.entries(totals)){
    out.areaBrier=out.preSum/6;
    for(const k of ['areaBrier','postSum','finalBrier','accuracy','confidentWrong']){
      const e=Math.abs(out[k]-reference.scores[policy][k]);assert(e<1e-11);maxError=Math.max(maxError,e);valueChecks++;
    }
  }
}
const out={scope:'Byte-exact replay, direct ordered likelihood check and separate backward scoring in six representative regimes; frozen forecasts, separate risk-budget compiler, zero-budget parity; frozen source hash verification, off-grid likelihoods and scores',inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-risk-budget-offgrid.mjs')).digest('hex'),replay:true,zeroParity,invalidBudgetTests:3,changedByDepth,frozenHashChecks,tieRuleTests:4,likelihoodChecks,maxLikelihoodError,valueChecks,maxError};
writeFileSync(dir+'risk-budget-offgrid-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
