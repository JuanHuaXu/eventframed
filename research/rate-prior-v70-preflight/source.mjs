// Algebraic prior sensitivity only: no stream labels, serving or adoption.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const source = 'research/rate-prior-v70-preflight.mjs';
const root = 'research/rate-prior-v70-preflight';
assert(!fs.existsSync(root));
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const raw = fs.readFileSync(source);
fs.mkdirSync(root, {mode: 0o700});
fs.writeFileSync(root + '/source.mjs', raw, {flag: 'wx', mode: 0o600});
const save = (p, x) => fs.writeFileSync(root + '/' + p, JSON.stringify(x, null, 2) + '\n', {flag: 'wx', mode: 0o600});
save('freeze.json', {time: new Date().toISOString(), source, sourceSHA256: hash(raw), scope: 'finite prior moment and same-Y measurement algebra, not a quality or performance experiment', empiricalLabelsOpened: false});
let checks = 0, maxDefect = 0;
function near(a, b) {
  const d = Math.abs(a - b);
  maxDefect = Math.max(maxDefect, d);
  checks++;
  assert(d < 2e-12, `${a} != ${b}`);
}
const sum = xs => xs.reduce((a, b) => a + b, 0);
const normalize = xs => {
  const s = sum(xs);
  assert(s > 0 && Number.isFinite(s));
  return xs.map(x => x / s);
};
const N = 20;
const configs = [
  {name: 'current', spike: .8, strength: 1},
  {name: 'no_spike_strength1', spike: 0, strength: 1},
  {name: 'no_spike_strength2', spike: 0, strength: 2},
];

// Candidate beta-binomial ratios; the reference instead enumerates the urn DP.
function prior(b, cfg) {
  const a = cfg.strength * b, c = cfg.strength * (1 - b);
  const w = [1];
  for (let j = 0; j < N; j++) w[0] *= (c + j) / (a + c + j);
  for (let j = 0; j < N; j++) w.push(w[j] * (N - j) / (j + 1) * (a + j) / (c + N - j - 1));
  return [{p: b, mass: cfg.spike}, ...normalize(w).map((mass, j) => ({p: j / N, mass: (1 - cfg.spike) * mass}))];
}
function urn(b, cfg) {
  const a = cfg.strength * b, c = cfg.strength * (1 - b);
  let w = [1];
  for (let n = 0; n < N; n++) {
    const next = Array(n + 2).fill(0);
    for (let j = 0; j <= n; j++) {
      next[j] += w[j] * (c + n - j) / (a + c + n);
      next[j + 1] += w[j] * (a + j) / (a + c + n);
    }
    w = next;
  }
  return [{p: b, mass: cfg.spike}, ...w.map((mass, j) => ({p: j / N, mass: (1 - cfg.spike) * mass}))];
}
const likelihood = (p, eta, bit) => bit ? eta + (1 - 2 * eta) * p : 1 - eta - (1 - 2 * eta) * p;
function posteriorMean(rows, eta, bit) {
  const masses = rows.map(x => x.mass * likelihood(x.p, eta, bit));
  return sum(rows.map((x, j) => x.p * masses[j])) / sum(masses);
}

