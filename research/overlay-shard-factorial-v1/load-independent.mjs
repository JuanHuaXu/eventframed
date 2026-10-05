import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';

// Independent arithmetic: no primary imports and raw traces, not receipt means.
export function independent(root){
 const run=JSON.parse(fs.readFileSync(path.join(root,'run-results.json'))),primary=JSON.parse(fs.readFileSync(path.join(root,'load-evaluation.json')));
 let samples=0,commands=0;const near=(a,b)=>assert.ok(Math.abs(a-b)<=1e-12);
 for(const c of run.commands){if(c.kind!=='load')continue;const text=fs.readFileSync(path.join(root,c.transcript),'utf8'),start=text.indexOf('PUBLIC_CAPTURE_TRACE=');assert.ok(start>=0);const t=JSON.parse(text.slice(start+'PUBLIC_CAPTURE_TRACE='.length).split('\n')[0]);
  const row=primary.rows.find(x=>x.arm===c.arm&&x.k===c.frontier&&x.pair===c.pair);assert.ok(row);
  for(const key of['recall','capture','feedback','live_age']){
   const values=t[key+'_ns'];assert.equal(values.length,key==='capture'?256:64);assert.ok(values.every(Number.isSafeInteger));
   const sorted=values.toSorted((a,b)=>a-b);for(const[k,p]of[['p50',.5],['p95',.95],['p99',.99],['max',1]])near(sorted[Math.ceil(p*values.length)-1]/1e6,row.metrics[key][k+'_ms']);samples+=values.length;
  }
  let overlap=0;for(let i=0;i<t.capture_ns.length;i++){let hit=false;for(let j=0;j<t.recall_ns.length;j++)hit||=t.capture_start_ns[i]<t.recall_end_ns[j]&&t.capture_end_ns[i]>t.recall_start_ns[j];overlap+=Number(hit);}assert.equal(overlap,row.overlaps);
  assert.equal(row.absolute,row.metrics.recall.p99_ms<100&&row.metrics.live_age.p99_ms<250);assert.equal(t.completed,64);assert.equal(t.replay_completed,64);assert.equal(t.ledger_rows,128);commands++;
 }
 assert.equal(commands,16);assert.equal(samples,7168);
 for(const x of primary.contrasts){const a=primary.rows.find(r=>r.k===x.k&&r.pair===x.pair&&r.arm===x.reference),b=primary.rows.find(r=>r.k===x.k&&r.pair===x.pair&&r.arm===x.candidate);near(b.metrics.recall.p99_ms/a.metrics.recall.p99_ms,x.recallP99Ratio);assert.equal(x.nonRegression,x.recallP99Ratio<=1.10);assert.equal(x.rescueImprovement,x.recallP99Ratio<=.90);}
 return{verified:true,commands,latencySamples:samples,contrasts:primary.contrasts.length,independentArithmetic:true,wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){const root=path.dirname(fileURLToPath(import.meta.url)),x=independent(root);fs.writeFileSync(path.join(root,'load-independent-results.json'),JSON.stringify(x,null,2)+'\n');console.log(JSON.stringify(x));}
