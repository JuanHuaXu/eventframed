// Isolated mathematical preflight. No production imports, corpus, hidden fixture
// rates, scientific confirmation, or claims of empirical gate completion.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';

const grid = Array.from({length: 21}, (_, i) => i / 20);
const noise = [0, .1, .2], noisePrior = [.8, .1, .1];
const offsets = Array.from({length: 10}, (_, i) => -.6 * (i + 1)).concat(Array.from({length: 10}, (_, i) => .6 * (i + 1)));
const sigmoid = x => 1 / (1 + Math.exp(-x));
const near = (a, b, tolerance = 2e-11) => assert(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) < tolerance, `${a} != ${b}`);
function logAdd(a, b) {
  if (a === -Infinity) return b;
  if (b === -Infinity) return a;
  const hi = Math.max(a, b), lo = Math.min(a, b);
  return hi + Math.log1p(Math.exp(lo - hi));
}
const logSum = xs => xs.reduce(logAdd, -Infinity);
function independent(b) {
  assert(b >= .25 && b <= .925);
  let lo = -80, hi = 80;
  for (let k = 0; k < 120; k++) {
    const t = (lo + hi) / 2, max = Math.max(...grid.map(g => t * g));
    const weights = grid.map(g => Math.exp(t * g - max)), sum = weights.reduce((a, b) => a + b, 0);
    const mean = grid.reduce((s, g, i) => s + g * weights[i] / sum, 0);
    if (mean < b) lo = t; else hi = t;
  }
  const t = (lo + hi) / 2, max = Math.max(...grid.map(g => t * g));
  const weights = grid.map(g => Math.exp(t * g - max)), sum = weights.reduce((a, b) => a + b, 0);
  const atoms = [{rate: b, prior: .8}, ...grid.map((rate, i) => ({rate, prior: .2 * weights[i] / sum}))];
  assert(atoms.every(x => x.prior > 0));
  near(atoms.reduce((s, x) => s + x.prior, 0), 1);
  near(atoms.reduce((s, x) => s + x.prior * x.rate, 0), b);
  return atoms;
}
function pooledRates(b) {
  let lo = -40, hi = 40;
  for (let k = 0; k < 100; k++) {
    const x = (lo + hi) / 2, mean = offsets.reduce((s, d) => s + sigmoid(x + d) / 20, 0);
    if (mean < b) lo = x; else hi = x;
  }
  return [b, ...offsets.map(d => sigmoid((lo + hi) / 2 + d))];
}
function factor(p, eta, e) {
  let mass = 0;
  for (let y = 0; y < 2; y++) {
    let value = y ? p : 1 - p;
    for (const observed of e) value *= observed === y ? 1 - eta : eta;
    mass += value;
  }
  return mass;
}
const logLikelihood = (p, eta, events) => events.reduce((s, e) => s + Math.log(factor(p, eta, e)), 0);
function individualSummary(atoms, eta, events) {
  const terms = atoms.map(x => Math.log(x.prior) + logLikelihood(x.rate, eta, events));
  const evidence = logSum(terms);
  return {evidence, mean: atoms.reduce((s, x, i) => s + x.rate * Math.exp(terms[i] - evidence), 0)};
}
function recursive(base, evidence) {
  const rates = base.map(pooledRates), atoms = base.map(independent), branches = [];
  for (let h = 0; h < 3; h++) {
    const eta = noise[h];
    const poolTerms = rates[0].map((_, z) => Math.log(z ? .01 : .8) + base.reduce((s, b, i) => s + logLikelihood(rates[i][z], eta, evidence[i]), 0));
    const poolLog = logSum(poolTerms);
    const poolMean = base.map((_, i) => poolTerms.reduce((s, term, z) => s + Math.exp(term - poolLog) * rates[i][z], 0));
    const individual = atoms.map((a, i) => individualSummary(a, eta, evidence[i]));
    const separateLog = individual.reduce((s, x) => s + x.evidence, 0);
    const combined = logAdd(poolLog, separateLog) - Math.log(2);
    const pooledWeight = Math.exp(poolLog - Math.log(2) - combined);
    branches.push({log: Math.log(noisePrior[h]) + combined, pooledWeight,
      mean: base.map((_, i) => pooledWeight * poolMean[i] + (1 - pooledWeight) * individual[i].mean)});
  }
  const normalizer = logSum(branches.map(x => x.log)), weights = branches.map(x => Math.exp(x.log - normalizer));
  return {logEvidence: normalizer, mean: base.map((_, i) => branches.reduce((s, x, h) => s + weights[h] * x.mean[i], 0)),
    weights, pooledWeight: branches.reduce((s, x, h) => s + weights[h] * x.pooledWeight, 0)};
}
function exhaustive(base, events) {
  // Enumerate the entire latent family (shared eta, pool/separate, parameter
  // indices). This does not multiply separately noise-marginalized members.
  const atoms = base.map(independent), rates = base.map(pooledRates), terms = [];
  for (let h = 0; h < 3; h++) {
    const eta = noise[h], common = Math.log(noisePrior[h]) - Math.log(2);
    for (let z = 0; z < 21; z++) terms.push({log: common + Math.log(z ? .01 : .8) +
      logLikelihood(rates[0][z], eta, events[0]) + logLikelihood(rates[1][z], eta, events[1]),
      p: [rates[0][z], rates[1][z]], h, pooled: true});
    for (const a of atoms[0]) for (const b of atoms[1]) terms.push({log: common + Math.log(a.prior) + Math.log(b.prior) +
      logLikelihood(a.rate, eta, events[0]) + logLikelihood(b.rate, eta, events[1]), p: [a.rate, b.rate], h, pooled: false});
  }
  const norm = logSum(terms.map(x => x.log)), mean = [0, 0], weights = [0, 0, 0];
  let pooledWeight = 0;
  for (const x of terms) {
    const w = Math.exp(x.log - norm);
    weights[x.h] += w;
    pooledWeight += w * +x.pooled;
    for (let i = 0; i < 2; i++) mean[i] += w * x.p[i];
  }
  return {logEvidence: norm, mean, weights, pooledWeight, enumeratedStates: terms.length};
}

