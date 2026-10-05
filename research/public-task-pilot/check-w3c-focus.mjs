import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const read = p => JSON.parse(fs.readFileSync(new URL(p, import.meta.url)));
const pairs = [['date-proposed','date-final'], ['status-proposed','status-final'],
  ['exclude-september','exclude-november'], ['before-2016','after-2016']];
const source = p => fs.readFileSync(new URL(p,import.meta.url),'utf8');
assert.equal(source('../../cmd/public-w3c-focus/main.go').split('func focusQuery(q string) string {')[1],
  source('../../cmd/public-contrast-focus/main.go').split('func focusQuery(q string) string {')[1]);
const summaries=[];
for(const set of ['w3c-focus-v1']) {
  const data=read(`./${set}/results.json`);
  const queries=read(`./${set}/queries.json`), oracle=read(`./${set}/oracle.json`);
  const facts=read('./w3c-focus-v1/corpus.json');
  for(const [p,h]of Object.entries(data.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',p))).digest('hex'),h);
  assert.equal(data.Results.length,queries.length*2);
  const keys=new Set();
  for(const r of data.Results) {
    assert(['baseline','focus'].includes(r.Arm));
    const q=queries.find(q=>q.case_id===r.Case); assert(q);
    assert.equal(r.Original,q.question);
    const pieces=q.question.split(', rather than ');
    const expected=r.Arm==='focus'&&pieces.length===2&&pieces.every(p=>p.trim())?pieces[0].trim().replace(/\?+$/,'')+'?':q.question;
    assert.equal(r.Effective,expected);
    assert(!keys.has(r.Arm+':'+r.Case));keys.add(r.Arm+':'+r.Case);
    assert.deepEqual(r.Candidates.map(c=>c.Fixture).sort(),facts.map(f=>f.fixture_id).sort());
    assert.equal(r.SupportRank,r.Candidates.findIndex(c=>oracle[r.Case].support.includes(c.Fixture))+1);
  }
  const samePrefix = ['exclude-september','exclude-november'].map(id=>data.Results.find(r=>r.Arm==='focus'&&r.Case===id));
  assert.equal(samePrefix[0].Effective,samePrefix[1].Effective);
  assert.deepEqual(samePrefix[0].Candidates,samePrefix[1].Candidates);
  assert.notDeepEqual(oracle[samePrefix[0].Case].support,oracle[samePrefix[1].Case].support);
  let maxLawMotion=0,changed=0;
  const rows=queries.map(q=>{
    const b=data.Results.find(r=>r.Arm==='baseline'&&r.Case===q.case_id);
    const f=data.Results.find(r=>r.Arm==='focus'&&r.Case===q.case_id);
    const changedQuery=b.Effective!==f.Effective;
    changed+=Number(changedQuery);
    if(!changedQuery) assert.deepEqual(f.Candidates,b.Candidates);
    for(const c of f.Candidates) {
      const old=b.Candidates.find(v=>v.Fixture===c.Fixture);
      for(const k of Object.keys(c.Law)) {
        assert(Number.isFinite(c.Law[k])&&Number.isFinite(old.Law[k]));
        maxLawMotion=Math.max(maxLawMotion,Math.abs(c.Law[k]-old.Law[k]));
      }
    }
    return {case:q.case_id,supported:oracle[q.case_id].support.length>0,changedQuery,baseRank:b.SupportRank,focusRank:f.SupportRank,baseTop:b.Candidates[0].Fixture,focusTop:f.Candidates[0].Fixture};
  });
  const positives=rows.filter(r=>r.supported);
  const base=positives.filter(r=>r.baseRank===1).length,focus=positives.filter(r=>r.focusRank===1).length;
  summaries.push({set,changed,base,focus,positive:positives.length,maxLawMotion,
    designScreen:focus>base&&positives.every(r=>r.baseRank!==1||r.focusRank===1),
    pairAccuracy:set==='w3c-focus-v1'?{
      base:pairs.filter(ids=>ids.every(id=>rows.find(r=>r.case===id).baseRank===1)).length,
      focus:pairs.filter(ids=>ids.every(id=>rows.find(r=>r.case===id).focusRank===1)).length,total:4}:null,
    temporalAdequacy:set==='w3c-focus-v1'?focus===8:null,rows});
}
const totalBase=summaries.reduce((n,s)=>n+s.base,0);
const totalFocus=summaries.reduce((n,s)=>n+s.focus,0);
const preserved=summaries.every(s=>s.rows.every(r=>!r.supported||r.baseRank!==1||r.focusRank===1));
console.log(JSON.stringify({summaries,pooledDesign:{base:totalBase,focus:totalFocus,
  preserved,passed:totalFocus>totalBase&&preserved,independentConfirmation:false},
  scope:'frozen rule, new public domain; no learning or agent generation'},null,2));
