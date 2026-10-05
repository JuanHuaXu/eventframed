import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [projectionPath,resultPath]=process.argv.slice(2),projection=JSON.parse(fs.readFileSync(projectionPath)),result=JSON.parse(fs.readFileSync(resultPath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');for(const [p,h]of Object.entries(result.hashes))assert.equal(sha(p),h);
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10,`${a} != ${b}`);
const train=projection.records.filter(r=>r.phase===0&&r.schedule===1),test=projection.records.filter(r=>r.phase===1&&r.schedule===1);
function coin(rep,r,candidate,design){const bytes=crypto.createHash('sha256').update(['mmm-paired-audit-v1',rep,r.phase,r.case,r.index,candidate,design].join(':')).digest();return bytes[0]&1;}
const rates=[128,64,32,16,8,4,2].map(n=>1/n),psi=rates.map(l=>{let s=0;for(let k=2;k<=80;k++)s+=l**k/k;return s;}),threshold=2/(.05/8);
const mixture=(b,v)=>rates.reduce((s,l,i)=>s+Math.exp(l*b-psi[i]*v),0)/rates.length;
let prefixes=0,finals=0;
assert.equal(result.runs.length,512);
for(const saved of result.runs){
  const candidate=saved.candidate==='random'?0:2,ids=[candidate,1],sums=[.5,.5],counts=[1,1];let ds=0,dn=1,tc=0,nonidentTrain=0;
  for(const r of train){
    if(r.origins[candidate]===r.origins[1])continue;nonidentTrain++;
    const choice=coin(saved.rep,r,candidate,'single');sums[choice]+=r.losses[ids[choice]];counts[choice]++;
    if(coin(saved.rep,r,candidate,'paired')){ds+=r.losses[1]-r.losses[candidate];dn++;}
  }
  const models=sums.map((s,i)=>s/counts[i]),md=ds/dn;
  models.forEach((m,i)=>near(m,saved.models[i]));near(md,saved.pairModel);assert.deepEqual(saved.trainingCounts,counts.map(n=>n-1));assert.equal(saved.pairTrainingCount,dn-1);
  tc=saved.design==='single'?nonidentTrain:2*(dn-1);assert.equal(saved.trainingCost,tc);assert.equal(saved.expectedTrainingCost,nonidentTrain);
  let sum=0,v=0,truth=0,nonident=0,cost=0,allExact=true,ever=false;const positive=[];
  for(let i=0;i<test.length;i++){
    const r=test[i],difference=r.losses[1]-r.losses[candidate],same=r.origins[candidate]===r.origins[1];let x=0;
    if(same)assert.equal(difference,0);
    else{
      allExact=false;nonident++;
      if(saved.design==='single'){
        const a=coin(saved.rep,r,candidate,'single'),m=saved.method==='DR'?models:[0,0];
        x=m[1]-m[0]+(a===1?2:-2)*(r.losses[ids[a]]-m[a]);cost++;
      }else{
        const include=coin(saved.rep,r,candidate,'paired'),m=saved.method==='DR'?md:0;
        x=include?2*difference-m:m;if(include)cost+=2;
      }
    }
    const prediction=i?Math.max(-1,Math.min(1,sum/i)):0;v+=(x-prediction)**2/16;sum+=x;truth+=difference;prefixes++;
    if(allExact)near(sum,truth);
    else {ever ||= mixture(Math.abs(sum-truth)/4,v)>threshold+1e-10;if(sum>0&&mixture(sum/4,v)>threshold)positive.push(i+1);}
  }
  assert.equal(saved.evaluationCost,cost);assert.equal(saved.expectedEvaluationCost,nonident);assert.equal(saved.everMiss,ever);assert.deepEqual(saved.positiveAt,positive);
  near(saved.mean,sum/672);near(saved.truth,truth/672);near(saved.residualSquares,v);assert.equal(saved.n,672);assert.equal(saved.scale,4);
  let lo=0,hi=1e4;for(let i=0;i<100;i++){const b=(lo+hi)/2;if(mixture(b,v)>=threshold)hi=b;else lo=b;}
  near(saved.lower,Math.max(-1,sum/672-4*hi/672));near(saved.upper,Math.min(1,sum/672+4*hi/672));finals++;
}
for(const s of result.summaries){
  const rs=result.runs.filter(r=>r.candidate===s.candidate&&r.design===s.design&&r.method===s.method),mean=f=>rs.reduce((x,r)=>x+f(r),0)/rs.length;assert.equal(rs.length,64);
  near(s.rmse,Math.sqrt(mean(r=>(r.mean-r.truth)**2)));near(s.meanWidth,mean(r=>r.upper-r.lower));near(s.meanTrainingCost,mean(r=>r.trainingCost));near(s.meanEvaluationCost,mean(r=>r.evaluationCost));
  assert.deepEqual(s.evaluationCostRange,[Math.min(...rs.map(r=>r.evaluationCost)),Math.max(...rs.map(r=>r.evaluationCost))]);
}
console.log(JSON.stringify({status:'PASS',prefixes,finals,scope:'Independent coin extraction, logged regressions, HT/DR increments, query counts, every-prefix coverage and final series-based EB inversion',hashes:{[resultPath]:sha(resultPath),[projectionPath]:sha(projectionPath),[import.meta.filename]:sha(import.meta.filename)}}));
