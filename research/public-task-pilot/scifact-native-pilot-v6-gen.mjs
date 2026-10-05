// Repair the owned launch boundary, not macOS protections or installed assets.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const source = 'research/public-task-pilot/scifact-native-pilot-v5-run.mjs';
const dest = 'research/public-task-pilot/scifact-native-pilot-v6-run.mjs';
const raw = fs.readFileSync(source, 'utf8');
const f = JSON.parse(fs.readFileSync('research/public-task-pilot/nativepool-v5/manifest.json'));
assert.equal(hash(raw), f.inputs[process.cwd() + '/' + source].sha256);
const oldConfig = 'research/public-task-pilot/scifact-native-pilot-v5-config.yaml';
const config = 'research/public-task-pilot/scifact-native-pilot-v6-config.yaml';
const c = fs.readFileSync(oldConfig, 'utf8').replaceAll('nativepool-v5', 'nativepool-v6');
let out = raw.replaceAll('nativepool-v5', 'nativepool-v6')
  .replaceAll('scifact-native-pilot-v5-config', 'scifact-native-pilot-v6-config')
  .replaceAll('scifact-native-pilot-v5-run', 'scifact-native-pilot-v6-run');
const pairs = [
  ["import fs from'node:fs';", "import{setPriority,getPriority}from'node:os';\nimport fs from'node:fs';"],
  ["const child=spawn('/usr/bin/nice',['-n','10',binary,'serve'],", "const child=spawn(binary,['serve'],"],
  ["command:['/usr/bin/nice','-n','10',binary,'serve']", "command:[binary,'serve'],requestedPriority:10,actualPriority:priority"],
];
for (const [a, b] of pairs) { assert.equal(out.split(a).length, 2); out = out.replace(a, b); }
const marker = 'const monitor=setInterval';
const priority = "let priority=null;try{setPriority(child.pid,10);priority=getPriority(child.pid);assert.equal(priority,10);}catch(e){watcherError='owned child priority failure: '+String(e);child.kill('SIGTERM');}\n";
assert.equal(out.split(marker).length, 2);
out = out.replace(marker, priority + marker);
let inverse = out.replace(priority + marker, marker);
for (const [a, b] of pairs.toReversed()) inverse = inverse.replace(b, a);
inverse = inverse.replaceAll('nativepool-v6', 'nativepool-v5')
  .replaceAll('scifact-native-pilot-v6-config', 'scifact-native-pilot-v5-config')
  .replaceAll('scifact-native-pilot-v6-run', 'scifact-native-pilot-v5-run');
assert.equal(inverse, raw);
fs.writeFileSync(config, c, { flag: 'wx', mode: 0o600 });
fs.writeFileSync(dest, out, { flag: 'wx', mode: 0o600 });
fs.writeFileSync('research/public-task-pilot/scifact-native-pilot-v6-generation.json', JSON.stringify({
  source, dest, sourceSHA256: hash(raw), destSHA256: hash(out), configSHA256: hash(c),
  completeInverseEquality: true, changed: 'direct owned daemon spawn, process-local priority 10 after lifecycle handlers',
  scientificClientBackendAndRssCeilingUnchanged: true,
  diagnosis: 'native-loader-probe-v1/result.json: direct retains DYLD_LIBRARY_PATH; nice wrapper drops it',
  falsifier: 'direct daemon still fails bare libllama lookup or leaves nonterminal owned job',
}, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log('runner inverse PASS; direct owned spawn and same priority only');
