import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const output=process.argv[2]??'research/public-task-pilot/partition-load-results.json';
const d=JSON.parse(fs.readFileSync(output));assert.equal(d.Dimension,768);
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
assert.equal(d.Arms.length,6);const seen=new Set();
for(const a of d.Arms){
 const key=[a.N,a.GapMS,a.Repeat].join('/');assert(!seen.has(key));seen.add(key);
 assert([800,3200,6400].includes(a.N)&&a.GapMS===5&&[0,1].includes(a.Repeat));
 assert.deepEqual(JSON.parse(fs.readFileSync(`${output}.arm-n${a.N}-gap${a.GapMS}-r${a.Repeat}.json`)),a);
 assert.equal(a.Reads.length,1024);assert.equal(a.Writes.length,512);
 if('RetirementDrainNS' in a)assert(Number.isFinite(a.RetirementDrainNS)&&a.RetirementDrainNS>0);
 assert([...a.Reads,...a.Writes].every(s=>s.NS>0));
 const re=a.Reads.filter(s=>s.Error).length,we=a.Writes.filter(s=>s.Error).length;
 const late=[...a.Reads,...a.Writes].filter(s=>s.NS>100e6).length;
 const misses=a.Reads.filter(s=>!s.Error&&!s.Hit).length;
 assert.equal(a.Acknowledged,512-we);
 const bs=a.Builds??[],be=bs.filter(b=>b.Error).length,audit=a.AuditErrors??[];
 let cleared=0;
 for(const b of bs){
  assert(b.Partition>=0&&b.Partition<8);assert(b.BeforeTotal>=32&&b.BeforeTotal<=64);
  assert(b.BeforeLocal>0&&b.BeforeLocal<=b.BeforeTotal);
  assert(b.AfterTotal>=0&&b.AfterTotal<=64);assert(b.AfterLocal>=0&&b.AfterLocal<=b.AfterTotal);
  assert(b.AfterRevision>=b.BeforeRevision);
  // Every acknowledged mutation is one fresh append in THIS frozen workload.
  // No other compactor runs. Revision advance therefore counts intervening adds.
  const removed=b.BeforeTotal+b.AfterRevision-b.BeforeRevision-b.AfterTotal;
  assert(removed>=0&&removed<=64);cleared+=removed;
 }
 assert(cleared<=a.Acknowledged);
 const pass=!(re||we||late||misses||be||audit.length||a.FailedPresent)&&a.Reopened===a.Acknowledged;
 if(!audit.length)assert.equal(a.DurableRevision,1+a.Acknowledged);
 console.log(JSON.stringify({key,pass,readErrors:re,writeErrors:we,late,misses,buildErrors:be,ack:a.Acknowledged,reopened:a.Reopened,cleared,builds:bs.length}));
}
console.log('All six arms, sidecars, hashes and 9216 operations accounted for.');
