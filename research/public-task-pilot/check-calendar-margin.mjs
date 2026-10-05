import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const read=p=>JSON.parse(fs.readFileSync(new URL(p,import.meta.url)));
const summaries=[];let maxLaw=0,maxCertaintyMovement=0;
for(const set of ['w3c-focus-v1','esa-date-v1','esa-iso-v1']) {
  const d=read(`./${set}/margin-results.json`),old=read(`./${set}/contract-results.json`),oracle=read(`./${set}/oracle.json`),corpus=read(`./${set}/corpus.json`);
  for(const [p,h]of Object.entries(d.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',p))).digest('hex'),h);
  assert.equal(d.Results.length,30);assert.equal(new Set(d.Results.map(r=>r.Arm+':'+r.Case)).size,30);
  for(const r of d.Results) {
    assert(['focus','calendar','margin'].includes(r.Arm));assert(oracle[r.Case]);
    assert.equal(r.Before.length,4);assert.equal(r.After.length,4);assert.equal(r.Candidates.length,1);
    assert.deepEqual(r.After.map(c=>c.ID).sort(),r.Before.map(c=>c.ID).sort());
    assert.deepEqual(Object.keys(r.Laws).sort(),corpus.map(c=>c.fixture_id).sort());
    assert.equal(r.SupportRank,oracle[r.Case].support.includes(r.Candidates[0].Fixture)?1:0);
    assert(Number.isFinite(r.PacketCertainty)&&r.PacketCertainty>=0&&r.PacketCertainty<=1);
    for(const c of r.After){const b=r.Before.find(b=>b.ID===c.ID);assert.equal(c.Text,b.Text);assert.equal(c.Metadata,b.Metadata);}
    if(r.Arm!=='margin') {
      const before=old.Results.find(o=>o.Arm===r.Arm&&o.Case===r.Case);assert(before);
      assert.equal(r.Candidates[0].Fixture,before.Candidates[0].Fixture);
    }
  }
  const rows=d.Results.filter(r=>r.Arm==='margin').map(m=>{
    const o=d.Results.find(r=>r.Arm==='calendar'&&r.Case===m.Case),f=d.Results.find(r=>r.Arm==='focus'&&r.Case===m.Case);
    assert.equal(m.Original,o.Original);assert.equal(m.Effective,o.Effective);
    assert.equal(m.Candidates[0].Fixture,o.Candidates[0].Fixture);
    assert.deepEqual(m.After.map(c=>c.ID),o.After.map(c=>c.ID));
    for(const id of Object.keys(m.Laws))for(const k of ['useful','not_useful'])maxLaw=Math.max(maxLaw,Math.abs(m.Laws[id][k]-f.Laws[id][k]));
    maxCertaintyMovement=Math.max(maxCertaintyMovement,Math.abs(m.PacketCertainty-o.PacketCertainty));
    return {case:m.Case,correct:m.SupportRank===1,supported:oracle[m.Case].support.length>0,focusCertainty:f.PacketCertainty,ordinalCertainty:o.PacketCertainty,marginCertainty:m.PacketCertainty};
  });
  summaries.push({set,correct:rows.filter(r=>r.supported&&r.correct).length,total:8,rows});
}
assert(maxLaw<=1e-12);
console.log(JSON.stringify({summaries,maxLaw,maxCertaintyMovement,winnersPreserved:true,
  scope:'empty-cache service composition; certainty movement is not calibration improvement'},null,2));
