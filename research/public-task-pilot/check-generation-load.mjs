import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const output=process.argv[2]??'research/public-task-pilot/generation-load-v2-results.json';
const d=JSON.parse(fs.readFileSync(output));
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
assert.equal(d.Arms.length,8);const seen=new Set();
for(const a of d.Arms){
 const key=[a.N,a.GapMS,a.Repeat].join('/');assert(!seen.has(key));seen.add(key);
 assert.deepEqual(JSON.parse(fs.readFileSync(`${output}.arm-n${a.N}-gap${a.GapMS}-r${a.Repeat}.json`)),a);
 assert([200,800].includes(a.N)&&[20,5].includes(a.GapMS)&&[0,1].includes(a.Repeat));
 assert.equal(a.Reads.length,256);assert.equal(a.Writes.length,128);
 assert([...a.Reads,...a.Writes].every(s=>s.NS>0));
 const re=a.Reads.filter(s=>s.Error).length,we=a.Writes.filter(s=>s.Error).length;
 const late=[...a.Reads,...a.Writes].filter(s=>s.NS>100e6).length;
 const misses=a.Reads.filter(s=>!s.Error&&!s.Hit).length;
 const errors={};for(const s of a.Writes)if(s.Error)errors[s.Error]=(errors[s.Error]??0)+1;
 assert.equal(a.Acknowledged,128-we);
 const buildErrors=(a.Builds??[]).filter(b=>b.Error).length;
 const audit=a.AuditErrors??[];
 const pass=!(re||we||late||misses||buildErrors||audit.length||a.FailedPresent)&&a.Reopened===a.Acknowledged;
 if(!audit.length)assert.equal(a.DurableRevision,1+a.Acknowledged);
 console.log(JSON.stringify({key,pass,readErrors:re,writeErrors:we,late,misses,buildErrors,audit,ack:a.Acknowledged,reopened:a.Reopened,failedPresent:a.FailedPresent,wallMS:a.WallNS/1e6,buildMS:(a.Builds??[]).map(b=>b.NS/1e6),readMaxMS:Math.max(...a.Reads.map(s=>s.NS))/1e6,writeMaxMS:Math.max(...a.Writes.map(s=>s.NS))/1e6,errors}));
}
console.log('All8 arms and3072 operation samples accounted for; pass is evaluated per arm.');
