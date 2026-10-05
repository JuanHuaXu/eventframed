// Preserve the raw diagnostic's Error field; classify its bad cardinality
// assertion separately from actual Recall failures using observable outputs.
import fs from 'node:fs';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./task-lexical-churn-results.json',import.meta.url)));
let serviceErrors=0,measurementErrors=0,successes=0;
const rows=d.Arms.map(a=>{
 let failed=0,miscount=0;
 for(const s of a.Samples){
  if(s.Error&&!s.Error.startsWith('frontier mismatch:')){failed++;assert.equal(s.Packed,0);continue;}
  assert(s.JournalMatched);assert.equal(s.FutureRecords,0);
  assert(s.Nominated>=a.N&&s.Nominated<=a.N+(a.Mode==='in-window'?16:0));
  if(s.Error)miscount++;
  successes++;
 }
 serviceErrors+=failed;measurementErrors+=miscount;
 return {repeat:a.Repeat,n:a.N,enabled:a.Enabled,mode:a.Mode,serviceErrors:failed,measurementErrors:miscount,staleRejections:a.StaleRejections,maxMS:Math.max(...a.Samples.map(s=>s.NS))/1e6};
});
console.log(JSON.stringify({serviceErrors,measurementErrors,successfulReturns:successes,rows},null,2));
