// Full fresh diagnostic, unchanged quality scope. No production service access.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import { spawn } from 'node:child_process';
const root = 'research/scored-v51-diagnostic';
assert(!fs.existsSync(root), 'exclusive diagnostic');
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function digest(p) { const h = crypto.createHash('sha256'); for await (const b of fs.createReadStream(p)) h.update(b); return h.digest('hex'); }
const old = JSON.parse(fs.readFileSync('research/memo-v49-normal/freeze.json')).files;
for (const [p, h] of Object.entries(old)) assert.equal(hash(fs.readFileSync(p)), h, 'unchanged old source ' + p);
const packages = ['internal/researchswitch', 'internal/researchmoment', 'internal/researchdispersion',
  'internal/researchhybridref', 'internal/researchswitchref', 'internal/researchbrier', 'internal/researchscoreref'];
const added = packages.flatMap(dir => fs.readdirSync(dir).filter(p => p.endsWith('.go')).map(p => dir + '/' + p));
for (const name of ['research-scored-gen', 'research-scored-study-gen', 'research-scored-fixture-gen', 'research-scored-inverse']) added.push('cmd/' + name + '/main.go');
added.push('research/scored-v51-generation.json', 'research/scored-v51-study-generation.json',
  'research/scored-v51-fixture-generation.json', 'research/scored-v51-inverse-audit.json',
  'research/scored-v51-diagnostic.mjs', 'research/scored-v51-readback.mjs',
  'docs/experiments/mmm-scored-v51-protocol.md', 'go.mod');
