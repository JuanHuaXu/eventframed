import fs from 'node:fs';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {runBudget} from './spike-budget-core.mjs';
import {delayedShare,testDelayedShare} from './delayed-fixed-share.mjs';
testDelayedShare();
const rows=Array.from({length:256},(_,i)=>({b:.1+.1*(i%8),c:.85-.05*(i%9),y:i%3===0,delay:i%32,missing:i%5===0}));
const fns={staticGlobal:()=>runBudget(rows),shareOnly:()=>delayedShare(rows),shareGlobal:()=>runBudget(rows,delayedShare(rows)),shareLocal:()=>{const w=delayedShare(rows);let out=[];for(let c=0;c<256;c+=32)out.push(...runBudget(rows.slice(c,c+32),w.slice(c,c+32)));return out;}};
const results={};let sink=0;
for(const [name,fn]of Object.entries(fns)){
  for(let j=0;j<20;j++)fn();
  results[name]=[];
  for(let trial=0;trial<3;trial++){const start=performance.now();for(let j=0;j<200;j++){const r=fn();sink+=typeof r[255]==='number'?r[255]:r[255].p;}const milliseconds=(performance.now()-start)/200;results[name].push({millisecondsPerTrajectory:milliseconds,amortizedMicrosecondsPerForecast:milliseconds*1000/256});}
}
assert(Number.isFinite(sink));const output=process.argv[2];assert(output);
fs.writeFileSync(output,JSON.stringify({results,limitation:'Single-process warm256-step fixture, not loaded/tail serving latency. Fit costs excluded. This measures O(T^2) reference methods; the research scorer also computes a discarded global guard for the local-ledger ablation.'},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(results,null,2));
