import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import os from 'node:os';
import {execFileSync} from 'node:child_process';

const dir = path.resolve('research/durable-witness-v23');
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const freezePath = path.join(dir, 'freeze.json');
const freeze = JSON.parse(fs.readFileSync(freezePath));
for (const [file, digest] of Object.entries(freeze.files)) {
  if (sha(fs.readFileSync(file)) !== digest) throw Error('source changed: ' + file);
}
const raw = path.join(dir, 'raw.ndjson');
if (fs.existsSync(raw) || fs.existsSync(path.join(dir,'run.json'))) throw Error('exclusive outputs already present');
const start = new Date().toISOString(), begin = performance.now();
let output, code = 0;
try {
  output = execFileSync('go', ['test','./internal/store/libravdbstore','-run','^TestResearchDurableWitnessLoadV23$','-count=1','-v','-timeout','5m'], {
    env:{...process.env,EVENTFRAME_DURABLE_WITNESS_V23_OUTPUT:raw}, encoding:'utf8', maxBuffer:16*1024*1024, timeout:310000,
  });
} catch (e) {
  code = e.status ?? -1; output = (e.stdout ?? '') + '\n' + (e.stderr ?? '') + '\n' + e.message;
}
fs.writeFileSync(path.join(dir,'load.log'),output,{flag:'wx'});
fs.writeFileSync(path.join(dir,'run.json'),JSON.stringify({start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,freeze_sha256:sha(fs.readFileSync(freezePath)),
  raw_sha256:fs.existsSync(raw)?sha(fs.readFileSync(raw)):null,log_sha256:sha(output),node:process.version,os:{platform:os.platform(),release:os.release(),arch:os.arch(),cpu:os.cpus()[0]?.model,logical_cpus:os.cpus().length,total_memory:os.totalmem()}},null,2)+'\n',{flag:'wx'});
console.log(output); process.exitCode = code ? 1 : 0;