const results = [];
let unsupportedPairBranches = 0, repeatedEvidenceNegativeControls = 0;
for (const b of [.25, .3, .4, .5, .7, .9, .925]) {
  for (const cfg of configs) {
    const rows = prior(b, cfg), ref = urn(b, cfg);
    for (let j = 0; j < rows.length; j++) near(rows[j].mass, ref[j].mass);
    near(sum(rows.map(x => x.mass)), 1);
    near(sum(rows.map(x => x.mass * x.p)), b);
    const variance = sum(rows.map(x => x.mass * (x.p - b) ** 2));
    const analyticVariance = (1 - cfg.spike) * b * (1 - b) * (N + cfg.strength) / (N * (cfg.strength + 1));
    near(variance, analyticVariance);
    for (const eta of [0, .1, .2]) {
      const q = eta + (1 - 2 * eta) * b;
      const yes = posteriorMean(rows, eta, true), no = posteriorMean(rows, eta, false);
      near(yes - b, (1 - 2 * eta) * variance / q);
      near(no - b, -(1 - 2 * eta) * variance / (1 - q));
      near(q * yes + (1 - q) * no, b);
      const stationaryExcess = q * (yes - b) ** 2 + (1 - q) * (no - b) ** 2;
      near(stationaryExcess, (1 - 2 * eta) ** 2 * variance ** 2 / (q * (1 - q)));
      for (const hazard of [0, 1 / 16, 1]) {
        const predYes = (1 - hazard) * yes + hazard * b;
        const predNo = (1 - hazard) * no + hazard * b;
        near(q * predYes + (1 - q) * predNo, b);
        near(q * (predYes - b) ** 2 + (1 - q) * (predNo - b) ** 2, (1 - hazard) ** 2 * stationaryExcess);
      }

      // Enumerate Y and BOTH noisy measurements; they share the original Y.
      for (const first of [false, true]) {
        let total = 0, mean = 0, averaged = 0;
        const branches = [];
        for (const second of [false, true]) {
          let mass = 0, moment = 0, analyticMass = 0, analyticMoment = 0;
          for (let j = 0; j < rows.length; j++) {
            const x = rows[j];
            let joint = 0;
            for (const y of [false, true]) {
              const py = y ? x.p : 1 - x.p;
              const pa = first === y ? 1 - eta : eta;
              const pb = second === y ? 1 - eta : eta;
              const m = ref[j].mass * py * pa * pb;
              mass += m;
              moment += m * x.p;
              joint += py * pa * pb;
            }
            analyticMass += x.mass * joint;
            analyticMoment += x.mass * joint * x.p;
          }
          near(mass, analyticMass);
          near(moment, analyticMoment);
          branches.push({mass, moment});
          total += mass;
          mean += moment;
          if (mass === 0) unsupportedPairBranches++;
        }
        near(total, likelihood(b, eta, first));
        near(mean / total, posteriorMean(rows, eta, first));
        for (const z of branches) if (z.mass > 0) averaged += z.mass / total * z.moment / z.mass;
        near(averaged, posteriorMean(rows, eta, first));
        if (eta === 0) {
          // Incorrect independent-Y multiplication must be detectably different.
          const masses = rows.map(x => x.mass * likelihood(x.p, eta, first) ** 2);
          const wrong = sum(rows.map((x, j) => x.p * masses[j])) / sum(masses);
          assert(Math.abs(wrong - posteriorMean(rows, eta, first)) > .001);
          repeatedEvidenceNegativeControls++;
        }
      }
      results.push({baseline: b, config: cfg.name, eta, variance, yes, no, stationaryExcess});
    }
  }
}
const ratios = [];
for (const old of results.filter(x => x.config === 'current')) {
  for (const cfg of configs.slice(1)) {
    const next = results.find(x => x.baseline === old.baseline && x.eta === old.eta && x.config === cfg.name);
    const responseRatio = (next.yes - old.baseline) / (old.yes - old.baseline);
    const harmRatio = next.stationaryExcess / old.stationaryExcess;
    near(harmRatio, responseRatio ** 2);
    assert(responseRatio > 1 && harmRatio > 1);
    ratios.push({baseline: old.baseline, eta: old.eta, config: cfg.name, responseRatio, stationaryHarmRatio: harmRatio});
  }
}
assert.equal(hash(fs.readFileSync(source)), hash(raw));
save('results.json', {checks, maxDefect, configurations: configs, rows: results, ratios, unsupportedPairBranches, repeatedEvidenceNegativeControls, conclusion: 'Broader moment-preserving priors react faster but increase one-observation stationary excess risk; not an unqualified rescue.', qualityRescueEstablished: false, uniqueCauseOfStreamFailureEstablished: false, equalTotalCostEstablished: false, goal: 'ACTIVE', goals: Array(7).fill('OPEN')});
save('completed.json', {exitCode: 0, sourceUnchanged: true, allJobsTerminal: true, resultsSHA256: hash(fs.readFileSync(root + '/results.json'))});
console.log(JSON.stringify({checks, maxDefect, unsupportedPairBranches, repeatedEvidenceNegativeControls, examples: ratios.filter(x => x.baseline === .5 && x.eta === .1), componentOnly: true}, null, 2));
