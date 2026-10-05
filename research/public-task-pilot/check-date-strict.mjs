import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const summaries=[];
for(const set of ['w3c-focus-v1','esa-date-v1','esa-iso-v1']) {
  const read=name=>JSON.parse(fs.readFileSync(path.join(import.meta.dirname,set,name)));
  const strict=read('normalized-strict-results.json'),old=read('normalized-results.json');
  for(const artifact of [strict,old])for(const [p,h]of Object.entries(artifact.hashes))
    assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(import.meta.dirname,p))).digest('hex'),h);
  assert.deepEqual(strict.records,old.records);
  assert.equal(strict.records.length,10);
  summaries.push({set,identicalRecords:10});
}
console.log(JSON.stringify({summaries,scope:'strict boundary repair preserves all30 public replay records'},null,2));
