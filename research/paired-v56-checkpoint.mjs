import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const root = 'research/checkpoint-2026-10-04-paired-v56';
const run = 'research/paired-v56-diagnostic';
const previous = 'research/checkpoint-2026-10-04-paired-v55/manifest.json';
const usage = Number(process.argv[2]);
assert(!fs.existsSync(root));
assert(Number.isFinite(usage) && usage >= 0 && usage <= 100);
async function fileHash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
assert.equal(await fileHash(previous), '48229e5dafe85d66d1202692282bcd46368826850fe40472b088127dba571fd4');
const old = JSON.parse(fs.readFileSync(previous));
const freeze = JSON.parse(fs.readFileSync(run + '/freeze.json'));
const completed = JSON.parse(fs.readFileSync(run + '/completed.json'));
const readback = JSON.parse(fs.readFileSync(run + '/readback.json'));
const equality = JSON.parse(fs.readFileSync(run + '/equivalence.json'));
assert.deepEqual(freeze.protectedFiles, old.trackedHashes);
assert.deepEqual(completed.checks.map(c => c.name), ['race-model', 'race-fixture', 'closure-helper', 'vet', 'allocation', 'experiment', 'audit']);
assert(completed.checks.every(c => c.exitCode === 0));
for (const [p, h] of Object.entries({ ...freeze.files, ...freeze.protectedFiles })) assert.equal(await fileHash(p), h, p);
for (const [p, h] of Object.entries(freeze.files)) assert.equal(await fileHash(run + '/source/' + p), h, p);
for (const [p, h] of Object.entries(completed.artifacts)) assert.equal(await fileHash(run + '/' + p), h, p);
for (const [p, h] of Object.entries(old.copies)) assert.equal(await fileHash(path.dirname(previous) + '/saved/' + p), h, p);
assert(readback.sourceClosureVerified);
assert.equal(readback.worlds, 40);
assert.equal(readback.arms, 840);
assert.equal(readback.distinctUnderlyingOutcomes, 96000);
assert.equal(equality.totals.arms, 840);
assert.equal(equality.numericTolerance, 0);
assert(equality.strictSelectionEquality && equality.executionEquivalent);
assert.equal(equality.totals.bitwiseNumericDifferences, 0);
assert.equal(equality.totals.changedDecisionBatches, 0);
const costFailures = Object.fromEntries(Object.entries(readback.summaries).map(([k, v]) => [k, v.over400MS]));
assert.deepEqual(costFailures, { random: 0, uncertainty: 4, information: 1, falsification: 1 });

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
for (const p of ['docs/experiments/mmm-paired-v56-results.md', 'research-direction.md', 'research/paired-v56-checkpoint.mjs', 'research/paired-v57-predictive-value-direction.md', 'research/paired-v57-predictive-value-identities.mjs', 'research/paired-v57-predictive-value-identities.json', 'research/decision-observation-sources/manifest.json']) await copy(p);
const manifest = {
  time: new Date().toISOString(),
  previousCheckpoint: { path: previous, sha256: await fileHash(previous), copiedArtifactsVerified: Object.keys(old.copies).length },
  copies, trackedHashes: freeze.protectedFiles,
  frozenCompilerInputs: freeze.compilerFiles.length,
  frozenSourceFiles: Object.keys(freeze.files).length,
  terminalCommands: completed.checks.length, allRequiredJobsTerminal: true,
  worlds: 40, arms: 840, distinctUnderlyingOutcomes: 96000,
  consumedV54Populations: true, executionEquivalent: equality.executionEquivalent,
  executionDifferences: equality.totals, costFailures,
  fullComputationalRescue: false, diagnosticOnly: true, nPerCell: 1,
  qualityAdoption: false, equalTotalCostSuperiorityEstablished: false,
  previousGoalTurn: 'PROGRESS (complete V56 collection and independent audit)',
  thisGoalTurn: 'PROGRESS (independent readback, bitwise comparison, result and checkpoint)',
  weeklyUsage: { usedPercent: usage, windowDurationMins: 10080, recordedAt: new Date().toISOString() },
  goals: Array(7).fill('OPEN'), goal: 'ACTIVE',
  productionPrivateWhitepaperUntouched: true, noCommitPushInstallDeployment: true,
  archiveIsChainedNotStandalone: true,
  copyrightedPrimaryReportNotVendored: true,
  next: 'Prediction-targeted observation value; preserve every original failure and sealed task cohort'
};
const fd = fs.openSync(root + '/manifest.json', 'wx', 0o600);
try { fs.writeFileSync(fd, JSON.stringify(manifest, null, 2) + '\n'); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
for (const [p, h] of Object.entries(copies)) assert.equal(await fileHash(root + '/saved/' + p), h, p);
console.log(JSON.stringify({ path: root + '/manifest.json', sha256: await fileHash(root + '/manifest.json'), copies: Object.keys(copies).length, executionEquivalent: equality.executionEquivalent, fullComputationalRescue: false, weeklyUsage: usage, wholeGoals: 'OPEN', goal: 'ACTIVE' }, null, 2));
