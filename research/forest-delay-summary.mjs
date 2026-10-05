import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(rows.length,192);assert.equal(header.PerCell,16);
assert.deepEqual(header.Seeds,[2027010151,2027010251]);
for(const [p,h] of Object.entries(header.Hashes)){assert.equal(hash(header.Sources[p]),h);assert.equal(hash(fs.readFileSync(p)),h);}
assert.equal(new Set(rows.map(r=>r.TrainSeed)).size,192);
const close=(a,b)=>assert(Math.abs(a-b)<1e-9,`${a} != ${b}`);
for(const r of rows){
 for(const schedule of ['Immediate','Delayed']){
  const s=r[schedule],frames=s.Frames;assert.equal(frames.length,512);
  assert.equal(s.AuditCost,18*frames.filter(f=>f.Audit).length);
  for(let a=0;a<2;a++)for(const part of ['Full','Post']){
   const fslice=frames.slice(part==='Full'?0:256);
   const m={N:fslice.length,Correct:0,Cost:0,Brier:0,LogLoss:0};
   for(const f of fslice){const p=f.Predictions[a];assert(Number.isFinite(p.p)&&p.p>0&&p.p<1);assert(Number.isInteger(p.mask)&&p.mask>=0&&p.mask<=511);assert.equal(p.values&~p.mask,0);assert(p.cost>=0&&p.cost<=6);m.Cost+=p.cost;m.Correct+=Number((p.p>=.5)===f.Y);m.Brier+=(p.p-Number(f.Y))**2;m.LogLoss-=f.Y?Math.log(p.p):Math.log1p(-p.p);}
   for(const k of Object.keys(m))close(m[k],s.Arms[a][part][k]);
  }
  assert.equal(s.Arms[0].SplitAt,s.Arms[1].SplitAt);
  const expectedFits=[],ready=Array(512).fill(false),expired=Array(512).fill(false);
  let audits=0,head=0,generation=1,applied=0,stale=0,censored=0;
  const issuedGeneration=[];
  const drain=()=>{while(head<issuedGeneration.length&&(ready[head]||expired[head])){if(expired[head])censored++;else if(issuedGeneration[head]===generation)applied++;else stale++;head++;}};
  for(let clock=0;clock<544;clock++){
   if(clock<512)issuedGeneration.push(generation);
   let fit=false;
   for(let origin=0;origin<Math.min(clock+1,512);origin++){
    const f=frames[origin];assert(f.Arrival>=origin&&f.Arrival<origin+32);
    if(schedule==='Immediate'){assert.equal(f.Arrival,origin);assert.equal(f.Missing,false);}
    if(!ready[origin]&&!expired[origin]&&!f.Missing&&f.Arrival===clock){ready[origin]=true;drain();if(f.Audit){audits++;if(audits>=32&&(audits-32)%16===0)fit=true;}}
   }
   for(let origin=head;origin<Math.min(Math.max(0,clock-31),512);origin++)if(!ready[origin])expired[origin]=true;
   drain();
   if(s.Arms[0].SplitAt===clock)generation++;
   if(fit){const ids=frames.flatMap((f,id)=>f.Audit&&!f.Missing&&f.Arrival<=clock?[id]:[]).slice(-256);expectedFits.push({Clock:clock,Origins:ids});generation++;}
  }
  assert.deepEqual(s.Fits,expectedFits);assert.equal(head,512);
  assert.deepEqual(s.Applied,[applied,applied]);assert.deepEqual(s.Stale,[stale,stale]);assert.deepEqual(s.Censored,[censored,censored]);
  assert.deepEqual(s.Released,frames.flatMap((f,id)=>!f.Missing?[id]:[]));
 }
 for(let n=0;n<512;n++)for(const k of ['X','RX','Y','RY','Audit','Seed','Baseline','Reference'])assert.equal(r.Immediate.Frames[n][k],r.Delayed.Frames[n][k]);
}
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['design','confirmation'])for(const name of ['copy_bit','copy_xor2','noise10_bit','reverses_xor2','stable_noise10','null_uniform'])for(const schedule of ['Immediate','Delayed']){
 const rs=rows.filter(r=>r.Phase===phase&&r.Case===name);assert.equal(rs.length,16);assert.deepEqual(rs.map(r=>r.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
 const contrast=part=>interval(rs.map(r=>{const a=r[schedule].Arms;return a[0][part].Brier/a[0][part].N-a[1][part].Brier/a[1][part].N;}));
 const post=contrast('Post'),full=contrast('Full');
 cells.push({phase,case:name,schedule,post,full,gain:post.mean>=.005&&post.lower>0,nonHarm:post.lower>=-.01&&full.lower>=-.01,
  brier:[0,1].map(a=>mean(rs.map(r=>r[schedule].Arms[a].Post.Brier/256))),
  foreground:[0,1].map(a=>mean(rs.map(r=>r[schedule].Arms[a].Full.Cost/512))),
  monitor:mean(rs.map(r=>r[schedule].MonitorCost/512)),audit:mean(rs.map(r=>r[schedule].AuditCost/512)),fits:mean(rs.map(r=>r[schedule].Fits.length)),
  applied:mean(rs.map(r=>r[schedule].Applied[0])),stale:mean(rs.map(r=>r[schedule].Stale[0])),censored:mean(rs.map(r=>r[schedule].Censored[0]))});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,trajectories:192,scheduleRuns:384,cells,limits:'Exploratory paired mean +/-3.5SE; no anytime guarantee. Schedules are paired, not independent trajectories. All forecasts including missing labels are scored.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({sourceHash:out.sourceHash,capturedSources:out.capturedSources,nonHarm:cells.filter(c=>c.nonHarm).length,total:cells.length,delayedGains:cells.filter(c=>c.schedule==='Delayed'&&!c.case.startsWith('stable')&&!c.case.startsWith('null')&&c.gain).length}));
