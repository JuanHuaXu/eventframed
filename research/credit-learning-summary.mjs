import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,shadow,output]=process.argv.slice(2);
const raw=fs.readFileSync(input),praw=fs.readFileSync(parent),sraw=fs.readFileSync(shadow);
const [h,...rows]=raw.toString().trim().split('\n').map(JSON.parse),[, ...ps]=praw.toString().trim().split('\n').map(JSON.parse),ss=JSON.parse(sraw);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(hash(praw),h.ParentSHA256);assert.equal(hash(sraw),h.ShadowSHA256);assert.equal(h.Consumed,true);assert.equal(rows.length,128);
for(const [p,s]of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),h.Hashes[p]);}
const key=r=>[r.Phase,r.Case,r.Index].join('/'),pm=new Map(ps.map(r=>[key(r),r])),sm=new Map(ss.Records.map(r=>[key(r)+'/'+r.Schedule,r]));
const close=(a,b)=>assert.ok(Math.abs(a-b)<1e-9,`${a} != ${b}`),records=[];
for(const r of rows)for(const [si,schedule]of ['Immediate','Delayed'].entries()){
  const c=r[schedule],p=pm.get(key(r))[schedule],s=sm.get(key(r)+'/'+schedule);
  assert.deepEqual(r.Acquisition[si],s.Acquisition);assert.equal(c.Arms[2].SplitAt,s.Candidate.SplitAt);
  assert.deepEqual(c.Fits,p.Fits);assert.deepEqual(c.Released,p.Released);
  assert.equal(c.MonitorCost+c.AuditCost,s.CandidateCost);assert.ok(s.CandidateCost<=s.OriginalCost);
  assert.equal(c.Frames.length,512);
  for(let i=0;i<512;i++)for(const k of ['X','RX','Y','RY','Audit','Missing','Arrival','Seed'])assert.equal(c.Frames[i][k],p.Frames[i][k]);
  for(const f of c.Fits)for(const id of f.Origins){const x=c.Frames[id];assert.ok(x.Audit&&!x.Missing&&x.Arrival<=f.Clock);}
  for(let a=0;a<4;a++)for(const [part,start]of [['Full',0],['Post',256]]){
    const m={N:512-start,Brier:0,Correct:0,Cost:0,LogLoss:0};
    for(const f of c.Frames.slice(start)){
      const x=f.Predictions[a];assert.ok(x.p>0&&x.p<1&&x.cost>=0&&x.cost<=6);assert.equal(x.values&~x.mask,0);
      m.Brier+=(x.p-Number(f.Y))**2;m.Correct+=Number((x.p>=.5)===f.Y);m.Cost+=x.cost;m.LogLoss-=f.Y?Math.log(x.p):Math.log1p(-x.p);
    }
    for(const k of Object.keys(m))close(m[k],c.Arms[a][part][k]);
  }
  let allowance=0,charge=0;
  for(const a of r.Acquisition[si]){allowance+=a.Allowance;charge+=a.Charge;assert.equal(a.Credit,allowance-charge);assert.ok(charge<=allowance);}
  close(allowance,p.MonitorCost+p.AuditCost);close(charge,c.MonitorCost+c.AuditCost);
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,full:[p,c].map(x=>x.Arms[2].Full.Brier/512),post:[p,c].map(x=>x.Arms[2].Post.Brier/256),accuracy:[p,c].map(x=>x.Arms[2].Post.Correct/256),foreground:[p,c].map(x=>x.Arms[2].Full.Cost/512),totalCoordinates:[p,c].map(x=>(x.Arms[2].Full.Cost+x.MonitorCost+x.AuditCost)/512),splits:[p,c].map(x=>x.Arms[2].SplitAt)});
}
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,v)=>s+(v-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
  const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
  const gain=interval(a.map(r=>r.post[0]-r.post[1])),harmFull=interval(a.map(r=>r.full[1]-r.full[0])),harmPost=interval(a.map(r=>r.post[1]-r.post[0]));
  cells.push({phase,case:name,schedule,post:[0,1].map(i=>mean(a.map(r=>r.post[i]))),accuracy:[0,1].map(i=>mean(a.map(r=>r.accuracy[i]))),foreground:[0,1].map(i=>mean(a.map(r=>r.foreground[i]))),totalCoordinates:[0,1].map(i=>mean(a.map(r=>r.totalCoordinates[i]))),gain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const targets=cells.filter(c=>c.case==='parity_to_majority');
const result={rawSHA256:hash(raw),sources:Object.keys(h.Sources).length,screens:{gains:targets.filter(c=>c.gainPass).length,gainTotal:targets.length,nonharm:cells.filter(c=>c.nonharmPass).length,nonharmTotal:cells.length},cells,records,limits:'Consumed closed-loop comparison; descriptive fixed-sample intervals, not research-wide coverage. Credit protects monitoring, not altered foreground cost. No fresh/agent/service validation.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({hash:result.rawSHA256,sources:result.sources,screens:result.screens,delayed:cells.filter(c=>c.schedule==='Delayed')}));
