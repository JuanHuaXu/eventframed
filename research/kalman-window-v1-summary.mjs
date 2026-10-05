import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) {
  throw new Error('usage: node research/kalman-window-v1-summary.mjs SPLIT.jsonl');
}

const rows = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const [manifest, ...records] = rows;
if (manifest.kind !== 'manifest') throw new Error('missing manifest');
for (const [path, expected] of Object.entries(manifest.hashes)) {
  const actual = createHash('sha256').update(readFileSync(path)).digest('hex');
  if (actual !== expected) throw new Error(`source changed: ${path}`);
}
const scenarios = ['stable05', 'shift128', 'gradual', 'delayed_missing', 'interaction'];
if (records.length !== 160) throw new Error(`wrong record count: ${records.length}`);
const pairs = new Set();
function metric(ticks, arm, start, end) {
  let brier = 0;
  let accuracy = 0;
  for (let i = start; i < end; i++) {
    const p = ticks[i].Predictions[arm];
    if (!Number.isFinite(p) || p < 0 || p > 1) throw new Error(`bad forecast at ${i}`);
    const y = Number(ticks[i].Outcome);
    brier += (p - y) ** 2;
    accuracy += Number((p >= .5) === ticks[i].Outcome);
  }
  return { Brier: brier / (end - start), Accuracy: accuracy / (end - start), N: end - start };
}
function stats(values) {
  const mean = values.reduce((a, b) => a + b, 0) / values.length;
  const variance = values.reduce((a, b) => a + (b - mean) ** 2, 0) / (values.length - 1);
  const margin = 3.5 * Math.sqrt(variance / values.length);
  return { mean, lower: mean - margin, upper: mean + margin };
}
for (const r of records) {
  if (r.Split !== manifest.split || !scenarios.includes(r.Scenario) ||
      r.Fit < 0 || r.Fit >= 4 || r.Stream < 0 || r.Stream >= 8 || r.Ticks.length !== 512) {
    throw new Error('invalid record coordinates');
  }
  const key = `${r.Scenario}:${r.Fit}:${r.Stream}`;
  if (pairs.has(key)) throw new Error(`duplicate record ${key}`);
  pairs.add(key);
  let delivered = 0;
  let audited = 0;
  let nonMissing = 0;
  for (let t = 0; t < 512; t++) {
    const tick = r.Ticks[t];
    nonMissing += Number(!tick.Missing);
    for (const origin of tick.Delivered ?? []) {
      if (origin > t || origin < 0 || r.Ticks[origin].Missing) {
        throw new Error(`invalid delivery at ${key}:${t} from ${origin}`);
      }
      if (r.Scenario === 'delayed_missing' && t - origin !== 16) {
        throw new Error(`bad delay at ${key}:${t}`);
      }
      delivered++;
      audited += Number(r.Ticks[origin].Audit);
    }
  }
  if (r.Available !== delivered || r.Audits !== audited || r.Updates !== audited ||
      r.Pending !== nonMissing - delivered || r.StateBytes > 4096) {
    throw new Error(`accounting mismatch at ${key}`);
  }
  const change = r.Scenario === 'stable05' ? 0 : r.Scenario === 'shift128' || r.Scenario === 'gradual' ? 128 : 256;
  for (let a = 0; a < 4; a++) {
    for (const [name, start, end] of [['Full', 0, 512], ['Tail', 384, 512], ['Early', change, change + 64]]) {
      const got = metric(r.Ticks, a, start, end);
      const stored = r[name][a];
      if (got.N !== stored.N || Math.abs(got.Brier - stored.Brier) > 1e-12 ||
          Math.abs(got.Accuracy - stored.Accuracy) > 1e-12) {
        throw new Error(`metric mismatch ${key}:${name}:${a}`);
      }
    }
  }
}
const report = { split: manifest.split, records: records.length, sourceHashes: manifest.hashes, cases: {} };
for (const scenario of scenarios) {
  const subset = records.filter(r => r.Scenario === scenario);
  const windowGain = subset.map(r => Math.min(r.Tail[1].Brier, r.Tail[2].Brier) - r.Tail[3].Brier);
  const earlyGain = subset.map(r => Math.min(r.Early[1].Brier, r.Early[2].Brier) - r.Early[3].Brier);
  const fullHarmBase = subset.map(r => r.Full[3].Brier - r.Full[0].Brier);
  const fullHarmWindow = subset.map(r => r.Full[3].Brier - Math.min(r.Full[1].Brier, r.Full[2].Brier));
  report.cases[scenario] = {
    meanTailBrier: [0, 1, 2, 3].map(a => subset.reduce((z, r) => z + r.Tail[a].Brier, 0) / subset.length),
    meanEarlyBrier: [0, 1, 2, 3].map(a => subset.reduce((z, r) => z + r.Early[a].Brier, 0) / subset.length),
    meanFullBrier: [0, 1, 2, 3].map(a => subset.reduce((z, r) => z + r.Full[a].Brier, 0) / subset.length),
    tailGainVsOracleWindow: stats(windowGain),
    earlyGainVsOracleWindow: stats(earlyGain),
    fullHarmVsBase: stats(fullHarmBase),
    fullHarmVsOracleWindow: stats(fullHarmWindow),
    maxUpdateP99NS: Math.max(...subset.map(r => r.UpdateP99NS)),
    meanAudits: subset.reduce((z, r) => z + r.Audits, 0) / subset.length,
    meanAvailable: subset.reduce((z, r) => z + r.Available, 0) / subset.length,
    meanCuts: subset.reduce((z, r) => z + r.Cuts, 0) / subset.length,
  };
}
const c = report.cases;
report.gates = {
  shift128: c.shift128.tailGainVsOracleWindow.mean >= .005 && c.shift128.tailGainVsOracleWindow.lower > 0,
  gradual: c.gradual.tailGainVsOracleWindow.mean >= .005 && c.gradual.tailGainVsOracleWindow.lower > 0,
  stable: c.stable05.fullHarmVsBase.upper < .01 && c.stable05.fullHarmVsOracleWindow.upper < .01,
  delayed: -c.delayed_missing.tailGainVsOracleWindow.lower < .01,
  cost: Math.max(...records.map(r => r.UpdateP99NS)) < 50000 && records.every(r => r.StateBytes < 4096),
};
report.pass = Object.values(report.gates).every(Boolean);
console.log(JSON.stringify(report, null, 2));
