import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

// A scheduling diagnostic only: no outcomes, forecasts, or quality are used.
function demand(frames, expireAge = null) {
  const busy = [];
  let issued = 0, completed = 0, expired = 0, peak = 0;
  for (let clock=0;clock<frames.length+32;clock++) {
    if(clock<frames.length) {
      let slot=busy.indexOf(null);
      if(slot<0) { slot=busy.length;busy.push(null); }
      busy[slot]=clock;issued++;
      peak=Math.max(peak,busy.filter(x=>x!==null).length);
    }
    // Feedback occurs after prediction, matching the captured driver.
    for(let slot=0;slot<busy.length;slot++) {
      const origin=busy[slot];if(origin===null)continue;
      const f=frames[origin];
      if(!f.Missing&&f.Arrival===clock) { busy[slot]=null;completed++; }
      else if(expireAge!==null&&clock-origin>=expireAge) { busy[slot]=null;expired++; }
    }
  }
  const pending=busy.filter(x=>x!==null).length;
  assert.equal(issued,completed+expired+pending);
  assert.equal(busy.length,peak);
  return {copies:busy.length,completed,expired,pending};
}

assert.equal(demand(Array.from({length:64},(_,i)=>({Arrival:i,Missing:false}))).copies,1);
assert.equal(demand(Array.from({length:64},(_,i)=>({Arrival:i+3,Missing:false}))).copies,4);
const absent=Array.from({length:64},(_,i)=>({Arrival:i,Missing:true}));
assert.deepEqual(demand(absent),{copies:64,completed:0,expired:0,pending:64});
assert.deepEqual(demand(absent,32),{copies:33,completed:0,expired:64,pending:0});

const [input,output]=process.argv.slice(2);
const raw=fs.readFileSync(input);
const hash=crypto.createHash('sha256').update(raw).digest('hex');
assert.equal(hash,'1533a049c2146cc7839799e9a394691d96c3fdf4a1a8e8d1c504db86053bf8e6');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
assert.equal(header.Consumed,true);assert.equal(rows.length,192);
const results=[];
for(const r of rows)for(const schedule of ['Immediate','Delayed']) {
  const frames=r[schedule].Frames;
  assert.equal(frames.length,512);
  for(let i=0;i<frames.length;i++)assert.ok(frames[i].Arrival>=i&&frames[i].Arrival<i+32);
  const original=demand(frames),bounded=demand(frames,32);
  const missing=frames.filter(f=>f.Missing).length;
  assert.equal(original.pending,missing);assert.equal(original.completed,512-missing);
  assert.equal(bounded.expired,missing);assert.equal(bounded.pending,0);
  results.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,missing,original,bounded});
}
assert.equal(new Set(results.map(r=>[r.phase,r.case,r.index,r.schedule].join('/'))).size,384);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const summaries=['Immediate','Delayed'].map(schedule=>{
  const a=results.filter(r=>r.schedule===schedule);
  return {schedule,n:a.length,meanMissing:mean(a.map(r=>r.missing)),
    originalCopies:{min:Math.min(...a.map(r=>r.original.copies)),mean:mean(a.map(r=>r.original.copies)),max:Math.max(...a.map(r=>r.original.copies))},
    boundedCopies:{min:Math.min(...a.map(r=>r.bounded.copies)),mean:mean(a.map(r=>r.bounded.copies)),max:Math.max(...a.map(r=>r.bounded.copies))}};
});
const result={sourceHash:hash,summaries,results,limits:'Consumed schedule-only diagnostic. Expiration is a modified protocol; no inherited regret or quality improvement claim.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({sourceHash:hash,summaries}));
