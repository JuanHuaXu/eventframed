import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {partition} from './date-constraints.mjs';
const root=import.meta.dirname;
const read=p=>JSON.parse(fs.readFileSync(path.join(root,p)));
const source=read('w3c-focus-v1/results.json');
for(const [p,h]of Object.entries(source.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(root,'../..',p))).digest('hex'),h);
const corpus=read('w3c-focus-v1/corpus.json');
const records=source.Results.filter(r=>r.Arm==='focus').map(r=>{
  const inputs=r.Candidates.map(c=>({text:corpus.find(f=>f.fixture_id===c.Fixture).text}));
  // The scorer sees only original question and ordered text, never oracle or IDs.
  const ranked=partition(r.Original,inputs);
  const order=ranked.ordered.map(a=>inputs.findIndex(i=>i.text===a.text));
  assert.equal(new Set(order).size,inputs.length);
  return {case:r.Case,original:r.Original,rules:ranked.rules,allContradicted:ranked.allContradicted,
    candidates:order.map((i,j)=>({...r.Candidates[i],Contradiction:ranked.ordered[j].contradiction}))};
});
// Only after ranking do support labels become visible to this runner.
const oracle=read('w3c-focus-v1/oracle.json');
for(const r of records) r.supportRank=r.candidates.findIndex(c=>oracle[r.case].support.includes(c.Fixture))+1;
const hashes={};
for(const p of ['DATE_CONSTRAINT_PROTOCOL.md','date-constraints.mjs','date-constraints.test.mjs','run-date-constraints.mjs','w3c-focus-v1/results.json','w3c-focus-v1/corpus.json','w3c-focus-v1/oracle.json'])
  hashes[p]=crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex');
fs.writeFileSync(path.join(root,'w3c-focus-v1/date-constraint-results.json'),JSON.stringify({hashes,records,scope:'offline consumed-design constraint replay; no daemon integration'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log('Saved ten constraint replay cases.');
