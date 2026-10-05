// Mechanical scaffolding; semantic inference changes are separate patches.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const parent = 'research/checkpoint-2026-10-04-tree-v59/manifest.json';
async function hashFile(p) {const h = crypto.createHash('sha256'); for await (const b of fs.createReadStream(p)) h.update(b); return h.digest('hex');}
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
assert.equal(await hashFile(parent), '6bff139a1658a268b7286dde9b27eef0ae5e8f416de7982857d1323bfb57780c');
const checkpoint = JSON.parse(fs.readFileSync(parent));
for (const [p, h] of Object.entries(checkpoint.copies)) assert.equal(await hashFile(path.dirname(parent) + '/saved/' + p), h, p);
for (const [p, h] of Object.entries(checkpoint.trackedHashes)) assert.equal(await hashFile(p), h, p);
const sources = {}, outputs = {};
function clone(src, dest) {
  assert(!fs.existsSync(dest));
  const raw = fs.readFileSync(src); sources[src] = hash(raw);
  let s = raw.toString().replaceAll('researchanchorref', 'researchlocalref').replaceAll('researchanchor', 'researchlocal')
    .replaceAll('V59', 'V60').replaceAll('tree-v59', 'tree-v60');
  if (src.endsWith('-run.mjs')) {
    s = s.replaceAll('research/checkpoint-2026-10-04-tree-v58/manifest.json', parent)
      .replaceAll('e6dcbb9a2eed68b4e7fb4866ecaa6878306805913b6b21e2c6d7d4678cdcb0d7', '6bff139a1658a268b7286dde9b27eef0ae5e8f416de7982857d1323bfb57780c');
  }
  fs.mkdirSync(path.dirname(dest), {recursive: true});
  fs.writeFileSync(dest, s, {flag: 'wx', mode: 0o600}); outputs[dest] = hash(s);
}
for (const file of ['inference.go', 'lifecycle.go', 'model_test.go', 'reference_test.go', 'benchmark_test.go'])
  clone('internal/researchanchor/' + file, 'internal/researchlocal/' + file);
for (const file of ['inference.go', 'lifecycle.go']) clone('internal/researchanchorref/' + file, 'internal/researchlocalref/' + file);
clone('internal/researchdispersion/tree_v59_generated_test.go', 'internal/researchdispersion/tree_v60_generated_test.go');
for (const file of ['run', 'readback', 'comparison']) clone('research/tree-v59-' + file + '.mjs', 'research/tree-v60-' + file + '.mjs');
fs.writeFileSync('research/tree-v60-generation.json', JSON.stringify({parent, parentCopiesVerified: Object.keys(checkpoint.copies).length,
  sources, outputs, manualInferenceAndPatchesSeparate: true}, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({outputs: Object.keys(outputs).length, parentCopiesVerified: Object.keys(checkpoint.copies).length}));
