import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const output=process.argv[2]??'research/public-task-pilot/partial-run-load-results.json';
const d=JSON.parse(fs.readFileSync(output));assert.equal(d.Dimension,768);
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
assert.equal(d.Arms.length,6);const seen=new Set();
for(const a of d.Arms){
 const key=[a.N,a.GapMS,a.Repeat].join('/');assert(!seen.has(key));seen.add(key);
 assert([800,3200,6400].includes(a.N)&&a.GapMS===5&&[0,1].includes(a.Repeat));
 assert.deepEqual(JSON.parse(fs.readFileSync(`${output}.arm-n${a.N}-gap${a.GapMS}-r${a.Repeat}.json`)),a);
 assert.equal(a.Reads.length,1024);assert.equal(a.Writes.length,512);
 assert(Number.isFinite(a.RetirementDrainNS)&&a.RetirementDrainNS>0);
 assert([...a.Reads,...a.Writes].every(s=>Number.isFinite(s.NS)&&s.NS>0));
 const re=a.Reads.filter(s=>s.Error).length,we=a.Writes.filter(s=>s.Error).length;
 const late=[...a.Reads,...a.Writes].filter(s=>s.NS>100e6).length;
 const misses=a.Reads.filter(s=>!s.Error&&!s.Hit).length;
 assert.equal(a.Acknowledged,512-we);
 const bs=a.Builds??[],be=bs.filter(b=>b.Error).length,audit=a.AuditErrors??[];
 let cleared=0,merges=0;
 for(const b of bs){
  assert(['flush','partial'].includes(b.Mode));assert(b.NS>0);
  assert(b.BeforeTotal>=0&&b.BeforeTotal<=64&&b.AfterTotal>=0&&b.AfterTotal<=64);
  assert.equal(b.BeforeLocal,b.BeforeTotal);assert.equal(b.AfterLocal,b.AfterTotal);
  assert(b.AfterRevision>=b.BeforeRevision);
  const removed=b.BeforeTotal+b.AfterRevision-b.BeforeRevision-b.AfterTotal;
  assert(removed>=0&&removed<=64);cleared+=removed;
  if(b.Mode==='flush'){
   assert(b.BeforeTotal>=32);assert([8,9].includes(b.BeforeRuns));
   if(!b.Error)assert.equal(b.AfterRuns,b.BeforeRuns+1);
  }else{
   assert.equal(b.BeforeRuns,10);
   if(!b.Error){merges++;assert.equal(b.AfterRuns,9);assert.equal(removed,0);}
  }
 }
 assert(cleared<=a.Acknowledged);
 const pass=!(re||we||late||misses||be||audit.length||a.FailedPresent)&&a.Reopened===a.Acknowledged&&merges>=2;
 if(!audit.length)assert.equal(a.DurableRevision,1+a.Acknowledged);
 console.log(JSON.stringify({key,pass,readErrors:re,writeErrors:we,late,misses,buildErrors:be,ack:a.Acknowledged,reopened:a.Reopened,cleared,merges}));
}
console.log('Six arms, sidecars, hashes and 9216 operations accounted for; partial merges drain zero delta entries.');
