import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';

// Research-only screen: masks are enumerated without teacher knowledge.
const degree = m => { let n = 0; for (; m; m &= m - 1) n++; return n; };
const masks = Array.from({length: 511}, (_, i) => i + 1)
  .filter(m => degree(m) >= 2 && degree(m) <= 4);
const phi = (x, m) => ((degree(m) - degree(x & m)) % 2 ? -1 : 1);
assert.equal(masks.length, 246);
for (let a = 1; a < 16; a++) for (let b = 1; b < 16; b++) {
  let dot = 0;
  for (let x = 0; x < 512; x++) dot += phi(x, a) * phi(x, b);
  assert.equal(dot, a === b ? 512 : 0);
}

function screen(samples) {
  const threshold = Math.sqrt(2 * Math.log(2 * masks.length / .1) / samples.length);
  return masks.map(mask => ({mask, correlation: samples.reduce((s, r) =>
    s + r.y * phi(r.x, mask), 0) / samples.length}))
    .filter(r => Math.abs(r.correlation) >= threshold)
    .sort((a, b) => Math.abs(b.correlation) - Math.abs(a.correlation) || a.mask - b.mask)
    .slice(0, 16);
}

// Separate deterministic random tapes for each replicate; no Math.random or
// shared stream whose consumption depends on learner decisions.
function uniform(key, index) {
  return createHash('sha256').update(`spectral-screen-v1:${key}:${index}`)
    .digest().readUInt32LE(0) / 4294967296;
}
const results = [];
for (const n of [32, 64]) for (const d of [0, 2, 3, 4]) {
  for (const noise of d === 0 ? [.5] : [0, .1, .3, .45]) {
    let recovered = 0, anySelected = 0, spuriousSelected = 0, selectedTotal = 0;
    for (let replicate = 0; replicate < 512; replicate++) {
      const key = `${n}:${d}:${noise}:${replicate}`;
      const order = Array.from({length: 9}, (_, i) => i);
      for (let i = 8; i > 0; i--) {
        const j = Math.floor(uniform(key, 10000 + i) * (i + 1));
        [order[i], order[j]] = [order[j], order[i]];
      }
      const target = order.slice(0, d).reduce((s, i) => s | (1 << i), 0);
      const samples = Array.from({length: n}, (_, i) => {
        const x = Math.floor(uniform(key, 2 * i) * 512);
        return {x, y: phi(x, target) * (uniform(key, 2 * i + 1) < noise ? -1 : 1)};
      });
      const before = JSON.stringify(samples);
      const selected = screen(samples);
      assert.equal(JSON.stringify(samples), before);
      assert(selected.length <= 16);
      recovered += Number(d > 0 && selected.some(r => r.mask === target));
      anySelected += Number(selected.length > 0);
      spuriousSelected += Number(selected.some(r => r.mask !== target));
      selectedTotal += selected.length;
    }
    results.push({n, degree: d, noise, replicates: 512, recovered,
      anySelected, spuriousSelected, selectedTotal});
  }
}
console.log(JSON.stringify({version: 1, candidates: masks.length, cap: 16,
  alpha: .1, seedNamespace: 'spectral-screen-v1', results}, null, 2));
