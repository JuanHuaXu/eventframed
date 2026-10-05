import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';

async function* records(path) {
  for await (const line of createInterface({input:createReadStream(path),crlfDelay:Infinity})) {
    if (line.trim()) yield JSON.parse(line);
  }
}
const dir='docs/experiments/';
const source=records(dir+'mmm-soft-learners-v120.jsonl')[Symbol.asyncIterator]();
const predictions=records(process.argv[2]??dir+'mmm-spectral-regression-v1-forecasts.jsonl')[Symbol.asyncIterator]();
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const rows=[];
let forecastChecks=0, controlChecks=0, maxResidual=0;
for (;;) {
  const a=await source.next(),b=await predictions.next();
  assert.equal(a.done,b.done);if(a.done) break;
  const s=a.value,r=b.value;
  for(const key of ['Phase','Case','Index','Schedule']) assert.equal(r[key],s[key]);
  assert.equal(r.Fits,48);assert.equal(r.P.length,256);
  assert(r.MaxResidual<=1e-8);maxResidual=Math.max(maxResidual,r.MaxResidual);
  const metrics=Array.from({length:21},()=>Array.from({length:2},()=>({brier:0,accuracy:0,realized:0})));
  for(let t=0;t<256;t++) {
    const step=s.Steps[t],ps=[...r.P[t],...step.P];assert.equal(ps.length,21);
    assert(Number.isFinite(step.Q)&&step.Q>=0&&step.Q<=1);
    for(let arm=0;arm<ps.length;arm++) {
      const p=ps[arm];assert(Number.isFinite(p)&&p>0&&p<1);
      const brier=(p-step.Q)**2+step.Q*(1-step.Q);
      const accuracy=p>=.5?step.Q:1-step.Q;
      const realized=(p-Number(step.Y))**2;
      for(let span=0;span<2;span++) if(span===0||t>=192) {
        const n=span===0?256:64;
        metrics[arm][span].brier+=brier/n;
        metrics[arm][span].accuracy+=accuracy/n;
        metrics[arm][span].realized+=realized/n;
      }
      if(arm<6) forecastChecks++;
    }
  }
  for(let k=0;k<15;k++) for(let span=0;span<2;span++) {
    assert(Math.abs(metrics[k+6][span].brier-s.Metrics[k][span].Brier)<1e-13);
    assert(Math.abs(metrics[k+6][span].accuracy-s.Metrics[k][span].Accuracy)<1e-13);controlChecks+=2;
  }
  rows.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,metrics,features:r.FeatureTotal});
}
assert.equal(rows.length,2688);
const cells=new Map();
for(const r of rows) {const key=[r.phase,r.case,r.schedule].join(':');if(!cells.has(key)) cells.set(key,[]);cells.get(key).push(r);}
const bounds=xs=>{
 assert.equal(xs.length,32);
 const mean=xs.reduce((a,b)=>a+b,0)/32;
 const se=Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/(31*32));
 return {mean,lower:mean-3.5*se,upper:mean+3.5*se};
};
const gates=[],averages=[];
const recovery=new Set([1,2,4,5,7,8,19,20]);
for(const [key,rs] of cells) {
 assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const [phase,scenario,schedule]=key.split(':').map(Number);
 for(let arm=0;arm<21;arm++) for(let span=0;span<2;span++) {
  const vals=rs.map(r=>r.metrics[arm][span]);
  averages.push({phase,case:scenario,schedule,arm,span,
    brier:vals.reduce((s,r)=>s+r.brier,0)/32,
    accuracy:vals.reduce((s,r)=>s+r.accuracy,0)/32,
    realized:vals.reduce((s,r)=>s+r.realized,0)/32});
 }
 for(let arm=2;arm<6;arm++) {
  const window=arm%2;
  const controls=[['linear',window],['logistic',10+window],['generic',6+2*window],['boolean',7+2*window],['markov',18]];
  for(const [name,control] of controls) for(let span=0;span<2;span++) {
   const interval=bounds(rs.map(r=>r.metrics[arm][span].brier-r.metrics[control][span].brier));
   gates.push({phase,case:scenario,schedule,arm,control:name,span,type:'nonharm',...interval,pass:interval.upper<=.01});
  }
  if(recovery.has(scenario)) for(const [name,control] of controls.filter(c=>c[0]!=='boolean')) {
   const interval=bounds(rs.map(r=>r.metrics[control][1].brier-r.metrics[arm][1].brier));
   gates.push({phase,case:scenario,schedule,arm,control:name,span:1,type:'gain',...interval,pass:interval.mean>=.005&&interval.lower>0});
  }
 }
}
const verdicts=[];
for(let arm=2;arm<6;arm++) {
 const counts={};for(const type of ['nonharm','gain']) {
  const subset=gates.filter(g=>g.arm===arm&&g.type===type);
  counts[type]={passed:subset.filter(g=>g.pass).length,total:subset.length};
 }
 verdicts.push({arm,...counts,pass:gates.filter(g=>g.arm===arm).every(g=>g.pass)});
}
console.log(JSON.stringify({records:rows.length,forecastChecks,controlChecks,maxResidual,verdicts,averages,gates},null,2));
