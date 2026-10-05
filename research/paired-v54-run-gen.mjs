// Mechanical harness clone only; all model/policy code is separately saved.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='research/noise-v53-run.mjs',original=fs.readFileSync(source,'utf8');
let s=original.replaceAll('noise-v53','paired-v54').replaceAll('NOISE_V53','PAIRED_V54').replaceAll('NoiseV53','PairedV54').replaceAll('researchnoisemomentref','researchpairedref').replaceAll('researchnoisemoment','researchpaired').replaceAll('2026105307','2026105407').replaceAll('2026105309','2026105409').replaceAll('2026105311','2026105411').replaceAll('18regimes','20regimes');
const start=s.indexOf('const extras='),end=s.indexOf('\nconst files=',start);assert(start>0&&end>start);
s=s.slice(0,start)+"const extras=['go.mod','go.sum','research/paired-v54-run-gen.mjs','research/paired-v54-generation.json','research/paired-v54-run.mjs','research/paired-v54-readback.mjs','research/paired-v54-preflight.md','docs/experiments/mmm-paired-v54-protocol.md','research/noise-v54-direction.md','research/noise-v54-identities.mjs','research/noise-v54-identities.json'];"+s.slice(end);
s=s.replace('FutureAndCorruptions|SeedSeparation','FutureAndCorruptions|SeedSeparation|ControlMetricGuard');
function save(p,b){if(fs.existsSync(p))assert.equal(fs.readFileSync(p,'utf8'),b,p);else fs.writeFileSync(p,b,{flag:'wx',mode:0o600});}
save('research/paired-v54-run.mjs','// Generated mechanically by paired-v54-run-gen.mjs.\n'+s);
const sources={[source]:hash(original),'research/paired-v54-run-gen.mjs':hash(fs.readFileSync('research/paired-v54-run-gen.mjs'))};
save('research/paired-v54-generation.json',JSON.stringify({sources,outputs:{'research/paired-v54-run.mjs':hash(fs.readFileSync('research/paired-v54-run.mjs'))},modelOrPolicyGeneration:false},null,2)+'\n');
