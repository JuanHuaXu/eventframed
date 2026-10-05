import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(process.argv[2] ?? 'research/public-task-pilot/batch-reconcile-results.json'));
const original=JSON.parse(fs.readFileSync(process.argv[3] ?? 'research/public-task-pilot/batch-queue-clean-results.json'));
const digest=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for(const [p,h]of Object.entries(d.Hashes))assert.equal(digest(p),h,p);
assert.equal(d.Results.length,32);
const seen=new Set();let failed=0,presentAfterError=0,retried=0,acknowledged=0;
const errorClasses={};
for(const r of d.Results) {
 assert(!seen.has(r.Source));seen.add(r.Source);
 assert(r.SourceUnchanged);assert.equal(digest(r.Source),r.SourceHash);
 assert.notEqual(r.Source,r.Copy);
 const a=original.Arms.find(a=>a.DatabasePath===r.Source);assert(a);
 assert.equal(r.Writes.length,16);assert.equal((r.Errors??[]).length,0);
 assert.equal(r.SearchCount,r.DirectCount);
 assert.equal((r.SearchMissingFromDirect??[]).length,0);
 assert.equal((r.DirectMissingFromSearch??[]).length,0);
 assert.equal(r.RecallError,'');
 assert.equal(r.RecallNominated,r.Mode==='future'?200:r.DirectCount);
 assert.equal(r.DirectCount,200+r.Writes.filter(w=>w.Present).length);
 // Only valid for this policy-bind + insert-only fixture.
 assert.equal(r.Before.evidence_epoch,r.DirectCount);
 assert.equal(r.Before.runtime_version,r.DirectCount+1);
 assert.equal(r.After.evidence_epoch,216);assert.equal(r.After.runtime_version,217);
 for(const [i,w]of r.Writes.entries()) {
  assert.equal(w.ID,`writer-${String(i).padStart(3,'0')}`);
  assert.equal(w.OriginalError,a.Writes[i].Error);
  assert.equal(w.Entered,a.Writes[i].Entered);
  assert(w.PresentAfter);assert.equal(w.RetryError,'');
  if(w.Present) {assert(w.PayloadMatches);assert.equal(w.LookupError,'');}
  if(w.OriginalError) {
   failed++;presentAfterError+=Number(w.Present);retried++;
   assert.equal(w.RetryDuplicate,w.Present);
   const c=errorClasses[w.OriginalError]??={count:0,present:0};c.count++;c.present+=Number(w.Present);
  } else {acknowledged++;assert(w.Present);}
 }
}
assert.equal(original.Arms.length,32);
const expectedFailed=original.Arms.reduce((n,a)=>n+a.Writes.filter(w=>w.Error).length,0);
assert.equal(failed,expectedFailed);assert.equal(acknowledged,512-expectedFailed);
console.log(JSON.stringify({status:'PASS',stores:32,failed,presentAfterError,retried,acknowledged,errorClasses,scope:'orderly reopen and retry on copies; not a rollback or crash theorem'},null,2));
