// Separate byte, execution-root and benchmark readback, not a quality test.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash = p => crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const roots = ['research/brier-v50-preflight', 'research/brier-v50-cost'];
const tests = ['TestSubstitutionReferenceAndDomination', 'TestExhaustiveImmediateRegret',
  'TestLifecycleAtomicityAndOwnedAdvice', 'TestInvalidAndRevival',
  'TestCommonOffsetAndExtremePriors', 'TestPredictionCannotReadUnrevealedOutcomes'];
for (const root of roots) {
  const freeze = JSON.parse(fs.readFileSync(root + '/freeze.json'));
  const done = JSON.parse(fs.readFileSync(root + '/completed.json'));
  assert(done.sourcesUnchanged);
  assert.equal(done.commands.length, root.endsWith('cost') ? 3 : 2);
  for (const [p, h] of Object.entries(freeze.sources)) {
    assert.equal(hash(p), h, 'current ' + p);
    assert.equal(hash(root + '/source/' + p), h, 'frozen ' + p);
  }
  for (const c of done.commands) {
    assert.equal(c.code, 0);
    assert.deepEqual(JSON.parse(fs.readFileSync(root + '/' + c.name + '-command.json')), c);
    assert.equal(hash(root + '/' + c.name + '.log'), c.logSHA256);
  }
  const log = fs.readFileSync(root + '/race.log', 'utf8');
  assert(!log.includes('SKIP'));
  for (const name of tests) assert(log.includes('--- PASS: ' + name + ' '), name);
}
const root = roots[1], freeze = JSON.parse(fs.readFileSync(root + '/freeze.json'));
assert.equal(hash('research/memo-v49-normal/completed.json'), freeze.normalCompletedSHA256);
const done = JSON.parse(fs.readFileSync(root + '/completed.json'));
const text = fs.readFileSync(root + '/bench.log', 'utf8');
const rows = text.split('\n').filter(x => x.startsWith('Benchmark')).map(line => {
  const [name, iterations, ns, unit, bytes, byteUnit, allocations, allocUnit] = line.trim().split(/\s+/);
  assert.equal(unit, 'ns/op'); assert.equal(byteUnit, 'B/op'); assert.equal(allocUnit, 'allocs/op');
  return { name: name.replace(/-\d+$/, ''), iterations: Number(iterations), nsPerOp: Number(ns),
    bytesPerOp: Number(bytes), allocationsPerOp: Number(allocations) };
});
assert.equal(rows.length, 9); assert.deepEqual(rows, done.samples);
const ranges = {};
for (const name of ['BenchmarkSubstitution3', 'BenchmarkSubstitution8', 'BenchmarkImmediateCycle3']) {
  const r = rows.filter(x => x.name === name); assert.equal(r.length, 3);
  assert(r.every(x => x.iterations > 0 && Number.isFinite(x.nsPerOp) && x.nsPerOp > 0));
  ranges[name] = { minNS: Math.min(...r.map(x => x.nsPerOp)), maxNS: Math.max(...r.map(x => x.nsPerOp)),
    maxBytesPerOp: Math.max(...r.map(x => x.bytesPerOp)), maxAllocationsPerOp: Math.max(...r.map(x => x.allocationsPerOp)) };
}
const out = { time: new Date().toISOString(), sourceAndCommandHashes: true,
  bothExecutedRaceSuites: tests, independentBenchmarkParsing: true, ranges,
  isolatedComponent: true, broadQualityUnproven: true, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' };
fs.writeFileSync(root + '/artifact-audit.json', JSON.stringify(out, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify(out, null, 2));
