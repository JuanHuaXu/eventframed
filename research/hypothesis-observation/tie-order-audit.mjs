import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';
import {acquisitionBelief,acquisitionChoice,TYPES} from './interval-acquisition.mjs';
import {lookaheadChoice} from './interval-lookahead.mjs';
import {canonicalHistory,countState,createCountPlanner} from './count-planning.mjs';

const vectors=[];
function enumerate(prefix,remaining) {
  if(prefix.length===7){vectors.push([...prefix,remaining]);return;}
  for(let n=0;n<=remaining;n++)enumerate([...prefix,n],remaining-n);
}
for(let n=0;n<=2;n++)enumerate([],n);
vectors.push([1,1,0,1,0,0,0,0],[1,0,0,1,1,0,0,2]);
const out={protocol:'TIE_ORDER_PROTOCOL.md',states:0,posteriorChecks:0,
  maxPosteriorError:0,policies:{},examples:[],sourceHashes:{}};
for(const policy of ['one','two','entropy'])out.policies[policy]={comparisons:0,
  canonicalDisagreements:0,reversedDisagreements:0,orderChanges:0,
  maxCanonicalExcess:0,aboveTolerance:0};
for(let roots=0;roots<16;roots++){
  const planner=createCountPlanner(roots);
  for(const counts of vectors){
    const h=canonicalHistory(roots,counts),reverse=h.slice(0,4),slots=[0,0,0,0];
    for(const s of h.slice(4).reverse()){
      const j=TYPES.indexOf(s.action[1]);
      reverse.push({action:[1,TYPES[j],slots[j]++],outcome:s.outcome});
    }
    assert.deepEqual(countState(reverse),{roots,counts});out.states++;
    const b=acquisitionBelief(h),r=acquisitionBelief(reverse);
    for(const error of [Math.abs(b.mass-r.mass),...b.forecast.map((p,i)=>Math.abs(p-r.forecast[i]))]){
      out.posteriorChecks++;out.maxPosteriorError=Math.max(out.maxPosteriorError,error);
      assert(error<1e-12);
    }
    for(const policy of ['one','two','entropy']){
      const choose=history=>TYPES.indexOf((policy==='two'?lookaheadChoice(history):
        acquisitionChoice(history,policy==='one'?'target':'entropy',()=>{throw Error('unexpected RNG');}))[1]);
      const a=choose(h),z=choose(reverse),reference=planner.weights(counts,policy).indexOf(1);
      const costs=policy==='entropy'?Array.from({length:4},(_,j)=>planner.branches(counts,j)
        .reduce((s,c)=>s+c.probability*Math.log(c.probability),0)):planner.plan(counts,policy==='one'?1:2).costs;
      const minimum=Math.min(...costs),excess=Math.max(costs[a]-minimum,costs[z]-minimum);
      const record=out.policies[policy];record.comparisons++;
      record.canonicalDisagreements+=Number(a!==reference);
      record.reversedDisagreements+=Number(z!==reference);
      record.orderChanges+=Number(a!==z);
      record.maxCanonicalExcess=Math.max(record.maxCanonicalExcess,excess);
      record.aboveTolerance+=Number(excess>1e-12);
      if(a!==z&&out.examples.length<12)out.examples.push({roots,counts,policy,a,z,reference,costs,excess});
    }
  }
}
for(const file of ['TIE_ORDER_PROTOCOL.md','tie-order-audit.mjs','count-planning.mjs',
  'interval-acquisition.mjs','interval-lookahead.mjs','noise-envelope.mjs']){
  out.sourceHashes[file]=crypto.createHash('sha256').update(fs.readFileSync(new URL(file,import.meta.url))).digest('hex');
}
const output=JSON.stringify(out,null,2)+'\n';
if(process.argv[2])fs.writeFileSync(process.argv[2],output,{flag:'wx',mode:0o600});
process.stdout.write(output);
