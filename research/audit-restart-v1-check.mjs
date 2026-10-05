import fs from 'node:fs';
import assert from 'node:assert/strict';

const [streamPath, resultPath] = process.argv.slice(2);
assert.ok(streamPath && resultPath);
const stream = JSON.parse(fs.readFileSync(streamPath));
const result = JSON.parse(fs.readFileSync(resultPath));
assert.equal(stream.records.length, 256);
assert.equal(result.records.length, stream.records.length);

function independentCrossing(scores) {
  for (let k = 0; k < scores.length; k++) {
    let sum = 0;
    for (const start of result.starts) {
      let product = 1;
      for (let j = start - 1; j <= k; j++) {
        const score = scores[j];
        product *= 1 + result.rate *
          (score.referenceCorrect - score.liveCorrect - result.diameter);
      }
      sum += product;
    }
    const wealth = sum / result.starts.length;
    if (wealth >= result.threshold) {
      return {
        clock: scores[k].release,
        origin: scores[k].origin,
        auditUpdates: k + 1,
        wealth,
      };
    }
  }
  return null;
}

let scoresChecked = 0;
for (let i = 0; i < stream.records.length; i++) {
  const source = stream.records[i];
  const recorded = result.records[i];
  for (const key of ['phase', 'case', 'index', 'schedule']) {
    assert.equal(recorded[key], source[key]);
  }
  for (let k = 0; k < source.scores.length; k++) {
    const score = source.scores[k];
    assert.ok(score.arrival <= score.release);
    assert.ok(score.origin <= score.release);
    if (k) {
      assert.ok(source.scores[k - 1].origin < score.origin);
      assert.ok(source.scores[k - 1].release < score.release);
    }
    scoresChecked++;
  }
  const computed = independentCrossing(source.scores);
  if (computed === null) {
    assert.equal(recorded.nomination, null);
  } else {
    assert.ok(recorded.nomination);
    for (const key of ['clock', 'origin', 'auditUpdates']) {
      assert.equal(recorded.nomination[key], computed[key]);
    }
    assert.ok(Math.abs(recorded.nomination.wealth - computed.wealth) < 1e-9);
  }
  assert.equal(recorded.beforeTrueChange,
    recorded.changed && recorded.nomination !== null && recorded.nomination.clock < 256);
}

const count = (rows, predicate) => rows.filter(predicate).length;
const changed = result.records.filter(row => row.changed);
const stable = result.records.filter(row => !row.changed);
assert.equal(changed.length, 128);
assert.equal(stable.length, 128);
const counts = {
  changedBy511: count(changed, row => row.nomination?.clock <= 511),
  changedBy543: count(changed, row => row.nomination?.clock <= 543),
  stableNominations: count(stable, row => row.nomination !== null),
  preChangeNominations: count(changed, row => row.beforeTrueChange),
  externalSplits: count(changed, row => row.externalSplitClock >= 0),
};
assert.deepEqual(counts, {
  changedBy511: 72,
  changedBy543: 82,
  stableNominations: 0,
  preChangeNominations: 0,
  externalSplits: 123,
});

const observations = stream.records.map(row => row.scores);
let sink = 0;
function arithmeticPass() {
  for (const scores of observations) {
    const products = Array(result.starts.length).fill(1);
    for (let i = 0; i < scores.length; i++) {
      const score = scores[i];
      const factor = 1 + result.rate *
        (score.referenceCorrect - score.liveCorrect - result.diameter);
      for (let j = 0; j < products.length; j++) {
        if (i + 1 >= result.starts[j]) products[j] *= factor;
      }
      sink += products.reduce((sum, value) => sum + value, 0);
    }
  }
}
for (let i = 0; i < 8; i++) arithmeticPass();
const trialNs = [];
for (let trial = 0; trial < 5; trial++) {
  const start = process.hrtime.bigint();
  for (let repeat = 0; repeat < 64; repeat++) arithmeticPass();
  trialNs.push(Number(process.hrtime.bigint() - start) /
    (64 * scoresChecked));
}
trialNs.sort((a, b) => a - b);
assert.ok(Number.isFinite(sink));
console.log(JSON.stringify({scoresChecked, counts,
  medianArithmeticNsPerScore: trialNs[2],
  arithmeticNsRange: [trialNs[0], trialNs[4]],
  benchmarkScope: 'eight-product JavaScript arithmetic only; excludes reads, fitting, synchronization, and service latency'}, null, 2));
