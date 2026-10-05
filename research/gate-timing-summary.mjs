import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...cells]=bytes.toString().trim().split('\n').map(JSON.parse);
assert.equal(header.Kind,'header');assert.equal(cells.length,12);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
for(const [path,digest]of Object.entries(header.Hashes)) {
  assert.equal(hash(header.Sources[path]),digest);assert.equal(hash(fs.readFileSync(path)),digest);
}
const sum=v=>v.reduce((a,b)=>a+b,0), percentile=(v,p)=>v.length?[...v].sort((a,b)=>a-b)[Math.ceil(p*v.length)-1]/1e6:null;
const summaries=cells.map(c=>{
  const r=c.Result,spans=c.Spans??[];assert.equal(c.Dropped,0);assert.equal((r.Errors??[]).length,0);
  assert.equal(r.ReadSchedule.length,192);assert.equal(r.WriteSchedule.length,96);
  for(const calls of [r.ReadSchedule,r.WriteSchedule])for(const s of calls)assert(s.EndNS>=s.StartNS&&s.StartNS>=s.DueNS);
  assert.equal(new Set(r.WriteSchedule.map(s=>s.Index)).size,96);
  if(!c.Measured)assert.equal(spans.length,0);
  const gates=[...new Set(spans.map(s=>s.Kind))].sort().map(kind=>{
    const all=spans.filter(s=>s.Kind===kind),entered=all.filter(s=>s.Entered);
    for(const s of all)assert(s.WaitNS>=0&&s.HeldNS>=0&&(s.Entered||s.HeldNS===0));
    return {kind,count:all.length,entered:entered.length,waitTotalMS:sum(all.map(s=>s.WaitNS))/1e6,heldTotalMS:sum(entered.map(s=>s.HeldNS))/1e6,waitP99MS:percentile(all.map(s=>s.WaitNS),.99),heldP99MS:percentile(entered.map(s=>s.HeldNS),.99)};
  });
  if(c.Measured)assert.equal(gates.find(g=>g.kind==='ingestion').entered,96);
  const callback=sum((r.Phases??[]).map(p=>p.CallbackNS));
  if(c.Measured&&c.Active)assert(gates.find(g=>g.kind==='asof').heldTotalMS*1e6>=callback-1);
  return {trial:c.Trial,active:c.Active,measured:c.Measured,accepted:r.Accepted,dropped:r.Dropped,expired:r.WaitExpired,admits:r.Admits,discards:r.Discards,
    scheduledReadP99MS:percentile(r.ReadSchedule.map(s=>s.EndNS-s.DueNS),.99),scheduledWriteP99MS:percentile(r.WriteSchedule.map(s=>s.EndNS-s.DueNS),.99),ageP95MS:percentile(r.AgeNS??[],.95),callbackTotalMS:callback/1e6,admissionTotalMS:sum((r.Phases??[]).map(p=>p.DurableAdmitNS))/1e6,gates};
});
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,summaries,
  limitations:'Direct bounded telemetry; durations include instrumentation cost. Rotated finite cells, not a latency non-inferiority guarantee. Idle controls retain the wrapper; not the earlier bare-native control. Cold observations, not training. Held totals and callback/admission totals overlap and must not be added.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(out,null,2));
