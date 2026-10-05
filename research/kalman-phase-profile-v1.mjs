import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const inputs = {
  design: {
    file: 'docs/experiments/mmm-kalman-early-v2-design.jsonl',
    sha256: 'fc7c60c6ed31326c126eba9bc8ddec39811b162ffeb37ae5eaaa9d6323ccb00b',
  },
  confirmation: {
    file: 'docs/experiments/mmm-kalman-early-v2-confirmation.jsonl',
    sha256: 'fe168278174624879307d2cbbe43e96ef4775da9eb56e178fa49bc51b2b3de1b',
  },
};
const scenarios = { shift128: 128, delayed_missing: 256 };
const arms = ['base', 'last64', 'q0005', 'q002', 'q01'];

function readTrials({ file, sha256 }, split) {
  const raw = readFileSync(root + file);
  const actual = createHash('sha256').update(raw).digest('hex');
  if (actual !== sha256) throw new Error(`${file}: source hash changed`);
  const [manifest, ...trials] = raw.toString().trim().split('\n').map(JSON.parse);
  if (manifest.kind !== 'manifest' || manifest.split !== split || trials.length !== 160) {
    throw new Error(`${file}: unexpected cohort`);
  }
  return trials;
}

function profile(trials, scenario, change) {
  const rows = trials.filter((r) => r.Scenario === scenario);
  if (rows.length !== 32) throw new Error(`${scenario}: expected 32 trajectories`);
  const phases = [];
  for (const [from, to] of [[0, 16], [16, 32], [32, 48], [48, 64]]) {
    const brier = Array(arms.length).fill(0);
    let arrivedAudits = 0;
    for (const row of rows) {
      if (row.Ticks.length !== 512) throw new Error('unexpected stream length');
      for (let t = change + from; t < change + to; t++) {
        const tick = row.Ticks[t];
        if (tick.P.length !== arms.length) throw new Error('unexpected arm count');
        for (let a = 0; a < arms.length; a++) {
          const p = tick.P[a];
          if (!(p >= 0 && p <= 1)) throw new Error('invalid probability');
          brier[a] += (p - Number(tick.Y)) ** 2;
        }
        for (const origin of tick.Delivered || []) {
          if (origin > t || origin < 0) throw new Error('future or invalid delivery');
          if (origin >= change && row.Ticks[origin].Audit) arrivedAudits++;
        }
      }
    }
    phases.push({
      clocksAfterChange: `${from}-${to - 1}`,
      meanBrier: Object.fromEntries(arms.map((arm, a) => [arm, Number((brier[a] / (rows.length * (to - from))).toFixed(5))])),
      arrivedPostChangeAudits: arrivedAudits,
    });
  }
  return { trajectories: rows.length, phases };
}

const result = {};
for (const [split, input] of Object.entries(inputs)) {
  const trials = readTrials(input, split);
  result[split] = Object.fromEntries(
    Object.entries(scenarios).map(([name, change]) => [name, profile(trials, name, change)]),
  );
}
process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
