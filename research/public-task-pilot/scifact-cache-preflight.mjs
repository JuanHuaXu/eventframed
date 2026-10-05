// Correctness-only preflight. No embedding service or benchmark data is opened.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {v43TimedProcesses, v43AuditProcesses} from '../research-process-guards.mjs';

assert.equal(v43TimedProcesses().length, 0, 'do not contend with timed forecast collection');
assert(v43AuditProcesses().length > 0, 'explicit verified offline audit window');
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const root = 'research/public-task-pilot/scifact-v1/cache-preflight';
assert(!fs.existsSync(root), 'exclusive preflight artifact root');
const inherited = 'research/eager-load-v44-preflight-authority-selector-repair/freeze.json';
const original = JSON.parse(fs.readFileSync(inherited));
const extras = ['research/public-task-pilot/scifactcache/memo.go',
  'research/public-task-pilot/scifactcache/memo_test.go',
  'research/public-task-pilot/scifact-cache-preflight.mjs'];
const files = {...original.files};
for (const file of extras) files[file] = hash(fs.readFileSync(file));
function unchanged() { for (const [file, digest] of Object.entries(files)) assert.equal(hash(fs.readFileSync(file)), digest, file); }
unchanged(); fs.mkdirSync(root, {mode: 0o700});
function save(name, value) { fs.writeFileSync(root + '/' + name, JSON.stringify(value, null, 2) + '\n', {flag: 'wx', mode: 0o600}); }
for (const file of extras) {
  const target = root + '/source/' + file;
  fs.mkdirSync(path.dirname(target), {recursive: true, mode: 0o700});
  fs.writeFileSync(target, fs.readFileSync(file), {flag: 'wx', mode: 0o600});
}
save('freeze.json', {time: new Date().toISOString(), files, inheritedFreeze: inherited,
  inheritedFreezeSHA256: hash(fs.readFileSync(inherited)), runtimeChanges: false});
const checks = [];
for (const [name, args] of [
  ['race', ['test', './research/public-task-pilot/scifactcache', '-race', '-v', '-count=3']],
  ['vet', ['vet', './research/public-task-pilot/scifactcache']],
]) {
  assert.equal(v43TimedProcesses().length, 0); unchanged();
  const start = new Date().toISOString(), begin = performance.now(); let code = 0, log = '';
  try { log = execFileSync('go', args, {encoding: 'utf8', timeout: 120000, maxBuffer: 4 << 20}); }
  catch (error) { code = error.status ?? -1; log = (error.stdout ?? '') + '\n' + (error.stderr ?? '') + '\n' + error.message; }
  fs.writeFileSync(root + '/' + name + '.log', log, {flag: 'wx', mode: 0o600});
  const row = {name, command: 'go', args, start, end: new Date().toISOString(), code,
    wallMS: performance.now() - begin, logSHA256: hash(log)};
  checks.push(row); save(name + '-command.json', row); console.log(log);
  unchanged(); assert.equal(code, 0, name);
}
const log = fs.readFileSync(root + '/race.log', 'utf8');
for (const test of ['TestRolesOwnershipAndWarmCapacity', 'TestPayloadCancellationIdentityAndNonfinite', 'TestConcurrentSameKeyAndConfiguration']) {
  assert.equal(log.split('--- PASS: ' + test + ' (').length - 1, 3, 'every root executes three times');
}
unchanged(); save('completed.json', {time: new Date().toISOString(), checks, sourceUnchanged: true,
  executedRoots: 3, repetitions: 3, noModelRuns: true, noDataOpened: true, noPerformanceClaim: true,
  payloadBoundNotRSS: true, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'});
