import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {guardedHead} from './guarded-head.mjs';
const[source,output]=process.argv.slice(2);assert(source&&output);
const tapes=[];let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  tapes.push([10,13].map(arm=>r.Steps.map(s=>({b:s.P[12],c:s.P[arm],y:s.Y,delay:s.Delay,missing:s.Missing}))));
}
assert.equal(tapes.length,672);const variants=[];let checksum=0;
for(let a=0;a<2;a++){
  const run=()=>{for(const pair of tapes){const r=guardedHead(pair[a]);checksum+=r.point[255]+r.local[255];}};
  run();const rounds=[];for(let j=0;j<3;j++){const start=performance.now();run();const milliseconds=performance.now()-start;rounds.push({milliseconds,microsecondsPerForecast:milliseconds*1000/172032});}
  variants.push({name:a===0?'segment64':'static64',rounds});
}
const report={variants,checksum,limitations:'Warm proposal plus BOTH guards on stored forecasts; excludes segment/static model fitting, I/O and loaded serving. Reference delayed filtering remains O(T^2). Not a latency comparison of model fitters.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
