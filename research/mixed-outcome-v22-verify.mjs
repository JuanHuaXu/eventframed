import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');

function assert(ok, message) {
  if (!ok) throw new Error(message);
}

function near(a, b, message) {
  assert(Number.isFinite(a) && Math.abs(a - b) <= 1e-10, `${message}: ${a} != ${b}`);
}

function risk(prediction, truth) {
  assert(prediction >= 0 && prediction <= 1 && truth > 0 && truth < 1, 'invalid probability');
  return truth * (1 - prediction) ** 2 + (1 - truth) * prediction ** 2;
}

function paired(values) {
  const mean = values.reduce((sum, x) => sum + x, 0) / values.length;
  const variance = values.reduce((sum, x) => sum + (x - mean) ** 2, 0) / (values.length - 1);
  const margin = 3.5 * Math.sqrt(variance / values.length);
  return { mean, lower: mean - margin, upper: mean + margin };
}

function readSplit(split) {
  const path = resolve(root, `docs/experiments/mmm-mixed-outcome-v22-${split}.jsonl`);
  const content = readFileSync(path);
  const [manifestLine, ...rowLines] = content.toString().trimEnd().split('\n');
  const manifest = JSON.parse(manifestLine);
  const seedBase = split === 'design' ? 2026102203 : 2026102204;
  assert(manifest.kind === 'manifest' && manifest.split === split &&
    manifest.seedBase === seedBase && manifest.worlds === 4, 'invalid manifest');
  for (const [name, wanted] of Object.entries(manifest.hashes)) {
    const actual = createHash('sha256').update(readFileSync(resolve(root, name))).digest('hex');
    assert(actual === wanted, `source hash mismatch: ${name}`);
  }
  assert(rowLines.length === 4, `${split} row count`);
  const rows = rowLines.map(JSON.parse);
  const worlds = new Set();
  for (const row of rows) {
    assert(row.Kind === 'trial' && row.Split === split && row.World >= 0 && row.World < 4 &&
      row.Seed === seedBase + row.World * 1000, `${split} identity`);
    assert(!worlds.has(row.World), `${split} duplicate world`);
    worlds.add(row.World);
    assert(row.InitialEpoch === 217 && row.LearnedEpoch === 217 && row.AfterWriteEpoch === 233,
      `${split} epoch transition`);
    assert(row.InitialCertified && row.LearnedCertified && row.AfterWriteCertified,
      `${split} synthetic certificate state`);
    assert(row.Candidates.length === 16 && row.OutcomeNS.length === 16, `${split} count`);
    const ids = new Set();
    let positive = 0;
    const sums = { Before: 0, Base: 0, Learned: 0, AfterWrite: 0 };
    for (const [index, candidate] of row.Candidates.entries()) {
      assert(candidate.EventID && !ids.has(candidate.EventID), `${split} duplicate selected event`);
      ids.add(candidate.EventID);
      near(candidate.Probability, index % 2 === 0 ? 0.8 : 0.2, `${split} hidden law`);
      if (candidate.TrainingUseful) positive++;
      assert(candidate.LearnedBelief && !candidate.AfterWriteBelief, `${split} belief gate`);
      near(candidate.AfterWrite, candidate.Base, `${split} postwrite base reversion`);
      for (const field of Object.keys(sums)) sums[field] += risk(candidate[field], candidate.Probability) / 16;
    }
    assert(positive === row.Positive && 16 - positive === row.Negative && positive > 0 && positive < 16,
      `${split} mixed labels`);
    for (const [field, value] of Object.entries(sums)) near(value, row[`${field}Brier`], `${split} ${field} risk`);
    assert(row.Candidates.filter((c) => c.PackedBefore).length === 10 &&
      row.Candidates.filter((c) => c.PackedLearned).length === 0 &&
      row.Candidates.filter((c) => c.PackedAfterWrite).length === 10,
    `${split} packet exposure`);
    for (const timing of [row.InitialRecallNS, row.LearnedRecallNS, row.AfterWriteRecallNS,
      row.WriteNS, row.RefreshNS, ...row.OutcomeNS]) {
      assert(Number.isFinite(timing) && timing >= 0, `${split} timing`);
    }
  }
  return { split, sha256: createHash('sha256').update(content).digest('hex'),
    worlds: rows.length, labels: rows.reduce((a, r) => a + r.Candidates.length, 0),
    positives: rows.reduce((a, r) => a + r.Positive, 0),
    before: rows.reduce((a, r) => a + r.BeforeBrier, 0) / rows.length,
    base: rows.reduce((a, r) => a + r.BaseBrier, 0) / rows.length,
    learned: rows.reduce((a, r) => a + r.LearnedBrier, 0) / rows.length,
    afterWrite: rows.reduce((a, r) => a + r.AfterWriteBrier, 0) / rows.length,
    learnedGain: paired(rows.map((r) => r.BaseBrier - r.LearnedBrier)),
    maxRecallMS: Math.max(...rows.flatMap((r) => [r.InitialRecallNS, r.LearnedRecallNS,
      r.AfterWriteRecallNS])) / 1e6,
    maxOutcomeMS: Math.max(...rows.flatMap((r) => r.OutcomeNS)) / 1e6,
  };
}

console.log(JSON.stringify({ design: readSplit('design'), confirmation: readSplit('confirmation') }, null, 2));
