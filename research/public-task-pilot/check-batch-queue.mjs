import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const data=JSON.parse(fs.readFileSync(process.argv[2] ?? 'research/public-task-pilot/batch-queue-clean-results.json'));
assert.equal(data.Arms.length,32);
for(const [p,h] of Object.entries(data.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
const keys=new Set(), totals={};
let journals=0,writes=0;
for(const a of data.Arms) {
 const key=[a.Repeat,a.Mode,a.ReadIntervalMS,a.Lease,a.Enabled].join('/');
 assert(!keys.has(key)); keys.add(key);
 assert([0,1].includes(a.Repeat)&&['future','in-window'].includes(a.Mode)&&[5,20].includes(a.ReadIntervalMS));
 assert.equal(typeof a.Lease,'boolean'); assert.equal(typeof a.Enabled,'boolean');
 assert.equal(a.N,200); assert.equal(a.Workers,4);
 assert.equal(a.Samples.length,32); assert.equal(a.Writes.length,16);
 assert.equal(a.Warmups.length,2);
 assert(a.DatabasePath && fs.existsSync(a.DatabasePath));
 for(const r of [...a.Samples,...a.Writes]) {
  assert(r.NS>0&&r.DispatchNS>=0&&r.WaitNS>=r.DispatchNS&&r.NS>=r.WaitNS);
  if(!r.Entered) assert(r.Error);
 }
 for(const r of a.Samples.filter(r=>!r.Error)) {
  assert(r.JournalMatched); assert.equal(r.FutureRecords,0); assert(r.Packed>0);
  assert(r.Nominated>=200&&r.Nominated<=(a.Mode==='future'?200:216));
  if(a.Enabled) assert(r.ExplanationBytes>0);
 }
 const sizes=a.BatchSizes??[];
 assert(sizes.every(n=>Number.isInteger(n)&&n>=1&&n<=16));
 assert(sizes.reduce((x,y)=>x+y,0)<=16);
 const re=a.Samples.filter(r=>r.Error).length,we=a.Writes.filter(r=>r.Error).length;
 const late=[...a.Samples,...a.Writes].filter(r=>r.NS>100e6).length;
 const warm=a.Warmups.filter(r=>r.Error).length;
 const reopen=a.ReopenErrors??[];
 if(!reopen.length) { assert.equal(a.ReopenedJournals,32-re); assert.equal(a.ReopenedWrites,16-we); }
 journals+=a.ReopenedJournals; writes+=a.ReopenedWrites;
 const pass=!(re||we||late||warm||reopen.length||a.StaleRejections);
 const group=`gap${a.ReadIntervalMS}/lease${a.Lease}`;
 const t=totals[group]??={arms:0,passes:0,readErrors:0,writeErrors:0,late:0,stale:0,batches:{}};
 t.arms++; t.passes+=Number(pass); t.readErrors+=re; t.writeErrors+=we; t.late+=late; t.stale+=a.StaleRejections;
 for(const n of sizes) t.batches[n]=(t.batches[n]??0)+1;
 console.log(JSON.stringify({key,pass,re,we,late,warm,reopen,batches:sizes}));
}
console.log(JSON.stringify({totals,journals,writes,integrity:'PASS'}));
