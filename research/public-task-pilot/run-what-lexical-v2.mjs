import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {lexicalOrder} from './what-lexical-v2.mjs';
const root=path.resolve(import.meta.dirname,'../..');
const read=p=>JSON.parse(fs.readFileSync(path.join(import.meta.dirname,p)));
const output=process.argv[2];if(!output||fs.existsSync(output))throw Error('NEW output required');
const hashes={};
function hash(p){hashes[p]=crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex');}
for(const name of ['what-lexical-v2.mjs','run-what-lexical-v2.mjs','WHAT_LEXICAL_V2_PROTOCOL.md'])hash('research/public-task-pilot/'+name);
const rows=[];
for(const set of ['semantic-role-v1','w3c-focus-v1','esa-date-v1','esa-iso-v1','landing-transfer-v1']){
 const file=set==='landing-transfer-v1'?'results.json':'task-results.json';
 const data=read(`${set}/${file}`),facts=read(`${set}/corpus.json`);
 for(const p of [file,'corpus.json','oracle.json'])hash(`research/public-task-pilot/${set}/${p}`);
 const fixture=id=>facts[Number(id.replace('public-record-',''))]?.fixture_id;
 const pending=[];
 for(const r of data.Results.filter(r=>r.Arm==='priority')){
  const plan=JSON.parse(r.Explanation).Plan;
  const raw=structuredClone(r.After);
  const lexical=lexicalOrder(r.Effective,raw);
  const combined=lexicalOrder(r.Effective,raw,plan);
  assert.deepEqual(raw,r.After);
  pending.push({set,case:r.Case,lexical:fixture(raw[lexical.order[0]].ID),combined:fixture(raw[combined.order[0]].ID),task:r.Candidates[0].Fixture,lexicalTrace:lexical,combinedTrace:combined,ids:raw.map(c=>c.ID)});
 }
 const oracle=read(`${set}/oracle.json`);
 for(const r of pending){r.positive=oracle[r.case].support.length>0;r.taskCorrect=oracle[r.case].support.includes(r.task);r.lexicalCorrect=oracle[r.case].support.includes(r.lexical);r.combinedCorrect=oracle[r.case].support.includes(r.combined);rows.push(r);}
}
const summaries=[...new Set(rows.map(r=>r.set))].map(set=>{
 const rs=rows.filter(r=>r.set===set&&r.positive);
 return {set,positive:rs.length,task:rs.filter(r=>r.taskCorrect).length,lexical:rs.filter(r=>r.lexicalCorrect).length,combined:rs.filter(r=>r.combinedCorrect).length,regressions:rs.filter(r=>r.taskCorrect&&!r.combinedCorrect).map(r=>r.case)};
});
const pass=summaries.every(s=>!s.regressions.length)&&summaries.find(s=>s.set==='landing-transfer-v1').combined>6;
fs.writeFileSync(output,JSON.stringify({hashes,rows,summaries,screen:pass?'PASS':'FAIL'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({summaries,screen:pass?'PASS':'FAIL'},null,2));
