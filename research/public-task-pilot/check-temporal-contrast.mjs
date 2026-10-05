import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const read = p => JSON.parse(fs.readFileSync(new URL(p, import.meta.url)));
const data = read('./temporal-contrast-v1/results.json');
const queries = read('./temporal-contrast-v1/queries.json');
const oracle = read('./temporal-contrast-v1/oracle.json');
const corpus = read('./context-v1/corpus.json');
for (const [p,h] of Object.entries(data.Hashes)) {
  assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',p))).digest('hex'), h);
}
const pairs = [['ceres-before','ceres-after'], ['pluto-before','pluto-after'],
  ['ceres-initial','ceres-subsequent'], ['count-proposed','count-adopted'],
  ['status-proposed','status-adopted']];
assert.equal(queries.length, 12);
assert.equal(new Set(queries.map(q=>q.case_id)).size, 12);
assert.equal(data.Results.length, 24);
const seen = new Set();
let maxDelta = 0;
for (const r of data.Results) {
  assert(['baseline','bounded'].includes(r.Arm));
  assert(queries.some(q=>q.case_id===r.Case));
  assert(!seen.has(`${r.Arm}:${r.Case}`)); seen.add(`${r.Arm}:${r.Case}`);
  assert.deepEqual(r.Candidates.map(c=>c.Fixture).sort(), corpus.map(c=>c.fixture_id).sort());
  const support = oracle[r.Case].support;
  assert(support.every(id=>corpus.some(c=>c.fixture_id===id)));
  assert.equal(r.SupportRank, r.Candidates.findIndex(c=>support.includes(c.Fixture))+1);
  if(r.Arm==='bounded') {
    assert.equal(r.Frontier, 8);
    const base = data.Results.find(b=>b.Arm==='baseline'&&b.Case===r.Case);
    for(const c of r.Candidates) {
      const old = base.Candidates.find(b=>b.Fixture===c.Fixture);
      assert.deepEqual(c.Law,old.Law);
      assert(Number.isFinite(c.Score) && Number.isFinite(old.Score));
      maxDelta = Math.max(maxDelta,Math.abs(c.Score-old.Score));
    }
  }
}
assert(maxDelta<=.005+1e-12);
const summaries = ['baseline','bounded'].map(arm=>{
  const rows = data.Results.filter(r=>r.Arm===arm);
  const positive = rows.filter(r=>oracle[r.Case].support.length>0);
  assert.equal(positive.length,10);
  const pairResults = pairs.map(ids=>{
    const selected = ids.map(id=>rows.find(r=>r.Case===id));
    const tops = selected.map(r=>r.Candidates[0].Fixture);
    return {ids,tops,ranks:selected.map(r=>r.SupportRank),bothCorrect:selected.every(r=>r.SupportRank===1),unchangedTop:tops[0]===tops[1]};
  });
  return {arm,correct:positive.filter(r=>r.SupportRank===1).length,total:10,
    passed:positive.every(r=>r.SupportRank===1),pairResults,
    absent:rows.filter(r=>oracle[r.Case].support.length===0).map(r=>({case:r.Case,top:r.Candidates[0].Fixture,abstentionImplemented:false}))};
});
console.log(JSON.stringify({summaries,maxDelta,lawsExactlyEqual:true,
  scope:'consumed public facts; new contrast queries; retrieval only, no learning or generation'},null,2));
