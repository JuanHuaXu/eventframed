// Mechanical isolated clone. The prediction-value implementation is separate.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const sources = {}, initialOutputs = {};
for (const name of ['model.go', 'model_test.go', 'reference_test.go', 'paths_test.go', 'cache_test.go']) {
  const src = 'internal/researchpairedmemo/' + name;
  const dest = 'internal/researchpairedvalue/' + name;
  assert(!fs.existsSync(dest));
  const b = fs.readFileSync(src);
  sources[src] = hash(b);
  const s = b.toString().replaceAll('researchpairedmemo', 'researchpairedvalue');
  fs.mkdirSync(path.dirname(dest), { recursive: true });
  fs.writeFileSync(dest, s, { flag: 'wx', mode: 0o600 });
  initialOutputs[dest] = hash(s);
}
fs.writeFileSync('research/paired-v57-module-generation.json', JSON.stringify({ sources, initialOutputs, predictionValueSeparate: true }, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify({ copies: 5 }));
