import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
const quantile=(a,p)=>[...a].sort((a,b)=>a-b)[Math.ceil(a.length*p)-1]/1e6;
export function summarize(c){
 const t=c.trace;assert.equal(t.functional,true);assert.equal(t.frontier,c.frontier);
 for(const[key,n]of[['recall_ns',64],['feedback_ns',64],['live_age_ns',64],['capture_ns',256],['recall_start_ns',64],['recall_end_ns',64],['capture_start_ns',256],['capture_end_ns',256],['candidate_counts',64],['packed_counts',64]]){assert.equal(t[key].length,n);assert.ok(t[key].every(x=>Number.isSafeInteger(x)&&x>=0));}
 assert.ok(t.candidate_counts.every(x=>x===c.frontier));assert.ok(t.packed_counts.every(x=>x>0&&x<=10));assert.equal(t.completed,64);assert.equal(t.replay_completed,64);assert.equal(t.ledger_rows,128);
 assert.ok(t.capture_bytes>0&&t.init_bytes>0);assert.ok(Number.isSafeInteger(t.init_ns)&&t.init_ns>0);assert.ok(Number.isSafeInteger(t.replay_ns)&&t.replay_ns>=0);
 for(const name of['recall','capture'])for(let i=0;i<t[name+'_ns'].length;i++){assert.ok(t[name+'_end_ns'][i]>=t[name+'_start_ns'][i]);assert.equal(t[name+'_end_ns'][i]-t[name+'_start_ns'][i],t[name+'_ns'][i]);if(i)assert.ok(t[name+'_start_ns'][i]>=t[name+'_end_ns'][i-1]);}
 let overlaps=0;for(let j=0;j<256;j++)if(t.recall_start_ns.some((r,i)=>t.capture_start_ns[j]<t.recall_end_ns[i]&&t.capture_end_ns[j]>r))overlaps++;
 assert.equal(t.overlaps,overlaps);assert.ok(overlaps>0);
 const metrics=Object.fromEntries(['recall','capture','feedback','live_age'].map(n=>[n,Object.fromEntries([['p50',.5],['p95',.95],['p99',.99],['max',1]].map(([k,p])=>[k+'_ms',quantile(t[n+'_ns'],p)]))]));
 const absolute=metrics.recall.p99_ms<100&&metrics.live_age.p99_ms<250;
 return{arm:c.arm,k:c.frontier,pair:c.pair,metrics,initMS:t.init_ns/1e6,replayMS:t.replay_ns/1e6,wallMS:c.wallMS,overlaps,captureBytes:t.capture_bytes,initBytes:t.init_bytes,completed:64,captures:256,ledgerRows:128,absolute,status:c.status};
}
export function evaluate(root){
 const run=JSON.parse(fs.readFileSync(path.join(root,'run-results.json')));assert.equal(run.completed,true);assert.equal(run.preflightPassed,true);
 const commands=run.commands.filter(c=>c.kind==='load');assert.equal(commands.length,16);
 const rows=commands.map(c=>{assert.equal(c.functional,true);assert.equal(c.signal,null);assert.equal(c.error,null);assert.equal(c.dataRaceReported,false);const b=fs.readFileSync(path.join(root,c.transcript));assert.equal(crypto.createHash('sha256').update(b).digest('hex'),c.sha256);assert.equal(b.length,c.bytes);const matches=[...b.toString().matchAll(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/g)];assert.equal(matches.length,1);assert.deepEqual(JSON.parse(matches[0][1]),c.trace);return summarize(c);});
 const keys=new Set(rows.map(x=>x.k+'/'+x.pair+'/'+x.arm));assert.equal(keys.size,16);
 const contrasts=[];for(const k of[50,200])for(const pair of[0,1])for(const[reference,candidate]of[['rebuild','overlay'],['rebuild','shard'],['rebuild','both'],['shard','both'],['overlay','both']]){
  const a=rows.find(x=>x.k===k&&x.pair===pair&&x.arm===reference),b=rows.find(x=>x.k===k&&x.pair===pair&&x.arm===candidate),ratio=b.metrics.recall.p99_ms/a.metrics.recall.p99_ms;
  contrasts.push({k,pair,reference,candidate,recallP99Ratio:ratio,nonRegression:ratio<=1.10,rescueImprovement:ratio<=.90});
 }
 const arms=Object.fromEntries(['rebuild','overlay','shard','both'].map(a=>[a,{absolute:rows.filter(x=>x.arm===a).every(x=>x.absolute),nonRegression:a==='rebuild'?null:contrasts.filter(x=>x.reference==='rebuild'&&x.candidate===a).every(x=>x.nonRegression),rescueImprovement:a==='rebuild'?null:contrasts.filter(x=>x.reference==='rebuild'&&x.candidate===a).every(x=>x.rescueImprovement)}]));
 return{verified:true,rows,contrasts,arms,closedLoopOnly:true,p99IsSampleMaximum:true,peakRSS:'not captured',wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url))){const root=path.dirname(fileURLToPath(import.meta.url)),x=evaluate(root);fs.writeFileSync(path.join(root,'load-evaluation.json'),JSON.stringify(x,null,2)+'\n');console.log(JSON.stringify({verified:x.verified,arms:x.arms}));}
