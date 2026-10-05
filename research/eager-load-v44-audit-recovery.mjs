// Audit the original completed workload, never regenerate or shrink it.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {v43NormalRunners, v43AuditProcesses, v43TimedProcesses} from './research-process-guards.mjs';
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const original = 'research/eager-load-v44', root = 'research/eager-load-v44-audit-stream-repair';
assert.equal(v43NormalRunners().length + v43AuditProcesses().length + v43TimedProcesses().length, 0);
assert.equal(JSON.parse(fs.readFileSync(original + '/load-command.json')).code, 0);
assert.equal(JSON.parse(fs.readFileSync(original + '/audit-command.json')).code, 1);
const freeze = JSON.parse(fs.readFileSync(original + '/freeze.json')), files = {...freeze.files};
const auditor = 'research/eager-load-v44-audit.mjs';
const changed = {[auditor]: {measuredFreeze: files[auditor], auditRepair: hash(fs.readFileSync(auditor))}};
assert.equal(hash(fs.readFileSync(original + '/source/' + auditor)), files[auditor]);
files[auditor] = changed[auditor].auditRepair;
for (const p of ['research/eager-load-v44-stream.mjs', 'research/eager-load-v44-stream-test.mjs',
  'research/eager-load-v44-audit-recovery.mjs']) files[p] = hash(fs.readFileSync(p));
function unchanged() { for (const [p, h] of Object.entries(files)) assert.equal(hash(fs.readFileSync(p)), h, p); }
async function fileHash(p) { const d = crypto.createHash('sha256'); for await (const bytes of fs.createReadStream(p)) d.update(bytes); return d.digest('hex'); }
const raw = original + '/raw.ndjson', rawSHA256 = await fileHash(raw);
unchanged(); assert(!fs.existsSync(root)); fs.mkdirSync(root, {mode: 0o700});
function save(name, value) { fs.writeFileSync(root + '/' + name, JSON.stringify(value, null, 2) + '\n', {flag: 'wx', mode: 0o600}); }
for (const p of [auditor, 'research/eager-load-v44-stream.mjs', 'research/eager-load-v44-stream-test.mjs',
  'research/eager-load-v44-audit-recovery.mjs']) {
  const target = root + '/source/' + p; fs.mkdirSync(path.dirname(target), {recursive: true, mode: 0o700});
  fs.writeFileSync(target, fs.readFileSync(p), {flag: 'wx', mode: 0o600});
}
save('freeze.json', {time: new Date().toISOString(), original, originalFreezeSHA256: hash(fs.readFileSync(original + '/freeze.json')),
  raw, rawSHA256, files, changed, originalMeasuredSourcesNotRerun: true});
const checks = [];
try {
  for (const [name, args] of [['stream-controls', ['research/eager-load-v44-stream-test.mjs']],
    ['scientific-selftest', [auditor, '--self-test']], ['audit', [auditor, raw]]]) {
    unchanged(); const start = new Date().toISOString(), begin = performance.now(); let code = 0, log = '';
    try { log = execFileSync('node', args, {encoding: 'utf8', timeout: 600000, maxBuffer: 32 << 20}); }
    catch (e) { code = e.status ?? -1; log = (e.stdout ?? '') + '\n' + (e.stderr ?? '') + '\n' + e.message; }
    fs.writeFileSync(root + '/' + name + '.log', log, {flag: 'wx', mode: 0o600});
    const row = {name, command: 'node', args, start, end: new Date().toISOString(), code,
      wallMS: performance.now() - begin, logSHA256: hash(log)};
    checks.push(row); save(name + '-command.json', row); console.log(log); unchanged(); assert.equal(code, 0, name);
  }
  const report = JSON.parse(fs.readFileSync(root + '/audit.log'));
  assert.equal(report.rawSHA256, rawSHA256); assert.equal(report.trials.length, 16);
  assert.equal(await fileHash(raw), rawSHA256); unchanged(); save('audit.json', report);
  save('completed.json', {time: new Date().toISOString(), checks, sourceUnchanged: true, rawSHA256,
    rawUnchanged: true, noWorkloadRerun: true, passedFiniteTrials: report.trials.filter(t => t.pass).length,
    allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'});
} catch (e) {
  save('failure.json', {time: new Date().toISOString(), error: e.message, checks, rawSHA256,
    allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'}); throw e;
}
