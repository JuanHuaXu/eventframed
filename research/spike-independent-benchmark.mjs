import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {runBudget} from './spike-budget-core.mjs';
import {delayedShare} from './delayed-fixed-share.mjs';
const [source,compact,output]=process.argv.slice(2);assert(source&&compact&&output);
const key=r=>[r.Phase,r.Case,r.Index,r.Schedule].join(':');
const candidates=new Map(fs.readFileSync(compact,'utf8').trim().split('\n').map(JSON.parse).map(r=>[key(r),r]));
const tapes=[];let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  const c=candidates.get(key(r));assert(c);candidates.delete(key(r));
  tapes.push(c.P.map((p,i)=>({b:r.Steps[i].P[12],c:p,y:r.Steps[i].Y,delay:r.Steps[i].Delay,missing:r.Steps[i].Missing})));
}
assert.equal(tapes.length,672);assert.equal(candidates.size,0);
let checksum=0;
function run(rows){
  const weights=delayedShare(rows);
  for(let c=0;c<256;c+=32){const f=runBudget(rows.slice(c,c+32),weights.slice(c,c+32));checksum+=f[31].p;}
}
for(const t of tapes)run(t);
const rounds=[];
for(let repeat=0;repeat<3;repeat++){
  const durations=[],start=performance.now();
  for(const tape of tapes){const s=performance.now();run(tape);durations.push(performance.now()-s);}
  const total=performance.now()-start;durations.sort((a,b)=>a-b);
  rounds.push({milliseconds:total,amortizedMicrosecondsPerForecast:1000*total/(672*256),trajectoryP50ms:durations[335],trajectoryP95ms:durations[638],trajectoryMaxMs:durations.at(-1)});
}
const result={records:672,forecastsPerRound:172032,rounds,checksum,limitations:'Warm in-memory 256-step trajectory replay of the frozen filter and local guard only. Includes reference O(T^2) filtering, excludes all model fits, retrieval, persistence, network and loaded serving. Amortized per-forecast values are not online tail latency.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(result,null,2));
