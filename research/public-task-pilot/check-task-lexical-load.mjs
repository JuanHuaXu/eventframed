import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../..');
const d=JSON.parse(fs.readFileSync(path.join(import.meta.dirname,'task-lexical-load-results.json')));
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex'),h);
assert.equal(d.Arms.length,16);assert.equal(d.Procs,4);
const unique=new Set(),rows=[];let errors=0;
for(const a of d.Arms){
 const key=[a.Repeat,a.N,a.Workers,a.Enabled].join(':');assert(!unique.has(key));unique.add(key);
 assert.equal(a.Samples.length,32);
 const times=a.Samples.map(s=>s.NS).sort((a,b)=>a-b);
 for(const s of a.Samples){assert(Number.isSafeInteger(s.NS)&&s.NS>0);if(s.Error)errors++;else{assert.equal(s.Nominated,a.N);assert(s.JournalMatched);assert(s.Packed>0);assert.equal(s.ExplanationBytes>0,a.Enabled);}}
 const ms=x=>Number((x/1e6).toFixed(3));
 rows.push({repeat:a.Repeat,n:a.N,workers:a.Workers,enabled:a.Enabled,p50ms:ms(times[15]),p95ms:ms(times[30]),maxms:ms(times[31]),wallMs:ms(a.WallNS),allocatedMB:Number((a.AllocatedBytes/1e6).toFixed(2))});
}
console.log(JSON.stringify({go:d.Go,samples:512,errors,rows,scope:'closed-loop read-only memory-store diagnostic; no p99 guarantee'},null,2));
