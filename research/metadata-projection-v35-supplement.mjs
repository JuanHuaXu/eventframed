// Post-hoc supplement only. Preserve the frozen prime-target failure and correct
// its corruption target to a delivered Recall with a recorded nomination oracle.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import vm from 'node:vm';
import {execFileSync} from 'node:child_process';
const dir = path.resolve('research/metadata-projection-v35');
const hash = x => crypto.createHash('sha256').update(x).digest('hex');
const need = (ok, why) => {if (!ok) throw Error(why)};
const verify = () => {
  const freeze = JSON.parse(fs.readFileSync(path.join(dir, 'freeze.json')));
  for (const [p, h] of Object.entries(freeze.files)) need(hash(fs.readFileSync(p)) === h, 'changed source ' + p);
  return freeze;
};
const freeze = verify();
let originalCode = 0, originalLog;
try {
  originalLog = execFileSync('node', ['research/metadata-projection-v35-audit.mjs'], {encoding: 'utf8', timeout: 120000, maxBuffer: 4*1024*1024});
} catch (e) {
  originalCode = e.status ?? -1;
  originalLog = (e.stdout ?? '') + '\n' + (e.stderr ?? '');
}
const originalPath = path.join(dir, 'audit-attempt.log');
if (fs.existsSync(originalPath)) need(hash(fs.readFileSync(originalPath)) === hash(originalLog), 'original failure log changed');
else fs.writeFileSync(originalPath, originalLog, {flag: 'wx'});
need(originalCode === 1 && originalLog.includes('corruption escaped trace-key'), 'original failure not reproduced');
const source = fs.readFileSync('research/metadata-projection-v35-audit.mjs', 'utf8');
const start = source.indexOf('const dir='), end = source.indexOf('\nfunction below(');
need(start >= 0 && end > start, 'checker extraction boundary');
const cs = source.indexOf('for(const[name,mutate]of['), ce = source.indexOf(']){const row=', cs);
need(cs >= 0 && ce > cs, 'controls extraction boundary');
const before = "r.MetadataReads.find(t=>t.Getters.length>2).Getters[2].Key='unknown'";
const after = "r.MetadataReads.find(t=>r.Reads.some(x=>x.Journal===t.Journal)&&t.Getters.length>2).Getters[2].Key='unknown'";
let list = source.slice(cs + 'for(const[name,mutate]of'.length, ce + 1);
need(list.split(before).length === 2, 'non-unique target repair');
list = list.replace(before, after);
const recode = 'const recode=(p,change)=>{const s=JSON.parse(p.Encoded);change(s);p.Encoded=JSON.stringify(s);p.SHA256=hash(p.Encoded)};';
const {check, mutations} = vm.runInNewContext(source.slice(start, end) + '\n' + recode + '\n({check,mutations:' + list + '})', {fs, path, crypto, vm, process, Buffer}, {timeout: 1000});
const run = JSON.parse(fs.readFileSync(path.join(dir, 'run.json')));
need(run.code === 0 && run.raw_sha256 === hash(fs.readFileSync(path.join(dir, 'raw.ndjson'))), 'raw/run mismatch');
const summaries = []; let header = false, footer = false, first;
for await (const line of readline.createInterface({input: fs.createReadStream(path.join(dir, 'raw.ndjson')), crlfDelay: Infinity})) {
  if (!line) continue;
  const row = JSON.parse(line);
  if (row.Type === 'header') {
    need(!header && row.Trials === 16 && row.FreezeSHA256 === hash(fs.readFileSync(path.join(dir, 'freeze.json'))), 'header');
    header = true; continue;
  }
  if (row.Type === 'footer') {need(!footer && summaries.length === 16 && row.Trials === 16, 'footer'); footer = true; continue;}
  need(header && !footer, 'envelope');
  summaries.push(check(row));
  if (row.Projected && row.Archived && !row.Visible && !first) first = row;
}
need(header && footer && summaries.length === 16 && new Set(summaries.map(r => [r.trial,r.mode,r.projected,r.visible].join('/'))).size === 16, 'complete factorial');
const originalTarget = first.MetadataReads.find(t => t.Getters.length > 2);
need(!first.Reads.some(r => r.Journal === originalTarget.Journal), 'prime-target diagnosis wrong');
const controls = [];
for (const [name, mutate] of mutations) {
  const row = structuredClone(first); mutate(row);
  let rejected = false; try {check(row)} catch {rejected = true}
  need(rejected, 'corruption escaped ' + name); controls.push(name);
}
verify();
const report = {
  type: 'post-hoc supplemental projection audit', time: new Date().toISOString(),
  original_audit_exit: originalCode, original_audit_log_sha256: hash(originalLog),
  original_control_error: 'trace-key targeted setup prime, whose full nomination was not saved',
  scope_limit: 'prime key nomination and independent prime acknowledgment remain unverified; all 128 delivered Recalls per trial have saved nomination and acknowledgment checks',
  source_files: Object.keys(freeze.files).length, raw_sha256: run.raw_sha256,
  freeze_sha256: hash(fs.readFileSync(path.join(dir, 'freeze.json'))),
  supplement_source_sha256: hash(fs.readFileSync('research/metadata-projection-v35-supplement.mjs')),
  controls, trials: summaries, all_seven_goals: 'OPEN'
};
fs.writeFileSync(path.join(dir, 'supplement.json'), JSON.stringify(report, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify({...report, trials: summaries.map(r => ({...r, freshness: {used: r.freshness.filter(x => x.used).length, censored: r.freshness.filter(x => !x.used).length, max_ns: Math.max(0, ...r.freshness.filter(x => x.used).map(x => x.first_use_ns))}}))}, null, 2));
