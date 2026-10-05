import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const file = 'docs/experiments/mmm-regime-query-batch-benchmark.txt';
const raw = fs.readFileSync(file, 'utf8');
assert(raw.includes('\nPASS\n'));
const records = [...raw.matchAll(/^BenchmarkRegimeQueryBatchPaired\/(reference_cold|batch_cold)-\d+\s+(\d+)\s+(\d+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/gm)]
  .map(m => ({arm: m[1], iterations: +m[2], nsPerOp: +m[3], bytesPerOp: +m[4], allocsPerOp: +m[5]}));
assert.equal(records.length, 6);
const median = xs => { assert.equal(xs.length, 3); return xs.toSorted((a, b) => a - b)[1]; };
const arms = Object.fromEntries(['reference_cold', 'batch_cold'].map(arm => {
  const rs = records.filter(r => r.arm === arm); assert.equal(rs.length, 3);
  return [arm, Object.fromEntries(['nsPerOp', 'bytesPerOp', 'allocsPerOp'].map(k => [k, median(rs.map(r => r[k]))]))];
}));
const speedup = arms.reference_cold.nsPerOp / arms.batch_cold.nsPerOp;
const allocationRatio = arms.batch_cold.bytesPerOp / arms.reference_cold.bytesPerOp;
const sources = [file, 'internal/observationlearners/regime_query_research_test.go',
  'internal/observationlearners/regime_query_batch_test.go',
  'internal/observationlearners/regime_query_batch_contract_test.go',
  'internal/observationlearners/segment_posterior.go',
  'internal/observationlearners/segment_batch_research_test.go',
  'docs/experiments/mmm-regime-query-batch-plan.md'];
const hashes = Object.fromEntries(sources.map(p => [p, crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')]));
console.log(JSON.stringify({scope: 'Isolated per-call cold fitted state; shared immutable lookup tables may be warm. Not serving p95/p99.',
  historyFrames: 80, retainedLabels: 63, candidateQueries: 8, probes: 8,
  records, medians: arms, speedup, allocationRatio, pass: speedup >= 4 && allocationRatio <= .5,
  hashes, scorerSHA256: crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex')}, null, 2));
