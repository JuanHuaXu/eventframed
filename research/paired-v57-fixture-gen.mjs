import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const sources = {}, initialOutputs = {};
function clone(src, dest, transform) {
  assert(!fs.existsSync(dest));
  const b = fs.readFileSync(src);
  sources[src] = hash(b);
  const s = transform(b.toString());
  fs.writeFileSync(dest, s, { flag: 'wx', mode: 0o600 });
  initialOutputs[dest] = hash(s);
}
fs.mkdirSync('internal/researchpairedvalueref', { recursive: true });
clone('internal/researchpairedref/reference.go', 'internal/researchpairedvalueref/reference.go', s => s.replaceAll('researchpairedref', 'researchpairedvalueref'));
clone('internal/researchdispersion/paired_v56_generated_test.go', 'internal/researchdispersion/paired_v57_generated_test.go', s => s.replaceAll('V56', 'V57').replaceAll('researchpairedmemo', 'researchpairedvalue').replaceAll('researchpairedref', 'researchpairedvalueref'));
clone('research/paired-v56-run.mjs', 'research/paired-v57-run.mjs', s => s.replaceAll('v56', 'v57').replaceAll('V56', 'V57').replaceAll('researchpairedmemo', 'researchpairedvalue').replaceAll('researchpairedref', 'researchpairedvalueref'));
clone('research/paired-v56-readback.mjs', 'research/paired-v57-readback.mjs', s => s.replaceAll('v56', 'v57').replaceAll("'falsification'];", "'falsification','predictive'];").replaceAll('length,21', 'length,24').replaceAll('j<21', 'j<24').replaceAll('j%7', 'j%8').replaceAll('schedule*7', 'schedule*8').replaceAll('schedule*8+7', 'schedule*8+8').replaceAll('information,falsification]', 'information,falsification,predictive]').replaceAll('information,falsification])', 'information,falsification,predictive])').replaceAll('arms:840', 'arms:960'));
clone('research/paired-v56-equivalence.mjs', 'research/paired-v57-control-equivalence.mjs', s => s.replaceAll('v56', 'v57').replaceAll('b.Arms.length===21', 'b.Arms.length===24').replace('structuredClone(b.Arms[j])', 'structuredClone(b.Arms[Math.floor(j/7)*8+j%7])'));
fs.writeFileSync('research/paired-v57-fixture-generation.json', JSON.stringify({ sources, initialOutputs, behavioralPatchesSeparate: true }, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify({ copies: Object.keys(initialOutputs).length }));
