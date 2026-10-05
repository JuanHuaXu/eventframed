import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {addRegimeSafePolicy} from './regime-safe-policy.mjs';
import {addTiePolicies,tieChoice} from './tie-policy.mjs';
import {actualJoint} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'regime-safe-evaluation.json'),d=JSON.parse(raw),model=JSON.parse(readFileSync(dir+'count-planning-exact.json'));
assert(raw.equals(execFileSync(process.execPath,[dir+'regime-safe-evaluation.mjs'],{maxBuffer:64*1024*1024})));
const compileStats=addTiePolicies(model);assert.deepEqual(compileStats,d.tieStats);
assert.deepEqual(addRegimeSafePolicy(model),d.safeStats);
const archived=JSON.parse(readFileSync(dir+'tie-policy-evaluation.json'));
let archivedChecks=0;
for(const row of archived.results){
  const current=d.results.find(r=>r.noise===row.noise&&r.mask===row.mask);
  for(const [policy,score]of Object.entries(row.scores)){assert.deepEqual(current.scores[policy],score);archivedChecks++;}
}
assert.equal(tieChoice([1+5e-13,1,2,3]),0);
assert.equal(tieChoice([1+2e-12,1,2,3]),1);
assert.equal(tieChoice([3,2,1,1]),2);
assert.throws(()=>tieChoice([NaN,1,2,3]));
const types=[0,1,2,7];let likelihoodChecks=0,maxLikelihoodError=0;
for(const root of model.roots)for(const s of root.states.filter((_,i)=>i%137===0))for(const noise of [.1,.2,.3])for(const mask of [0,3,7,15]){
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
for(const[noise,mask]of [[.1,0],[.3,3],[.25,7],[.2,15],[.2,5],[.3,10]]){
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
const out={scope:'Byte-exact replay, direct ordered likelihood check and separate backward scoring in six representative regimes; frozen forecasts, separate regime-safe compiler; archived-policy score parity',inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-regime-safe.mjs')).digest('hex'),replay:true,archivedChecks,tieRuleTests:4,likelihoodChecks,maxLikelihoodError,valueChecks,maxError};
writeFileSync(dir+'regime-safe-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
