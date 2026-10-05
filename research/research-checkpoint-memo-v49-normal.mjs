// Research checkpoint only: preserve exact sources, records and linked raw tapes.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const root = 'research/checkpoint-2026-10-04-memo-v49-normal';
assert(!fs.existsSync(root), 'exclusive checkpoint');
const usage = Number(process.env.EVENTFRAME_WEEKLY_USAGE);
assert(Number.isFinite(usage) && usage >= 0 && usage <= 80, 'fresh weekly usage');
const normal = 'research/memo-v49-normal';
const preflight = 'research/brier-v50-preflight';
const cost = 'research/brier-v50-cost';
const previous = 'research/checkpoint-2026-10-04-memo-v49/manifest.json';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function digest(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
const sources = {}, records = {}, artifacts = {};
for (const dir of [normal, preflight, cost]) {
  const freeze = JSON.parse(fs.readFileSync(dir + '/freeze.json'));
  const done = JSON.parse(fs.readFileSync(dir + '/completed.json'));
  const commands = done.checks ?? done.commands;
  assert(commands.every(x => x.code === 0));
  assert.equal(commands.length, dir === normal ? 11 : dir === preflight ? 2 : 3);
  assert(done.sourceUnchanged || done.sourcesUnchanged);
  for (const [p, h] of Object.entries(freeze.files ?? freeze.sources)) {
    assert.equal(hash(fs.readFileSync(p)), h, 'current ' + p);
    assert.equal(hash(fs.readFileSync(dir + '/source/' + p)), h, 'frozen ' + p);
    if (sources[p]) assert.equal(sources[p], h);
    sources[p] = h;
  }
  for (const c of commands) {
    assert.equal(await digest(dir + '/' + c.name + '.log'), c.logSHA256);
    assert.deepEqual(JSON.parse(fs.readFileSync(dir + '/' + c.name + '-command.json')), c);
  }
  for (const [p, h] of Object.entries(done.artifacts ?? {})) {
    assert.equal(await digest(dir + '/' + p), h, p);
    if (p.endsWith('.jsonl')) artifacts[dir + '/' + p] = { sha256: h, bytes: fs.statSync(dir + '/' + p).size };
  }
  for (const name of fs.readdirSync(dir)) if (/\.(json|log)$/.test(name)) {
    records[dir + '/' + name] = hash(fs.readFileSync(dir + '/' + name));
  }
}
const readback = JSON.parse(fs.readFileSync(normal + '/readback.json'));
assert(readback.sourceAndRawHashes && readback.independentMetricArithmetic && readback.stage === 'normal');
for (const s of Object.values(readback.summaries)) {
  assert.equal(s.worlds, 448); assert.equal(s.candidateArms, 4032);
  assert.equal(s.distinctLabels, 1075200); assert.equal(Object.keys(s.cells).length, 84);
}
assert(JSON.parse(fs.readFileSync(normal + '/artifact-audit.json')).sourceCopiesAndRawHashes);
assert.equal(JSON.parse(fs.readFileSync(preflight + '/artifact-audit.json')).executedTestRoots.length, 6);
assert(JSON.parse(fs.readFileSync(cost + '/artifact-audit.json')).sourceAndCommandHashes);
const samples = JSON.parse(fs.readFileSync(cost + '/completed.json')).samples;
assert.equal(samples.length, 9);
for (const p of ['research-direction.md', 'docs/experiments/mmm-memo-v49-normal-results.md',
  'docs/experiments/mmm-brier-v50-cost-results.md', 'research/research-checkpoint-memo-v49-normal.mjs']) {
  sources[p] = hash(fs.readFileSync(p));
}
for (const p of ['research/brier-v50-artifact-audit.mjs', 'research/memo-v49-normal-summary.mjs']) {
  sources[p] = hash(fs.readFileSync(p));
}
records[previous] = hash(fs.readFileSync(previous));
fs.mkdirSync(root, { mode: 0o700 });
function copy(p, folder, expected) {
  const b = fs.readFileSync(p); assert.equal(hash(b), expected);
  const dest = root + '/' + folder + '/' + p;
  fs.mkdirSync(path.dirname(dest), { recursive: true, mode: 0o700 });
  const fd = fs.openSync(dest, 'wx', 0o600);
  try { fs.writeFileSync(fd, b); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
  assert.equal(hash(fs.readFileSync(dest)), expected);
}
for (const [p, h] of Object.entries(sources)) copy(p, 'source', h);
for (const [p, h] of Object.entries(records)) copy(p, 'records', h);
for (const [p, x] of Object.entries(artifacts)) assert.equal(await digest(p), x.sha256, 'raw ' + p);
const manifest = { time: new Date().toISOString(), source: sources, records, artifacts,
  previousCheckpoint: { path: previous, sha256: records[previous] },
  previousResearchTurn: 'PROGRESS', thisTurn: 'PROGRESS', goal: 'ACTIVE', allSevenWholeGoals: 'OPEN',
  weeklyUsageAtLastReadPercent: usage, allRequiredManagedJobsTerminal: true,
  productionPrivateCorporaWhitepaperUntouched: true, pushOrDeploymentPerformed: false,
  qualityAdoption: readback.qualityAdoption,
  scope: 'V49 unchanged full design/confirmation with independent dense audit; V50 isolated Brier primitive correctness and standalone cost, no broad quality result',
  next: 'Fresh bounded Brier integration and public source-preserving retrieval; loaded serving and whole-goal requirements remain open' };
const fd = fs.openSync(root + '/manifest.json', 'wx', 0o600);
try { fs.writeFileSync(fd, JSON.stringify(manifest, null, 2) + '\n'); fs.fsyncSync(fd); }
finally { fs.closeSync(fd); }
console.log(JSON.stringify({ root, sources: Object.keys(sources).length, records: Object.keys(records).length,
  artifacts: Object.keys(artifacts).length, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' }, null, 2));
