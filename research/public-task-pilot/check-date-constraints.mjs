import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const read=p=>JSON.parse(fs.readFileSync(new URL(p,import.meta.url)));
const d=read('./w3c-focus-v1/date-constraint-results.json');
const src=read('./w3c-focus-v1/results.json');
const oracle=read('./w3c-focus-v1/oracle.json');
for(const [p,h]of Object.entries(d.hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(import.meta.dirname,p))).digest('hex'),h);
assert.equal(d.records.length,10);assert.equal(new Set(d.records.map(r=>r.case)).size,10);
const rows=d.records.map(r=>{
  const f=src.Results.find(s=>s.Case===r.case&&s.Arm==='focus');assert(f);
  assert.equal(r.original,f.Original);
  assert.deepEqual(r.candidates.map(c=>c.Fixture).sort(),f.Candidates.map(c=>c.Fixture).sort());
  assert.equal(r.supportRank,r.candidates.findIndex(c=>oracle[r.case].support.includes(c.Fixture))+1);
  for(const c of r.candidates){const old=f.Candidates.find(o=>o.Fixture===c.Fixture);assert.deepEqual(c.Law,old.Law);assert.equal(c.Score,old.Score);}
  return {case:r.case,supported:oracle[r.case].support.length>0,focus:f.SupportRank,constraint:r.supportRank,rules:r.rules,allContradicted:r.allContradicted};
});
// Explicit independent expected partitions for the two known exclusion dates.
for(const [id,excluded]of [['exclude-september','html51-proposed'],['exclude-november','html51-rec']]) {
  const r=d.records.find(r=>r.case===id);
  assert.deepEqual(r.candidates.filter(c=>c.Contradiction).map(c=>c.Fixture),[excluded]);
  assert.equal(r.candidates.at(-1).Fixture,excluded);
}
for(const [id,compatible]of [['before-2016','html5-rec'],['after-2016','html52-rec']]) {
  const r=d.records.find(r=>r.case===id);
  assert.deepEqual(r.candidates.filter(c=>!c.Contradiction).map(c=>c.Fixture),[compatible]);
}
const positives=rows.filter(r=>r.supported);
const base=positives.filter(r=>r.focus===1).length, improved=positives.filter(r=>r.constraint===1).length;
const preserved=positives.every(r=>r.focus!==1||r.constraint===1);
const exclusions=['exclude-september','exclude-november'].every(id=>rows.find(r=>r.case===id).constraint===1);
console.log(JSON.stringify({rows,base,improved,total:8,preserved,exclusions,
  passed:improved>base&&preserved&&exclusions,lawsPreserved:true,scope:d.scope},null,2));
