import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';

const d = JSON.parse(fs.readFileSync('research/public-task-pilot/static-recall-results.json'));
const vector = id => Array.from({length: 24}, (_, block) =>
  [...crypto.createHash('sha256').update(`${id}/block-${block}`).digest()].map(x => x - 127.5)).flat();
const cosine = (a, b) => a.reduce((s, x, i) => s + x * b[i], 0) /
  Math.sqrt(a.reduce((s, x) => s + x*x, 0) * b.reduce((s, x) => s + x*x, 0));
assert.equal(d.N, 6400);
assert.equal(d.Dimension, 768);
assert.equal(d.Arms.length, 2);
for (const [p, h] of Object.entries(d.Hashes)) {
  assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'), h, p);
}
for (const [repeat, a] of d.Arms.entries()) {
  assert.equal(a.Repeat, repeat);
  assert.equal(a.Queries.length, 1024);
  let misses = 0;
  for (const [i, q] of a.Queries.entries()) {
    assert.equal(q.ID, `seed-${i}`);
    assert.equal(q.OwnedExact, true);
    assert.equal(q.Error, '');
    assert.equal(q.Candidates.length, 10);
    assert.equal(new Set(q.Candidates.map(c => c.ID)).size, 10);
    assert.equal(q.Hit, q.Candidates.some(c => c.ID === q.ID));
    assert(q.NS >= 0);
    const v = vector(q.ID);
    for (const [j, c] of q.Candidates.entries()) {
      assert(/^seed-\d+$/.test(c.ID));
      assert(Number(c.ID.slice(5)) < d.N);
      assert(Math.abs(c.Score - cosine(v, vector(c.ID))) < 1e-12);
      if (j) assert(q.Candidates[j-1].Score >= c.Score);
    }
    if (!q.Hit) misses++;
  }
  console.log(JSON.stringify({repeat, queries: a.Queries.length, misses}));
}
