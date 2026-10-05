import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const [summaryPath, fullPath, manifestPath, outputPath] = process.argv.slice(2);
assert.ok(summaryPath && fullPath && manifestPath && outputPath);
const summaryRaw = fs.readFileSync(summaryPath);
const summary = JSON.parse(summaryRaw);
const manifestRaw = fs.readFileSync(manifestPath);
const manifest = JSON.parse(manifestRaw);
const original = fs.readFileSync(manifest.originalPath, 'utf8');
const generated = fs.readFileSync(manifest.generatedPath, 'utf8');
assert.equal(hash(original), manifest.originalSHA256);
assert.equal(hash(generated), manifest.generatedSHA256);
const changes = [
  ['uint16(rng.Intn(512))', 'researchTruthInput(rng)', 2],
  ['truth(x, t, s, rng)', 'researchTruthLabel(x, t, s, rng)', 1],
  ['truth(x, -1, s, rng)', 'researchTruthLabel(x, -1, s, rng)', 1],
  ['2026093001*1000000 + int64(j*1000+fit)',
    '(int64(2026100122+k)*1000000 + int64(j*1000+fit))', 1],
  ['int64(2026093002+k)', 'int64(2026100124+k)', 1],
];
assert.equal(manifest.replacementCounts.length, changes.length);
let prefix = original;
for (let i = 0; i < changes.length; i++) {
  const [oldText, newText, count] = changes[i];
  assert.equal(manifest.replacementCounts[i].oldText, oldText);
  assert.equal(manifest.replacementCounts[i].count, count);
  const parts = prefix.split(oldText);
  assert.equal(parts.length, count + 1);
  prefix = parts.join(newText);
}
assert.ok(generated.startsWith(prefix));
assert.ok(generated.slice(prefix.length).startsWith('\n// Experiment-only outcome family'));
assert.equal(summary.Hashes['original retained.go'], manifest.originalSHA256);
assert.equal(summary.Hashes['generated retained.go'], manifest.generatedSHA256);
assert.equal(summary.Hashes['outcome protocol'], manifest.protocolSHA256);
assert.equal(summary.Records.length, 480);
assert.equal(summary.Comparisons.length, 120);
assert.equal(summary.Verdicts.length, 3);

const scenarios = [
  'stable05', 'stable20', 'shift128', 'shift256', 'shift384', 'gradual',
  'recurring', 'delayed_missing', 'interaction', 'null',
];
const arms = ['replacement_mmm', 'adaptive_retained', 'static_retained'];
const seen = new Set();
for (const record of summary.Records) {
  assert.ok(['design', 'confirmation'].includes(record.Split));
  assert.ok(scenarios.includes(record.Scenario));
  assert.ok(Number.isInteger(record.Fit) && record.Fit >= 0 && record.Fit < 3);
  assert.ok(Number.isInteger(record.Stream) && record.Stream >= 0 && record.Stream < 8);
  const key = [record.Split, record.Scenario, record.Fit, record.Stream].join('/');
  assert.ok(!seen.has(key));
  seen.add(key);
}

const comparisons = [];
for (const split of ['design', 'confirmation']) {
  for (const scenario of scenarios) {
    const records = summary.Records.filter(row => row.Split === split && row.Scenario === scenario);
    assert.equal(records.length, 24);
    for (const arm of arms) {
      const index = arms.indexOf(arm) + 1;
      for (const window of ['full', 'post']) {
        const gains = records.map(row => {
          const metrics = window === 'full' ? row.Full : row.Post;
          return metrics[0].Brier - metrics[index].Brier;
        });
        const gain = gains.reduce((sum, value) => sum + value, 0) / 24;
        const variance = gains.reduce((sum, value) => sum + (value - gain) ** 2, 0) / 23;
        const margin = 3.6 * Math.sqrt(variance / 24);
        const perFit = [0, 1, 2].map(fit =>
          records.reduce((sum, row, j) => sum + (row.Fit === fit ? gains[j] / 8 : 0), 0));
        comparisons.push({Split: split, Scenario: scenario, Arm: arm, Window: window,
          Gain: gain, Lower: gain - margin, Upper: gain + margin, PerFit: perFit});
      }
    }
  }
}

for (const computed of comparisons) {
  const reported = summary.Comparisons.find(row =>
    row.Split === computed.Split && row.Scenario === computed.Scenario &&
    row.Arm === computed.Arm && row.Window === computed.Window);
  assert.ok(reported);
  for (const key of ['Gain', 'Lower', 'Upper']) {
    assert.ok(Math.abs(reported[key] - computed[key]) < 1e-12,
      `${computed.Arm} ${computed.Scenario} ${key}`);
  }
  for (let fit = 0; fit < 3; fit++) {
    assert.ok(Math.abs(reported.PerFit[fit] - computed.PerFit[fit]) < 1e-12);
  }
}

const verdicts = [];
for (const arm of arms) {
  let shift = true, protection = true, noMeanHarm = true, seedSigns = true;
  for (const row of comparisons) {
    if (row.Split !== 'confirmation' || row.Arm !== arm) continue;
    noMeanHarm &&= row.Gain >= -0.01;
    if (['stable05', 'stable20'].includes(row.Scenario) && row.Window === 'full') {
      protection &&= -row.Lower <= 0.01;
    }
    if (['shift128', 'shift256'].includes(row.Scenario) && row.Window === 'post') {
      shift &&= row.Gain >= 0.005 && row.Lower > 0;
      seedSigns &&= row.PerFit.every(gain => gain > 0);
    }
  }
  verdicts.push({Arm: arm, Shift: shift, Protection: protection,
    NoMeanHarm: noMeanHarm, SeedSigns: seedSigns,
    Pass: shift && protection && noMeanHarm && seedSigns});
}
assert.deepEqual(summary.Verdicts, verdicts);

const mean = values => values.reduce((sum, value) => sum + value, 0) / values.length;
const notable = comparisons.filter(row => row.Split === 'confirmation' && row.Window === 'post' &&
  ['shift128', 'shift256', 'interaction', 'recurring', 'delayed_missing'].includes(row.Scenario));
const output = {
  mode: manifest.mode,
  summarySHA256: hash(summaryRaw),
  fullJournalSHA256: hash(fs.readFileSync(fullPath)),
  manifestSHA256: hash(manifestRaw),
  checkerSHA256: hash(fs.readFileSync(new URL(import.meta.url))),
  records: summary.Records.length,
  comparisonsChecked: comparisons.length,
  verdicts,
  notable,
  cost: {
    meanCountFitMsPerStream: mean(summary.Records.map(row => row.FitNS / 1e6)),
    meanTreeFitMsPerStream: mean(summary.Records.map(row => row.TreeNS / 1e6)),
    meanTreeUpdatesPerStream: mean(summary.Records.map(row => row.TreeUpdates)),
    maxTreeNodes: Math.max(...summary.Records.map(row => row.TreeNodes)),
  },
  limits: 'Independent paired arithmetic and source-prefix audit. Full journal chronology and per-frame scores were checked in Go. Synthetic outcome-family test only.',
};
fs.writeFileSync(outputPath, JSON.stringify(output, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify({mode: output.mode, verdicts, cost: output.cost}, null, 2));
