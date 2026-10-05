import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../..');
const read=p=>JSON.parse(fs.readFileSync(path.join(import.meta.dirname,p)));
const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const sources=read('priority-overlay-v2/sources.json');
const overlay=read('priority-overlay-v2/overlay.json');
for(const [p,h]of Object.entries(sources)) {
 assert.equal(hash(path.join(root,p)),h.original,'original runtime source changed');
 assert.equal(hash(overlay.Replace[path.join(root,p)]),h.overlay,'overlay changed');
}
const summaries=[];
// Equivalence against the earlier actual service, not just an offline helper.
for(const [set,minimum]of [['w3c-focus-v1',7],['esa-date-v1',8],['esa-iso-v1',7]]) {
 const data=read(`${set}/priority-fast-results.json`),oracle=read(`${set}/oracle.json`),corpus=read(`${set}/corpus.json`);
 for(const [p,h]of Object.entries(data.Hashes)) assert.equal(hash(path.join(root,p)),h);
 const previous=read(`${set}/priority-overlay-results.json`);
 for(const row of data.Results) {
 const old=previous.Results.find(x=>x.Arm===row.Arm&&x.Case===row.Case);assert(old);
 const clean=r=>Object.fromEntries(Object.entries(r).filter(([k])=>k!=='RecallNS'));
 assert.deepEqual(clean(row),clean(old),'full recorded result changed');
 }
 assert.equal(data.Results.length,20);
 assert.equal(new Set(data.Results.map(r=>r.Arm+':'+r.Case)).size,20);
 const ids=Object.fromEntries(corpus.map((c,i)=>[`public-record-${String(i).padStart(3,'0')}`,c.fixture_id]));
 let baseline=0,priority=0,rescues=0,regressions=0,absent=0;
 for(const p of data.Results.filter(r=>r.Arm==='priority')) {
  const f=data.Results.find(r=>r.Arm==='focus'&&r.Case===p.Case); assert(f);
  assert(p.JournalMatched&&f.JournalMatched);
  assert.equal(p.CalibrationStatus,'not_evaluated');assert.equal(p.PacketCertainty,0);
  assert.equal(f.CalibrationStatus,'');assert.equal(f.Explanation,'');
  assert.deepEqual(p.Laws,f.Laws);
  assert.deepEqual(Object.keys(p.Laws).sort(),corpus.map(c=>c.fixture_id).sort());
  assert.equal(p.Candidates.length,1);assert.equal(f.Candidates.length,1);
  assert.equal(p.Before.length,4);assert.equal(p.After.length,4);
  assert.deepEqual(p.Before,f.Before);assert.deepEqual(p.After,f.After);
  const explanation=JSON.parse(p.Explanation);
  assert.equal(explanation.Plan.CalibrationStatus,'not_evaluated');
  assert.equal(explanation.Plan.Method,'research/calendar-priority-v1');
  assert.deepEqual(explanation.PackedIDs.map(id=>ids[id]),p.Candidates.map(c=>c.Fixture));
  const digest=crypto.createHash('sha256').update(JSON.stringify(['research',p.Original])).digest('hex');
  assert.equal(explanation.OriginalQueryDigest,digest);
  const input=p.After.find(c=>ids[c.ID]===p.Candidates[0].Fixture);assert(input);
  assert.equal(p.Candidates[0].Score,input.Score);
  const positive=oracle[p.Case].support.length>0;
  for(const r of [f,p]) assert.equal(r.SupportRank,oracle[r.Case].support.includes(r.Candidates[0].Fixture)?1:0);
  if(positive) {baseline+=f.SupportRank;priority+=p.SupportRank;rescues+=p.SupportRank>f.SupportRank;regressions+=p.SupportRank<f.SupportRank;}
  else absent++;
 }
 assert.equal(regressions,0);assert(priority>=minimum);assert.equal(absent,2);
 summaries.push({set,baseline,priority,positive:8,rescues,regressions,absentStillPacked:absent});
}
console.log(JSON.stringify({summaries,fullFrontierLawsUnchanged:true,originalScoresPreserved:true,journalExplanationsMatched:30,scope:'consumed-data service screen, not fresh learning validation'},null,2));
