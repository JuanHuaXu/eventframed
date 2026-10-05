import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';

// Semantic check only. This does not construct or validate confidence sets.
function certifiedSuffix(candidates, evidence, clock) {
  if (!candidates || candidates.length === 0) return null;
  assert.ok(candidates.every(t => Number.isInteger(t) && t >= 0 && t <= clock));
  const upper = Math.max(...candidates);
  return evidence.filter(e => e.origin >= upper && e.origin <= clock &&
    e.arrival <= clock && e.arrival >= e.origin && e.audit && !e.missing);
}

let checked = 0;
for (let lower = 0; lower < 32; lower++) for (let upper = lower; upper < 32; upper++) {
  for (let truth = lower; truth <= upper; truth++) {
    const candidates = [...new Set([lower, truth, upper])];
    const evidence = Array.from({ length: 40 }, (_, origin) => ({
      origin, arrival: origin + origin % 4, audit: origin % 3 !== 0, missing: origin % 7 === 0,
    }));
    for (const clock of [upper, 39]) {
      const kept = certifiedSuffix(candidates, evidence, clock);
      assert.ok(kept.every(e => e.origin >= truth && e.arrival <= clock && e.audit && !e.missing));
      const reference = evidence.filter(e => e.origin >= upper && e.origin <= clock && e.arrival <= clock && e.audit && !e.missing);
      assert.deepEqual(kept, reference);
      checked++;
    }
  }
}
assert.equal(certifiedSuffix([], [], 9), null);
assert.equal(certifiedSuffix(null, [], 9), null);
assert.throws(() => certifiedSuffix([10], [], 9));
const revealing = { origin: 4, arrival: 4, audit: true, missing: false };
assert.deepEqual(certifiedSuffix([4], [revealing], 4), [revealing]);
// A lower-bound reset includes origin4 although true onset may be8.
assert.ok(revealing.origin >= Math.min(3, 8) && revealing.origin < 8);
assert.deepEqual(certifiedSuffix([3, 8], [revealing], 8), []);
const result = { checked, scriptSHA256: crypto.createHash('sha256').update(fs.readFileSync(new URL(import.meta.url))).digest('hex'),
  limits: 'Deterministic coverage-event implication only. No confidence-set estimator, statistical coverage, localization accuracy, forecast rescue or production policy is validated.' };
fs.writeFileSync(process.argv[2], JSON.stringify(result, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify(result));
