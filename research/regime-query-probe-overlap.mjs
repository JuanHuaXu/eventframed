import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';

// Post-hoc covariate diagnostic only. No query Y, future Y, or Q enters these
// counts. They measure exact-input overlap, not posterior covariance or cause.
const summary = JSON.parse(fs.readFileSync(process.argv[2]));
const key = r => `${r.phase}:${r.case}:${r.index}:${r.schedule}`;
const decisions = new Map(summary.records.map(r => [key(r), r]));
const stream = fs.createReadStream(process.argv[3]), hash = crypto.createHash('sha256');
stream.on('data', b => hash.update(b));
const records = []; let header, seen = 0;
for await (const line of createInterface({input: stream, crlfDelay: Infinity})) {
  const s = JSON.parse(line); if (!header) { header = s; assert.equal(s.Version, 'soft-learners-v120'); continue; }
  seen++;
  const d = decisions.get(`${s.Phase}:${s.Case}:${s.Index}:${s.Schedule}`); assert(d);
  if (s.Schedule !== 1) continue;
  const probes = s.Steps.slice(153, 161).map(x => x.X), future = s.Steps.slice(161, 192).map(x => x.X);
  records.push({phase:s.Phase,case:s.Case,index:s.Index,
    values:d.selected.slice(1).map(j => {
      assert(j >= 152 && j < 160); const x = s.Steps[j].X;
      return {origin:j,positionInProbes:Number(j>=153),virtualExactRate:probes.filter(v=>v===x).length/8,
        futureExactRate:future.filter(v=>v===x).length/31};
    })});
}
assert.equal(seen,2688);assert.equal(records.length,1344);assert.equal(hash.digest('hex'),summary.inputSHA256);
const mean = xs => xs.reduce((a,b)=>a+b,0)/xs.length;
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++) {
  const rs=records.filter(r=>r.phase===phase&&r.case===c);assert.equal(rs.length,32);
  groups.push({phase,case:c,means:[0,1,2].map(a=>Object.fromEntries(['positionInProbes','virtualExactRate','futureExactRate'].map(k=>[k,mean(rs.map(r=>r.values[a][k]))])))});
}
const overall=[0,1,2].map(a=>Object.fromEntries(['positionInProbes','virtualExactRate','futureExactRate'].map(k=>[k,mean(records.map(r=>r.values[a][k]))])));
console.log(JSON.stringify({scope:'Post-hoc exact-input exposure diagnostic, not causal or posterior-covariance attribution',
  sourceSHA256:summary.inputSHA256,summarySHA256:crypto.createHash('sha256').update(fs.readFileSync(process.argv[2])).digest('hex'),
  scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),overall,groups,records}));
