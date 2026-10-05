// Isolated math/lifecycle and batch-cost evidence, not a cohort experiment.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
const root = 'research/paired-v57-unit-preflight';
const parent = 'research/checkpoint-2026-10-04-paired-v56/manifest.json';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
assert(!fs.existsSync(root));
assert.equal(hash(fs.readFileSync(parent)), 'd8e840f1ed8928303646c622b11b55a516a2ec2fda5e283a1aa98667ed36cca3');
const protectedFiles = JSON.parse(fs.readFileSync(parent)).trackedHashes;
const closureRaw = execFileSync('go', ['run', './cmd/research-go-list-closure', './internal/researchpairedvalue', './internal/researchpairedref'], { maxBuffer: 128*1024*1024 });
const closure = JSON.parse(closureRaw), inputs = new Set();
for (const row of closure) {
  if (!row.Dir?.startsWith(process.cwd() + '/')) continue;
  for (const key of ['GoFiles', 'CgoFiles', 'EmbedFiles']) for (const name of row[key] ?? []) {
    const p = path.resolve(row.Dir, name);
    if (p.startsWith(process.cwd() + '/')) inputs.add(path.relative(process.cwd(), p));
  }
}
const support = ['research/paired-v57-module-gen.mjs', 'research/paired-v57-module-generation.json', 'research/paired-v57-unit-preflight.mjs', 'research/paired-v57-prepatch.md', 'research/paired-v57-predictive-value-direction.md', 'research/paired-v57-predictive-value-identities.mjs', 'research/paired-v57-predictive-value-identities.json'];
const files = Object.fromEntries([...new Set([...inputs, ...support])].sort().map(p => [p, hash(fs.readFileSync(p))]));
function verify() {
  for (const [p, h] of Object.entries({ ...files, ...protectedFiles })) assert.equal(hash(fs.readFileSync(p)), h, p);
}
verify();
fs.mkdirSync(root, { mode: 0o700 });
for (const p of Object.keys(files)) {
  const dest = root + '/source/' + p;
  fs.mkdirSync(path.dirname(dest), { recursive: true, mode: 0o700 });
  fs.copyFileSync(p, dest, fs.constants.COPYFILE_EXCL);
  fs.chmodSync(dest, 0o600);
}
fs.writeFileSync(root + '/closure.json', JSON.stringify(closure, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
const toolchain = JSON.parse(execFileSync('go', ['env', '-json', 'GOVERSION', 'GOOS', 'GOARCH', 'GOROOT', 'GOEXPERIMENT'], { encoding: 'utf8' }));
fs.writeFileSync(root + '/freeze.json', JSON.stringify({ time: new Date().toISOString(), files, protectedFiles, compilerInputs: inputs.size, closureSHA256: hash(closureRaw), toolchain, scope: 'module math/race/vet and isolated 150-origin batch cost; no cohort or serving experiment' }, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
const env = { ...process.env };
for (const k of Object.keys(env)) if (k.startsWith('EVENTFRAME_')) delete env[k];
const commands = [];
for (const [name, args] of [
  ['race', ['test', '-race', './internal/researchpairedvalue', './internal/researchpairedref', '-count=1', '-v', '-timeout=15m']],
  ['vet', ['vet', './internal/researchpairedvalue', './internal/researchpairedref']],
  ['batch-cost', ['test', './internal/researchpairedvalue', '-run=^$', '-bench=^BenchmarkPredictionValueFrontier$', '-benchmem', '-benchtime=3x', '-count=3', '-timeout=15m']]
]) {
  verify();
  const start = new Date().toISOString();
  let output, exitCode = 0;
  try { output = execFileSync('go', args, { env, maxBuffer: 32*1024*1024 }); }
  catch (e) { output = Buffer.concat([e.stdout ?? Buffer.alloc(0), e.stderr ?? Buffer.alloc(0)]); exitCode = e.status ?? -1; }
  fs.writeFileSync(root + '/' + name + '.log', output, { flag: 'wx', mode: 0o600 });
  process.stdout.write(output);
  commands.push({ name, args, start, end: new Date().toISOString(), exitCode, logSHA256: hash(output) });
  verify();
  if (exitCode !== 0) {
    fs.writeFileSync(root + '/failure.json', JSON.stringify({ commands, sourceUnchanged: true }, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
    throw Error(name + ' failed');
  }
}
fs.writeFileSync(root + '/completed.json', JSON.stringify({ commands, compilerInputs: inputs.size, sourceUnchanged: true, fullExperimentRun: false, batchCostMeasured: true, servingLatencyEstablished: false, wholeGoals: 'OPEN', goal: 'ACTIVE' }, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify({ compilerInputs: inputs.size, cohortRun: false, wholeGoals: 'OPEN' }));
