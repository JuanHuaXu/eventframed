// Mechanical reuse of the frozen-command runner; only isolated target changes.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const source='research/regime-v77-run.mjs',out='research/regime-frozen-v78-run.mjs';
let s=fs.readFileSync(source,'utf8');
s=s.replaceAll('regime-v77-', 'regime-frozen-v78-').replaceAll('./internal/researchregime\'', './internal/researchregimefrozen\'')
  .replaceAll('research/regime-v77-run.mjs','research/regime-frozen-v78-run.mjs')
  .replaceAll('docs/experiments/mmm-regime-v77-protocol.md','docs/experiments/mmm-regime-frozen-v78-protocol.md')
  .replaceAll('EVENTFRAME_REGIME_V77_REPORT','EVENTFRAME_REGIME_V78_REPORT');
const marker="const closure = JSON.parse";
assert.equal(s.split(marker).length-1,1);
s=s.replace(marker,`const priorRun='research/regime-v77-initial',priorFreeze=JSON.parse(fs.readFileSync(priorRun+'/freeze.json')),priorDone=JSON.parse(fs.readFileSync(priorRun+'/completed.json'));
assert(priorDone.allJobsTerminal&&priorDone.allChecksPass&&priorDone.checks.length===4);
for(const[p,h]of Object.entries(priorFreeze.files))assert.equal(await hash(p),h,p);
for(const x of priorDone.checks)assert.equal(await hash(priorRun+'/'+x.name+'.log'),x.logSHA256);
${marker}`);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.writeFileSync(out,s,{flag:'wx',mode:0o600});
fs.writeFileSync('research/regime-frozen-v78-tools-generation.json',JSON.stringify({source,sourceSHA256:hash(fs.readFileSync(source)),out,initialSHA256:hash(s),scope:'same immutable parent/protected-files gate; original V77 attempt verified terminal/unchanged before fixed-support tests'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({out}));
