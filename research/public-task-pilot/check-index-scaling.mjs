import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const d=JSON.parse(fs.readFileSync('research/public-task-pilot/index-scaling-results.json'));
assert.equal(d.status,0);assert(!d.signal&&!d.error);
for(const [p,h] of Object.entries(d.hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
const rows=[...d.stdout.matchAll(/^BenchmarkResearchIndexScaling\/n(\d+)\/flat(true|false)\/batch(\d+)-4\s+1\s+(\d+) ns\/op\s+(\d+) search-ns\/query\s+([\d.]+) self-recall@10\s+(\d+) B\/op\s+(\d+) allocs\/op$/gm)];
assert.equal(rows.length,24);
for(const n of [200,800,1600]) for(const flat of ['false','true']) for(const size of [1,16]) {
 const r=rows.filter(r=>+r[1]===n&&r[2]===flat&&+r[3]===size);assert.equal(r.length,2);
 for(const a of r) assert.equal(+a[6],1);
 console.log(JSON.stringify({n,flat:flat==='true',batch:size,meanWriteMS:(+r[0][4]+ +r[1][4])/2e6,meanSearchMS:(+r[0][5]+ +r[1][5])/2e6,writeRunsMS:r.map(x=>+x[4]/1e6)}));
}
console.log('PASS:24 observations and frozen source hashes. Self-vector recall is not semantic recall.');
