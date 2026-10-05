import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const sha256 = value => crypto.createHash('sha256').update(value).digest('hex');
const [streamPath, zeroPath, output] = process.argv.slice(2);
assert.ok(streamPath && zeroPath && output);
const streamRaw = fs.readFileSync(streamPath);
const zeroRaw = fs.readFileSync(zeroPath);
assert.equal(sha256(streamRaw), 'c9a21a020a2321ca46481d0c145205cd8e3f78349cd4475494f87d72fd337451');
const stream = JSON.parse(streamRaw);
const zero = JSON.parse(zeroRaw);
assert.equal(stream.records.length, 256);
assert.equal(zero.records.length, 256);
assert.equal(zero.lambda, .5);
assert.equal(zero.threshold, 20);

function firstCrossing(scores, margin) {
  let wealth = 1;
  for (let j = 0; j < scores.length; j++) {
    const s = scores[j];
    const w = s.referenceCorrect - s.liveCorrect;
    assert.ok(Number.isInteger(w) && w >= -1 && w <= 1);
    const multiplier = 1 + .5 * (w - margin);
    assert.ok(multiplier >= 0);
    wealth *= multiplier;
    if (wealth >= 20) return {clock:s.release,origin:s.origin,auditUpdates:j+1,wealth};
  }
  return null;
}

assert.equal(firstCrossing([{origin:0,release:0,referenceCorrect:1,liveCorrect:0}],.15),null);
assert.equal(firstCrossing(Array.from({length:9},(_,i)=>({origin:i,release:i,referenceCorrect:1,liveCorrect:0})),.15)?.auditUpdates,9);
assert.equal(firstCrossing(Array.from({length:32},(_,i)=>({origin:i,release:i,referenceCorrect:0,liveCorrect:1})),.15),null);

const records=[];
for(let i=0;i<stream.records.length;i++){
  const source=stream.records[i],old=zero.records[i];
  for(const key of ['phase','case','index','schedule'])assert.equal(source[key],old[key]);
  const zeroCheck=firstCrossing(source.scores,0);
  assert.deepEqual(zeroCheck,old.nomination);
  const margin=firstCrossing(source.scores,.15);
  if(margin&&zeroCheck)assert.ok(margin.clock>=zeroCheck.clock);
  records.push({phase:source.phase,case:source.case,index:source.index,schedule:source.schedule,
    externalSplitClock:source.split,zeroGap:zeroCheck,margin015:margin,
    changed:source.case.includes('_to_')});
}
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of ['stable_majority3','stable_parity4','majority_to_parity','parity_to_majority']){
  for(const schedule of ['Immediate','Delayed']){
    const rows=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);
    assert.equal(rows.length,16);
    const mean=a=>a.length?a.reduce((s,x)=>s+x,0)/a.length:null;
    cells.push({phase,case:name,schedule,zeroBy511:rows.filter(r=>r.zeroGap&&r.zeroGap.clock<=511).length,
      marginBy511:rows.filter(r=>r.margin015&&r.margin015.clock<=511).length,
      marginBy543:rows.filter(r=>r.margin015).length,
      marginBeforeChange:rows.filter(r=>r.changed&&r.margin015&&r.margin015.clock<256).length,
      meanMarginClockAmongDetected:mean(rows.flatMap(r=>r.margin015?[r.margin015.clock]:[]))});
  }
}
const result={streamSHA256:sha256(streamRaw),zeroGapSHA256:sha256(zeroRaw),scriptSHA256:sha256(fs.readFileSync(new URL(import.meta.url))),
  lambda:.5,margin:.15,threshold:20,cells,records,
  limits:'Posthoc comparison on consumed synthetic trajectories. A one-sided scalar margin e-process is not the external Anti-Pigeon contract or a target-law diameter certificate. The zero-gap nomination cannot authorize splitting. No fresh confirmation or forecast benefit.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells));
