// Independent descriptive pairing; consumed results are not a new selection set.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {readEnvelope} from './eager-load-v44-stream.mjs';

const root=process.argv[2]??'research/cohort-batch-v46';
assert(/^research\/cohort-batch-v46(?:-[a-z0-9-]+)?$/.test(root));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const doneBytes=fs.readFileSync(root+'/completed.json'), done=JSON.parse(doneBytes);
const auditBytes=fs.readFileSync(root+'/audit.json'), audit=JSON.parse(auditBytes);
assert(done.sourceUnchanged && done.checks.every(c=>c.code===0));
assert.equal(audit.trials.length,32);
for(const cap of [4,8]) assert.equal(audit.controls[cap].length,54);
const counts=[];
const envelope=await readEnvelope(root+'/raw.ndjson',row=>{
  assert([4,8].includes(row.BatchCap));
  const sizes=row.Batches.map(b=>b.IDs.length);
  assert.equal(sizes.reduce((a,b)=>a+b,0),129);
  assert(sizes.every(n=>n>=1 && n<=row.BatchCap));
  const matched=audit.trials.find(t=>t.trial===row.Trial && t.visible===row.Visible && t.archived===row.Archived && t.eager===row.EagerEnabled && t.batch_cap===row.BatchCap);
  assert(matched);assert.equal(matched.journal_batches,sizes.length);
  const histogram=Array.from({length:8},(_,i)=>sizes.filter(n=>n===i+1).length);
  assert.deepEqual(matched.batch_histogram,histogram);
  counts.push({trial:row.Trial,visible:row.Visible,archived:row.Archived,eager:row.EagerEnabled,cap:row.BatchCap,
    batches:sizes.length,maximumBatch:Math.max(...sizes),meanBatch:129/sizes.length,aboveFour:sizes.filter(n=>n>4).length});
},32);
assert.equal(envelope.rawSHA256,audit.rawSHA256);assert.equal(envelope.rawSHA256,done.rawSHA256);
const gates={call:['call_p99_ns',100e6],offer:['offer_p99_ns',100e6],outcome:['outcome_p99_ns',100e6],outcomeMax:['outcome_max_ns',250e6],write:['write_p99_ns',250e6],view:['view_max_ns',250e6]};
const paired=[];
for(const trial of [1,2]) for(const visible of [false,true]) for(const archived of [false,true]) for(const eager of [false,true]){
  const arms=[4,8].map(cap=>audit.trials.find(t=>t.trial===trial && t.visible===visible && t.archived===archived && t.eager===eager && t.batch_cap===cap));
  assert(arms.every(Boolean));const [control,candidate]=arms;
  const delta=Object.fromEntries(Object.entries(gates).map(([name,[key]])=>[name,(candidate.metrics[key]-control.metrics[key])/1e6]));
  paired.push({trial,visible,archived,eager,control:control.metrics,candidate:candidate.metrics,deltaMS:delta,
    batches4:control.journal_batches,batches8:candidate.journal_batches,pass4:control.pass,pass8:candidate.pass});
}
const cells=[];
for(const visible of [false,true]) for(const archived of [false,true]) for(const eager of [false,true]) for(const cap of [4,8]){
  const rows=audit.trials.filter(t=>t.visible===visible && t.archived===archived && t.eager===eager && t.batch_cap===cap);
  assert.equal(rows.length,2);const range=key=>{const values=rows.map(t=>t.metrics[key]/1e6);return [Math.min(...values),Math.max(...values)]};
  cells.push({visible,archived,eager,cap,callP99MS:range('call_p99_ns'),offerP99MS:range('offer_p99_ns'),outcomeP99MS:range('outcome_p99_ns'),outcomeMaxMS:range('outcome_max_ns'),batches:rows.map(t=>t.journal_batches),passes:rows.filter(t=>t.pass).length});
}
const arms=Object.fromEntries([4,8].map(cap=>{
  const rows=audit.trials.filter(t=>t.batch_cap===cap);
  return [cap,{trials:rows.length,passed:rows.filter(t=>t.pass).length,failedGates:Object.fromEntries(Object.entries(gates).map(([name,[key,bound]])=>[name,rows.filter(t=>!(t.metrics[key]<bound)).length])),
    greaterThanFourBatches:counts.filter(t=>t.cap===cap).reduce((n,t)=>n+t.aboveFour,0)}];
}));
const pairedDirection=Object.fromEntries(Object.keys(gates).map(name=>[name,{better:paired.filter(p=>p.deltaMS[name]<0).length,worse:paired.filter(p=>p.deltaMS[name]>0).length,equal:paired.filter(p=>p.deltaMS[name]===0).length}]));
const report={time:new Date().toISOString(),scriptSHA256:hash(fs.readFileSync('research/cohort-batch-v46-readback.mjs')),completedSHA256:hash(doneBytes),auditSHA256:hash(auditBytes),rawSHA256:envelope.rawSHA256,
  arms,cells,counts,paired,pairedDirection,originalGatesUnchanged:true,alternatedOrderTwoRepeatsNotPopulationCertification:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync(root+'/readback.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(report,null,2));
