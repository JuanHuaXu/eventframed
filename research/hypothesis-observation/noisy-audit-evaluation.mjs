import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {compileAudit,auditPaths,auditLikelihood} from './noisy-audit.mjs';
import {actualJoint} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',policies=['target','entropy','random'],roots=[];
for(let p=0;p<16;p++)for(let a=0;a<2;a++)roots.push(compileAudit(p,a));
const prepared=roots.map(root=>({root,paths:Object.fromEntries(policies.map(p=>[p,auditPaths(root,p)]))}));
const loss=(joint,forecast)=>joint.reduce((a,v,y)=>a+v*forecast.reduce((b,p,z)=>b+(p-Number(y===z))**2,0),0);
const results=[],comparisons=[];let massChecks=0;
for(const noise of [.1,.15,.2,.25,.3])for(let mask=0;mask<16;mask++){
  const scores=Object.fromEntries(policies.map(p=>[p,{finalBrier:0,areaBrier:0,mass:Array(6).fill(0)}]));
  for(const{root,paths}of prepared){
    const auditP=auditLikelihood(mask,root.audit),initial=actualJoint(root.pattern,Array(8).fill(0),noise,mask).map(v=>v*auditP);
    for(const p of policies)scores[p].areaBrier+=loss(initial,root.preAudit)/6;
    root.states.forEach((s,i)=>{
      const step=s.counts.reduce((a,v)=>a+v,0),joint=actualJoint(root.pattern,s.counts,noise,mask).map(v=>v*auditP),mass=joint.reduce((a,v)=>a+v,0),risk=loss(joint,s.forecast);
      for(const p of policies){const w=paths[p][step].get(i)||0;if(!w)continue;scores[p].mass[step]+=w*mass;
        if(step<5)scores[p].areaBrier+=w*risk/6;else scores[p].finalBrier+=w*risk;
      }
    });
  }
  for(const out of Object.values(scores))for(const mass of out.mass){assert(Math.abs(mass-1)<1e-11);massChecks++;}
  for(const control of ['random','entropy']){
    const finalGain=scores[control].finalBrier-scores.target.finalBrier,areaGain=scores[control].areaBrier-scores.target.areaBrier;
    comparisons.push({noise,mask,control,finalGain,areaGain,finalNonharm:finalGain>=-.01,areaPositive:mask===15?null:areaGain>0});
  }results.push({noise,mask,scores});
}
const files=['NOISY_AUDIT_PROTOCOL.md','noisy-audit.mjs','noisy-audit-evaluation.mjs','noise-envelope.mjs','count-planning.mjs','interval-acquisition.mjs','tie-policy.mjs','exact-regime.mjs'];
console.log(JSON.stringify({scope:'Hypothetical noisy source audit; equal16-credit pilot with five renewals; no validated real-world audit reliability',hashes:Object.fromEntries(files.map(f=>[f,createHash('sha256').update(readFileSync(dir+f)).digest('hex')])),massChecks,roots,results,comparisons}));
