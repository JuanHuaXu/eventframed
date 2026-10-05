import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {addTiePolicies} from './tie-policy.mjs';
import {addRiskBudgetPolicy} from './risk-budget-policy.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'copy-trigger-audit.json'),d=JSON.parse(raw);
assert(raw.equals(execFileSync(process.execPath,[dir+'copy-trigger-audit.mjs'],{maxBuffer:64*1024*1024})));
const model=JSON.parse(readFileSync(dir+'count-planning-exact.json'));addTiePolicies(model);addRiskBudgetPolicy(model);
const unequal=Array.from({length:16},()=>Array(7).fill(0));let enumeratedPrefixes=0,oddTraceChecks=0;
for(const root of model.roots){
  const states=new Map(root.states.map(s=>[s.counts.join(','),s]));
  for(let mask=0;mask<16;mask++){
    const runs={};
    for(const policy of ['risk_budget','tie_entropy']){
      const maps=Array.from({length:7},()=>new Map()),traces=[];
      // Enumerate individual ordered paths, unlike the audit's merged propagation.
      function visit(counts,trace){
        enumeratedPrefixes++;const k=trace.length,key=counts.join(',');
        maps[k].set(key,(maps[k].get(key)||0)+1);
        if(k===6){traces.push(JSON.stringify(trace));return;}
        const a=states.get(key).actions[policy];
        for(let y=0;y<2;y++){
          if((mask&(1<<a))&&y!==((root.pattern>>a)&1))continue;
          const next=counts.slice();next[2*a+y]++;visit(next,[...trace,[a,y]]);
        }
      }
      visit(Array(8).fill(0),[]);runs[policy]={maps,traces:traces.sort()};
    }
    for(let k=0;k<=6;k++){
      const x=runs.risk_budget.maps[k],y=runs.tie_entropy.maps[k];
      if([...new Set([...x.keys(),...y.keys()])].some(key=>(x.get(key)||0)!==(y.get(key)||0)))unequal[mask][k]++;
    }
    if(mask&1){assert.deepEqual(runs.risk_budget.traces,runs.tie_entropy.traces);oddTraceChecks++;}
  }
}
for(let mask=0;mask<16;mask++)assert.deepEqual(unequal[mask],d.masks[mask].unequalLayers);
assert(d.masks.filter(m=>!(m.mask&1)).every(m=>m.unequalLayers[6]>0));
assert.equal(d.counterexamples.length,0);
const out={scope:'Byte-exact replay plus separate individual ordered-path enumeration; odd masks have identical complete action/outcome traces',enumeratedPrefixes,oddTraceChecks,layerChecks:16*7,replay:true,inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-copy-trigger.mjs')).digest('hex')};
writeFileSync(dir+'copy-trigger-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
