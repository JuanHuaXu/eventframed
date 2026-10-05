import fs from 'node:fs';
import readline from 'node:readline';
import crypto from 'node:crypto';

// Post-hoc evaluator diagnostic only. True rates never enter a learner;
// neither the original gates nor their failed classifications are changed.
const interval = values => {
  const mean = values.reduce((a, v) => a + v / values.length, 0);
  const se = Math.sqrt(values.reduce((a, v) => a + (v - mean) ** 2, 0) / (values.length - 1) / values.length);
  return { mean, se, lower: mean - 3.5 * se, upper: mean + 3.5 * se };
};
const results = { study: 'shape-v35-oracle-headroom', postHoc: true, changesOriginalGates: false, learnerUsesTrueRates: false, splits: {} };
for (const split of ['design', 'confirmation']) {
  const audit = JSON.parse(fs.readFileSync(`docs/experiments/mmm-shape-v35-${split}-audit.json`));
  const stream = fs.createReadStream(`docs/experiments/mmm-shape-v35-${split}.jsonl`);
  const digest = crypto.createHash('sha256');
  stream.on('data', bytes => digest.update(bytes));
  const groups = {};
  let worlds = 0, manifest;
  for await (const line of readline.createInterface({ input: stream, crlfDelay: Infinity })) {
    const w = JSON.parse(line);
    if (!manifest) {
      manifest = w;
      if (w.Kind !== 'manifest' || w.Split !== split || w.Worlds !== 1024) throw Error('manifest');
      continue;
    }
    if (w.Kind !== 'world' || w.Rates.length !== 150) throw Error('world domain');
    const floors = { Brier: 0, PriorityBrier: 0 };
    w.Rates.forEach((p, i) => {
      if (!(p >= 0 && p <= 1)) throw Error('rate');
      const risk = p * (1 - p);
      floors.Brier += risk / 150;
      floors.PriorityBrier += (i < 10 ? 3 : 1) * risk / 170;
    });
    const fixed = w.Snapshots.find(s => s.Model === 'fixed2' && s.Budget === 2400);
    const shape = w.ShapeSnapshots.find(s => s.Model === 'shape' && s.Budget === 2400);
    if (!fixed || !shape) throw Error('missing final laws');
    const name = `${w.Geometry}/${w.Regime}`;
    const group = groups[name] ??= { Brier: [], PriorityBrier: [] };
    for (const field of Object.keys(group)) {
      const oracleGain = fixed[field] - floors[field];
      const remaining = shape[field] - floors[field];
      if (oracleGain < -1e-12 || remaining < -1e-12) throw Error('below Bernoulli floor');
      group[field].push({ oracleGain, remaining, achieved: fixed[field] - shape[field] });
    }
    worlds++;
  }
  if (worlds !== 1024 || digest.digest('hex') !== audit.SHA256) throw Error('raw source changed');
  const cells = {};
  for (const [name, fields] of Object.entries(groups)) {
    cells[name] = {};
    for (const [field, rows] of Object.entries(fields)) {
      if (rows.length !== 32) throw Error('cell trajectories');
      const maximum = interval(rows.map(r => r.oracleGain));
      const exactMeanGateUnattainableOnThisCohort = maximum.mean < .005;
      cells[name][field] = {
        knownRateOracleGain: maximum,
        achievedShapeGain: interval(rows.map(r => r.achieved)),
        remainingShapeExcess: interval(rows.map(r => r.remaining)),
        fixed005MeanGateUnattainableOnThisCohort: exactMeanGateUnattainableOnThisCohort,
      };
    }
  }
  results.splits[split] = { sha256: audit.SHA256, worlds, cells };
  for (const name of ['tight/calibrated', 'tight/mean_shared']) console.log(JSON.stringify({ split, name, headroom: cells[name] }));
}
fs.writeFileSync('docs/experiments/mmm-shape-v35-headroom.json', JSON.stringify(results, null, 2) + '\n', { mode: 0o600 });
