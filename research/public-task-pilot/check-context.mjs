import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./context-v1/results.json',import.meta.url)));
const oracle=JSON.parse(fs.readFileSync(new URL('./context-v1/oracle.json',import.meta.url)));
for(const [f,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',f))).digest('hex'),h);
assert.equal(d.Results.length,20);const seen=new Set();let maxLaw=0,maxDelta=0;
for(const r of d.Results){const key=[r.Arm,r.Case].join(':');assert(!seen.has(key));seen.add(key);assert.equal(r.Candidates.length,8);assert.equal(new Set(r.Candidates.map(c=>c.Fixture)).size,8);
 const support=oracle[r.Case].support;const rank=r.Candidates.findIndex(c=>support.includes(c.Fixture))+1;assert.equal(rank,r.SupportRank);
 if(r.Arm==='bounded'){const b=d.Results.find(b=>b.Arm==='baseline'&&b.Case===r.Case);assert(b);assert.equal(r.Frontier,8);
  for(const c of r.Candidates){const old=b.Candidates.find(v=>v.Fixture===c.Fixture);maxDelta=Math.max(maxDelta,Math.abs(old.Score-c.Score));for(const k of Object.keys(old.Law))maxLaw=Math.max(maxLaw,Math.abs(old.Law[k]-c.Law[k]));}
 }
}
assert(maxDelta<=.005+1e-12);assert(maxLaw<=1e-12);
const rows=d.Results.filter(r=>r.Arm==='baseline').map(b=>{const c=d.Results.find(r=>r.Arm==='bounded'&&r.Case===b.Case);return {case:b.Case,supported:oracle[b.Case].support.length>0,baseRank:b.SupportRank,boundedRank:c.SupportRank,baseTop:b.Candidates[0].Fixture,boundedTop:c.Candidates[0].Fixture};});
const positives=rows.filter(r=>r.supported),base=positives.filter(r=>r.baseRank===1).length,bounded=positives.filter(r=>r.boundedRank===1).length;
console.log(JSON.stringify({rows,base,bounded,maxDelta,maxLaw,passed:bounded>base&&positives.every(r=>r.baseRank!==1||r.boundedRank===1),scope:'public context retrieval; no generation or learning'},null,2));
