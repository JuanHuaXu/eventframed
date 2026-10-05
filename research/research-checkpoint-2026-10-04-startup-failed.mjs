// Exact source/evidence checkpoint; no commit, push, deployment or private data.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {v43NormalRunners, v43AuditProcesses, v43TimedProcesses} from './research-process-guards.mjs';
const root = 'research/checkpoint-2026-10-04-switch-load-public';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p) { const h = crypto.createHash('sha256'); for await (const b of fs.createReadStream(p)) h.update(b); return h.digest('hex'); }
assert.equal(v43NormalRunners().length + v43AuditProcesses().length + v43TimedProcesses().length, 0);
const usage = Number(process.env.EVENTFRAME_WEEKLY_USAGE);
assert(Number.isFinite(usage) && usage >= 0 && usage <= 80, 'actual current usage');
const freezes = ['research/switch-v43-study-normal/freeze.json', 'research/eager-load-v44-owner-time/freeze.json'];
const files = {};
for (const record of freezes) for (const [p, h] of Object.entries(JSON.parse(fs.readFileSync(record)).files)) {
  assert(files[p] === undefined || files[p] === h, 'inconsistent active snapshots'); files[p] = h;
}
const extras = ['research/switch-v43-policy-verdict.mjs', 'research/switch-v43-quality-breakdown.mjs',
  'research/eager-load-v44-readback.mjs', 'research/eager-load-v44-audit-recovery.mjs',
  'research/public-task-pilot/scifact-source.mjs', 'research/public-task-pilot/scifact-source-test.mjs',
  'research/public-task-pilot/scifact-groups.mjs', 'research/public-task-pilot/scifact-groups-test.mjs',
  'research/public-task-pilot/scifact-groups-audit.mjs', 'research/public-task-pilot/scifact-cache-preflight.mjs',
  'research/public-task-pilot/scifactcache/memo.go', 'research/public-task-pilot/scifactcache/memo_test.go',
  'research/public-task-pilot/SCIFACT_SOURCE_DIRECTION.md',
  'research/public-task-pilot/metrology-headroom.mjs', 'research/public-task-pilot/metrology-headroom-test.mjs',
  'research/public-task-pilot/metrology-headroom-verify.mjs', 'research/research-checkpoint-2026-10-04.mjs',
  'docs/experiments/mmm-switch-v43-normal-results.md', 'docs/experiments/mmm-eager-load-v44-results.md',
  'docs/experiments/mmm-eager-load-v44-stream-repair.md', 'docs/experiments/mmm-eager-load-v44-owner-time-repair.md',
  'docs/experiments/scifact-source-preparation-results.md',
  'docs/experiments/research-checkpoint-2026-10-04-switch-load-public.md', 'research-direction.md'];
for (const p of extras) files[p] = hash(fs.readFileSync(p));
function unchanged() { for (const [p, h] of Object.entries(files)) assert.equal(hash(fs.readFileSync(p)), h, p); }
unchanged(); assert(!fs.existsSync(root)); fs.mkdirSync(root, {mode: 0o700});
function save(p, b) { fs.mkdirSync(path.dirname(p), {recursive: true, mode: 0o700}); fs.writeFileSync(p, b, {flag: 'wx', mode: 0o600}); }
for (const [p, h] of Object.entries(files)) {
  const b = fs.readFileSync(p); assert.equal(hash(b), h); save(root + '/source/' + p, b);
  assert.equal(hash(fs.readFileSync(root + '/source/' + p)), h);
}
const runs = ['research/switch-v43-study-normal', 'research/eager-load-v44',
  'research/eager-load-v44-audit-stream-repair', 'research/eager-load-v44-preflight-owner-time-stream-repair',
  'research/eager-load-v44-owner-time', 'research/public-task-pilot/scifact-v1/cache-preflight'];
const records = {};
for (const run of runs) for (const name of fs.readdirSync(run)) {
  const p = run + '/' + name;
  if (fs.statSync(p).isFile() && /\.(json|log)$/.test(name)) {
    const b = fs.readFileSync(p); records[p] = hash(b); save(root + '/records/' + p, b);
  }
}
const publicRoot = 'research/public-task-pilot/scifact-v1';
for (const name of ['source.json', 'corpus.json', 'queries.json', 'labels-train.json', 'labels-test.json', 'split-plan.json', 'split-audit.json']) {
  const p = publicRoot + '/' + name, b = fs.readFileSync(p); records[p] = hash(b); save(root + '/records/' + p, b);
}
for (const name of ['headroom.json', 'headroom-verified.json', 'HEADROOM_RESULTS.md']) {
  const p = 'research/public-task-pilot/metrology-transfer-v1/' + name, b = fs.readFileSync(p);
  records[p] = hash(b); save(root + '/records/' + p, b);
}
const historical = {};
for (const run of ['research/eager-load-v44', 'research/eager-load-v44-audit-stream-repair']) {
  const old = JSON.parse(fs.readFileSync(run + '/freeze.json'));
  for (const [p, h] of Object.entries(old.files)) if (files[p] !== h) {
    const original = run + '/source/' + p;
    // Recovery snapshots only its changed audit files. Unchanged inherited
    // files remain in the original measured snapshot, not invented copies.
    const location = fs.existsSync(original) ? original : 'research/eager-load-v44/source/' + p;
    const b = fs.readFileSync(location); assert.equal(hash(b), h, 'historical source authority');
    const target = root + '/historical/' + run + '/' + p; save(target, b); historical[location] = h;
  }
}
const rawPaths = ['research/switch-v43-study-normal/design-fixture.jsonl', 'research/switch-v43-study-normal/design.jsonl',
  'research/switch-v43-study-normal/confirmation-fixture.jsonl', 'research/switch-v43-study-normal/confirmation.jsonl',
  'research/eager-load-v44/raw.ndjson', 'research/eager-load-v44-owner-time/raw.ndjson', publicRoot + '/source.zip'];
const raw = {};
for (const p of rawPaths) raw[p] = {bytes: fs.statSync(p).size, sha256: await fileHash(p)};
const switchDone = JSON.parse(fs.readFileSync('research/switch-v43-study-normal/completed.json'));
for (const p of rawPaths.slice(0, 4)) assert.equal(raw[p].sha256, switchDone.artifacts[path.basename(p)]);
const loadDone = JSON.parse(fs.readFileSync('research/eager-load-v44-owner-time/completed.json'));
assert.equal(raw['research/eager-load-v44-owner-time/raw.ndjson'].sha256, loadDone.rawSHA256);
const verdict = JSON.parse(fs.readFileSync('research/switch-v43-study-normal/policy-verdict.json'));
assert.equal(verdict.qualityAdoption, false); assert.equal(loadDone.passedFiniteTrials, 0);
unchanged();
save(root + '/manifest.json', Buffer.from(JSON.stringify({time: new Date().toISOString(), files, records, historical, raw,
  frozenSourcesVerified: true, copiedCurrentSources: Object.keys(files).length, noPrivateCorporaCopied: true,
  noDeploymentOrPush: true, weeklyUsage: usage, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE',
  negativeResultsPreserved: true, v43Adoption: false, v44PassedFiniteTrials: 0,
  next: 'Prospective context-local routing, durable queue-critical-path profiling, public task formulation; never tune consumed confirmation.'}, null, 2) + '\n')));
console.log(JSON.stringify({root, copiedCurrentSources: Object.keys(files).length, historicalVersions: Object.keys(historical).length,
  rawArtifactsHashed: rawPaths.length, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'}));
