import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const streamPath = process.argv[2], output = process.argv[3];
assert.ok(streamPath && output);
const raw = fs.readFileSync(streamPath);
assert.equal(hash(raw), 'c9a21a020a2321ca46481d0c145205cd8e3f78349cd4475494f87d72fd337451');
const stream = JSON.parse(raw);
assert.equal(stream.records.length, 256);

function nominate(scores) {
  let wealth = 1;
  for (let j = 0; j < scores.length; j++) {
    const item = scores[j];
    const w = item.referenceCorrect - item.liveCorrect;
    assert.ok(Number.isInteger(w) && w >= -1 && w <= 1);
    wealth *= 1 + 0.5 * w;
    if (wealth >= 20) {
      return {clock: item.release, origin: item.origin, auditUpdates: j + 1, wealth};
    }
  }
  return null;
}
assert.equal(nominate([{origin:0,release:0,liveCorrect:1,referenceCorrect:1}]), null);
assert.equal(nominate(Array.from({length:8},(_,i)=>({origin:i,release:i,liveCorrect:0,referenceCorrect:1})))?.auditUpdates, 8);
assert.equal(nominate(Array.from({length:20},(_,i)=>({origin:i,release:i,liveCorrect:1,referenceCorrect:0}))), null);

const records = [];
for (const r of stream.records) {
  const delayed = r.schedule === 'Delayed';
  assert.ok(delayed || r.schedule === 'Immediate');
  let last = -1;
  for (const s of r.scores) {
    assert.ok(s.origin > last && s.arrival >= s.origin && s.release === s.origin + (delayed ? 31 : 0));
    assert.ok(s.arrival <= s.release && s.release <= 542);
    last = s.origin;
  }
  const alarm = nominate(r.scores);
  const changed = r.case.includes('_to_');
  records.push({phase:r.phase,case:r.case,index:r.index,schedule:r.schedule,
    externalSplitClock:r.split,nomination:alarm,
    beforeTrueChange:changed && alarm!==null && alarm.clock<256,
    by511:alarm!==null && alarm.clock<=511,
    by543:alarm!==null && alarm.clock<=543,
    delay:changed && alarm!==null && alarm.clock>=256 ? alarm.clock-256 : null});
}

const mean = a => a.length ? a.reduce((s,x)=>s+x,0)/a.length : null;
const cells = [];
for (const phase of ['cohort1','cohort2']) for (const name of ['stable_majority3','stable_parity4','majority_to_parity','parity_to_majority']) {
  for (const schedule of ['Immediate','Delayed']) {
    const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);
    assert.equal(a.length,16);
    cells.push({phase,case:name,schedule,
      nominated:a.filter(r=>r.nomination).length,
      by511:a.filter(r=>r.by511).length,
      by543:a.filter(r=>r.by543).length,
      beforeTrueChange:a.filter(r=>r.beforeTrueChange).length,
      externalSplits:a.filter(r=>r.externalSplitClock>=0).length,
      meanDelayAmongDetected:mean(a.flatMap(r=>r.delay===null?[]:[r.delay])),
      meanAuditUpdatesAmongDetected:mean(a.flatMap(r=>r.nomination?[r.nomination.auditUpdates]:[])),
      meanClockAmongDetected:mean(a.flatMap(r=>r.nomination?[r.nomination.clock]:[]))});
  }
}
const result={inputSHA256:hash(raw),scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),lambda:.5,threshold:20,records,cells,
  limits:'Consumed synthetic nomination screen. Zero-gap e-process has a narrower null than Anti-Pigeon; no split authorization or onset confidence set. Idealized martingale assumptions require independent paired outcomes and predictable audit selection. Detection counts are not future predictive benefit or real-world error guarantees.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells));
