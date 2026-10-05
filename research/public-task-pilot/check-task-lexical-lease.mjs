import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../..');
const d=JSON.parse(fs.readFileSync(path.join(import.meta.dirname,'task-lexical-lease-results.json')));
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex'),h);
assert.equal(d.Arms.length,32);
const rows=[];let leasedErrors=0,unleasedErrors=0,writeErrors=0,futureLeaks=0,leasedStale=0;
const ms=x=>Number((x/1e6).toFixed(3));
for(const a of d.Arms){
 assert.equal(a.Samples.length,32);assert.equal(a.Writes.length,16);
 const errors=a.Samples.filter(s=>s.Error).length;
 if(a.Lease){leasedErrors+=errors;leasedStale+=a.StaleRejections;}else unleasedErrors+=errors;
 writeErrors+=a.Writes.filter(s=>s.Error).length;
 for(const s of a.Samples){assert(s.NS>0);futureLeaks+=s.FutureRecords;if(!s.Error){assert(s.JournalMatched);assert(s.Nominated>=a.N&&s.Nominated<=a.N+(a.Mode==='in-window'?16:0));}}
 const durations=a.Samples.map(s=>s.NS).sort((a,b)=>a-b);
 rows.push({repeat:a.Repeat,n:a.N,mode:a.Mode,enabled:a.Enabled,lease:a.Lease,errors,stale:a.StaleRejections,p95ms:ms(durations[30]),readMaxMS:ms(durations[31]),readWaitMaxMS:ms(Math.max(...a.Samples.map(s=>s.WaitNS))),writeMaxMS:ms(Math.max(...a.Writes.map(s=>s.NS))),writeWaitMaxMS:ms(Math.max(...a.Writes.map(s=>s.WaitNS))),readWallMS:ms(a.ReadWallNS),totalWallMS:ms(a.WallNS)});
}
const leased=rows.filter(r=>r.lease);
console.log(JSON.stringify({reads:1024,writes:512,leasedErrors,unleasedErrors,writeErrors,futureLeaks,leasedStale,leasedReadMaxMS:Math.max(...leased.map(r=>r.readMaxMS)),leasedWriteMaxMS:Math.max(...leased.map(r=>r.writeMaxMS)),rows},null,2));
