import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {contextMixture} from './context-mixture.mjs';
import {pointwiseGuard} from './pointwise-brier-guard.mjs';
const[source,original,input,output]=process.argv.slice(2);assert(source&&original&&input&&output);
const data=JSON.parse(fs.readFileSync(input));assert.equal(data.summary.cohort,'consumed-context-mixture-v1');assert.equal(data.results.length,672);
let state=2463534242;const rnd=()=>{state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;};
const mean=v=>v.reduce((a,b)=>a+b,0)/v.length;
function ci(v){const m=[];for(let i=0;i<10000;i++)m.push(mean(v.map(()=>v[Math.floor(rnd()*v.length)])));m.sort((a,b)=>a-b);return[m[249],m[9749]];}
const comparisons=[];
for(const regime of ['all','stationary','changing'])for(const period of ['whole','terminal64']){
  const rows=data.results.filter(r=>{const c=Number(r.key.split(':')[1]),ch=c<9?c%3!==0:c>=19;return regime==='all'||ch===(regime==='changing');});
  const metric=r=>period==='whole'?r.expected:r.expected.map((_,a)=>(r.blocks[6].expected[a]+r.blocks[7].expected[a])/2);
  for(const ref of [1,2,4]){
    const clusters=Array.from({length:8},(_,i)=>rows.filter(r=>Number(r.key.split(':')[2])===i));assert(clusters.every(c=>c.length===clusters[0].length));
    const v=clusters.map(c=>mean(c.map(r=>metric(r)[0]-metric(r)[ref])));
    comparisons.push({regime,period,reference:data.summary.arms[ref].arm,delta:mean(v),pointwise95:ci(v)});
  }
}
const old=JSON.parse(fs.readFileSync(original)),map=new Map(old.results.map(r=>[r.key,r])),tapes=[];let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const s=JSON.parse(line);if(header){assert.equal(s.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[s.Phase,s.Case,s.Index,s.Schedule].join(':'),r=map.get(key);assert(r);map.delete(key);
  tapes.push(s.Steps.map((v,i)=>({x:v.X,b:v.P[12],c:r.armPredictions[6][i],y:v.Y,delay:v.Delay,missing:v.Missing})));
}
assert.equal(tapes.length,672);assert.equal(map.size,0);const benchmark=[];let checksum=0;
for(const context of [false,true]){
  const run=()=>{for(const rows of tapes){const ws=contextMixture(rows,{context});for(let i=0;i<256;i++)checksum+=pointwiseGuard(rows[i].b,rows[i].c,ws[i].w).p;}};
  run();const rounds=[];
  for(let j=0;j<3;j++){const start=performance.now();run();const milliseconds=performance.now()-start;rounds.push({milliseconds,microsecondsPerForecast:milliseconds*1000/172032});}
  benchmark.push({context,rounds});
}
const report={comparisons,benchmark,checksum,limitations:'Consumed-data pointwise eight-index intervals, not independent confirmation. Warm 10D normal-system rebuild/solve plus guard benchmark; excludes original fitting, I/O, persistence and serving. Reference O(T^2+TWp^2+Tp^3), W=64,p=10.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
