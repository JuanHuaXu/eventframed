import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';
const rows=[];
for await(const line of createInterface({input:createReadStream(process.argv[2]??'docs/experiments/mmm-degree-bfgs-v1.jsonl'),crlfDelay:Infinity}))if(line.trim())rows.push(JSON.parse(line));
assert.equal(rows.length,1008);
const seen=new Set(),groups=Array.from({length:4},()=>({n:0,converged:0,oldConverged:0,objectiveNonharm:0,objectiveGain:0,brierGain:0,realizedGain:0,newBrier:0,oldBrier:0,iterations:0,evaluations:0,projectedGradient:0,resets:0,patterns:0,stops:{}}));
let forecastChecks=0,traceChecks=0,maxObjectiveRegression=0;
const failures=[];
for(const r of rows){
 const key=[r.Phase,r.Case,r.Index,r.Schedule,r.Clock,r.Window,r.LearnNoise].join(':');assert(!seen.has(key));seen.add(key);
 const a=r.Stats,b=r.OldStats,group=groups[Number(r.LearnNoise)*2+r.Window];group.n++;
 assert(a.Iterations<=64&&a.Evaluations<=1345);assert.equal(a.Trace.length,a.Iterations+1);
 assert(a.Final<=a.Initial);assert.equal(a.Initial,b.Initial);
 for(let i=1;i<a.Trace.length;i++){assert(a.Trace[i]<=a.Trace[i-1]);traceChecks++;}
 assert(Math.abs(a.Final-a.Trace.at(-1))<1e-11);
 assert(Number.isFinite(a.ProjectedGradient)&&a.ProjectedGradient>=0);
 for(let i=0;i<5;i++){const lo=Math.log(i===4?.01:1e-4),hi=Math.log(i===4?4:16),tol=4*Number.EPSILON*Math.max(1,Math.abs(lo),Math.abs(hi));assert(Number.isFinite(a.Logs[i])&&a.Logs[i]>=lo-tol&&a.Logs[i]<=hi+tol);}
 if(!r.LearnNoise)assert.equal(a.Logs[4],0);
 const converged=a.Stop==='projected-gradient'&&a.ProjectedGradient<=1e-6;
 group.converged+=Number(converged);group.oldConverged+=Number(b.Stop==='projected-gradient'&&b.ProjectedGradient<=1e-6);
 group.objectiveNonharm+=Number(a.Final<=b.Final+1e-8);group.objectiveGain+=b.Final-a.Final;
 maxObjectiveRegression=Math.max(maxObjectiveRegression,a.Final-b.Final);
 if(!converged||a.Final>b.Final+1e-8)failures.push({key,stop:a.Stop,pg:a.ProjectedGradient,delta:a.Final-b.Final});
 group.iterations+=a.Iterations;group.evaluations+=a.Evaluations;group.projectedGradient+=a.ProjectedGradient;group.resets+=a.HessianResets;group.patterns+=a.QPPatterns;
 group.stops[a.Stop]=(group.stops[a.Stop]??0)+1;
 assert.equal(r.Predictions.length,32);assert.equal(r.OldPredictions.length,32);
 for(let i=0;i<32;i++){
  const p=r.Predictions[i],q=r.Q[i],old=r.OldPredictions[i],y=Number(r.Y[i]);
  assert([p,q,old].every(Number.isFinite)&&p>0&&p<1&&old>0&&old<1&&q>=0&&q<=1);
  const nb=(p-q)**2+q*(1-q),ob=(old-q)**2+q*(1-q);
  group.newBrier+=nb/32;group.oldBrier+=ob/32;group.brierGain+=(ob-nb)/32;group.realizedGain+=((old-y)**2-(p-y)**2)/32;forecastChecks++;
 }
}
const converged=groups.reduce((s,g)=>s+g.converged,0),objectiveNonharm=groups.reduce((s,g)=>s+g.objectiveNonharm,0);
for(const g of groups){assert.equal(g.n,252);for(const k of ['objectiveGain','brierGain','realizedGain','newBrier','oldBrier','iterations','evaluations','projectedGradient','resets','patterns'])g[k]/=g.n;}
console.log(JSON.stringify({records:rows.length,forecastChecks,traceChecks,converged,objectiveNonharm,maxObjectiveRegression,screenPass:converged/rows.length>=.95&&objectiveNonharm===rows.length,groups,failures},null,2));
