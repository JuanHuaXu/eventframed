import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {partition} from './date-normalized-strict.mjs';
const read=p=>JSON.parse(fs.readFileSync(new URL(p,import.meta.url)));
const summaries=[];let lawMax=0,totalRescued=0;
for(const set of ['w3c-focus-v1','esa-date-v1','esa-iso-v1']) {
  const d=read(`./${set}/contract-results.json`),corpus=read(`./${set}/corpus.json`),oracle=read(`./${set}/oracle.json`);
  const offline=read(`./${set}/normalized-strict-results.json`);
  for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',p))).digest('hex'),h);
  assert.equal(d.Results.length,20);assert.equal(new Set(d.Results.map(r=>r.Arm+':'+r.Case)).size,20);
  const identity=new Map(corpus.map((f,i)=>[`public-record-${String(i).padStart(3,'0')}`,f.fixture_id]));
  for(const r of d.Results) {
    assert(['focus','calendar'].includes(r.Arm));assert(oracle[r.Case]);
    assert.equal(r.Candidates.length,1);assert.equal(r.Before.length,4);assert.equal(r.After.length,4);
    assert.deepEqual(r.Before.map(c=>c.ID).sort(),[...identity.keys()].sort());
    assert.deepEqual(r.After.map(c=>c.ID).sort(),[...identity.keys()].sort());
    assert.deepEqual(Object.keys(r.Laws).sort(),corpus.map(f=>f.fixture_id).sort());
    assert.equal(r.SupportRank,oracle[r.Case].support.includes(r.Candidates[0].Fixture)?1:0);
    for(const c of r.Before){const f=corpus.find(f=>f.fixture_id===identity.get(c.ID));assert(c.Text.includes(f.text));}
    for(const c of r.After){const b=r.Before.find(b=>b.ID===c.ID);assert.equal(c.Text,b.Text);assert.equal(c.Metadata,b.Metadata);}
    if(r.Arm==='calendar') {
      // Independent JS partition consumes the same actual what fields as Go.
      const input=r.Before.map(c=>({id:c.ID,text:c.Text.split('\n').find(l=>l.startsWith('what: ')).slice(6)}));
      const expected=partition(r.Original,input).ordered.map(c=>c.id);
      assert.deepEqual(r.After.map(c=>c.ID),expected);
      assert.equal(r.Candidates[0].Fixture,identity.get(expected[0]));
      const b=d.Results.find(b=>b.Arm==='focus'&&b.Case===r.Case);
      for(const id of identity.values())for(const k of ['useful','not_useful']) {
        assert(Number.isFinite(r.Laws[id][k])&&Number.isFinite(b.Laws[id][k]));
        lawMax=Math.max(lawMax,Math.abs(r.Laws[id][k]-b.Laws[id][k]));
      }
      const old=offline.records.find(o=>o.case===r.Case);
      assert.equal(r.Candidates[0].Fixture,old.candidates[0].Fixture);
    }
  }
  const rows=d.Results.filter(r=>r.Arm==='calendar').map(c=>{
    const b=d.Results.find(b=>b.Arm==='focus'&&b.Case===c.Case);
    return {case:c.Case,supported:oracle[c.Case].support.length>0,control:b.SupportRank,calendar:c.SupportRank,
      winnerInputRank:c.Before.findIndex(v=>identity.get(v.ID)===c.Candidates[0].Fixture)+1};
  });
  const positive=rows.filter(r=>r.supported),rescued=positive.filter(r=>r.control===0&&r.calendar===1);
  totalRescued+=rescued.length;
  assert(rescued.every(r=>r.winnerInputRank>1));
  summaries.push({set,control:positive.filter(r=>r.control===1).length,calendar:positive.filter(r=>r.calendar===1).length,total:8,
    preserved:positive.every(r=>r.control!==1||r.calendar===1),rescued,rows});
}
assert(lawMax<=1e-12);
console.log(JSON.stringify({summaries,lawMax,totalRescued,passed:totalRescued>0&&summaries.every(s=>s.preserved),
  scope:'actual pre-packing contract; pack1; isolated empty-cache research, not production qualification'},null,2));
