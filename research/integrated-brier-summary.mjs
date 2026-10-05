import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input, output] = process.argv.slice(2); assert(input && output);
const bytes = fs.readFileSync(input);
const [header, ...rows] = bytes.toString().trim().split('\n').map(JSON.parse);
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind, 'header'); assert.equal(rows.length, 160);
assert.deepEqual(header.Seeds, [2026092131, 2026092132]); assert.equal(header.PerCell, 16);
for (const [name, digest] of Object.entries(header.Hashes)) {
  assert.equal(hash(header.Sources[name]), digest);
  assert.equal(hash(fs.readFileSync(name)), digest);
}
assert.equal(new Set(rows.map(r => r.FitSeed)).size, 160);
assert.equal(new Set(rows.map(r => r.FitHash)).size, 160);
const mean = a => a.reduce((s, x) => s + x, 0) / a.length;
const interval = a => {
  const m = mean(a), se = Math.sqrt(a.reduce((s, x) => s + (x - m) ** 2, 0) / (a.length - 1) / a.length);
  return {mean: m, lower: m - 3.5 * se, upper: m + 3.5 * se};
};
const summaries = [];
for (const phase of ['design', 'confirmation']) for (const scenario of ['stable', 'member_shift', 'common_shift', 'recurring', 'null']) {
  const rs = rows.filter(r => r.Original.Split === phase && r.Original.Scenario === scenario);
  assert.equal(rs.length, 16);
  assert.deepEqual(rs.map(r => r.Original.Index).sort((a,b) => a-b), Array.from({length:16}, (_,i) => i));
  for (const r of rs) {
    assert(Number.isInteger(r.Fits) && r.Fits >= 0);
    assert.match(r.FitHash, /^[a-f0-9]{64}$/);
    assert.equal(r.Tapes.length, 3);
    for (const tape of r.Tapes) assert.match(tape, /^[a-f0-9]{64}$/);
    for (const a of ['Control', 'Strong', 'Linear']) {
      assert.equal(r[a].SplitAt, r.Control.SplitAt);
      for (const p of ['Full', 'Post']) {
        const m = r[a][p];
        assert.equal(m.N, p === 'Full' ? 512 : scenario === 'recurring' ? 384 : 256);
        assert.equal(m.Cost, r.Control[p].Cost);
        assert(Number.isFinite(m.Brier) && m.Brier >= 0 && m.Brier <= m.N);
        assert(Number.isFinite(m.LogLoss) && m.LogLoss >= 0);
        assert(Number.isInteger(m.Correct) && m.Correct >= 0 && m.Correct <= m.N);
      }
    }
  }
  const loss = (r,a,p) => r[a][p].Brier / r[a][p].N;
  const arms = Object.fromEntries(['Control','Strong','Linear'].map(a => [a,mean(rs.map(r => loss(r,a,'Post')))]));
  const contrasts = {};
  for (const [name,a,b] of [['strongVsControl','Control','Strong'],['linearVsControl','Control','Linear'],['strongVsLinear','Linear','Strong']]) {
    const post = interval(rs.map(r => loss(r,a,'Post') - loss(r,b,'Post')));
    const full = interval(rs.map(r => loss(r,a,'Full') - loss(r,b,'Full')));
    contrasts[name] = {post,full,gainScreen:post.mean >= .005 && post.lower > 0,nonHarmScreen:post.lower >= -.01 && full.lower >= -.01};
  }
  summaries.push({phase,scenario,n:16,arms,contrasts});
}
const result = {sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,trajectories:rows.length,summaries,
  limits:'Frozen exploratory synthetic pilot; mean +/-3.5SE over trajectories, not an anytime confidence sequence or broad robustness certificate. Observations fixed to control.'};
fs.writeFileSync(output, JSON.stringify(result,null,2)+'\n', {flag:'wx'});
console.log(JSON.stringify(result,null,2));
