import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

// Mirrors fit eligibility, not forecast or outcome generation. Labels are never
// inspected: only audit status, missingness, origin and arrival are needed.
export function reconstructFits(frames, end=544) {
  const batches=Array.from({length:end},()=>[]);
  frames.forEach((f,i)=>{
    assert.ok(Number.isInteger(f.Arrival)&&f.Arrival>=i&&f.Arrival<end);
    if(!f.Missing&&f.Audit)batches[f.Arrival].push(i);
  });
  const arrived=[],fits=[];
  for(let clock=0;clock<end;clock++){
    let fit=false;
    for(const i of batches[clock]){
      arrived.push(i);
      if(arrived.length>=32&&(arrived.length-32)%16===0)fit=true;
    }
    // Multiple triggers at one clock produce ONE publication using every audit
    // delivered at that clock, not just the count at the first trigger.
    if(fit){arrived.sort((a,b)=>a-b);fits.push({Clock:clock,Origins:arrived.slice(-256)});}
  }
  return fits;
}

function selfTest(){
  const f=Array.from({length:48},(_,i)=>({Arrival:i<16?31:47,Audit:true,Missing:false}));
  assert.deepEqual(reconstructFits(f,48),[{Clock:47,Origins:Array.from({length:48},(_,i)=>i)}]);
  const g=Array.from({length:48},(_,i)=>({Arrival:i,Audit:true,Missing:false}));
  assert.deepEqual(reconstructFits(g,48).map(x=>[x.Clock,x.Origins.length]),[[31,32],[47,48]]);
  g[0].Missing=true;
  assert.deepEqual(reconstructFits(g,48).map(x=>[x.Clock,x.Origins.length]),[[32,32]]);
  g[1].Audit=false;
  assert.deepEqual(reconstructFits(g,48).map(x=>[x.Clock,x.Origins.length]),[[33,32]]);
  assert.throws(()=>reconstructFits([{Arrival:-1,Audit:true}],2));
  assert.throws(()=>reconstructFits([{Arrival:2,Audit:true}],2));
}
selfTest();
if(process.argv[2]==='--test'){console.log('training exposure contracts PASS');process.exit(0);}
const [input,output]=process.argv.slice(2),raw=fs.readFileSync(input);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(hash(raw),'4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
assert.equal(rows.length,128);
const records=[];
for(const r of rows)for(const schedule of ['Immediate','Delayed']){
  const p=r[schedule],fits=reconstructFits(p.Frames);
  assert.deepEqual(fits,p.Fits);
  for(const f of fits)for(const i of f.Origins){assert.ok(p.Frames[i].Audit&&!p.Frames[i].Missing&&p.Frames[i].Arrival<=f.Clock);}
  const post=fits.filter(f=>f.Clock>=256&&f.Clock<511).map(f=>{
    const ids=f.Origins.slice(-64),newCount=ids.filter(i=>i>=256).length;
    return {clock:f.Clock,nextUsable:f.Clock+1,ids,count:ids.length,newCount,newFraction:newCount/ids.length,span:f.Clock-ids[0],event64Count:f.Origins.filter(i=>i>f.Clock-64).length};
  });
  let fi=-1;const serving=[];
  for(let clock=0;clock<512;clock++){
    while(fi+1<fits.length&&fits[fi+1].Clock<clock)fi++;
    if(clock<256)continue;
    const ids=fi<0?[]:fits[fi].Origins.slice(-64);
    serving.push({clock,fitClock:fi<0?null:fits[fi].Clock,count:ids.length,newFraction:ids.length?ids.filter(i=>i>=256).length/ids.length:null,oldestAge:ids.length?clock-ids[0]:null});
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,post,serving,firstPostFit:post[0]?.nextUsable??null,firstHalfNew:post.find(f=>f.newFraction>=.5)?.nextUsable??null,firstAllNew:post.find(f=>f.newFraction===1)?.nextUsable??null});
}
const mean=a=>a.length?a.reduce((s,x)=>s+x,0)/a.length:null;
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
  const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
  const times=field=>{const x=a.flatMap(r=>r[field]===null?[]:[r[field]]);return {observed:x.length,censored:16-x.length,meanAmongObserved:mean(x)};};
  // Average within each trajectory FIRST, not across fit counts.
  cells.push({phase,case:name,schedule,firstPostFit:times('firstPostFit'),firstHalfNew:times('firstHalfNew'),firstAllNew:times('firstAllNew'),meanFitSpan:mean(a.map(r=>mean(r.post.map(f=>f.span)))),meanEvent64Support:mean(a.map(r=>mean(r.post.map(f=>f.event64Count)))),meanServingNewFraction:mean(a.map(r=>mean(r.serving.flatMap(f=>f.newFraction===null?[]:[f.newFraction]))))});
}
const result={inputSHA256:hash(raw),scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),fitDriverSHA256:header.Hashes['internal/observationgate/credit_learning_run_test.go'],verifiedSchedules:records.length,cells,records,limits:'Consumed metadata-only exposure diagnosis. Origin256 is used only to label evaluator-known post-change evidence, never as a learner signal. Stable cases use the same artificial time cut as controls. No shorter-window predictions fitted, no accuracy or causation claim. Event64Support counts arrived audit labels, not all frames or selected full-monitor inputs.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({verifiedSchedules:records.length,changed:cells.filter(c=>c.case.includes('_to_'))},null,2));
