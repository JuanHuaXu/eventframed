// Additional audit of unchanged collected data, NOT a rerun or retuned model.
import fs from 'node:fs';import assert from 'node:assert/strict';
let s=fs.readFileSync('research/retention-v64b-run.mjs','utf8').replaceAll('retention-v64b-diagnostic','retention-v64c-audit').replaceAll('retention-v64b-run.mjs','retention-v64c-run.mjs').replaceAll('retention-v64b-readback.mjs','retention-v64c-readback.mjs');
s=s.replace("'docs/experiments/mmm-retention-v64-protocol.md'];","'docs/experiments/mmm-retention-v64-protocol.md','docs/experiments/mmm-retention-v64-ordering-audit.md','research/retention-v64c-gen.mjs','research/retention-v64c-run-gen.mjs','research/retention-v64-tie-gen.mjs','research/retention-v64-tie-generation.json','research/retention-v64-tie-diagnose.mjs','research/retention-v64-tie-diagnostic.json','research/retention-v64-tie-diagnostic.log','research/retention-v64-tie-diagnostic-source.go'];");
s=s.replace("'research/retention-v63-integration/allocation.json']","'research/retention-v63-integration/allocation.json','research/retention-v64b-diagnostic/diagnostic.jsonl','research/retention-v64b-diagnostic/freeze.json','research/retention-v64b-diagnostic/failure.json','research/retention-v64b-diagnostic/experiment-command.json']");
const begin=s.indexOf('try {\n await run('),end=s.indexOf(' const artifacts={}',begin);assert(begin>=0&&end>begin);
s=s.slice(0,begin)+`try {
 await run('race-fixture',['test','-race','./internal/researchdispersion','-run','^TestRetentionV64TieCorruptions$','-count=1','-v','-timeout=20m']);
 await run('vet',['vet',...packages]);
 await run('audit',['test','./internal/researchdispersion','-run','^TestRetentionV64TieAudit$','-count=1','-v','-timeout=50m'],{EVENTFRAME_RETENTION_V64_DATA_FREEZE:path.resolve('research/retention-v64b-diagnostic/freeze.json'),EVENTFRAME_RETENTION_V64_AUDIT:path.resolve('research/retention-v64b-diagnostic/diagnostic.jsonl'),EVENTFRAME_RETENTION_V64_TIE_STATS:path.resolve(root+'/ordering.json')});
`+s.slice(end);
s=s.replace("scope:'all40 consumed worlds x3delays x5arms; no fresh confirmation or goal7 claim'","scope:'additional independent replay of unchanged original600-arm data; numerical near-tie clarification, NOT a model/data/gate change',originalCollection:'research/retention-v64b-diagnostic',dataReusedNoRerun:true");
fs.writeFileSync('research/retention-v64c-run.mjs',s,{flag:'wx',mode:0o600});