const baselines = Array.from({length: 28}, (_, i) => .25 + i * .025);
baselines[27] = .925;
let pairedMassChecks = 0, quantizationChecks = 0;
for (const b of baselines) {
  const atoms = independent(b);
  assert(Math.min(...atoms.map(x => x.rate)) === 0 && Math.max(...atoms.map(x => x.rate)) === 1);
  for (const eta of noise) for (const x of atoms) {
    const pair = [[0, 0], [0, 1], [1, 0], [1, 1]].map(e => factor(x.rate, eta, e));
    near(pair.reduce((a, b) => a + b, 0), 1);
    near(pair[0] + pair[1], factor(x.rate, eta, [0]));
    near(pair[2] + pair[3], factor(x.rate, eta, [1]));
    pairedMassChecks++;
  }
}
for (let k = 0; k <= 10000; k++) {
  const p = k / 10000, distance = Math.min(...grid.map(g => Math.abs(g - p)));
  assert(distance * distance <= .000625 + 1e-15);
  quantizationChecks++;
}
const histories = [
  [[], []], [[[0]], [[1]]], [[[0, 1]], [[1, 1]]],
  [Array.from({length: 4}, () => [0, 0]), Array.from({length: 4}, () => [1, 1])],
  [Array.from({length: 20}, () => [0, 0]), Array.from({length: 20}, () => [1, 1])],
  [[[0], [1, 1], [0, 1]], [[1], [0, 0]]]
];
const cases = [];
for (const base of [[.925, .925], [.25, .925], [.47, .78]]) for (const events of histories) {
  const r = recursive(base, events), x = exhaustive(base, events);
  near(r.logEvidence, x.logEvidence);
  near(r.pooledWeight, x.pooledWeight);
  for (let i = 0; i < 2; i++) near(r.mean[i], x.mean[i]);
  for (let h = 0; h < 3; h++) near(r.weights[h], x.weights[h]);
  if (events.every(a => a.length === 0)) for (let i = 0; i < 2; i++) near(r.mean[i], base[i]);
  cases.push({base, observations: events.map(e => e.length), ...r, enumeratedStates: x.enumeratedStates});
}
const divergent = cases.find(c => c.base[0] === .925 && c.base[1] === .925 && c.observations[0] === 20);
assert(divergent.mean[0] < .25 && divergent.mean[1] > .75 && divergent.pooledWeight < .01);
// Wrong common-noise marginalization is a nonvacuous adjacent-path control.
const a = independent(.25), b = independent(.925), ea = [[0, 1], [1, 1]], eb = [[0, 0], [0, 0]];
const correct = noise.reduce((s, eta, h) => s + noisePrior[h] * Math.exp(individualSummary(a, eta, ea).evidence + individualSummary(b, eta, eb).evidence), 0);
const wrong = noise.reduce((s, eta, h) => s + noisePrior[h] * Math.exp(individualSummary(a, eta, ea).evidence), 0) *
  noise.reduce((s, eta, h) => s + noisePrior[h] * Math.exp(individualSummary(b, eta, eb).evidence), 0);
assert(Math.abs(correct - wrong) > 1e-6);
// Removing a discordant factor revives zero-noise support; rebuilding from the
// remaining ledger avoids subtraction of negative infinity in this preflight.
const withDiscordance = recursive([.47, .78], [[[0, 1]], [[1, 1]]]);
const afterRemoval = recursive([.47, .78], [[], [[1, 1]]]);
assert(withDiscordance.weights[0] === 0 && afterRemoval.weights[0] > 0);

const result = {stage: 'Mathematical preflight only, no Go implementation or empirical rescue',
  baselineIdentities: baselines.length, pairedMassChecks, quantizationChecks, jointComparisons: cases.length,
  statesPerJointComparison: 1515, cases, divergent, commonNoiseNegativeControl: {correct, incorrectlySeparated: wrong},
  zeroNoiseRevival: {before: withDiscordance.weights[0], after: afterRemoval.weights[0]},
  oracleSquaredProbabilityApproximationBound: .000625, noTrajectoryRegretOrCoverageClaim: true,
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync('research/tree-v60-identities.mjs')).digest('hex'),
  wholeGoals: Array(7).fill('OPEN')};
fs.writeFileSync('research/tree-v60-identities.json', JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({baselineIdentities: result.baselineIdentities, pairedMassChecks, quantizationChecks,
  jointComparisons: cases.length, divergent, commonNoiseDifference: correct - wrong, zeroNoiseRevival: result.zeroNoiseRevival}, null, 2));
