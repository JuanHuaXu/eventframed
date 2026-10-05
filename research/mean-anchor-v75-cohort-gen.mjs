// Mechanical harness reuse, not an edit to the frozen model or prior study.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const source='internal/researchdispersion/mean_joint_v74_test.go',raw=fs.readFileSync(source),out='internal/researchdispersion/mean_anchor_v75_test.go',sha=b=>crypto.createHash('sha256').update(b).digest('hex');
let s=raw.toString().replaceAll('MeanJointV74','MeanAnchorV75').replaceAll('meanJointV74','meanAnchorV75').replaceAll('researchmeanjoint"','researchmeananchor"').replaceAll('EVENTFRAME_MEANJOINT_V74','EVENTFRAME_MEANANCHOR_V75');
assert(s.includes('researchmeanjointref"'));assert(!s.includes('researchmeanjoint"'));
fs.writeFileSync(out,s,{flag:'wx',mode:0o600});fs.writeFileSync('research/mean-anchor-v75-cohort-generation.json',JSON.stringify({source,sourceSHA256:sha(raw),out,initialSHA256:sha(s),modes:JSON.parse(fs.readFileSync('research/mean-joint-v74-study-generation.json')).modes,scope:'all30 native-policy arms; same40 worlds/three delays/public consumed base; unchanged controls and gates; independent reference function retained but no fresh all-arm dense replay claimed'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({out,sourceSHA256:sha(raw)}));
