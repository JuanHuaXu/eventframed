import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const [input, output] = process.argv.slice(2);
const raw = fs.readFileSync(input);
const [header, ...cells] = raw.toString().trim().split('\n').map(JSON.parse);
const hash = value => crypto.createHash('sha256').update(value).digest('hex');
for (const [name, source] of Object.entries(header.Sources)) {
  assert.equal(hash(source), header.Hashes[name]);
  assert.equal(hash(fs.readFileSync(name)), header.Hashes[name]);
}
assert.equal(cells.length, 12);
const seen = new Set();
const means = cells.map(c => {
  assert.ok([0, 1, 2].includes(c.Trial));
  assert.ok([50, 200].includes(c.Size));
  assert.equal(typeof c.Indexed, 'boolean');
  const key = `${c.Trial}/${c.Size}/${c.Indexed}`;
  assert.ok(!seen.has(key)); seen.add(key);
  assert.equal(c.Samples.length, 32);
  assert.equal(c.Verified, 32 * c.Size);
  for (const s of c.Samples) {
    for (const v of Object.values(s)) assert.ok(Number.isSafeInteger(v) && v >= 0);
    assert.ok(s.TotalNS >= s.PreflightNS + s.BeginNS + s.PrepareNS + s.RowsNS + s.CommitNS);
  }
  const meanMS = {};
  for (const phase of ['TotalNS', 'PreflightNS', 'BeginNS', 'PrepareNS', 'RowsNS', 'CommitNS']) {
    meanMS[phase] = c.Samples.reduce((sum, s) => sum + s[phase], 0) / 32 / 1e6;
  }
  return { trial: c.Trial, size: c.Size, indexed: c.Indexed, meanMS };
});
const pairs = means.filter(c => c.indexed).map(c => {
  const control = means.find(x => !x.indexed && x.trial === c.trial && x.size === c.size);
  const extraMS = Object.fromEntries(Object.entries(c.meanMS).map(([p, v]) => [p, v - control.meanMS[p]]));
  return { trial: c.trial, size: c.size, extraMS, totalRatio: c.meanMS.TotalNS / control.meanMS.TotalNS };
});
const result = { rawHash: hash(raw), sourceCount: Object.keys(header.Sources).length,
  verified: cells.reduce((s, c) => s + c.Verified, 0), means, pairs,
  limits: 'Absent index violates source uniqueness. Diagnostic only, no valid rescue or serving-speedup claim.' };
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify({rawHash: result.rawHash, verified: result.verified, pairs}));
