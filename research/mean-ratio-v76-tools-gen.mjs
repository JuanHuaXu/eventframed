// Mechanical preflight reuse with separately frozen source and evidence boundary.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const source='research/mean-anchor-v75-preflight.mjs',raw=fs.readFileSync(source),sha=b=>crypto.createHash('sha256').update(b).digest('hex');
let s=raw.toString().replaceAll('mean-anchor-v75','mean-ratio-v76').replaceAll('researchmeananchor','researchmeanratio').replaceAll('EVENTFRAME_MEANANCHOR_V75','EVENTFRAME_MEANRATIO_V76');
function rep(a,b){assert.equal(s.split(a).length-1,1,a);s=s.replace(a,b)}
rep('research/checkpoint-2026-10-04-dynvarcache-v72-mean-v73/manifest.json','research/checkpoint-2026-10-05-mean-joint-v74-anchor-v75/manifest.json');
rep('d4045ac945e8044155f54fb0871d8b2fffb30fa8a4806d431beaf028ea22cc54','f4f547b44b8f2ac25bcd58155d0882c410cbb37dcc63a1c7806cc1afa18e27b4');
rep("const v74='research/mean-joint-v74-diagnostic'","const v74='research/mean-anchor-v75-diagnostic'");
rep('done.checks.length===6','done.checks.length===4');
rep('no test concurrency with V74 timed screen','no test concurrency with V75 timed screen');
rep('same full27-mean joint model, anchored revealed prefix with complete retained journal; independent dense and full64 audits, not whole-study equivalence or a quality rescue','same full27-mean joint model, exact at-anchor conditional pair replacement with full older-pair replay; dense/full64/support/branch audits, not whole-study equivalence or a quality rescue');
const out='research/mean-ratio-v76-preflight.mjs';fs.writeFileSync(out,s,{flag:'wx',mode:0o600});fs.writeFileSync('research/mean-ratio-v76-tools-generation.json',JSON.stringify({source,sourceSHA256:sha(raw),out,initialSHA256:sha(s),scope:'source archived before unit/race/vet/serial benchmarks; no overlap with timedV75 cohort; no whole76cohort yet'},null,2)+'\n',{flag:'wx',mode:0o600});
