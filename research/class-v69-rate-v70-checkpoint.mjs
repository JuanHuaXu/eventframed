// Chained isolated-research checkpoint; does not mutate git or production.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const root = 'research/checkpoint-2026-10-04-class-v69-rate-v70';
const previous = 'research/checkpoint-2026-10-04-shared-v68-class-v69/manifest.json';
const usage = Number(process.argv[2]);
assert(!fs.existsSync(root));
assert(Number.isFinite(usage) && usage >= 0 && usage <= 100);
async function hash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
const json = p => JSON.parse(fs.readFileSync(p));
assert.equal(await hash(previous), '409e55ab5208c107077a30957fd33720d7bbb3b6260ade01ee45e34a6bdd9d53');
const parent = json(previous);
for (const [p, h] of Object.entries(parent.copies)) assert.equal(await hash(path.dirname(previous) + '/saved/' + p), h, p);
for (const [p, h] of Object.entries(parent.generatedCompilerCopies)) assert.equal(await hash(path.dirname(previous) + '/' + p), h, p);
const tracked = execFileSync('git', ['diff', '--name-only'], {encoding: 'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked, Object.keys(parent.trackedHashes).sort());
assert.equal(tracked.length, 14);
for (const [p, h] of Object.entries(parent.trackedHashes)) assert.equal(await hash(p), h, p);

const run = 'research/class-v69-diagnostic';
const f = json(run + '/freeze.json'), c = json(run + '/completed.json'), r = json(run + '/readback.json');
assert(c.allJobsTerminal && c.sourceUnchanged && c.checks.length === 8 && c.checks.every(x => x.exitCode === 0));
for (const [p, h] of Object.entries({...f.files, ...f.protectedFiles, ...f.data})) assert.equal(await hash(p), h, p);
for (const [p, h] of Object.entries(f.files)) assert.equal(await hash(run + '/source/' + p), h, p);
for (const [p, h] of Object.entries(c.artifacts)) assert.equal(await hash(run + '/' + p), h, p);
for (const [p, h] of Object.entries(f.generatedCompilerCopies)) assert.equal(await hash(run + '/' + p), h, p);
for (const x of c.checks) assert.equal(await hash(run + '/' + x.name + '.log'), x.logSHA256, x.name);
assert.equal(r.arms, 1200);
assert.equal(r.modelArms, 960);
assert.equal(r.independentIssuedPackets, 2304000);
assert.equal(r.scalarIssuedComparisons, 4608000);
assert.equal(r.controlsBitwiseEqual, 960);
assert.equal(r.populationsIdentical, 40);
assert(r.allocationGate);
assert(!r.equalTotalCostSuperiorityEstablished && !r.loadedServingEstablished);
assert.equal(json(run + '/cost-audit.json').dataSHA256, await hash(run + '/diagnostic.jsonl'));
assert.equal(json(run + '/cost-audit.json').sourceSHA256, await hash('research/class-v69-cost-audit.mjs'));
const headroom = json(run + '/recovery-headroom.json');
assert.equal(headroom.dataSHA256, await hash(run + '/diagnostic.jsonl'));
assert.equal(headroom.sourceSHA256, await hash('research/class-v69-recovery-headroom.mjs'));
assert.equal(headroom.worlds, 40);
assert(!headroom.originalFailedVerdictsChanged && !headroom.futureLabelsSuppliedToLearner);
for (const mode of ['hybrid_model_class', 'hybrid_noise_class']) {
  assert(!r.summaries[mode].gates.gain01 && !r.summaries[mode].gates.noAdaptiveHarm01 && !r.summaries[mode].gates.recovery);
}

const failure = json('research/rate-prior-v70-preflight/failure.json');
assert.equal(failure.exitCode, 1);
assert(failure.allJobsTerminal && !failure.scientificGatesChanged);
const original = json('research/rate-prior-v70-preflight/freeze.json');
assert.equal(await hash(original.source), original.sourceSHA256);
assert.equal(await hash('research/rate-prior-v70-preflight/source.mjs'), original.sourceSHA256);
assert.equal(failure.sourceSHA256, original.sourceSHA256);
const pf = json('research/rate-prior-v70-recheck/freeze.json');
const pc = json('research/rate-prior-v70-recheck/completed.json');
const pr = json('research/rate-prior-v70-recheck/results.json');
assert.equal(await hash(pf.source), pf.sourceSHA256);
assert.equal(await hash('research/rate-prior-v70-recheck/source.mjs'), pf.sourceSHA256);
assert(pc.allJobsTerminal && pc.sourceUnchanged && pc.exitCode === 0);
assert.equal(await hash('research/rate-prior-v70-recheck/results.json'), pc.resultsSHA256);
assert.equal(pr.checks, 2079);
assert(!pr.qualityRescueEstablished && !pr.uniqueCauseOfStreamFailureEstablished);
const af = json('research/rate-prior-v70-audit/freeze.json');
const ac = json('research/rate-prior-v70-audit/completed.json');
const ar = json('research/rate-prior-v70-audit/results.json');
for (const [p, h] of Object.entries(af.files)) assert.equal(await hash(p), h, p);
assert.equal(await hash('research/rate-prior-v70-audit/source.mjs'), af.files['research/rate-prior-v70-audit.mjs']);
assert(ac.allJobsTerminal && ac.exitCode === 0);
assert.equal(await hash('research/rate-prior-v70-audit/results.json'), ac.resultsSHA256);
assert.equal(ar.checks, 399);
assert.equal(ar.corruptionControlsRejected, 9);

fs.mkdirSync(root, {mode: 0o700});
const copies = {}, generatedCompilerCopies = {};
async function copy(src) {
  if (Object.hasOwn(copies, src)) {assert.equal(await hash(src), copies[src]); return;}
  const out = root + '/saved/' + src;
  fs.mkdirSync(path.dirname(out), {recursive: true, mode: 0o700});
  fs.copyFileSync(src, out, fs.constants.COPYFILE_EXCL | fs.constants.COPYFILE_FICLONE);
  fs.chmodSync(out, 0o600);
  copies[src] = await hash(src);
  assert.equal(await hash(out), copies[src]);
}
async function tree(dir) {
  for (const n of fs.readdirSync(dir)) {
    const p = dir + '/' + n;
    if (fs.statSync(p).isDirectory()) await tree(p); else await copy(p);
  }
}
for (const p of Object.keys(f.files)) await copy(p);
for (const p of fs.readdirSync(run).filter(p => /\.(json|jsonl|log)$/.test(p))) await copy(run + '/' + p);
for (const [p, h] of Object.entries(f.generatedCompilerCopies)) {
  const out = 'generated/' + h + '-test-main.go';
  fs.mkdirSync(root + '/generated', {recursive: true, mode: 0o700});
  fs.copyFileSync(run + '/' + p, root + '/' + out, fs.constants.COPYFILE_EXCL);
  assert.equal(await hash(root + '/' + out), h);
  generatedCompilerCopies[out] = h;
}
for (const dir of ['research/rate-prior-v70-preflight', 'research/rate-prior-v70-recheck', 'research/rate-prior-v70-audit']) await tree(dir);
for (const n of fs.readdirSync('research').filter(n => /^(class-v69|rate-prior-v70).*\.(mjs|json|md)$/.test(n))) await copy('research/' + n);
for (const p of ['docs/experiments/mmm-class-v69-results.md', 'docs/experiments/mmm-rate-prior-v70-results.md', 'research/dynamic-dispersion-v71-direction.md', 'research-direction.md']) await copy(p);

const manifest = {
  time: new Date().toISOString(), previousCheckpoint: {path: previous, sha256: await hash(previous), copiesVerified: Object.keys(parent.copies).length},
  copies, generatedCompilerCopies, trackedHashes: parent.trackedHashes,
  goal: 'ACTIVE', goals: Array(7).fill('OPEN'), weeklyUsage: {usedPercent: usage, windowDurationMins: 10080}, allRequiredJobsTerminal: true,
  completedCommands: {path: run, completedSHA256: copies[run + '/completed.json']},
  classStudy: {worlds: 40, arms: 1200, modelArms: 960, independentIssuedPackets: 2304000, scalarIssuedComparisons: 4608000, controlsBitwiseEqual: 960, populationsIdentical: 40, totals: r.totals, costComparison: r.costComparison, gates: Object.fromEntries(Object.entries(r.summaries).map(([m, s]) => [m, s.gates])), equalTotalCostSuperiorityEstablished: false, loadedServingEstablished: false},
  recoveryFeasibility: {worlds: headroom.worlds, shiftedWorlds: headroom.shiftedWorlds, phases: headroom.phases, unreachablePhases: headroom.unreachablePhases, oracleMeanIssuedRisk: headroom.oracleMeanIssuedRisk, meanMaximumRiskGainFull: headroom.meanMaximumRiskGainFull, meanMaximumRecoveryGainAdaptiveShifted: headroom.meanMaximumRecoveryGainAdaptiveShifted, evaluatorOnly: true, originalFailedVerdictsChanged: false},
  priorAlgebra: {checks: pr.checks, maxDefect: pr.maxDefect, independentReadbackChecks: ar.checks, independentReadbackMaxDefect: ar.maxDefect, corruptionControls: ar.corruptionControlsRejected, impossiblePairBranches: pr.unsupportedPairBranches, repeatedEvidenceNegativeControls: pr.repeatedEvidenceNegativeControls, originalCheckerFailureRetained: true, originalStderrIsTranscribed: true, streamQualityRescueEstablished: false},
  previousGoalTurn: 'PROGRESS(V69 class implementation, component audits and live full experiment); intervening social clarification NO_RESEARCH_PROGRESS',
  thisGoalTurn: 'PROGRESS(full1200-arm collection and960-arm independent replay, immutable control checks, complete measured cost readback, preserved negative science and independently audited prior tradeoff)',
  productionPrivateWhitepaperUntouched: true, noCommitPushInstallDeployment: true, privateOrSealedLabelsOpened: false, populationConfirmationDispatched: false, allSevenEmpiricalGoalsValidated: false, scientificAdoption: false, archiveIsChainedNotStandalone: true,
  residualLimitations: ['consumed worlds/n1 per cell are not fresh confirmation', 'structural class scores do not inherit EC2 competitiveness or certify external truth', 'equal requests are not equal TOTAL cost', 'core elapsed is not loaded serving/persistence/freshness', 'finite prior algebra is not a stream rescue', 'valid useful splits, untouched agent outcomes and durable loaded freshness remain required'],
  next: 'Coherent dynamic/noisy integration of existing learnable-dispersion components, not unconditional prior widening; preserve all seven whole criteria and pursue valid useful splits, untouched agent tasks and durable loaded freshness.'
};
fs.writeFileSync(root + '/manifest.json', JSON.stringify(manifest, null, 2) + '\n', {flag: 'wx', mode: 0o600});
for (const [p, h] of Object.entries(copies)) assert.equal(await hash(root + '/saved/' + p), h, p);
for (const [p, h] of Object.entries(parent.trackedHashes)) assert.equal(await hash(p), h, p);
console.log(JSON.stringify({path: root + '/manifest.json', sha256: await hash(root + '/manifest.json'), copies: Object.keys(copies).length, allJobsTerminal: true, goal: 'ACTIVE', wholeGoals: 'seven OPEN', weeklyUsage: usage}, null, 2));
