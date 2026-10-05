import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const fit=JSON.parse(fs.readFileSync(process.argv[2],'utf8')),target=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
assert.equal(fit.targetSHA256,sha(process.argv[3]));assert.equal(fit.records.length,target.records.length);
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length,phases=[];
for(const phase of [0,1]){
  const rows=[];
  for(let i=0;i<fit.records.length;i++){
    const r=fit.records[i],t=target.records[i];for(const k of ['phase','case','index','schedule'])assert.equal(r[k],t[k]);
    if(r.phase!==phase||r.schedule!==1)continue;
    const ys=t.branches.slice(1).map(b=>t.branches[0].value-b.value),ps=r.decision.scores;assert.equal(ys.length,ps.length);
    rows.push({ys,ps,y:mean(ys),p:mean(ps)});
  }
  assert.equal(rows.length,672);const overall=mean(rows.map(r=>r.y));
  const withinVariance=mean(rows.map(r=>mean(r.ys.map(y=>(y-r.y)**2))));
  const betweenVariance=mean(rows.map(r=>(r.y-overall)**2));
  const withinSSE=mean(rows.map(r=>mean(r.ys.map((y,i)=>((y-r.y)-(r.ps[i]-r.p))**2))));
  const betweenSSE=mean(rows.map(r=>(r.y-r.p)**2));
  const totalSSE=mean(rows.map(r=>mean(r.ys.map((y,i)=>(y-r.ps[i])**2))));
  assert(Math.abs(totalSSE-withinSSE-betweenSSE)<1e-14);
  phases.push({phase,withinVariance,betweenVariance,withinVarianceFraction:withinVariance/(withinVariance+betweenVariance),withinSSE,betweenSSE,totalSSE,
    withinR2:1-withinSSE/withinVariance,betweenR2:1-betweenSSE/betweenVariance});
}
console.log(JSON.stringify({scope:'Post-hoc training-objective diagnostic; not a rescue or acceptance gate',fitSHA256:sha(process.argv[2]),targetSHA256:sha(process.argv[3]),scriptSHA256:sha(import.meta.filename),phases}));
