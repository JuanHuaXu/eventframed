// Mechanical clones of the checker regression and readback; no learner edits.
import fs from 'node:fs';import assert from 'node:assert/strict';
const src=fs.readFileSync('internal/researchdispersion/retention_v64_test.go','utf8');
const test=src.slice(src.indexOf('func TestRetentionV64FutureAndCorruptions(')).replaceAll('TestRetentionV64FutureAndCorruptions','TestRetentionV64TieCorruptions').replaceAll('independentRetentionV64','independentRetentionV64TieAudit');
assert(test.startsWith('func TestRetentionV64TieCorruptions'));
fs.writeFileSync('internal/researchdispersion/retention_v64_tie_corruptions_test.go','package researchdispersion\nimport("encoding/json";"testing")\n'+test,{flag:'wx',mode:0o600});
let s=fs.readFileSync('research/retention-v64b-readback.mjs','utf8').replaceAll('retention-v64b-diagnostic','retention-v64c-audit');
s=s.replace("load(root+'/diagnostic.jsonl')","load('research/retention-v64b-diagnostic/diagnostic.jsonl')");
s=s.replace('assert.deepEqual(rows[0].Sources,freeze.files);',"assert.deepEqual(rows[0].Sources,JSON.parse(fs.readFileSync('research/retention-v64b-diagnostic/freeze.json')).files);");
s=s.replace("independentIssueComparisons:864000","independentIssueComparisons:864000,ordering:JSON.parse(fs.readFileSync(root+'/ordering.json')),originalCollectionUnchanged:true,originalAuditFailurePreserved:true");
fs.writeFileSync('research/retention-v64c-readback.mjs',s,{flag:'wx',mode:0o600});
