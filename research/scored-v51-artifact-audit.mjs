// Separate source/log/raw verification; preserve and annotate stale metadata.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root = 'research/scored-v51-diagnostic';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function digest(p) { const h = crypto.createHash('sha256'); for await (const b of fs.createReadStream(p)) h.update(b); return h.digest('hex'); }
const freeze = JSON.parse(fs.readFileSync(root + '/freeze.json')), done = JSON.parse(fs.readFileSync(root + '/completed.json'));
const bytes = fs.readFileSync(root + '/readback.json'), readback = JSON.parse(bytes);
assert(done.sourceUnchanged && done.checks.length === 16 && done.checks.every(c => c.code === 0));
for (const [p, h] of Object.entries(freeze.files)) { assert.equal(hash(fs.readFileSync(p)), h); assert.equal(hash(fs.readFileSync(root + '/source/' + p)), h); }
for (const [p, h] of Object.entries(done.artifacts)) assert.equal(await digest(root + '/' + p), h, p);
for (const c of done.checks) { assert.deepEqual(JSON.parse(fs.readFileSync(root + '/' + c.name + '-command.json')), c); assert.equal(await digest(root + '/' + c.name + '.log'), c.logSHA256); }
const testRoots = ['TestScoredV51DenseDelayedReference', 'TestScoredV51LegacyExactAndCancellation',
  'TestScoredV51LifecycleAndFutureFork', 'TestScoredV51HybridLegacyExactAndFencing', 'TestScoredV51StreamIdentity',
  'TestHybridV48CompactEquivalent', 'TestHybridV48CompactExtremeRevival',
  'TestHybridV48DelayedHeadsAndOriginalServedLaw', 'TestHybridV48OwnershipCapsAndGlobalFence'];
const log = fs.readFileSync(root + '/race.log', 'utf8'); assert(!log.includes('SKIP'));
for (const name of testRoots) assert(log.includes('--- PASS: ' + name + ' '));
for (const name of ['seed-race', 'fixture', 'fixture-checks', 'allocation', ...freeze.styles, ...freeze.styles.map(x => x + '-audit')]) assert(!fs.readFileSync(root + '/' + name + '.log', 'utf8').includes('SKIP'));
const fixtureChecks = JSON.parse(fs.readFileSync(root + '/fixture-checks.json'));
assert.equal(fixtureChecks.futurePrefixCases, 36); assert.equal(fixtureChecks.semanticCorruptionsRejected, 68);
assert(readback.sourceAndRawHashes && readback.independentMetricArithmetic && !readback.qualityAdoption && !readback.intervalsClaimed);
assert.equal(readback.worlds, 28); assert.equal(readback.distinctLabels, 67200);
assert.equal(readback.candidateArms, 1008); assert.equal(readback.legacyControlArms, 252);
assert.equal(readback.exactLegacyPairs, 252); assert.equal(readback.exactSharedAdvicePairs, 1008);
const usage = Number(process.env.EVENTFRAME_WEEKLY_USAGE), reset = Number(process.env.EVENTFRAME_WEEKLY_RESET);
assert(Number.isFinite(usage) && usage >= 0 && usage <= 80 && Number.isSafeInteger(reset) && reset > 0);
if (usage !== readback.weeklyUsage) {
  assert.equal(readback.weeklyUsage, 20, 'specific stale field remains preserved');
  const correction = { time: new Date().toISOString(), readbackSHA256: hash(bytes),
    preservedReportedPercent: readback.weeklyUsage, authoritativeFreshPercent: usage, resetsAt: reset,
    reason: 'Caller passed stale 20 after the usage tool returned 21. Raw readback is preserved; only usage metadata is superseded.',
    scientificFieldsChanged: false, criterionAbove80Triggered: false };
  fs.writeFileSync(root + '/readback-usage-correction.json', JSON.stringify(correction, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
}
const result = { time: new Date().toISOString(), sources: Object.keys(freeze.files).length,
  artifacts: Object.keys(done.artifacts).length, commands: done.checks.length, sourceCopiesAndRawHashes: true,
  commandHashes: true, executedRaceRoots: testRoots, futurePrefixes: 36, semanticCorruptions: 68,
  exactLegacyPairs: 252, exactSharedAdvicePairs: 1008, readbackSHA256: hash(bytes), weeklyUsage: usage,
  allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' };
fs.writeFileSync(root + '/artifact-audit.json', JSON.stringify(result, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify(result, null, 2));
