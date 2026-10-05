import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const path = 'docs/experiments/mmm-brier-kernel-v94.json';
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const raw = fs.readFileSync(path);
const a = JSON.parse(raw);
assert.equal(a.Protocol, 'mmm-brier-kernel-v94');
assert.equal(Object.keys(a.Hashes).length, 6);
for (const [p, hash] of Object.entries(a.Hashes)) assert.equal(sha(fs.readFileSync(p)), hash, p);
const cases = ['random', 'adaptive_adversary', 'alternating_extremes', 'regime_flip', 'changing_experts', 'generic_better', 'challenger_better'];
assert.equal(a.Records.length, 448);
const seen = new Set();
for (const r of a.Records) {
  assert(cases.includes(r.Case));
  assert(Number.isInteger(r.Index) && r.Index >= 0 && r.Index < 32);
  assert([0, .001].includes(r.Share));
  const key = `${r.Case}/${r.Index}/${r.Share}`;
  assert(!seen.has(key)); seen.add(key);
  assert.equal(r.Steps, 4096);
  assert(r.Loss.length === 3 && r.Loss.every(x => Number.isFinite(x) && x >= 0 && x <= 4096));
  assert(r.MaxBoundViolation >= 0 && r.MaxBoundViolation <= 1e-8);
  assert(r.MaxJensenViolation >= 0 && r.MaxJensenViolation <= 1e-12);
  assert(r.FinalWeights.length === 2 && r.FinalWeights.every(x => Number.isFinite(x) && x >= 0 && x <= 1));
  assert(Math.abs(r.FinalWeights[0] + r.FinalWeights[1] - 1) < 1e-12);
  assert(/^[0-9a-f]{64}$/.test(r.Tape));
  assert(Number.isInteger(r.RecoveryLag) && r.RecoveryLag >= -1 && r.RecoveryLag < 2048);
  assert(Number.isInteger(r.PostCorrect) && r.PostCorrect >= 0 && r.PostCorrect <= 2048);
  if (r.Case !== 'regime_flip') { assert.equal(r.RecoveryLag, -1); assert.equal(r.PostCorrect, 0); }
}
const groups = [];
for (const scenario of cases) for (const share of [0, .001]) {
  const rows = a.Records.filter(r => r.Case === scenario && r.Share === share);
  assert.equal(rows.length, 32);
  const mean = f => rows.reduce((s, r) => s + f(r), 0) / rows.length;
  groups.push({scenario, share, records: rows.length, distinctTapes: new Set(rows.map(r => r.Tape)).size,
    meanBrier: [0, 1, 2].map(k => mean(r => r.Loss[k] / r.Steps)),
    maxRegret: Math.max(...rows.map(r => r.MaxRegret)),
    maxBoundViolation: Math.max(...rows.map(r => r.MaxBoundViolation)),
    maxJensenViolation: Math.max(...rows.map(r => r.MaxJensenViolation)),
    recoveryLag: [...new Set(rows.map(r => r.RecoveryLag))],
    postFlipAccuracy: scenario === 'regime_flip' ? mean(r => r.PostCorrect / 2048) : null});
}
const result = {protocol: a.Protocol, artifactSHA256: sha(raw), evaluatorSHA256: sha(fs.readFileSync(import.meta.filename)),
  runtime: a.Runtime, records: 448, pairedStreams: 224,
  caveat: 'Deterministic duplicate tapes are not independent confirmation; cumulative realized-loss checks do not establish future risk or learned-model quality.', groups};
const output = JSON.stringify(result, null, 2) + '\n';
if (process.argv[2]) fs.writeFileSync(process.argv[2], output, {flag: 'wx', mode: 0o600});
else process.stdout.write(output);
