// Retrospective final-suffix audit and fixed-stream counterfactual. Uses only
// actually observed W1/W2 to reconstruct beliefs; hidden rates are score-only.
// This is not a prospective policy run, new gate, or proof of practical rescue.
import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const window = Number(process.argv[2]); assert([1200, 2400].includes(window));
const root = 'research/tree-v60-diagnostic', etas = [0, .1, .2], etaPrior = [.8, .1, .1];
const near = (a, b) => assert(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) < 2e-10, `${a} != ${b}`);
function lse(xs) {const max = Math.max(...xs); if (max === -Infinity) return max; return max + Math.log(xs.reduce((s, x) => s + Math.exp(x - max), 0));}
const logistic = x => 1 / (1 + Math.exp(-x));
function prior(b) {
  let urn = [1];
  for (let t = 0; t < 20; t++) {
    const next = Array(t + 2).fill(0);
    for (let z = 0; z <= t; z++) {next[z + 1] += urn[z] * (b + z) / (1 + t); next[z] += urn[z] * (1 - b + t - z) / (1 + t);}
    urn = next;
  }
  const localRates = [b, ...urn.map((_, z) => z / 20)], localPrior = [.8, ...urn.map(p => .2 * p)];
  let lo = -40, hi = 40;
  const ds = Array.from({length: 10}, (_, i) => -.6 * (i + 1)).concat(Array.from({length: 10}, (_, i) => .6 * (i + 1)));
  for (let k = 0; k < 100; k++) {const a = (lo + hi) / 2; if (ds.reduce((s, d) => s + logistic(a + d) / 20, 0) < b) lo = a; else hi = a;}
  const rates = [b, ...ds.map(d => logistic((lo + hi) / 2 + d))];
  near(localPrior.reduce((s, p) => s + p, 0), 1);
  near(localRates.reduce((s, p, z) => s + p * localPrior[z], 0), b);
  return {rates, localRates, localPrior};
}
function factor(p, eta, k) {
  let sum = 0;
  for (let y = 0; y < 2; y++) {let mass = y ? p : 1 - p; const a = k < 2 ? k : Math.floor((k - 2) / 2);
    mass *= a === y ? 1 - eta : eta; if (k >= 2) mass *= (k - 2) % 2 === y ? 1 - eta : eta; sum += mass;}
  return sum;
}
function summary(prior, contributions, rates) {
  const terms = prior.map((p, z) => Math.log(p) + contributions[z]), evidence = lse(terms);
  const weights = evidence === -Infinity ? prior.slice() : terms.map(x => Math.exp(x - evidence));
  return {evidence, weights, mean: rates.reduce((s, p, z) => s + p * weights[z], 0)};
}
function quality(q, rates) {
  let brier = 0; for (let i = 0; i < 150; i++) brier += (rates[i] * (1 - rates[i]) + (q[i] - rates[i]) ** 2) / 150;
  const order = q.map((p, i) => [p, i]).sort((a, b) => b[0] - a[0] || a[1] - b[1]);
  return {brier, utility: order.slice(0, 10).reduce((s, [, i]) => s + rates[i] / 10, 0)};
}
const cells = []; let first = true, forecastChecks = 0;
for await (const line of readline.createInterface({input: fs.createReadStream(root + '/diagnostic.jsonl'), crlfDelay: Infinity})) {
  const x = JSON.parse(line); if (first) {first = false; continue;}
  const w = x.Population.World, priors = w.Base.map(prior), sorted = w.Base.slice().sort((a, b) => a - b), groups = Array.from({length: 255}, () => []);
  const ranks = w.Base.map(b => Math.floor(sorted.indexOf(b) * 128 / 150));
  for (let i = 0; i < 150; i++) for (let n = 127 + ranks[i]; ; n = Math.floor((n - 1) / 2)) {groups[n].push(i); if (!n) break;}
  for (const a of x.Arms) {
    if (a.Mode === 'full' || a.Mode === 'adaptive') continue;
    // A larger final issued suffix is being probed. All first measurements have
    // arrived. Requested available W2 is valid if its original slot is retained
    // under the counterfactual window, even if the OLD 600-position model expired it.
    const paired = new Map((a.AuditOutcomes ?? []).filter(o => o.Available).map(o => [o.Trial, o.Value]));
    const counts = Array.from({length: 150}, () => Array(6).fill(0));
    for (let k = 2400 - window; k < 2400; k++) {const i = k % 150, one = +w.Outcomes[Math.floor(k / 150)][i]; counts[i][paired.has(k) ? 2 + 2 * one + +paired.get(k) : one]++;}
    const local = [], nodes = [], noiseLogs = [], independentLogs = [];
    for (let h = 0; h < 3; h++) {
      const eta = etas[h], memberContributions = priors.map((p, i) => p.rates.map(rate => counts[i].reduce((s, c, k) => c ? s + c * Math.log(factor(rate, eta, k)) : s, 0)));
      local[h] = priors.map((p, i) => summary(p.localPrior, p.localRates.map(rate => counts[i].reduce((s, c, k) => c ? s + c * Math.log(factor(rate, eta, k)) : s, 0)), p.localRates));
      nodes[h] = [];
      for (let n = 254; n >= 0; n--) {
        const contributions = Array.from({length: 21}, (_, z) => groups[n].reduce((s, i) => s + memberContributions[i][z], 0));
        const pool = summary([.8, ...Array(20).fill(.01)], contributions, priors[0].rates);
        const separate = n >= 127 ? groups[n].reduce((s, i) => s + local[h][i].evidence, 0) : nodes[h][2 * n + 1].evidence + nodes[h][2 * n + 2].evidence;
        const evidence = lse([pool.evidence - Math.log(2), separate - Math.log(2)]);
        const stop = evidence === -Infinity ? 1 : Math.exp(pool.evidence - Math.log(2) - evidence);
        nodes[h][n] = {...pool, evidence, stop};
      }
      noiseLogs[h] = Math.log(etaPrior[h]) + nodes[h][0].evidence;
      independentLogs[h] = Math.log(etaPrior[h]) + local[h].reduce((s, m) => s + m.evidence, 0);
    }
    const denom = lse(noiseLogs), indDenom = lse(independentLogs), mixedDenom = lse(noiseLogs.map((_, h) => Math.log(etaPrior[h]) + lse([nodes[h][0].evidence, independentLogs[h] - Math.log(etaPrior[h])]) - Math.log(2)));
    const weights = noiseLogs.map(t => Math.exp(t - denom)), indWeights = independentLogs.map(t => Math.exp(t - indDenom));
    const reconstructed = [], independent = [], rootMixture = [];
    for (let i = 0; i < 150; i++) {
      let q = 0, ind = 0, mixed = 0;
      for (let h = 0; h < 3; h++) {
        let n = 127 + ranks[i], item = nodes[h][n];
        let mean = item.weights.reduce((s, p, z) => s + p * priors[i].rates[z], 0);
        let v = item.stop * mean + (1 - item.stop) * local[h][i].mean;
        while (n > 0) {n = Math.floor((n - 1) / 2); item = nodes[h][n]; mean = item.weights.reduce((s, p, z) => s + p * priors[i].rates[z], 0); v = item.stop * mean + (1 - item.stop) * v;}
        q += weights[h] * v; ind += indWeights[h] * local[h][i].mean;
        // A .5 direct-root independent alternative to the ENTIRE V60 tree.
        // This counterfactual is coherent but has not selected its own evidence.
        mixed += Math.exp(noiseLogs[h] - Math.log(2) - mixedDenom) * v + Math.exp(independentLogs[h] - Math.log(2) - mixedDenom) * local[h][i].mean;
      }
      assert(q >= 0 && q <= 1 && Number.isFinite(q)); forecastChecks++;
      reconstructed.push(q); independent.push(ind); rootMixture.push(mixed);
    }
    const actual = quality(reconstructed, w.Rates[15]), alternate = quality(independent, w.Rates[15]), mixture = quality(rootMixture, w.Rates[15]);
    const published600Brier = a.Snapshots.at(-1).Brier;
    cells.push({geometry: w.Geometry, regime: w.Regime, schedule: a.Schedule, mode: a.Mode,
      rootPoolPosterior: weights.reduce((s, p, h) => s + p * nodes[h][0].stop, 0),
      directRootIndependentPosterior: Math.exp(indDenom - Math.log(2) - mixedDenom),
      actual, published600Brier, retainedRiskGain: published600Brier - actual.brier, independentCounterfactual: alternate, rootMixtureCounterfactual: mixture,
      independentRiskGain: actual.brier - alternate.brier, rootMixtureRiskGain: actual.brier - mixture.brier});
  }
}
assert.equal(cells.length, 720); assert.equal(forecastChecks, 108000);
const mean = (xs, field) => xs.reduce((s, x) => s + x[field] / xs.length, 0);
const summaries = Object.fromEntries(['no_pair', 'random', 'uncertainty', 'information', 'falsification', 'predictive'].map(m => {const c = cells.filter(x => x.mode === m); return [m,
  {cells: c.length, retainedRiskGain: mean(c, 'retainedRiskGain'), retentionWins: c.filter(x => x.retainedRiskGain > 1e-12).length, retentionLosses: c.filter(x => x.retainedRiskGain < -1e-12).length, rootPoolPosteriorMean: mean(c, 'rootPoolPosterior'), rootPoolOver99: c.filter(x => x.rootPoolPosterior > .99).length,
    directRootIndependentPosteriorMean: mean(c, 'directRootIndependentPosterior'),
    independentRiskGain: mean(c, 'independentRiskGain'), independentWins: c.filter(x => x.independentRiskGain > 1e-12).length,
    rootMixtureRiskGain: mean(c, 'rootMixtureRiskGain'), rootMixtureWins: c.filter(x => x.rootMixtureRiskGain > 1e-12).length}];}));
const result = {forecastChecksAreProbabilityShapesNotPublished600Equality: true, oldExpiredRepliesIncludedOnlyIfWithinExpandedSuffix: true, stage: 'Retrospective suffix probe on fixed old nominations, not a prospective policy', retainedIssuedPositions: window,
  scope: 'NOT issued/prequential risk, recovery, runtime, a new policy, fresh confirmation, or an empirical rescue. Hidden rates used only for outcome scoring.',
  forecastChecks, summaries, cells, wholeGoals: Array(7).fill('OPEN'),
  sourceSHA256: crypto.createHash('sha256').update(fs.readFileSync('research/tree-v60-window-probe.mjs')).digest('hex')};
fs.writeFileSync(root + '/window-probe-' + window + '.json', JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({forecastChecks, summaries}, null, 2));
