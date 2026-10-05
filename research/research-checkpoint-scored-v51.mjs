// Preserve exact research sources, evidence and negatives without publishing.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root = 'research/checkpoint-2026-10-04-scored-v51', study = 'research/scored-v51-diagnostic';
assert(!fs.existsSync(root), 'exclusive checkpoint');
const usage = Number(process.env.EVENTFRAME_WEEKLY_USAGE); assert(Number.isFinite(usage) && usage >= 0 && usage <= 80);
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function digest(p) { const h = crypto.createHash('sha256'); for await (const b of fs.createReadStream(p)) h.update(b); return h.digest('hex'); }
const freeze = JSON.parse(fs.readFileSync(study + '/freeze.json')), done = JSON.parse(fs.readFileSync(study + '/completed.json'));
const audit = JSON.parse(fs.readFileSync(study + '/artifact-audit.json')), readback = JSON.parse(fs.readFileSync(study + '/readback.json'));
assert(done.sourceUnchanged && done.checks.length === 16 && done.checks.every(x => x.code === 0));
assert(audit.sourceCopiesAndRawHashes && audit.executedRaceRoots.length === 9 && audit.weeklyUsage >= 0 && audit.weeklyUsage <= 80);
assert(readback.sourceAndRawHashes && readback.independentMetricArithmetic && !readback.qualityAdoption);
assert.equal(readback.worlds, 28); assert.equal(readback.exactLegacyPairs, 252); assert.equal(readback.exactSharedAdvicePairs, 1008);
const sources = { ...freeze.files }, records = {}, artifacts = {};
for (const [p, h] of Object.entries(freeze.files)) { assert.equal(hash(fs.readFileSync(p)), h, p); assert.equal(hash(fs.readFileSync(study + '/source/' + p)), h); }
for (const [p, h] of Object.entries(done.artifacts)) {
  assert.equal(await digest(study + '/' + p), h, p);
  if (p.endsWith('.jsonl')) artifacts[study + '/' + p] = { sha256: h, bytes: fs.statSync(study + '/' + p).size };
}
for (const c of done.checks) { assert.equal(await digest(study + '/' + c.name + '.log'), c.logSHA256); assert.deepEqual(JSON.parse(fs.readFileSync(study + '/' + c.name + '-command.json')), c); }
for (const dir of [study, 'research/scored-v51-first-preflight-failure']) {
  for (const p of fs.readdirSync(dir).filter(p => /\.(json|log)$/.test(p))) records[dir + '/' + p] = hash(fs.readFileSync(dir + '/' + p));
}
const bad = JSON.parse(fs.readFileSync('research/scored-v51-first-preflight-failure/failure.json')); assert.equal(bad.code, 1);
for (const [p, h] of Object.entries(bad.sources)) assert.equal(hash(fs.readFileSync('research/scored-v51-first-preflight-failure/source/' + p)), h);
for (const p of ['research-direction.md', 'docs/experiments/mmm-scored-v51-results.md', 'research/scored-v51-artifact-audit.mjs', 'research/research-checkpoint-scored-v51.mjs']) sources[p] = hash(fs.readFileSync(p));
const previous = 'research/checkpoint-2026-10-04-memo-v49-normal/manifest.json'; records[previous] = hash(fs.readFileSync(previous));
const correction = JSON.parse(fs.readFileSync(study + '/readback-usage-correction.json'));
assert.equal(correction.readbackSHA256, hash(fs.readFileSync(study + '/readback.json')));
assert.equal(correction.authoritativeFreshPercent, audit.weeklyUsage); assert.equal(correction.scientificFieldsChanged, false);
fs.mkdirSync(root, { mode: 0o700 });
function copy(p, folder, h) {
  const b = fs.readFileSync(p); assert.equal(hash(b), h);
  const dest = root + '/' + folder + '/' + p; fs.mkdirSync(path.dirname(dest), { recursive: true, mode: 0o700 });
  const fd = fs.openSync(dest, 'wx', 0o600); try { fs.writeFileSync(fd, b); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
  assert.equal(hash(fs.readFileSync(dest)), h);
}
for (const [p, h] of Object.entries(sources)) copy(p, 'source', h);
for (const [p, h] of Object.entries(records)) copy(p, 'records', h);
for (const [p, x] of Object.entries(artifacts)) assert.equal(await digest(p), x.sha256, 'linked raw ' + p);
const manifest = { time: new Date().toISOString(), source: sources, records, artifacts,
  preservedFailure: { path: 'research/scored-v51-first-preflight-failure', source: bad.sources },
  previousCheckpoint: { path: previous, sha256: records[previous] }, previousResearchTurn: 'PROGRESS', thisTurn: 'PROGRESS',
  goal: 'ACTIVE', allSevenWholeGoals: 'OPEN', weeklyUsageAtLastReadPercent: usage,
  allRequiredManagedJobsTerminal: true, productionPrivateCorporaWhitepaperUntouched: true, pushOrDeploymentPerformed: false,
  scope: 'V51 full fresh outer-head factorial and exact legacy compatibility; negative recovery/quality rescue evidence, no normal confirmation/adoption',
  next: 'Public source-preserving retrieval and loaded serving work; fresh all-head alignment can test the remaining loss hypothesis without changing V51 evidence or gates' };
const fd = fs.openSync(root + '/manifest.json', 'wx', 0o600); try { fs.writeFileSync(fd, JSON.stringify(manifest, null, 2) + '\n'); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
console.log(JSON.stringify({ root, sources: Object.keys(sources).length, records: Object.keys(records).length, artifacts: Object.keys(artifacts).length, goal: 'ACTIVE', allSevenWholeGoals: 'OPEN' }, null, 2));
