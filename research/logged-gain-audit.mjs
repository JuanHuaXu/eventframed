import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [projectionPath,resultPath]=process.argv.slice(2),projection=JSON.parse(fs.readFileSync(projectionPath)),result=JSON.parse(fs.readFileSync(resultPath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for(const [p,h]of Object.entries(result.hashes))assert.equal(sha(p),h);
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10,`${a} != ${b}`);
const training=projection.records.filter(r=>r.phase===0&&r.schedule===1),evaluation=projection.records.filter(r=>r.phase===1&&r.schedule===1);
function draw(rep,r){let counter=0;for(;;){const h=crypto.createHash('sha256').update(['mmm-logged-gain-v1',rep,r.phase,r.case,r.index,counter++].join(':')).digest();const n=h[0]+256*h[1]+65536*h[2]+16777216*h[3];if(n!==4294967295)return n%3;}}
let increments=0,intervals=0;
for(let rep=0;rep<64;rep++){
  const sums=[.5,.5,.5],counts=[1,1,1];
  for(const r of training){const a=draw(rep,r);for(let j=0;j<3;j++)if(r.origins[j]===r.origins[a]){sums[j]+=r.losses[a];counts[j]++;}}
  const models=sums.map((s,j)=>s/counts[j]);
  for(const saved of result.runs.filter(r=>r.rep===rep)){
    models.forEach((m,i)=>near(m,saved.models[i]));assert.deepEqual(saved.trainingCounts,counts.map(n=>n-1));
    const candidate=saved.candidate==='random'?0:2;let total=0,v=0,truth=0;
    for(const r of evaluation){
      const selected=draw(rep,r),actions=[...new Set(r.origins)],regression=[],ps=[],delta=[];
      for(const origin of actions){
        const indexes=[0,1,2].filter(j=>r.origins[j]===origin);
        regression.push(saved.method==='IPS'?0:indexes.reduce((s,j)=>s+models[j],0)/indexes.length);
        ps.push(indexes.length/3);delta.push(Number(r.origins[1]===origin)-Number(r.origins[candidate]===origin));
      }
      const estimate=actions.map((_,a)=>regression[a]+(actions[a]===r.origins[selected]?(r.losses[selected]-regression[a])/ps[a]:0));
      total+=delta.reduce((s,d,a)=>s+d*estimate[a],0);
      const endpoints=[];
      for(let logged=0;logged<actions.length;logged++)for(const y of [0,1])endpoints.push(delta.reduce((s,d,a)=>s+d*(regression[a]+(a===logged?(y-regression[a])/ps[a]:0)),0));
      const width=Math.max(...endpoints)-Math.min(...endpoints);v+=(width===0?0:Math.max(width,1e-12))**2;
      truth+=r.losses[1]-r.losses[candidate];increments++;
    }
    const n=evaluation.length,m=total/n;near(saved.mean,m);near(saved.truth,truth/n);near(saved.widthSum,v);near(saved.finalError,m-truth/n);
    const rates=[.01,.02,.05,.1,.2,.5,1,2],threshold=2/(.05/4);
    let lo=0,hi=1e4;
    for(let step=0;step<100;step++){
      const b=(lo+hi)/2,sum=rates.reduce((s,l)=>s+Math.exp(l*b-l*l*v/8),0)/rates.length;
      if(sum>=threshold)hi=b;else lo=b;
    }
    near(saved.lower,Math.max(-1,m-hi/n));near(saved.upper,Math.min(1,m+hi/n));intervals++;
  }
}
for(const s of result.summaries){
  const rs=result.runs.filter(r=>r.candidate===s.candidate&&r.method===s.method),average=f=>rs.reduce((x,r)=>x+f(r),0)/rs.length;
  near(s.meanEstimate,average(r=>r.mean));near(s.rmse,Math.sqrt(average(r=>r.finalError**2)));near(s.meanWidth,average(r=>r.upper-r.lower));
}
console.log(JSON.stringify({status:'PASS',increments,finalIntervals:intervals,scope:'Independent byte decoding, phase0 regressions, per-action DR value difference, predictable endpoint ranges and final mixture inversion; per-prefix coverage checked by experiment replay, not independently reconstructed',hashes:{[resultPath]:sha(resultPath),[projectionPath]:sha(projectionPath),[import.meta.filename]:sha(import.meta.filename)}}));
