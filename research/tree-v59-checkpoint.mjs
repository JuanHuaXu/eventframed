// Exclusive chained checkpoint; verifies the real frozen evidence and preserves
// the negative result. No git commit, upload, production mutation or goal closure.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';

const root = 'research/checkpoint-2026-10-04-tree-v59';
const run = 'research/tree-v59-diagnostic';
const previous = 'research/checkpoint-2026-10-04-tree-v58/manifest.json';
const usage = Number(process.argv[2]);
assert(!fs.existsSync(root));
assert(Number.isFinite(usage) && usage >= 0 && usage <= 100);
async function hash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
const json = p => JSON.parse(fs.readFileSync(p));
assert.equal(await hash(previous), 'e6dcbb9a2eed68b4e7fb4866ecaa6878306805913b6b21e2c6d7d4678cdcb0d7');
const parent = json(previous), freeze = json(run + '/freeze.json'), complete = json(run + '/completed.json');
const readback = json(run + '/readback.json'), c57 = json(run + '/comparison.json'), c58 = json(run + '/comparison-v58.json');
const capacity = json(run + '/capacity.json'), order = json(run + '/order-capacity.json');
assert.deepEqual(freeze.protectedFiles, parent.trackedHashes);
const tracked = execFileSync('git', ['diff', '--name-only'], {encoding: 'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked, Object.keys(freeze.protectedFiles).sort());
assert.equal(tracked.length, 14);
assert.equal(freeze.compilerFiles.length, 77);
assert.equal(Object.keys(freeze.files).length, 89);
assert.deepEqual(complete.checks.map(c => c.name), ['race-model', 'race-fixture', 'closure-helper', 'vet', 'microbench', 'allocation', 'experiment', 'audit']);
assert(complete.checks.every(c => c.exitCode === 0));
assert(!fs.existsSync(run + '/failure.json'));
for (const [p, h] of Object.entries({...freeze.files, ...freeze.protectedFiles})) assert.equal(await hash(p), h, p);
for (const [p, h] of Object.entries(freeze.files)) assert.equal(await hash(run + '/source/' + p), h, p);
for (const [p, h] of Object.entries(complete.artifacts)) assert.equal(await hash(run + '/' + p), h, p);
for (const [p, h] of Object.entries(parent.copies)) assert.equal(await hash(path.dirname(previous) + '/saved/' + p), h, p);
assert.equal(readback.worlds, 40);
assert.equal(readback.arms, 960);
assert.equal(readback.distinctUnderlyingOutcomes, 96000);
assert(readback.sourceClosureVerified);
for (const c of [c57, c58]) { assert.equal(c.controlsBitwiseEqual, 240); assert(c.allPopulationsIdentical); }
for (const [name, s] of Object.entries(readback.summaries)) {
  assert.equal(s.adaptiveHarmOver01, 105);
  assert.equal(s.over400MS, name === 'predictive' ? 76 : 0);
  assert(s.fullRiskGain < 0 && s.recoveryGainAdaptive < 0);
}
for (const name of ['no_pair', 'random', 'uncertainty', 'information', 'falsification', 'predictive']) {
  assert.equal(c58.changes[name].wins, 3);
  assert.equal(c58.changes[name].losses, 117);
}
assert.equal(capacity.unrepresentable, 26874);
assert.equal(capacity.forecastRangeChecks, 1728000);
assert.equal(order.snapshotChecks, 12000);
assert.equal(order.pairChecks, 276000);
assert.equal(order.equalityChecks, 12000);
assert.equal(capacity.sourceSHA256, await hash('research/tree-v59-capacity.mjs'));
assert.equal(order.sourceSHA256, await hash('research/tree-v59-order-capacity.mjs'));
for (const name of ['tree-v60-identities', 'tree-v60-beta-identities']) {
  const m = json('research/' + name + '.json');
  assert.equal(m.sourceSHA256, await hash('research/' + name + '.mjs'));
  assert.equal(m.baselineIdentities, 28);
  assert.equal(m.pairedMassChecks, 1848);
  assert.equal(m.quantizationChecks, 10001);
  assert.equal(m.jointComparisons, 18);
  assert(m.cases.every(c => c.enumeratedStates === 1515));
}
assert.equal(await hash('research/tree-v60-preflight/failed-identities.mjs'), '1d4c4f5e45c0091f14ac6b78d76156efc59169e898fc449e6259856fcc5f50ca');
assert.equal(await hash('research/tree-v60-preflight/second-identities.mjs'), '9957bd5a217b3f50a7806243c0deddbbd5f831cb399750805571935b2181329b');

fs.mkdirSync(root, {mode: 0o700});
const copies = {};
async function copy(src) {
  if (Object.hasOwn(copies, src)) { assert.equal(await hash(src), copies[src]); return; }
  const dest = root + '/saved/' + src;
  fs.mkdirSync(path.dirname(dest), {recursive: true, mode: 0o700});
  fs.copyFileSync(src, dest, fs.constants.COPYFILE_EXCL);
  fs.chmodSync(dest, 0o600);
  copies[src] = await hash(src);
  assert.equal(await hash(dest), copies[src]);
}
for (const p of Object.keys(freeze.files)) await copy(p);
for (const p of fs.readdirSync(run).filter(p => /\.(json|jsonl|log)$/.test(p))) await copy(run + '/' + p);
for (const p of ['docs/experiments/mmm-tree-v59-results.md', 'research-direction.md', 'research/tree-v59-checkpoint.mjs',
  'research/tree-v59-capacity.mjs', 'research/tree-v59-order-capacity.mjs', 'research/tree-v60-direction.md',
  'research/tree-v60-identities.mjs', 'research/tree-v60-identities.json',
  'research/tree-v60-beta-gen.mjs', 'research/tree-v60-beta-generation.json',
  'research/tree-v60-beta-identities.mjs', 'research/tree-v60-beta-identities.json',
  'research/tree-v60-preflight/failure.md', 'research/tree-v60-preflight/failed-identities.mjs',
  'research/tree-v60-preflight/second-identities.mjs']) await copy(p);
const manifest = {
  time: new Date().toISOString(), previousCheckpoint: {path: previous, sha256: await hash(previous), copiedArtifactsVerified: Object.keys(parent.copies).length},
  copies, trackedHashes: freeze.protectedFiles, frozenCompilerInputs: 77, frozenSourceFiles: 89,
  terminalCommands: 8, allRequiredJobsTerminal: true, worlds: 40, arms: 960, distinctUnderlyingOutcomes: 96000,
  consumedV54Populations: true, controlsBitwiseEqualBothV57V58: 240, scientificAdoption: false,
  predictive400MSFailures: 76, constructorAllocatedBytes: readback.allocationBytes.modelAllocatedBytes,
  notLoadedServingLatency: true, notEmpiricalGoalCompletion: true, summary: readback.summaries.predictive,
  comparisonV58: c58.changes.predictive, comparisonV57: c57.changes.predictive,
  capacity: {unrepresentable: capacity.unrepresentable, fraction: capacity.fractionUnrepresentable, excessFloor: capacity.minimumCapacityExcess, forecastChecks: capacity.forecastRangeChecks},
  order: {meanExcessFloor: order.meanRelaxedOrderExcess, snapshots: order.snapshotChecks, pairs: order.pairChecks, scope: order.scope},
  expiredObservationAccounting: c58.expiry, nextModelMathematicsOnly: true, populationConfirmationDispatched: false,
  goals: Array(7).fill('OPEN'), goal: 'ACTIVE',
  previousGoalTurn: 'PROGRESS(V59 isolated implementation/full experiment/audit); intervening social turn NO_RESEARCH_PROGRESS',
  thisGoalTurn: 'PROGRESS(terminal readback, two structural capacity audits, preserved rejection, finite member-local prototypes and exact mathematical tests)',
  weeklyUsage: {usedPercent: usage, windowDurationMins: 10080, recordedAt: new Date().toISOString()},
  productionPrivateWhitepaperUntouched: true, noCommitPushInstallDeployment: true, archiveIsChainedNotStandalone: true,
  next: 'Implement member-local independent branch and finite Polya prior with atomic dual-factor lifecycle, independent reconstruction and complete original seven-goal gates'
};
const fd = fs.openSync(root + '/manifest.json', 'wx', 0o600);
try { fs.writeFileSync(fd, JSON.stringify(manifest, null, 2) + '\n'); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
for (const [p, h] of Object.entries(copies)) assert.equal(await hash(root + '/saved/' + p), h, p);
console.log(JSON.stringify({path: root + '/manifest.json', sha256: await hash(root + '/manifest.json'),
  copies: Object.keys(copies).length, allRequiredJobsTerminal: true, wholeGoals: 'OPEN', goal: 'ACTIVE', weeklyUsage: usage}, null, 2));
