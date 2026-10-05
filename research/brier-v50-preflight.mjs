// Primitive correctness only. Run apart from scientific timed collection.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';

const root = 'research/brier-v50-preflight';
assert(!fs.existsSync(root), 'exclusive preflight root');
const names = [
  'internal/researchbrier/aggregate.go',
  'internal/researchbrier/aggregate_test.go',
  'docs/experiments/mmm-brier-v50-exploration.md',
  'research/brier-v50-preflight.mjs',
];
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const sources = Object.fromEntries(names.map(p => [p, hash(fs.readFileSync(p))]));
fs.mkdirSync(root, { mode: 0o700 });
function save(name, value) {
  const fd = fs.openSync(root + '/' + name, 'wx', 0o600);
  try { fs.writeFileSync(fd, JSON.stringify(value, null, 2) + '\n'); fs.fsyncSync(fd); }
  finally { fs.closeSync(fd); }
}
function unchanged() {
  for (const [p, digest] of Object.entries(sources)) {
    assert.equal(hash(fs.readFileSync(p)), digest, p);
    assert.equal(hash(fs.readFileSync(root + '/source/' + p)), digest, 'copy ' + p);
  }
}
for (const p of names) {
  const dest = root + '/source/' + p;
  fs.mkdirSync(path.dirname(dest), { recursive: true, mode: 0o700 });
  const fd = fs.openSync(dest, 'wx', 0o600);
  try { fs.writeFileSync(fd, fs.readFileSync(p)); fs.fsyncSync(fd); }
  finally { fs.closeSync(fd); }
}
save('freeze.json', { time: new Date().toISOString(), sources, primitiveOnly: true, qualityStudy: false });
const commands = [];
async function run(name, args) {
  unchanged();
  const start = new Date().toISOString(), fd = fs.openSync(root + '/' + name + '.log', 'wx', 0o600);
  const log = crypto.createHash('sha256');
  let code;
  try {
    code = await new Promise((resolve, reject) => {
      const child = spawn('go', args, { stdio: ['ignore', 'pipe', 'pipe'] });
      const timer = setTimeout(() => child.kill('SIGTERM'), 10 * 60 * 1000);
      child.on('error', err => { clearTimeout(timer); reject(err); });
      for (const stream of [child.stdout, child.stderr]) stream.on('data', b => {
        fs.writeSync(fd, b); log.update(b); process.stdout.write(b);
      });
      child.on('close', c => { clearTimeout(timer); resolve(c ?? -1); });
    });
  } finally { fs.fsyncSync(fd); fs.closeSync(fd); }
  const row = { name, args, start, end: new Date().toISOString(), code, logSHA256: log.digest('hex') };
  commands.push(row); save(name + '-command.json', row); unchanged();
  assert.equal(code, 0, name);
}
try {
  await run('race', ['test', '-race', './internal/researchbrier', '-v', '-count=1', '-timeout=10m']);
  await run('vet', ['vet', './internal/researchbrier']);
  save('completed.json', { time: new Date().toISOString(), commands, sourcesUnchanged: true,
    primitiveOnly: true, broadQualityUnproven: true, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' });
} catch (err) {
  save('failure.json', { time: new Date().toISOString(), commands, error: err.message });
  throw err;
}
