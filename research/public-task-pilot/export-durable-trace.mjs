import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const input='research/public-task-pilot/durable-trace-results.json';
const d=JSON.parse(fs.readFileSync(input));
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
assert.equal(d.Arms.length,2);
const targets=d.Arms.flatMap(a=>['sched','sync'].map(kind=>({trace:a.TraceFile,kind,out:a.TraceFile+'.'+kind+'.pprof'})));
for(const t of targets)assert(!fs.existsSync(t.out),'derivative exists: '+t.out);
for(const a of d.Arms){assert.equal(a.Samples.length,128);for(const r of a.Samples){assert.equal(r.Error,'');assert.equal(r.Nominated,200);assert(r.JournalMatched);}}
for(const t of targets){const bytes=execFileSync('go',['tool','trace','-pprof='+t.kind,t.trace],{maxBuffer:64*1024*1024});fs.writeFileSync(t.out,bytes,{flag:'wx',mode:0o600});console.log(t.out);}
console.log('Verified256 returns and source hashes; exported scheduling/synchronization profiles');
