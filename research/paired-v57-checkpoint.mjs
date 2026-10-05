import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root = 'research/checkpoint-2026-10-04-paired-v57';
const run = 'research/paired-v57-diagnostic';
const previous = 'research/checkpoint-2026-10-04-paired-v56/manifest.json';
const usage = Number(process.argv[2]);
assert(!fs.existsSync(root));
assert(Number.isFinite(usage) && usage >= 0 && usage <= 100);
async function fileHash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
assert.equal(await fileHash(previous), 'd8e840f1ed8928303646c622b11b55a516a2ec2fda5e283a1aa98667ed36cca3');
const parent = JSON.parse(fs.readFileSync(previous));
const freeze = JSON.parse(fs.readFileSync(run + '/freeze.json'));
const completed = JSON.parse(fs.readFileSync(run + '/completed.json'));
const readback = JSON.parse(fs.readFileSync(run + '/readback.json'));
const equality = JSON.parse(fs.readFileSync(run + '/equivalence.json'));
assert.deepEqual(freeze.protectedFiles, parent.trackedHashes);
assert.deepEqual(completed.checks.map(c => c.name), ['race-model', 'race-fixture', 'closure-helper', 'vet', 'allocation', 'experiment', 'audit']);
assert(completed.checks.every(c => c.exitCode === 0));
for (const [p, h] of Object.entries({ ...freeze.files, ...freeze.protectedFiles })) assert.equal(await fileHash(p), h, p);
for (const [p, h] of Object.entries(freeze.files)) assert.equal(await fileHash(run + '/source/' + p), h, p);
for (const [p, h] of Object.entries(completed.artifacts)) assert.equal(await fileHash(run + '/' + p), h, p);
for (const [p, h] of Object.entries(parent.copies)) assert.equal(await fileHash(path.dirname(previous) + '/saved/' + p), h, p);
assert(readback.sourceClosureVerified);
assert.equal(readback.worlds, 40);
assert.equal(readback.arms, 960);
assert.equal(readback.distinctUnderlyingOutcomes, 96000);
assert.equal(equality.totals.arms, 840);
assert.equal(equality.numericTolerance, 0);
assert(equality.strictSelectionEquality && equality.executionEquivalent);
assert.equal(equality.totals.bitwiseNumericDifferences, 0);
assert.equal(equality.totals.changedDecisionBatches, 0);
const unitRoot = 'research/paired-v57-unit-preflight';
const unitFreeze = JSON.parse(fs.readFileSync(unitRoot + '/freeze.json'));
const unit = JSON.parse(fs.readFileSync(unitRoot + '/completed.json'));
assert.deepEqual(unit.commands.map(c => c.name), ['race', 'vet', 'batch-cost']);
assert(unit.commands.every(c => c.exitCode === 0));
// The initial unit source version precedes the added full-fixture reference.
// Validate its saved version, never silently rewrite it from current sources.
for (const [p, h] of Object.entries(unitFreeze.files)) assert.equal(await fileHash(unitRoot + '/source/' + p), h, p);
for (const c of unit.commands) assert.equal(await fileHash(unitRoot + '/' + c.name + '.log'), c.logSHA256);

fs.mkdirSync(root, { mode: 0o700 });
const copies = {};
async function copy(src) {
  const dest = root + '/saved/' + src;
  fs.mkdirSync(path.dirname(dest), { recursive: true, mode: 0o700 });
  fs.copyFileSync(src, dest, fs.constants.COPYFILE_EXCL);
  fs.chmodSync(dest, 0o600);
  copies[src] = await fileHash(src);
  assert.equal(await fileHash(dest), copies[src]);
}
for (const p of Object.keys(freeze.files)) await copy(p);
for (const p of fs.readdirSync(run).filter(p => /\.(json|jsonl|log)$/.test(p))) await copy(run + '/' + p);
for (const p of fs.readdirSync(unitRoot).filter(p => /\.(json|log)$/.test(p))) await copy(unitRoot + '/' + p);
for (const p of Object.keys(unitFreeze.files)) await copy(unitRoot + '/source/' + p);
for (const p of ['docs/experiments/mmm-paired-v57-results.md', 'research-direction.md', 'research/paired-v57-checkpoint.mjs', 'research/decision-observation-sources/manifest.json']) await copy(p);
const manifest = {
  time: new Date().toISOString(),
  previousCheckpoint: { path: previous, sha256: await fileHash(previous), copiedArtifactsVerified: Object.keys(parent.copies).length },
  copies, trackedHashes: freeze.protectedFiles,
  frozenCompilerInputs: freeze.compilerFiles.length,
  frozenSourceFiles: Object.keys(freeze.files).length,
  terminalCommands: completed.checks.length, allRequiredJobsTerminal: true,
  worlds: 40, arms: 960, distinctUnderlyingOutcomes: 96000,
  consumedV54Populations: true, original840ControlArmsExecutionEquivalent: true,
  executionDifferences: equality.totals, predictiveSummary: readback.summaries.predictive,
  predictiveTotals: readback.totals.predictive,
  diagnosticOnly: true, nPerCell: 1, qualityAdoption: false,
  equalTotalCostSuperiorityEstablished: false,
  previousGoalTurn: 'PROGRESS (V56 full collection and independent audit)',
  thisGoalTurn: 'PROGRESS (V56 verification/checkpoint; V57 research, implementation, math audits and complete experiment)',
  weeklyUsage: { usedPercent: usage, windowDurationMins: 10080, recordedAt: new Date().toISOString() },
  goals: Array(7).fill('OPEN'), goal: 'ACTIVE',
  productionPrivateWhitepaperUntouched: true, noCommitPushInstallDeployment: true,
  archiveIsChainedNotStandalone: true, copyrightedPrimaryReportNotVendored: true,
  next: 'Use the complete all-regime result to choose the next viable lead without relaxing gates or consuming sealed labels'
};
const fd = fs.openSync(root + '/manifest.json', 'wx', 0o600);
try { fs.writeFileSync(fd, JSON.stringify(manifest, null, 2) + '\n'); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
for (const [p, h] of Object.entries(copies)) assert.equal(await fileHash(root + '/saved/' + p), h, p);
console.log(JSON.stringify({ path: root + '/manifest.json', sha256: await fileHash(root + '/manifest.json'), copies: Object.keys(copies).length, controlEquivalent: true, predictive: readback.summaries.predictive, weeklyUsage: usage, wholeGoals: 'OPEN', goal: 'ACTIVE' }, null, 2));
