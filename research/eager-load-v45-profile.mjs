// Same full workload, profiler overhead explicit; no serving adoption claim.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync} from 'node:child_process';
import {v43TimedProcesses, v43AuditProcesses, v43NormalRunners} from './research-process-guards.mjs';
const root = 'research/eager-load-v45-profile', hash = b => crypto.createHash('sha256').update(b).digest('hex');
assert.equal(v43TimedProcesses().length + v43AuditProcesses().length + v43NormalRunners().length, 0);
const prior = 'research/eager-load-v44-owner-time';
const done = JSON.parse(fs.readFileSync(prior + '/completed.json')); assert(done.sourceUnchanged && done.checks.every(c => c.code === 0));
const files = {...JSON.parse(fs.readFileSync(prior + '/freeze.json')).files};
for (const p of ['research/eager-load-v45-profile.mjs', 'docs/experiments/mmm-eager-load-v45-profile-protocol.md']) files[p] = hash(fs.readFileSync(p));
function unchanged() { for (const [p, h] of Object.entries(files)) assert.equal(hash(fs.readFileSync(p)), h, p); }
unchanged(); assert(!fs.existsSync(root)); fs.mkdirSync(root, {mode: 0o700});
function save(name, value) { fs.writeFileSync(root + '/' + name, JSON.stringify(value, null, 2) + '\n', {flag: 'wx', mode: 0o600}); }
for (const [p, h] of Object.entries(files)) {
  const target = root + '/source/' + p; fs.mkdirSync(path.dirname(target), {recursive: true, mode: 0o700});
  const b = fs.readFileSync(p); assert.equal(hash(b), h); fs.writeFileSync(target, b, {flag: 'wx', mode: 0o600});
}
save('freeze.json', {time: new Date().toISOString(), files, originalWorkloadUnchanged: true,
  profileOverheadNotLatencyAdoption: true, host: {cpu: os.cpus()[0]?.model, logical: os.cpus().length, memory: os.totalmem(), load: os.loadavg()}});
const checks = [];
function run(name, cmd, args, env = {}, timeout = 600000) {
  unchanged(); const start = new Date().toISOString(), begin = performance.now(); let code = 0, log = '';
  try { log = execFileSync(cmd, args, {encoding: 'utf8', env: {...process.env, ...env}, timeout, maxBuffer: 32 << 20}); }
  catch (e) { code = e.status ?? -1; log = (e.stdout ?? '') + '\n' + (e.stderr ?? '') + '\n' + e.message; }
  fs.writeFileSync(root + '/' + name + '.log', log, {flag: 'wx', mode: 0o600});
  const row = {name, command: cmd, args, env, start, end: new Date().toISOString(), code, wallMS: performance.now() - begin, logSHA256: hash(log)};
  checks.push(row); save(name + '-command.json', row); console.log(log); unchanged(); assert.equal(code, 0, name);
}
try {
  const binary = path.resolve(root + '/subject.test');
  run('load', 'go', ['test', './internal/store/libravdbstore', '-run', '^TestResearchEagerLoadV44$', '-v', '-count=1', '-timeout=5m',
    '-o', binary, '-cpuprofile=' + path.resolve(root + '/cpu.pprof'), '-blockprofile=' + path.resolve(root + '/block.pprof'), '-blockprofilerate=1',
    '-mutexprofile=' + path.resolve(root + '/mutex.pprof'), '-mutexprofilefraction=1', '-memprofile=' + path.resolve(root + '/alloc.pprof'), '-memprofilerate=32768'],
    {EVENTFRAME_EAGER_LOAD_V44_OUTPUT: path.resolve(root + '/raw.ndjson')});
  run('audit', 'node', ['research/eager-load-v44-audit.mjs', root + '/raw.ndjson']);
  const audit = JSON.parse(fs.readFileSync(root + '/audit.log')); assert.equal(audit.trials.length, 16); assert.equal(audit.controls.length, 52); save('audit.json', audit);
  for (const profile of ['cpu', 'block', 'mutex', 'alloc']) {
    for (const cumulative of [false, true]) {
      run(profile + (cumulative ? '-cum' : '-flat'), 'go', ['tool', 'pprof', '-top', ...(cumulative ? ['-cum'] : []), '-nodecount=35', binary, path.resolve(root + '/' + profile + '.pprof')]);
    }
  }
  run('cpu-recall', 'go', ['tool', 'pprof', '-top', '-cum', '-nodecount=35', '-focus=service.*Recall', binary, path.resolve(root + '/cpu.pprof')]);
  run('cpu-outcome', 'go', ['tool', 'pprof', '-top', '-cum', '-nodecount=35', '-focus=ApplyBayesianOutcome', binary, path.resolve(root + '/cpu.pprof')]);
  run('block-recall', 'go', ['tool', 'pprof', '-top', '-cum', '-nodecount=35', '-focus=service.*Recall', binary, path.resolve(root + '/block.pprof')]);
  unchanged(); save('completed.json', {time: new Date().toISOString(), checks, sourceUnchanged: true,
    profiledRawSHA256: audit.rawSHA256, profiles: Object.fromEntries(['cpu', 'block', 'mutex', 'alloc'].map(p => [p, hash(fs.readFileSync(root + '/' + p + '.pprof'))])),
    binarySHA256: hash(fs.readFileSync(binary)), originalWorkloadUnchanged: true, profileOverheadNotLatencyAdoption: true,
    allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'});
} catch (e) { save('failure.json', {time: new Date().toISOString(), error: e.message, checks, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'}); throw e; }
