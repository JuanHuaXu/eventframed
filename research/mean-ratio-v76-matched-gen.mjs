// Public matched benchmark and fixture, unchanged observations and trial positions.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex'),outputs={};
function emit(out,source,s){assert(!fs.existsSync(out));fs.writeFileSync(out,s,{flag:'wx',mode:0o600});outputs[out]={source,sourceSHA256:sha(fs.readFileSync(source)),initialSHA256:sha(s)}}
let source='internal/researchmeanperfcheck/performance_test.go',s=fs.readFileSync(source,'utf8');s=s.replaceAll('v74','__previous__').replaceAll('v75','v76').replaceAll('__previous__','v75').replaceAll('researchmeananchor','researchmeanratio').replaceAll('researchmeanjoint','researchmeananchor').replaceAll('researchmeanperfcheck','researchmeanratioperfcheck');
fs.mkdirSync('internal/researchmeanratioperfcheck',{mode:0o700});emit('internal/researchmeanratioperfcheck/performance_test.go',source,s);
source='research/mean-anchor-v75-matched.mjs';s=fs.readFileSync(source,'utf8').replaceAll('mean-anchor-v75','mean-ratio-v76').replaceAll('researchmeanperfcheck','researchmeanratioperfcheck');
assert.equal(s.split('research/mean-joint-v74-diagnostic').length-1,2);s=s.replaceAll('research/mean-joint-v74-diagnostic','research/mean-anchor-v75-diagnostic');
emit('research/mean-ratio-v76-matched.mjs',source,s);fs.writeFileSync('research/mean-ratio-v76-matched-generation.json',JSON.stringify({outputs,scope:'same oldest/latest row, same150-member/16-round public fixture, priorV75 versus conditional-factorV76; no fullcohort/loadedserving/equalcost claim'},null,2)+'\n',{flag:'wx',mode:0o600});
