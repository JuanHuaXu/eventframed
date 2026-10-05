import fs from 'node:fs';
import crypto from 'node:crypto';

const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const path = 'docs/experiments/mmm-boolean-specialist-v90.json';
const raw = fs.readFileSync(path), a = JSON.parse(raw);
if (a.Protocol !== 'mmm-boolean-specialist-v90' || a.Records.length !== 3840) throw Error('artifact contract');
for (const [path, hash] of Object.entries(a.Hashes)) if (sha(fs.readFileSync(path)) !== hash) throw Error(`source mismatch ${path}`);
const cases = ['parity1', 'parity2', 'parity3', 'parity4', 'complement4', 'majority3', 'mux3', 'constant', 'null', 'dependent4'];
const seen = new Set(), pools = new Map();
for (const r of a.Records) {
  const key = [r.Phase, r.Case, r.Index, r.N].join(':');
  if (seen.has(key) || !['design', 'confirmation'].includes(r.Phase) || !cases.includes(r.Case) || !Number.isInteger(r.Index) || r.Index < 0 || r.Index >= 64 || ![16, 32, 64].includes(r.N)) throw Error('record key');
  seen.add(key);
  if (![r.TrainSHA256, r.PredictionSHA256].every(x => /^[0-9a-f]{64}$/.test(x)) || !Number.isInteger(r.Rule) || r.Rule < 0 || r.Rule >= 512) throw Error('record hash/rule');
  if (r.Metrics.length !== 2 || r.Oracle.length !== 2) throw Error('dimensions');
  for (const arm of r.Metrics) {
    if (arm.length !== 2) throw Error('views');
    arm.forEach((m, view) => {
      if (![m.Brier, m.Accuracy, r.Oracle[view]].every(x => Number.isFinite(x) && x >= 0 && x <= 1) || m.Brier + 1e-12 < r.Oracle[view]) throw Error('score bounds');
      if (r.Case === 'null' && Math.abs(m.Accuracy - .5) > 1e-12) throw Error('null accuracy');
    });
  }
  const poolKey = `${r.Case}:${r.Phase}`;
  if (!pools.has(poolKey)) pools.set(poolKey, new Set());
  pools.get(poolKey).add(r.Rule);
}
for (const name of cases.filter(x => !['constant', 'null'].includes(x))) {
  const design = pools.get(`${name}:design`), confirmation = pools.get(`${name}:confirmation`);
  if ([...design].some(x => confirmation.has(x))) throw Error('overlapping rule pools');
}
const mean = xs => xs.reduce((s, x) => s + x, 0) / xs.length;
const interval = xs => {
  const value = mean(xs), se = Math.sqrt(xs.reduce((s, x) => s + (x - value) ** 2, 0) / (xs.length - 1) / xs.length);
  return {mean: value, lower: value - 3.5 * se, upper: value + 3.5 * se};
};
const summaries = [];
for (const phase of ['design', 'confirmation']) for (const name of cases) for (const n of [16, 32, 64]) {
  const rs = a.Records.filter(r => r.Phase === phase && r.Case === name && r.N === n);
  if (rs.length !== 64) throw Error('cell count');
  const views = [0, 1].map(view => ({
    name: ['full', 'fixed_mask63'][view],
    oracle: mean(rs.map(r => r.Oracle[view])),
    subset: {brier: mean(rs.map(r => r.Metrics[0][view].Brier)), accuracy: mean(rs.map(r => r.Metrics[0][view].Accuracy))},
    boolean: {brier: mean(rs.map(r => r.Metrics[1][view].Brier)), accuracy: mean(rs.map(r => r.Metrics[1][view].Accuracy))},
    gain: interval(rs.map(r => r.Metrics[0][view].Brier - r.Metrics[1][view].Brier)),
  }));
  const required = n >= 32 && ['parity3', 'parity4', 'complement4'].includes(name);
  summaries.push({phase, case: name, n, fits: 64, required, advancePass: required ? views[0].gain.mean >= .01 && views[0].gain.lower > 0 : null, views});
}
const output = JSON.stringify({artifactSHA256: sha(raw), sourceCount: Object.keys(a.Hashes).length, evaluatorSHA256: sha(fs.readFileSync('research/boolean-specialist-v90-summary.mjs')), runtime: a.Runtime, records: a.Records.length, underlyingTrainingSets: 1280, advanceToIntegration: summaries.filter(s => s.required).every(s => s.advancePass), classification: 'finite component evidence only; not delayed-stream rescue or adoption', summaries}, null, 2) + '\n';
if (process.argv[2]) fs.writeFileSync(process.argv[2], output, {flag: 'wx', mode: 0o600});
else process.stdout.write(output);
