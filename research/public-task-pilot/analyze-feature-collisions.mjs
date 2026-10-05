import fs from 'node:fs';
import assert from 'node:assert/strict';
import path from 'node:path';
import crypto from 'node:crypto';
const d=JSON.parse(fs.readFileSync(new URL('./feature-collisions.json',import.meta.url)));
for(const [file,hash] of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',file))).digest('hex'),hash);
const cases=new Map();
for(const row of d.Rows){if(!cases.has(row.Case))cases.set(row.Case,[]);cases.get(row.Case).push(row);}
const results=[];
for(const [id,rows]of cases){
 assert.equal(new Set(rows.map(r=>r.Fixture)).size,rows.length);
 const positives=rows.filter(r=>r.Relevant);if(!positives.length)continue;assert.equal(positives.length,1);
 const p=positives[0];
 const bits=rows.filter(r=>r.Bits===p.Bits).length;
 const sparse=rows.filter(r=>JSON.stringify(r.Sparse)===JSON.stringify(p.Sparse)).length;
 results.push({id,candidates:rows.length,binarySupportBucket:bits,sparseSupportBucket:sparse,binaryTieUniformOracle:1/bits,sparseTieUniformOracle:1/sparse});
}
console.log(JSON.stringify({rows:d.Rows.length,cases:cases.size,positiveCases:results.length,results,binaryUniformOracleMean:results.reduce((s,r)=>s+r.binaryTieUniformOracle,0)/results.length,sparseUniformOracleMean:results.reduce((s,r)=>s+r.sparseTieUniformOracle,0)/results.length,scope:'per-query oracle bucket selection, uniform within ties; NOT generalization or actual rank accuracy'},null,2));
