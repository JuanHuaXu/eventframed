import assert from 'node:assert/strict';
import {oneChangeReference, auditOneChangeSnapshot} from './one-change-reference.mjs';

for (const n of [0, 1, 15, 16, 18, 64]) {
  const samples = Array.from({length: n}, (_, i) => [i * 17 % 512, i % 3 === 0]);
  const r = oneChangeReference(samples, [0, 17, 511]);
  assert.equal(r.cuts.length, 1 + Math.max(0, n - 15));
  assert(Math.abs(r.weights.reduce((a, b) => a + b, 0) - 1) < 1e-12);
  assert(r.predictions.every(p => p > 0 && p < 1));
  if (!n) assert.deepEqual(r.predictions, [.5, .5, .5]);
  if (n < 16) assert.deepEqual(r.predictions, r.full);
}
const source = {
  Initial: Array.from({length: 16}, (_, i) => ({Bits: i, Outcome: i % 3 === 0})),
  Steps: Array.from({length: 256}, () => ({X: 0, Y: false, Q: .2, Delay: 0, Missing: false}))
};
const ref = oneChangeReference(source.Initial.map(s => [s.Bits, s.Outcome]), [0]);
const record = {Clock: 0, Origins: Array.from({length: 16}, (_, i) => i - 16), Fit: {
  Cuts: ref.cuts, Weights: ref.weights, LeftLog: ref.leftLog, RightLog: ref.rightLog,
  LogEvidence: ref.logEvidence, Predictions: Array(512).fill(ref.predictions[0])
}, Predictions: Array(32).fill(ref.predictions[0])};
auditOneChangeSnapshot(record, source);
for (const mutate of [
  r => r.Origins[0] = 0,
  r => r.Fit.Cuts[1] = 9,
  r => r.Fit.Weights[0] += .01,
  r => r.Fit.LeftLog[1] += .01,
  r => r.Predictions[0] += .01
]) {
  const bad = structuredClone(record); mutate(bad);
  assert.throws(() => auditOneChangeSnapshot(bad, source));
}
const future = structuredClone(source);
future.Steps.forEach(s => { s.Y = true; s.Q = .9; });
assert.deepEqual(auditOneChangeSnapshot(record, source), auditOneChangeSnapshot(record, future));
console.log('PASS: six reference supports, five rejected record mutations, current/future Y and Q independence');
