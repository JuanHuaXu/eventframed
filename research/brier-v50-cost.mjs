// Run only after V49 timed collection AND its required audits are terminal.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import { spawn } from 'node:child_process';

const root = 'research/brier-v50-cost';
const normal = JSON.parse(fs.readFileSync('research/memo-v49-normal/completed.json'));
assert(normal.sourceUnchanged && normal.checks.length === 11 && normal.checks.every(c => c.code === 0));
assert(!fs.existsSync(root), 'exclusive cost root');
const names = [
  'go.mod',
  'internal/researchbrier/aggregate.go',
  'internal/researchbrier/aggregate_test.go',
  'internal/researchbrier/benchmark_test.go',
  'docs/experiments/mmm-brier-v50-exploration.md',
  'docs/experiments/mmm-brier-v50-preflight-results.md',
  'research/brier-v50-cost.mjs',
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
save('freeze.json', { time: new Date().toISOString(), sources,
  host: { cpu: os.cpus()[0]?.model, logical: os.cpus().length, load: os.loadavg() },
  isolatedComponent: true, loadedServingStudy: false,
  normalCompletedSHA256: hash(fs.readFileSync('research/memo-v49-normal/completed.json')) });
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
  await run('bench', ['test', './internal/researchbrier', '-run', '^$', '-bench', '^Benchmark(Substitution3|Substitution8|ImmediateCycle3)$', '-benchmem', '-benchtime=1s', '-count=3']);
  const text = fs.readFileSync(root + '/bench.log', 'utf8');
  const pattern = /^(Benchmark\S+)-\d+\s+(\d+)\s+([\d.]+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/gm;
  const samples = [...text.matchAll(pattern)].map(m => ({ name: m[1], iterations: Number(m[2]),
    nsPerOp: Number(m[3]), bytesPerOp: Number(m[4]), allocationsPerOp: Number(m[5]) }));
  assert.equal(samples.length, 9, 'all three repeated benchmarks executed');
  for (const name of ['BenchmarkSubstitution3', 'BenchmarkSubstitution8', 'BenchmarkImmediateCycle3']) {
    const rows = samples.filter(x => x.name === name);
    assert.equal(rows.length, 3);
    assert(rows.every(x => x.iterations > 0 && Number.isFinite(x.nsPerOp) && x.nsPerOp > 0));
  }
  save('completed.json', { time: new Date().toISOString(), commands, sourcesUnchanged: true, samples,
    isolatedComponent: true, loadedServingStudy: false, broadQualityUnproven: true,
    allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' });
} catch (err) {
  save('failure.json', { time: new Date().toISOString(), commands, error: err.message });
  throw err;
}
