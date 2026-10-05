// Recover the interrupted diagnostic without changing its frozen model or evidence.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';

const root = 'research/prior-v40-diagnostic';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const freeze = JSON.parse(fs.readFileSync(root + '/freeze.json'));
assert.equal(freeze.stage, 'diagnostic');
assert(!fs.existsSync(root + '/failure.json'), 'inspect prior failure before recovery');
assert(!fs.existsSync(root + '/completed.json'), 'diagnostic already complete');
const save = (p, x) => fs.writeFileSync(root + '/' + p, JSON.stringify(x, null, 2) + '\n', {flag: 'wx', mode: 0o600});
function unchanged() {
  for (const [p, h] of Object.entries(freeze.files)) assert.equal(hash(fs.readFileSync(p)), h, 'source changed ' + p);
}
unchanged();
save('recovery.json', {
  time: new Date().toISOString(),
  missing_session: 63347,
  authoritative_process_check: 'pgrep found no prior-v40/TestPrior/researchdispersion.test/go test process',
  reason: 'interruption after completed model-race, before experimental outcomes',
  helper_sha256: hash(fs.readFileSync('research/prior-v40-resume.mjs')),
  frozen_sources_changed: false,
  gate_changes: false,
});
const checks = [];
function run(name, args, env = {}, timeout = 2100000) {
  unchanged();
  const command = root + '/' + name + '-command.json';
  if (fs.existsSync(command)) {
    const row = JSON.parse(fs.readFileSync(command));
    assert.equal(row.code, 0);
    assert.deepEqual(row.args, args);
    assert.deepEqual(row.env, env);
    assert.equal(hash(fs.readFileSync(root + '/' + name + '.log')), row.log_sha256);
    checks.push(row);
    console.log('Verified completed check:', name);
    return;
  }
  assert(!fs.existsSync(root + '/' + name + '.log'), 'unrecorded output requires inspection');
  const start = new Date().toISOString(), begin = performance.now();
  let code = 0, log = '';
  try {
    log = execFileSync('go', args, {encoding: 'utf8', env: {...process.env, ...env}, timeout, maxBuffer: 16 * 1024 * 1024});
  } catch (e) {
    code = e.status ?? -1;
    log = (e.stdout ?? '') + '\n' + (e.stderr ?? '') + '\n' + e.message;
  }
  fs.writeFileSync(root + '/' + name + '.log', log, {flag: 'wx', mode: 0o600});
  const row = {name, args, env, start, end: new Date().toISOString(), code, wall_ms: performance.now() - begin, log_sha256: hash(log)};
  save(name + '-command.json', row);
  checks.push(row);
  console.log(log);
  unchanged();
  if (code !== 0) throw Error(name + ' failed');
}
try {
  run('model-race', ['test', '-race', './internal/researchprior', './internal/researchpriorref', '-v', '-count=1']);
  run('integration-race', ['test', '-race', './internal/researchdispersion', '-run', '^TestPrior(AuditorNegative|FuturePrefix)V40$', '-v', '-count=1', '-timeout=10m']);
  run('vet', ['vet', './internal/researchprior', './internal/researchpriorref', './internal/researchdispersion']);
  run('benchmarks', ['test', './internal/researchprior', '-run', '^$', '-bench', '^BenchmarkPrior', '-benchmem', '-benchtime=50ms', '-count=3', '-timeout=5m']);
  const raw = path.resolve(root + '/diagnostic.jsonl');
  run('diagnostic', ['test', './internal/researchdispersion', '-run', '^TestPriorExperimentV40$', '-count=1', '-v', '-timeout=35m'], {EVENTFRAME_PRIOR_V40_OUT: raw, EVENTFRAME_PRIOR_V40_SPLIT: 'diagnostic'});
  run('diagnostic-audit', ['test', './internal/researchdispersion', '-run', '^TestPriorStudyAuditV40$', '-count=1', '-v', '-timeout=35m'], {EVENTFRAME_PRIOR_V40_AUDIT: raw, EVENTFRAME_PRIOR_V40_BENCHMARK: path.resolve(root + '/benchmarks.log')});
  const artifacts = {};
  for (const p of fs.readdirSync(root).filter(p => /\.(jsonl|json|log)$/.test(p))) {
    const h = crypto.createHash('sha256');
    for await (const b of fs.createReadStream(root + '/' + p)) h.update(b);
    artifacts[p] = h.digest('hex');
  }
  save('completed.json', {time: new Date().toISOString(), checks, artifacts, source_unchanged: true, stage: 'diagnostic', recovery: true, normal_cohorts_consumed: false, whole_goal_complete: false, goal: 'ACTIVE', all_seven_whole_goals: 'OPEN'});
} catch (e) {
  save('recovery-failure.json', {time: new Date().toISOString(), error: e.message, checks, whole_goal_complete: false});
  throw e;
}
