import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root=path.resolve(import.meta.dirname,'../..');
const read=p=>JSON.parse(fs.readFileSync(path.join(import.meta.dirname,p)));
const d=read('semantic-role-v1/results.json'),oracle=read('semantic-role-v1/oracle.json');
for(const [p,h]of Object.entries(d.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,p))).digest('hex'),h);
assert.equal(d.Results.length,24);assert.equal(new Set(d.Results.map(r=>r.Arm+':'+r.Case)).size,24);
const rows=[];let baseline=0,priority=0,rescues=0,regressions=0;
for(const p of d.Results.filter(r=>r.Arm==='priority')) {
 const f=d.Results.find(r=>r.Arm==='focus'&&r.Case===p.Case);assert(f);
 assert.equal(p.Candidates.length,1);assert.equal(f.Candidates.length,1);
 assert.deepEqual(p.Laws,f.Laws);assert.equal(Object.keys(p.Laws).length,2);
 assert.deepEqual(p.Before,f.Before);assert.deepEqual(p.After,f.After);
 assert.equal(p.Before.length,2);assert(p.JournalMatched&&f.JournalMatched);
 assert.equal(p.CalibrationStatus,'not_evaluated');assert.equal(p.PacketCertainty,0);
 for(const r of [f,p]) assert.equal(r.SupportRank,oracle[r.Case].support.includes(r.Candidates[0].Fixture)?1:0);
 baseline+=f.SupportRank;priority+=p.SupportRank;rescues+=p.SupportRank>f.SupportRank;regressions+=p.SupportRank<f.SupportRank;
 rows.push({case:p.Case,baseline:f.SupportRank,priority:p.SupportRank});
}
console.log(JSON.stringify({baseline,priority,total:12,rescues,regressions,rows,universalPrioritySafety:regressions===0?'not_falsified_here':'falsified',scope:'designed paired public-fact retrieval, not population prevalence'},null,2));
