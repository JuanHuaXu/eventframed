import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const read=p=>JSON.parse(fs.readFileSync(new URL(p,import.meta.url)));
const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const summaries=[];
for(const set of ['w3c-focus-v1','esa-date-v1','esa-iso-v1']) {
  const d=read(`./${set}/normalized-results.json`),old=read(`./${set}/date-constraint-results.json`),src=read(`./${set}/results.json`);
  const oracle=read(`./${set}/oracle.json`),queries=read(`./${set}/queries.json`);
  for(const artifact of [d,old])for(const [p,h]of Object.entries(artifact.hashes))assert.equal(hash(path.join(import.meta.dirname,p)),h);
  for(const [p,h]of Object.entries(src.Hashes))assert.equal(hash(path.resolve(import.meta.dirname,'../..',p)),h);
  assert.equal(d.records.length,10);assert.equal(new Set(d.records.map(r=>r.case)).size,10);
  const rows=queries.map(q=>{
    const r=d.records.find(r=>r.case===q.case_id),previous=old.records.find(r=>r.case===q.case_id);
    const focus=src.Results.find(r=>r.Arm==='focus'&&r.Case===q.case_id);assert(r&&previous&&focus);
    assert.equal(r.original,q.question);assert.deepEqual(r.rules,previous.rules);
    assert.deepEqual(r.candidates.map(c=>c.Fixture).sort(),focus.Candidates.map(c=>c.Fixture).sort());
    assert.equal(r.supportRank,r.candidates.findIndex(c=>oracle[q.case_id].support.includes(c.Fixture))+1);
    for(const c of r.candidates){const f=focus.Candidates.find(v=>v.Fixture===c.Fixture);assert.deepEqual(c.Law,f.Law);assert.equal(c.Score,f.Score);}
    if(set!=='esa-iso-v1')assert.deepEqual(r,previous);
    return {case:q.case_id,supported:oracle[q.case_id].support.length>0,old:previous.supportRank,normalized:r.supportRank};
  });
  const positive=rows.filter(r=>r.supported);assert.equal(positive.length,8);
  const before=positive.filter(r=>r.old===1).length,after=positive.filter(r=>r.normalized===1).length;
  const preserved=positive.every(r=>r.old!==1||r.normalized===1);
  summaries.push({set,before,after,preserved,rows});
}
const iso=summaries.find(s=>s.set==='esa-iso-v1');
const exclusions=iso.rows.filter(r=>r.case.startsWith('exclude-')).every(r=>r.normalized===1);
console.log(JSON.stringify({summaries,exclusions,passed:iso.after>iso.before&&summaries.every(s=>s.preserved)&&exclusions,
  scope:'consumed-data normalization rescue; unchanged query grammar; no runtime integration'},null,2));
