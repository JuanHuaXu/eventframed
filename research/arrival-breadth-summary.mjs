import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2),raw=fs.readFileSync(input);
const [h,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(h.Consumed,true);assert.equal(rows.length,384);
for(const [name,s] of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[name]);assert.equal(hash(fs.readFileSync(name)),h.Hashes[name]);}
const close=(a,b)=>assert.ok(Math.abs(a-b)<1e-9,`${a} != ${b}`),records=[];
for(const r of rows){
 for(let i=0;i<512;i++)for(const k of ['X','RX','Y','RY','Audit','Seed'])assert.equal(r.Immediate.Frames[i][k],r.Delayed.Frames[i][k]);
 for(const schedule of ['Immediate','Delayed']){
  const s=r[schedule];assert.equal(s.Frames.length,512);
  for(const fit of s.Fits){assert.ok(fit.Origins.length>=32&&fit.Origins.length<=256);let last=-1;for(const id of fit.Origins){assert.ok(id>last);last=id;const f=s.Frames[id];assert.ok(f.Audit&&!f.Missing&&f.Arrival<=fit.Clock);}}
  for(let a=0;a<4;a++)for(const [part,start] of [['Full',0],['Post',256]]){
   const frames=s.Frames.slice(start),m={N:frames.length,Brier:0,Correct:0,Cost:0,LogLoss:0};
   for(const f of frames){const p=f.Predictions[a];assert.ok(p.p>0&&p.p<1&&p.cost>=0&&p.cost<=6);assert.equal(p.values&~p.mask,0);m.Brier+=(p.p-Number(f.Y))**2;m.Correct+=Number((p.p>=.5)===f.Y);m.Cost+=p.cost;m.LogLoss-=f.Y?Math.log(p.p):Math.log1p(-p.p);if(schedule==='Immediate')close(p.p,f.Predictions[0].p);}
   for(const k of Object.keys(m))close(m[k],s.Arms[a][part][k]);
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,full:s.Arms.map(a=>a.Full.Brier/512),post:s.Arms.map(a=>a.Post.Brier/256),cost:s.Arms.map(a=>a.Full.Cost/512)});
 }
}
assert.equal(new Set(records.map(r=>[r.phase,r.case,r.index,r.schedule].join('/'))).size,768);
const mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
const interval=x=>{const m=mean(x),se=Math.sqrt(x.reduce((s,v)=>s+(v-m)**2,0)/(x.length-1)/x.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['design','confirmation'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
 const gain=interval(a.map(r=>r.post[0]-r.post[2]));
 const harmFull=interval(a.map(r=>r.full[2]-r.full[0])),harmPost=interval(a.map(r=>r.post[2]-r.post[0]));
 cells.push({phase,case:name,schedule,post:[0,1,2,3].map(i=>mean(a.map(r=>r.post[i]))),cost:[0,1,2,3].map(i=>mean(a.map(r=>r.cost[i]))),gain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const changed=cells.filter(c=>c.schedule==='Delayed'&&!['stable_noise10','null_uniform'].includes(c.case));
const result={rawHash:hash(raw),sources:Object.keys(h.Sources).length,screens:{gains:changed.filter(c=>c.gainPass).length,gainTotal:changed.length,nonharm:cells.filter(c=>c.nonharmPass).length,nonharmTotal:cells.length},cells,records,limits:'Consumed fixed-sample breadth screen; intervals not research-wide coverage. No equal-cost or production claim.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({hash:result.rawHash,sources:result.sources,screens:result.screens,confirmation:changed.filter(c=>c.phase==='confirmation')}));
