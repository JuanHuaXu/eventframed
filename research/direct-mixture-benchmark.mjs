import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {directMixture} from './direct-mixture.mjs';
import {pointwiseGuard} from './pointwise-brier-guard.mjs';
const[source,input,output]=process.argv.slice(2);assert(source&&input&&output);
const old=JSON.parse(fs.readFileSync(input));const map=new Map(old.results.map(r=>[r.key,r]));let header=true;const tapes=[];
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const s=JSON.parse(line);if(header){assert.equal(s.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[s.Phase,s.Case,s.Index,s.Schedule].join(':'),r=map.get(key);assert(r);map.delete(key);
  tapes.push(s.Steps.map((v,i)=>({b:v.P[12],c:r.armPredictions[6][i],y:v.Y,delay:v.Delay,missing:v.Missing})));
}
assert.equal(tapes.length,672);assert.equal(map.size,0);
let checksum=0;const results=[];
for(const[name,config]of [['ridgeAll',{}],['ridge64',{window:64}],['currentCovAll',{currentCovariance:true}],['currentCov64',{window:64,currentCovariance:true}]]){
  const run=()=>{for(const rows of tapes){const ws=directMixture(rows,config);for(let i=0;i<256;i++)checksum+=pointwiseGuard(rows[i].b,rows[i].c,ws[i].w).p;}};
  run();const rounds=[];for(let r=0;r<3;r++){const start=performance.now();run();const milliseconds=performance.now()-start;rounds.push({milliseconds,microsecondsPerForecast:milliseconds*1000/172032});}
  results.push({name,rounds});
}
const report={results,checksum,limitations:'Warm reference 256-step trajectory replay including proposal and pointwise guard, excluding fitting, I/O and loaded serving. History scans are O(T^2), with origin arrays materialized; no O(1) implementation claim.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));