const files = Object.fromEntries([...new Set([...Object.keys(old), ...added])].sort().map(p => [p, hash(fs.readFileSync(p))]));
const proof = JSON.parse(fs.readFileSync('research/scored-v51-inverse-audit.json'));
assert(proof.allInverseSyntaxEqual && proof.checks.length === 4);
for (const [p, h] of Object.entries(proof.sources)) assert.equal(hash(fs.readFileSync(p)), h);
fs.mkdirSync(root, { mode: 0o700 });
function save(name, value) {
  const fd = fs.openSync(root + '/' + name, 'wx', 0o600);
  try { fs.writeFileSync(fd, JSON.stringify(value, null, 2) + '\n'); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
}
for (const [p, h] of Object.entries(files)) {
  const b = fs.readFileSync(p), dest = root + '/source/' + p; assert.equal(hash(b), h);
  fs.mkdirSync(path.dirname(dest), { recursive: true, mode: 0o700 });
  const fd = fs.openSync(dest, 'wx', 0o600);
  try { fs.writeFileSync(fd, b); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
}
function unchanged() { for (const [p, h] of Object.entries(files)) assert.equal(hash(fs.readFileSync(p)), h, p); }
const styles = ['legacy', 'log_mean', 'log_strong', 'brier_mean', 'brier_strong'];
save('freeze.json', { time: new Date().toISOString(), stage: 'diagnostic', files, styles,
  seedBases: { diagnostic: 2026105107, design: 2026105109, confirmation: 2026105111 },
  host: { cpu: os.cpus()[0]?.model, logical: os.cpus().length, load: os.loadavg() },
  candidateCollectionSerialized: true, futureNormalGatesUnchanged: true, qualityAdoptionUnproven: true });
const checks = [];
async function run(name, args, env = {}) {
  unchanged(); const start = new Date().toISOString(), begin = performance.now();
  const fd = fs.openSync(root + '/' + name + '.log', 'wx', 0o600), log = crypto.createHash('sha256');
  console.log('START ' + name + ' ' + start); let code;
  try {
    code = await new Promise((resolve, reject) => {
      const child = spawn('go', args, { env: { ...process.env, ...env }, stdio: ['ignore', 'pipe', 'pipe'] });
      const timer = setTimeout(() => child.kill('SIGTERM'), 60 * 60 * 1000);
      child.on('error', e => { clearTimeout(timer); reject(e); });
      for (const stream of [child.stdout, child.stderr]) stream.on('data', b => { fs.writeSync(fd, b); log.update(b); process.stdout.write(b); });
      child.on('close', c => { clearTimeout(timer); resolve(c ?? -1); });
    });
  } finally { fs.fsyncSync(fd); fs.closeSync(fd); }
  const row = { name, args, env, start, end: new Date().toISOString(), code, wallMS: performance.now() - begin, logSHA256: log.digest('hex') };
  checks.push(row); save(name + '-command.json', row); unchanged(); assert.equal(code, 0, name); console.log('DONE ' + name);
}
const fixture = path.resolve(root + '/diagnostic-fixture.jsonl'), common = { EVENTFRAME_SCORED_V51_FIXTURE: fixture };
try {
  await run('race', ['test', '-race', './internal/researchswitch', '-run', '^(TestScoredV51(DenseDelayedReference|LegacyExactAndCancellation|LifecycleAndFutureFork|HybridLegacyExactAndFencing|StreamIdentity)|TestHybridV48(CompactEquivalent|CompactExtremeRevival|DelayedHeadsAndOriginalServedLaw|OwnershipCapsAndGlobalFence))$', '-v', '-count=1', '-timeout=10m']);
  await run('seed-race', ['test', '-race', './internal/researchdispersion', '-run', '^TestScoredV51SeedSeparation$', '-v', '-count=1', '-timeout=10m']);
  await run('vet', ['vet', ...packages.map(p => './' + p), './cmd/research-scored-gen', './cmd/research-scored-study-gen', './cmd/research-scored-fixture-gen', './cmd/research-scored-inverse']);
  await run('fixture', ['test', './internal/researchdispersion', '-run', '^TestScoredV51Fixture$', '-v', '-count=1', '-timeout=30m'], { ...common, EVENTFRAME_SCORED_V51_SPLIT: 'diagnostic' });
  assert(fs.statSync(fixture).size > 0);
  await run('fixture-checks', ['test', '-race', './internal/researchswitch', '-run', '^TestScoredV51FixtureChecks$', '-v', '-count=1', '-timeout=30m'], { ...common, EVENTFRAME_SCORED_V51_CHECKS: path.resolve(root + '/fixture-checks.json') });
  const f = JSON.parse(fs.readFileSync(root + '/fixture-checks.json')); assert.equal(f.futurePrefixCases, 36); assert.equal(f.semanticCorruptionsRejected, 68);
  await run('allocation', ['test', './internal/researchswitch', '-run', '^TestScoredV51Allocation$', '-v', '-count=1'], { ...common, EVENTFRAME_SCORED_V51_ALLOCATION: path.resolve(root + '/allocation.json') });
  for (const style of styles) {
    const raw = path.resolve(root + '/' + style + '.jsonl');
    await run(style, ['test', './internal/researchswitch', '-run', '^TestScoredV51Study$', '-v', '-count=1', '-timeout=30m'], { ...common, EVENTFRAME_SCORED_V51_STYLE: style, EVENTFRAME_SCORED_V51_OUT: raw });
    assert(fs.statSync(raw).size > 0);
    await run(style + '-audit', ['test', './internal/researchswitch', '-run', '^TestScoredV51StudyAudit$', '-v', '-count=1', '-timeout=30m'], { ...common, EVENTFRAME_SCORED_V51_STYLE: style, EVENTFRAME_SCORED_V51_AUDIT: raw });
  }
  unchanged(); const artifacts = {};
  for (const p of fs.readdirSync(root).filter(p => /\.(json|jsonl|log)$/.test(p))) artifacts[p] = await digest(root + '/' + p);
  save('completed.json', { time: new Date().toISOString(), checks, artifacts, sourceUnchanged: true,
    qualityAdoptionUnproven: true, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE', productionChanged: false });
} catch (e) {
  save('failure.json', { time: new Date().toISOString(), checks, error: e.message, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' });
  throw e;
}
