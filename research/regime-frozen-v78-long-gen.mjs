// External full-history audit: immutable candidate, independent dense oracle.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const dir='internal/researchregimefrozencheck';
if(fs.existsSync(dir))assert.deepEqual(fs.readdirSync(dir).sort(),['long_test.go']);
else fs.mkdirSync(dir,{mode:0o700});
const source='internal/researchregimefrozen/oracle_test.go';
const s=fs.readFileSync(source,'utf8').replace('package researchregimefrozen','package researchregimefrozencheck');
fs.writeFileSync(dir+'/oracle_test.go',s,{flag:'wx',mode:0o600});
let runner=fs.readFileSync('research/regime-frozen-v78-run.mjs','utf8').replaceAll('./internal/researchregimefrozen\'','./internal/researchregimefrozencheck\'')
  .replaceAll('research/regime-frozen-v78-run.mjs','research/regime-frozen-v78-long-run.mjs').replace('`research/regime-frozen-v78-${stage}`','`research/regime-frozen-v78-long-${stage}`');
fs.writeFileSync('research/regime-frozen-v78-long-run.mjs',runner,{flag:'wx',mode:0o600});
const h=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.writeFileSync('research/regime-frozen-v78-long-generation.json',JSON.stringify({source,sourceSHA256:h(fs.readFileSync(source)),externalOracleSHA256:h(s),runnerSHA256:h(runner),scope:'full64 unknown/reverse-arrival histories and fixed-support pair branches; no candidate edits or inherited empirical goal pass'},null,2)+'\n',{flag:'wx',mode:0o600});
