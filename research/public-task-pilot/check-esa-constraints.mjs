import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {recordDate,contradiction,constraints} from './date-constraints.mjs';
const read=p=>JSON.parse(fs.readFileSync(new URL(p,import.meta.url)));
const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const text=p=>fs.readFileSync(new URL(p,import.meta.url),'utf8');
assert.equal(text('../../cmd/public-esa-date/main.go').split('func focusQuery(q string) string {')[1],text('../../cmd/public-contrast-focus/main.go').split('func focusQuery(q string) string {')[1]);
const previous=read('./w3c-focus-v1/date-constraint-results.json');
assert.equal(hash(path.join(import.meta.dirname,'date-constraints.mjs')),previous.hashes['date-constraints.mjs']);
const compound='Rosetta launched on 2 March 2004 and arrived at its comet on 6 August 2014.';
assert.equal(recordDate(compound),null);
assert(!contradiction(constraints('before 2010'),compound));
assert.equal(recordDate('Rosetta was the first mission to rendezvous with a comet.'),null);
const summaries=[];
for(const set of ['esa-date-v1','esa-iso-v1']) {
  const d=read(`./${set}/date-constraint-results.json`),src=read(`./${set}/results.json`);
  const oracle=read(`./${set}/oracle.json`),queries=read(`./${set}/queries.json`),corpus=read(`./${set}/corpus.json`);
  for(const [p,h]of Object.entries(d.hashes)) assert.equal(hash(path.join(import.meta.dirname,p)),h);
  for(const [p,h]of Object.entries(src.Hashes)) assert.equal(hash(path.resolve(import.meta.dirname,'../..',p)),h);
  assert.equal(src.Results.length,20);assert.equal(d.records.length,10);
  assert.equal(new Set(src.Results.map(r=>r.Arm+':'+r.Case)).size,20);
  assert.equal(new Set(d.records.map(r=>r.case)).size,10);
  const rows=queries.map(q=>{
    const b=src.Results.find(r=>r.Arm==='baseline'&&r.Case===q.case_id);
    const f=src.Results.find(r=>r.Arm==='focus'&&r.Case===q.case_id);
    const c=d.records.find(r=>r.case===q.case_id);assert(b&&f&&c);
    assert.equal(c.original,q.question);assert.equal(b.Effective,q.question);
    for(const [r,candidates,rank]of [[b,b.Candidates,b.SupportRank],[f,f.Candidates,f.SupportRank],[c,c.candidates,c.supportRank]]) {
      assert.deepEqual(candidates.map(c=>c.Fixture).sort(),corpus.map(c=>c.fixture_id).sort());
      assert.equal(rank,candidates.findIndex(v=>oracle[q.case_id].support.includes(v.Fixture))+1);
    }
    for(const row of c.candidates){const old=f.Candidates.find(v=>v.Fixture===row.Fixture);assert.deepEqual(row.Law,old.Law);assert.equal(row.Score,old.Score);}
    if(set==='esa-iso-v1'||['earlier','later','absent-2020','absent-cost'].includes(q.case_id))
      assert.deepEqual(c.candidates.map(v=>v.Fixture),f.Candidates.map(v=>v.Fixture));
    return {case:q.case_id,supported:oracle[q.case_id].support.length>0,baseline:b.SupportRank,focus:f.SupportRank,constraint:c.supportRank,rules:c.rules.length};
  });
  const positive=rows.filter(r=>r.supported);assert.equal(positive.length,8);
  const baseline=positive.filter(r=>r.baseline===1).length,focus=positive.filter(r=>r.focus===1).length,correct=positive.filter(r=>r.constraint===1).length;
  const preserved=positive.every(r=>r.focus!==1||r.constraint===1),exclusions=rows.filter(r=>r.case.startsWith('exclude-')).every(r=>r.constraint===1);
  summaries.push({set,baseline,focus,constraint:correct,total:8,preserved,exclusions,passed:correct>focus&&preserved&&exclusions,unknownDates:corpus.filter(c=>!recordDate(c.text)).length,rows});
}
console.log(JSON.stringify({summaries,ambiguityChecksPassed:true,goal5Complete:false},null,2));
