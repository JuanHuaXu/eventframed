// Post-hoc diagnostic only. These metrics do not alter the frozen v26 screens.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const risk = (q, p) => p * (1 - q) ** 2 + (1 - p) * q ** 2;
const mean = xs => xs.reduce((a, b) => a + b, 0) / xs.length;
const interval = xs => {
  const m = mean(xs);
  const se = Math.sqrt(xs.reduce((s, x) => s + (x - m) ** 2, 0) / (xs.length - 1) / xs.length);
  return { mean: m, lower: m - 3.5 * se, upper: m + 3.5 * se };
};
const primary = JSON.parse(readFileSync(resolve(root, 'docs/experiments/mmm-packet-prior-v26-summary.json')));
const result = { kind: 'post-hoc-packed-risk-diagnostic', splits: {} };
for (const split of ['design', 'confirmation']) {
  const bytes = readFileSync(resolve(root, `docs/experiments/mmm-packet-prior-v26-${split}.jsonl`));
  const sha256 = createHash('sha256').update(bytes).digest('hex');
  assert.equal(sha256, primary[split].sha256);
  const [, ...rows] = bytes.toString('utf8').trim().split('\n').map(JSON.parse);
  const cases = [];
  for (const geometry of ['tight', 'wide']) {
    for (const regime of ['independent', 'aligned', 'reversed', 'calibrated']) {
      const worlds = rows.filter(r => r.Geometry === geometry && r.Regime === regime);
      assert.equal(worlds.length, 32);
      const before = [], after = [], sameLaw = [], bias = [], forecasts = [], truths = [];
      for (const row of worlds) {
        const candidates = new Map(row.Candidates.map(c => [c.ID, c]));
        const base = row.Controls.find(c => c.Name === 'baseline').IDs.map(id => candidates.get(id));
        const packed = row.ActualPacket.map(id => candidates.get(id));
        assert.equal(base.length, 10);
        assert.equal(packed.length, 10);
        before.push(mean(base.map(c => risk(c.Base, c.Probability))));
        after.push(mean(packed.map(c => risk(c.Law, c.Probability))));
        sameLaw.push(mean(base.map(c => risk(c.Law, c.Probability))));
        forecasts.push(mean(packed.map(c => c.Law)));
        truths.push(mean(packed.map(c => c.Probability)));
        bias.push(forecasts.at(-1) - truths.at(-1));
      }
      const metrics = {
        geometry, regime, worlds: 32,
        baselinePacketBaselineLawBrier: mean(before),
        servedPacketServedLawBrier: mean(after),
        baselinePacketServedLawBrier: mean(sameLaw),
        totalPackedRiskHarm: interval(after.map((x, i) => x - before[i])),
        selectionPackedRiskHarm: interval(after.map((x, i) => x - sameLaw[i])),
        meanServedForecast: mean(forecasts), meanServedTruth: mean(truths),
        meanProbabilityBias: interval(bias),
      };
      cases.push(metrics);
      console.log(JSON.stringify({ split, ...metrics }));
    }
  }
  result.splits[split] = { sha256, cases };
}
const index = process.argv.indexOf('--out');
if (index >= 0) writeFileSync(resolve(root, process.argv[index + 1]), JSON.stringify(result, null, 2) + '\n', { flag: 'wx' });
