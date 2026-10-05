import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import path from 'node:path';
import {createInterface} from 'node:readline';
import {auditOneChangeSnapshot} from './one-change-reference.mjs';

const id = r => `${r.Phase}:${r.Case}:${r.Index}:${r.Schedule}`;
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const sourceStream = fs.createReadStream(process.argv[3]), sourceHash = crypto.createHash('sha256');
sourceStream.on('data', b => sourceHash.update(b));
const sources = new Map(); let sourceHeader, sourceCount = 0;
for await (const line of createInterface({input: sourceStream, crlfDelay: Infinity})) {
  const r = JSON.parse(line);
  if (!sourceHeader) { sourceHeader = r; assert.equal(r.Version, 'soft-learners-v120'); continue; }
  sourceCount++;
  if ([0, 19, 20].includes(r.Case)) { assert(!sources.has(id(r))); sources.set(id(r), r); }
}
assert.equal(sourceCount, 2688); assert.equal(sources.size, 384);
const inputSHA256 = sourceHash.digest('hex');
assert.equal(inputSHA256, '5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
const stream = fs.createReadStream(process.argv[2]), hash = crypto.createHash('sha256');
stream.on('data', b => hash.update(b));
let header, maxError = 0, checks = 0, weightChecks = 0;
const records = [], identities = new Set();
for await (const line of createInterface({input: stream, crlfDelay: Infinity})) {
  const r = JSON.parse(line);
  if (!header) {
    header = r; assert.equal(r.Version, 'one-change-v1'); assert.equal(r.InputSHA256, inputSHA256);
    assert.equal(r.Cap, 64); assert.equal(r.MinimumSide, 8); assert.equal(r.NoChangePrior, .9); assert.equal(r.GenericPrior, .95);
    assert.equal(r.Records, 1152); assert.equal(r.Workers, 4);
    for (const [file, expected] of Object.entries(r.Hashes)) assert.equal(sha(fs.readFileSync(path.resolve('internal/observationlearners', file))), expected);
    continue;
  }
  assert([160, 192, 224].includes(r.Clock));
  const key = `${id(r)}:${r.Clock}`; assert(!identities.has(key)); identities.add(key);
  const source = sources.get(id(r)); assert(source);
  const audited = auditOneChangeSnapshot(r, source); maxError = Math.max(maxError, audited.maxError);
  assert.deepEqual(r.Origins, source.Fits[r.Clock / 32].Origins[0]);
  const losses = [0, 0, 0];
  for (let i = 0; i < 32; i++) {
    const s = source.Steps[r.Clock + i];
    assert(Math.abs(audited.ref.full[i] - s.P[13]) < 1e-10);
    const ps = [r.Predictions[i], s.P[13], s.P[10]];
    ps.forEach((p, arm) => { assert(Number.isFinite(p) && p > 0 && p < 1); losses[arm] += ((p - s.Q) ** 2 + s.Q * (1 - s.Q)) / 32; });
    checks++;
  }
  weightChecks += r.Fit.Weights.length;
  records.push({phase: r.Phase, case: r.Case, index: r.Index, schedule: r.Schedule, clock: r.Clock,
    brier: losses, changeProbability: 1 - r.Fit.Weights[0]});
}
assert.equal(records.length, 1152); assert.equal(checks, 36864);
const mean = xs => xs.reduce((a, b) => a + b, 0) / xs.length;
const ci = xs => { assert.equal(xs.length, 32); const m = mean(xs), se = Math.sqrt(xs.reduce((s, x) => s + (x - m) ** 2, 0) / 31 / 32); return {mean: m, lower: m - 3.5 * se, upper: m + 3.5 * se}; };
const groups = []; let nonharm = 0, gains = 0, gainTotal = 0;
for (let phase = 0; phase < 2; phase++) for (const c of [0, 19, 20]) for (let schedule = 0; schedule < 2; schedule++) for (const clock of [160, 192, 224]) {
  const rs = records.filter(r => r.phase === phase && r.case === c && r.schedule === schedule && r.clock === clock);
  assert.equal(rs.length, 32); assert.equal(new Set(rs.map(r => r.index)).size, 32);
  const contrasts = [1, 2].map(control => ci(rs.map(r => r.brier[control] - r.brier[0])));
  const target = c !== 0 && (clock === 160 || (schedule === 1 && clock === 192));
  for (const delta of contrasts) { if (delta.lower >= -.01) nonharm++; if (target) { gainTotal++; if (delta.mean >= .005 && delta.lower > 0) gains++; } }
  groups.push({phase, case: c, schedule, clock, brier: [0, 1, 2].map(a => mean(rs.map(r => r.brier[a]))),
    changeProbability: mean(rs.map(r => r.changeProbability)), controlMinusCandidate: contrasts, target});
}
assert.equal(gainTotal, 24);
const hashes = Object.fromEntries(['one-change-summary.mjs', 'one-change-reference.mjs', 'regime-predictive-reference.mjs'].map(f => [f, sha(fs.readFileSync(new URL(f, import.meta.url)))]));
console.log(JSON.stringify({scope: 'Consumed-data snapshot screen, not full-stream or serving validation', artifactSHA256: hash.digest('hex'), inputSHA256, hashes,
  checks, weightChecks, maxError, nonharm, nonharmTotal: 72, gains, gainTotal, pass: nonharm === 72 && gains === 24, records, groups}));
