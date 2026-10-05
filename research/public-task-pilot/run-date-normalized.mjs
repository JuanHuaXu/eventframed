import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {partition} from './date-normalized.mjs';
const root=import.meta.dirname;
const set=process.argv[2];
assert(['w3c-focus-v1','esa-date-v1','esa-iso-v1'].includes(set));
const read=p=>JSON.parse(fs.readFileSync(path.join(root,p)));
const source=read(`${set}/results.json`);
for(const [p,h]of Object.entries(source.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(root,'../..',p))).digest('hex'),h);
const corpus=read(`${set}/corpus.json`);
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
const oracle=read(`${set}/oracle.json`);
for(const r of records) r.supportRank=r.candidates.findIndex(c=>oracle[r.case].support.includes(c.Fixture))+1;
const hashes={};
for(const p of ['DATE_NORMALIZATION_PROTOCOL.md','date-constraints.mjs','date-constraints.test.mjs','date-normalized.mjs','date-normalized.test.mjs','run-date-normalized.mjs',`${set}/results.json`,`${set}/corpus.json`,`${set}/oracle.json`])
  hashes[p]=crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex');
fs.writeFileSync(path.join(root,`${set}/normalized-results.json`),JSON.stringify({hashes,records,scope:'date normalization consumed-data replay; no daemon integration'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log('Saved ten constraint replay cases.');
