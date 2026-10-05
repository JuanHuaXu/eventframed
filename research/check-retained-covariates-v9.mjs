import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const [summaryPath, fullPath, manifestPath, outputPath] = process.argv.slice(2);
assert.ok(summaryPath && fullPath && manifestPath && outputPath);
const summaryRaw = fs.readFileSync(summaryPath);
const summary = JSON.parse(summaryRaw);
const manifest = JSON.parse(fs.readFileSync(manifestPath));
const original = fs.readFileSync(manifest.originalPath, 'utf8');
const generated = fs.readFileSync(manifest.generatedPath, 'utf8');
assert.equal(hash(original), manifest.originalSHA256);
assert.equal(hash(generated), manifest.generatedSHA256);
const sites = original.split('uint16(rng.Intn(512))');
assert.equal(sites.length, 3);
assert.ok(generated.startsWith(sites.join('researchCovariateDraw(rng)')));
assert.equal(summary.Hashes['original retained.go'], manifest.originalSHA256);
assert.equal(summary.Hashes['generated retained.go'], manifest.generatedSHA256);
assert.equal(summary.Hashes['covariate protocol'], manifest.protocolSHA256);
assert.equal(summary.Records.length, 480);
assert.equal(summary.Comparisons.length, 120);
assert.equal(summary.Verdicts.length, 3);

const scenarios = [
  'stable05', 'stable20', 'shift128', 'shift256', 'shift384', 'gradual',
  'recurring', 'delayed_missing', 'interaction', 'null',
];
const arms = ['replacement_mmm', 'adaptive_retained', 'static_retained'];
const seen = new Set();
for (const row of summary.Records) {
  assert.ok(['design', 'confirmation'].includes(row.Split));
  assert.ok(scenarios.includes(row.Scenario));
  assert.ok(Number.isInteger(row.Fit) && row.Fit >= 0 && row.Fit < 3);
  assert.ok(Number.isInteger(row.Stream) && row.Stream >= 0 && row.Stream < 8);
  const key = [row.Split, row.Scenario, row.Fit, row.Stream].join('/');
  assert.ok(!seen.has(key));
  seen.add(key);
}

const comparisons = [];
for (const split of ['design', 'confirmation']) {
  for (const scenario of scenarios) {
    const rows = summary.Records.filter(row => row.Split === split && row.Scenario === scenario);
    assert.equal(rows.length, 24);
    for (const arm of arms) {
      const armIndex = arms.indexOf(arm) + 1;
      for (const window of ['full', 'post']) {
        const gains = rows.map(row => {
          const metrics = window === 'full' ? row.Full : row.Post;
          return metrics[0].Brier - metrics[armIndex].Brier;
        });
        const gain = gains.reduce((total, value) => total + value, 0) / 24;
        const variance = gains.reduce((total, value) => total + (value - gain) ** 2, 0) / 23;
        const margin = 3.6 * Math.sqrt(variance / 24);
        const perFit = [0, 1, 2].map(fit =>
          rows.reduce((total, row, index) => total + (row.Fit === fit ? gains[index] / 8 : 0), 0));
        comparisons.push({Split: split, Scenario: scenario, Arm: arm, Window: window,
          Gain: gain, Lower: gain - margin, Upper: gain + margin, PerFit: perFit});
      }
    }
  }
}

const tolerance = 1e-12;
for (const computed of comparisons) {
  const reported = summary.Comparisons.find(row =>
    row.Split === computed.Split && row.Scenario === computed.Scenario &&
    row.Arm === computed.Arm && row.Window === computed.Window);
  assert.ok(reported);
  for (const key of ['Gain', 'Lower', 'Upper']) {
    assert.ok(Math.abs(reported[key] - computed[key]) < tolerance,
      `${computed.Arm} ${computed.Scenario} ${key}`);
  }
  for (let fit = 0; fit < 3; fit++) {
    assert.ok(Math.abs(reported.PerFit[fit] - computed.PerFit[fit]) < tolerance);
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

const mean = values => values.reduce((total, value) => total + value, 0) / values.length;
const notable = comparisons.filter(row => row.Split === 'confirmation' && row.Window === 'post' &&
  ['shift128', 'shift256', 'interaction', 'recurring', 'delayed_missing'].includes(row.Scenario));
const output = {
  mode: manifest.mode,
  summarySHA256: hash(summaryRaw),
  fullJournalSHA256: hash(fs.readFileSync(fullPath)),
  manifestSHA256: hash(fs.readFileSync(manifestPath)),
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
  limits: 'Independent summary arithmetic and overlay-source check. Full journal chronology and per-frame scores were checked in Go. This is offline synthetic input-generator stress, not daemon latency or agent-task validation.',
};
fs.writeFileSync(outputPath, JSON.stringify(output, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify({mode: output.mode, verdicts, cost: output.cost}, null, 2));
