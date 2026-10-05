import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../..');
const read=p=>JSON.parse(fs.readFileSync(path.join(import.meta.dirname,p)));
const summaries=[];let calls=0,pass=true;
for(const [set,minimum]of [['semantic-role-v1',7],['w3c-focus-v1',7],['esa-date-v1',8],['esa-iso-v1',7]]) {
 const d=read(`${set}/task-results.json`),oracle=read(`${set}/oracle.json`),facts=read(`${set}/corpus.json`);
 for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex'),h);
 assert.equal(d.Results.length,2*Object.keys(oracle).length);calls+=d.Results.length;
 let baseline=0,task=0,regressions=0,rescues=0,positive=0;const rows=[];
 for(const p of d.Results.filter(r=>r.Arm==='priority')) {
  const f=d.Results.find(r=>r.Arm==='focus'&&r.Case===p.Case);assert(f);
  assert(p.JournalMatched&&f.JournalMatched);assert.deepEqual(p.Laws,f.Laws);
  assert.deepEqual(p.Before,f.Before);assert.deepEqual(p.After,f.After);
  assert.equal(Object.keys(p.Laws).length,facts.length);
  assert.equal(p.Candidates.length,1);assert.equal(f.Candidates.length,1);
  const plan=JSON.parse(p.Explanation).Plan;assert.equal(plan.Method,'research/task-role-v1');
  assert.equal(p.CalibrationStatus,'not_evaluated');assert.equal(p.PacketCertainty,0);
  for(const r of [p,f])assert.equal(r.SupportRank,oracle[r.Case].support.includes(r.Candidates[0].Fixture)?1:0);
  if(oracle[p.Case].support.length){positive++;baseline+=f.SupportRank;task+=p.SupportRank;regressions+=p.SupportRank<f.SupportRank;rescues+=p.SupportRank>f.SupportRank;}
  rows.push({case:p.Case,baseline:f.SupportRank,task:p.SupportRank,role:plan.Role,relation:plan.TargetRelation});
 }
 const screen=task>=minimum&&regressions===0;pass&&=screen;
 summaries.push({set,baseline,task,positive,rescues,regressions,screen,rows});
}
assert.equal(calls,84);
console.log(JSON.stringify({calls,summaries,combinedScreen:pass?'PASS':'FAIL',scope:'consumed design replay, not learned or general semantic understanding'},null,2));
