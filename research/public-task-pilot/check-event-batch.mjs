import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import assert from 'node:assert/strict';

const raw = JSON.parse(readFileSync('research/public-task-pilot/event-batch-screen.json', 'utf8'));
assert.equal(raw.status, 0);
assert.ok(!raw.signal && !raw.error);
for (const [path, digest] of Object.entries(raw.hashes)) {
  assert.equal(createHash('sha256').update(readFileSync(path)).digest('hex'), digest, path);
}
const rows = [...raw.stdout.matchAll(/^BenchmarkResearchEventBatch\/batch(\d+)-4\s+1\s+(\d+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/gm)];
assert.equal(rows.length, 15);
const medians = {};
for (const size of [1, 2, 4, 8, 16]) {
  const runs = rows.filter(r => Number(r[1]) === size).map(r => Number(r[2])).sort((a,b) => a-b);
  assert.equal(runs.length, 3);
  medians[size] = runs[1];
}
for (const [size, ns] of Object.entries(medians)) {
  console.log(JSON.stringify({size: Number(size), median_ms_per128: ns/1e6, amortized_ms_per_event: ns/128e6, speedup: medians[1]/ns, screen_gain_pass: size !== '1' && ns <= 0.8*medians[1]}));
}
console.log('Artifact hashes and all 15 measured operations verified. Timing screen only.');
