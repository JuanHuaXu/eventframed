import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const mean=a=>a.length?a.reduce((s,x)=>s+x,0)/a.length:null;
// This simulator has delay<=31. The watermark is metadata-only and fixed;
// never treat arbitrary production delays as bounded by this fixture constant.
function stream(frames,clock,delayed){
 const watermark=clock-(delayed?31:0);
 return frames.flatMap((f,origin)=>origin<=watermark&&f.Audit&&!f.Missing&&f.Arrival<=clock?[origin]:[]);
}
const test=[{Audit:true,Missing:false,Arrival:31},{Audit:true,Missing:false,Arrival:1},{Audit:true,Missing:true,Arrival:2}];
assert.deepEqual(stream(test,30,true),[]);
assert.deepEqual(stream(test,31,true),[0]);
assert.deepEqual(stream(test,32,true),[0,1]);
assert.deepEqual(stream(test,2,false),[1]);
const [input,output]=process.argv.slice(2),raw=fs.readFileSync(input);
assert.equal(hash(raw),'4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
assert.equal(rows.length,128);const records=[];
for(const r of rows)for(const [si,schedule]of ['Immediate','Delayed'].entries()){
 const tape=r[schedule],frames=tape.Frames;assert.equal(frames.length,512);
 let prior=[];for(let clock=0;clock<544;clock++){
  const ids=stream(frames,clock,si===1);assert.deepEqual(ids.slice(0,prior.length),prior);
  for(const i of ids){assert.ok(frames[i].Arrival<=clock);assert.ok(r.Acquisition[si][i].Full);}
  prior=ids;
 }
 const scores=prior.map(i=>({origin:i,arrival:frames[i].Arrival,release:i+(si===1?31:0),
  liveCorrect:Number((frames[i].Baseline>=.5)===frames[i].Y),referenceCorrect:Number((frames[i].Reference>=.5)===frames[i].RY)}));
 const split=tape.Arms[2].SplitAt;
 const snapshots=[128,256,384,480,511].map(clock=>{
  const ids=stream(frames,clock,si===1),available=frames.flatMap((f,i)=>i<=clock&&f.Audit&&!f.Missing&&f.Arrival<=clock?[i]:[]);
  return {clock,count:ids.length,availableCount:available.length,buffered:available.length-ids.length,post256:ids.filter(i=>i>=256).length};
 });
 const atSplit=split<0?null:stream(frames,split,si===1);
 const arrivedAtSplit=split<0?[]:frames.flatMap((f,i)=>i<=split&&f.Audit&&!f.Missing&&f.Arrival<=split?[i]:[]);
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,split,atSplitCount:atSplit?.length??null,
  gateHasLaterAuditLabels:split<0?null:arrivedAtSplit.some(i=>!atSplit.includes(i)),snapshots,scores});
}
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
 const split=a.filter(r=>r.split>=0);
 cells.push({phase,case:name,schedule,splits:split.length,withUnfinalizedAuditAtGate:split.filter(r=>r.gateHasLaterAuditLabels).length,
  meanAuditCountAtGate:mean(split.map(r=>r.atSplitCount)),snapshots:[0,1,2,3,4].map(j=>({clock:a[0].snapshots[j].clock,
   count:mean(a.map(r=>r.snapshots[j].count)),buffered:mean(a.map(r=>r.snapshots[j].buffered)),post256:mean(a.map(r=>r.snapshots[j].post256))}))});
}
const result={inputSHA256:hash(raw),driverSHA256:header.Hashes['internal/observationgate/credit_learning_run_test.go'],scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),cells,records,
 limits:'Offline candidate-stream feasibility only. Pseudorandom generator separation is a simulator design assumption, not an empirical independence proof. Full audit scores use frozen baseline; gate also uses non-audits and may observe buffered labels. No inherited stopping-time or exchangeability guarantee, localization certificate, efficacy, or production delay bound.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.case.includes('_to_')&&c.schedule==='Delayed')));
