import fs from 'node:fs';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {pointwiseGuard,testPointwiseGuard} from './pointwise-brier-guard.mjs';
testPointwiseGuard();
const[input,output]=process.argv.slice(2);assert(input&&output);
const x=JSON.parse(fs.readFileSync(input));assert.equal(x.results.length,672);
const rows=x.results.flatMap(r=>r.forecasts.map((f,i)=>[r.armPredictions[2][i],r.armPredictions[6][i],f.proposed]));
let checksum=0;
const run=()=>{for(const [b,c,w]of rows)checksum+=pointwiseGuard(b,c,w).p;};
for(let i=0;i<3;i++)run();
const rounds=[];
for(let i=0;i<3;i++){const start=performance.now();for(let j=0;j<10;j++)run();const milliseconds=performance.now()-start;rounds.push({milliseconds,forecasts:rows.length*10,microsecondsPerForecast:1000*milliseconds/(rows.length*10)});}
const report={rounds,checksum,limitations:'Warm scalar pointwise guard only, measured over stored as-of proposals. Excludes Fixed Share filtering, all fitting, retrieval, persistence and loaded serving. O(1) arithmetic per forecast.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
