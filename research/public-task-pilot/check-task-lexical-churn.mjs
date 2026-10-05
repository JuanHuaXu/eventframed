import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../..');
const d=JSON.parse(fs.readFileSync(path.join(import.meta.dirname,'task-lexical-churn-results.json')));
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex'),h);
assert.equal(d.Arms.length,16);
const rows=[];let reads=0,writes=0,errors=0,writeErrors=0,futureLeaks=0;
for(const a of d.Arms){
 assert.equal(a.Samples.length,32);assert.equal(a.Writes.length,16);reads+=32;writes+=16;
 const times=a.Samples.map(s=>s.NS).sort((a,b)=>a-b);
 const failures=a.Samples.filter(s=>s.Error);errors+=failures.length;
 writeErrors+=a.Writes.filter(w=>w.Error).length;
 for(const s of a.Samples){futureLeaks+=s.FutureRecords;assert(s.NS>0);if(!s.Error){assert(s.JournalMatched);assert.equal(s.Nominated,a.N);}}
 const ms=x=>Number((x/1e6).toFixed(3));
 rows.push({repeat:a.Repeat,n:a.N,mode:a.Mode,enabled:a.Enabled,p50ms:ms(times[15]),p95ms:ms(times[30]),maxms:ms(times[31]),attempts:a.JournalAttempts,stale:a.StaleRejections,overlapWrites:a.Writes.filter(w=>w.ActiveReads>0).length,errors:failures.map(s=>s.Error)});
}
console.log(JSON.stringify({reads,writes,errors,writeErrors,futureLeaks,rows,scope:'memory-store ingestion, no durable or production guarantee'},null,2));
