import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2),raw=fs.readFileSync(input);
const [h,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(rows.length,192);assert.equal(h.Consumed,true);
for(const [name,s] of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[name]);assert.equal(hash(fs.readFileSync(name)),h.Hashes[name]);}
const parent=fs.readFileSync('docs/experiments/mmm-coupled-arrival-v1.jsonl');assert.equal(hash(parent),h.ParentHash);
const tracking=JSON.parse(fs.readFileSync('docs/experiments/mmm-delay-tracking-selector-v1.json'));
const close=(a,b)=>assert.ok(Math.abs(a-b)<1e-10,`${a} != ${b}`);
const records=[];
for(const r of rows)for(const schedule of ['Immediate','Delayed']){
 const s=r[schedule],f=s.Frames;assert.equal(f.length,512);
 const reference=tracking.results.find(x=>x.phase===r.Phase&&x.case===r.Case&&x.index===r.Index&&x.schedule===schedule);assert.ok(reference);
 close(s.Arms[3].Full.Brier/512,reference.scores.full.arrival);close(s.Arms[3].Post.Brier/256,reference.scores.post.arrival);
 for(let a=0;a<4;a++)for(const [part,start] of [['Full',0],['Post',256]]){
  const selected=f.slice(start),m={N:selected.length,Cost:0,Correct:0,Brier:0,LogLoss:0};
  for(const row of selected){const p=row.Predictions[a];assert.ok(p.p>0&&p.p<1);assert.ok(p.cost<=6&&p.cost>=0);m.Cost+=p.cost;m.Correct+=Number((p.p>=.5)===row.Y);m.Brier+=(p.p-Number(row.Y))**2;m.LogLoss-=row.Y?Math.log(p.p):Math.log1p(-p.p);}
  for(const key of Object.keys(m))close(m[key],s.Arms[a][part][key]);
 }
 for(const row of f){assert.equal(row.Predictions[0].mask,row.Predictions[3].mask);assert.equal(row.Predictions[0].cost,row.Predictions[3].cost);if(schedule==='Immediate')for(let a=1;a<4;a++)close(row.Predictions[0].p,row.Predictions[a].p);}
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,post:s.Arms.map(a=>a.Post.Brier/256),cost:s.Arms.map(a=>a.Full.Cost/512),changedMasks:f.filter(x=>x.Predictions[0].mask!==x.Predictions[1].mask).length});
}
assert.equal(new Set(records.map(r=>[r.phase,r.case,r.index,r.schedule].join('/'))).size,384);
const mean=x=>x.reduce((s,a)=>s+a,0)/x.length;
const cells=[];
for(const phase of ['design','confirmation'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
 cells.push({phase,case:name,schedule,n:a.length,post:[0,1,2,3].map(i=>mean(a.map(r=>r.post[i]))),cost:[0,1,2,3].map(i=>mean(a.map(r=>r.cost[i]))),meanChangedMasks:mean(a.map(r=>r.changedMasks))});
}
const result={rawHash:hash(raw),sources:Object.keys(h.Sources).length,cells,records,limits:'Consumed inner/outer-selector diagnostic; not fresh full-system validation.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(cells.filter(c=>c.phase==='confirmation'&&c.schedule==='Delayed')));
