import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/public-task-pilot/';
let calls=0;
for(const set of ['semantic-role-v1','w3c-focus-v1','esa-date-v1','esa-iso-v1','landing-transfer-v1']) {
 const old=JSON.parse(fs.readFileSync(root+set+'/lexical-service-results.json'));
 const fresh=JSON.parse(fs.readFileSync(root+set+'/incremental-service-results.json'));
 assert.equal(old.Model,fresh.Model);
 for(const [p,h]of Object.entries(fresh.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
 assert.equal(fresh.Results.length,old.Results.length);
 for(let i=0;i<old.Results.length;i++) {
  const a={...old.Results[i]},b={...fresh.Results[i]};
  delete a.RecallNS;delete b.RecallNS;
  assert(a.JournalMatched&&b.JournalMatched);
  if(a.Explanation) {a.Explanation=JSON.parse(a.Explanation);b.Explanation=JSON.parse(b.Explanation);}
  assert.deepEqual(b,a,`${set}/${i}`);calls++;
 }
 console.log(`${set}: ${fresh.Results.length} exact outputs`);
}
assert.equal(calls,108);
console.log('PASS:108 service outputs equal, excluding duration only');
