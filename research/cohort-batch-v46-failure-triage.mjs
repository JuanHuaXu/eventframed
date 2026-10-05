// Preserve per-arm failures without laundering a partial workload into a pass.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import {check} from './cohort-batch-v46-audit.mjs';
import {readEnvelope} from './eager-load-v44-stream.mjs';
const root='research/cohort-batch-v46';
assert(fs.existsSync(root+'/failure.json'));
const rows=[];
const envelope=await readEnvelope(root+'/raw.ndjson',row=>{
  let accepted=false,reason='';try{check(row);accepted=true}catch(e){reason=e.message}
  rows.push({trial:row.Trial,cap:row.BatchCap,visible:row.Visible,archived:row.Archived,eager:row.EagerEnabled,
    reads:row.Reads.length,writes:row.Writes.length,outcomes:row.Outcomes.length,
    acceptedByFullChecker:accepted,rejection:reason,latencyPass:accepted && row.Pass,
    firstFailedBatch:row.Batches.find(b=>b.Error),maximumBatch:Math.max(...row.Batches.map(b=>b.IDs.length)),
    errors:[...new Set(row.Errors??[])]});
},32);
const report={time:new Date().toISOString(),rawSHA256:envelope.rawSHA256,rows,
  controlAccepted:rows.filter(r=>r.cap===4&&r.acceptedByFullChecker).length,
  candidateAccepted:rows.filter(r=>r.cap===8&&r.acceptedByFullChecker).length,
  fullStudyAudit:'FAIL',allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync(root+'/failure-triage.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(report,null,2));
