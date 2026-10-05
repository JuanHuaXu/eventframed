import assert from 'node:assert/strict';
import {regimeReference} from './regime-predictive-reference.mjs';

// Direct probability-space enumeration, independent of Go's interval table,
// transforms and log-sum recursion. The sample cap keeps these products finite.
export function oneChangeReference(samples, queries) {
  const n = samples.length;
  const full = regimeReference(samples, queries);
  const cuts = [0], leftLog = [0], rightLog = [full.logEvidence];
  const priors = [n >= 16 ? .9 : 1], tails = [full.predictions];
  for (let cut = 8; cut <= n - 8; cut++) {
    const left = regimeReference(samples.slice(0, cut), []);
    const right = regimeReference(samples.slice(cut), queries);
    cuts.push(cut); leftLog.push(left.logEvidence); rightLog.push(right.logEvidence);
    priors.push(.1 / (n - 15)); tails.push(right.predictions);
  }
  const unnormalized = priors.map((p, i) => p * Math.exp(leftLog[i]) * Math.exp(rightLog[i]));
  const z = unnormalized.reduce((a, b) => a + b, 0);
  assert(z > 0 && Number.isFinite(z));
  const weights = unnormalized.map(w => w / z);
  const predictions = queries.map((_, q) => weights.reduce((s, w, i) => s + w * tails[i][q], 0));
  return {cuts, leftLog, rightLog, weights, logEvidence: Math.log(z), predictions, full: full.predictions};
}

export function auditOneChangeSnapshot(r, source) {
  assert(!r.Error && r.Fit);
  const clock = r.Clock;
  assert(Number.isInteger(clock) && clock >= 0 && clock <= 224);
  assert.equal(source.Steps.length, 256);
  const origins = [];
  for (let j = -16; j < clock; j++) {
    if (j < 0 || (!source.Steps[j].Missing && j + source.Steps[j].Delay <= clock)) origins.push(j);
  }
  const retained = origins.slice(-64);
  assert.deepEqual(r.Origins, retained);
  const samples = retained.map(j => j < 0 ? [source.Initial[j + 16].Bits, source.Initial[j + 16].Outcome] : [source.Steps[j].X, source.Steps[j].Y]);
  const queries = source.Steps.slice(clock, clock + 32).map(s => s.X);
  const ref = oneChangeReference(samples, queries), fit = r.Fit;
  assert.deepEqual(fit.Cuts, ref.cuts);
  let maxError = 0;
  const check = (a, b) => { assert(Number.isFinite(a) && Number.isFinite(b)); const e = Math.abs(a - b); maxError = Math.max(maxError, e); assert(e < 1e-10); };
  for (const [actual, expected] of [[fit.Weights, ref.weights], [fit.LeftLog, ref.leftLog], [fit.RightLog, ref.rightLog]]) {
    assert.equal(actual.length, expected.length);
    actual.forEach((x, i) => check(x, expected[i]));
  }
  check(fit.LogEvidence, ref.logEvidence);
  check(fit.Weights.reduce((a, b) => a + b, 0), 1);
  assert.equal(fit.Predictions.length, 512);
  assert.equal(r.Predictions.length, 32);
  r.Predictions.forEach((p, i) => { assert(p > 0 && p < 1); check(p, ref.predictions[i]); check(p, fit.Predictions[queries[i]]); });
  return {ref, maxError};
}
