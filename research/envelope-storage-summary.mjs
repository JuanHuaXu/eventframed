import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const [input, output] = process.argv.slice(2);
assert(input && output);
const bytes = fs.readFileSync(input);
const [header, ...cells] = bytes.toString().trim().split('\n').map(JSON.parse);
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind, 'header');
assert.equal(cells.length, 18);
for (const [name, digest] of Object.entries(header.Hashes)) {
  assert.equal(hash(header.Sources[name]), digest);
  assert.equal(hash(fs.readFileSync(name)), digest);
}
assert.equal(new Set(cells.map(c => [c.Trial,c.Size,c.Mode].join(':'))).size,18);
const mean = v => v.reduce((a,b)=>a+b,0)/v.length/1e6;
const p95 = v => [...v].sort((a,b)=>a-b)[Math.ceil(.95*v.length)-1]/1e6;
const results = cells.map(c => {
  assert([0,1,2].includes(c.Trial)); assert([50,200].includes(c.Size));
  assert(['rows','envelope','staged'].includes(c.Mode));
  assert.equal(c.Batches,32);
  assert.equal(c.PrepareNS.length,32); assert.equal(c.FinalNS.length,32);
  assert.equal(c.Verified,32*(c.Mode==='rows'?c.Size:c.Mode==='staged'?2:1));
  assert(c.FinalNS.every(x=>Number.isSafeInteger(x)&&x>0));
  assert(c.PrepareNS.every(x=>Number.isSafeInteger(x)&&(c.Mode==='staged'?x>0:x===0)));
  const total = c.FinalNS.map((n,i)=>n+c.PrepareNS[i]);
  return {trial:c.Trial,size:c.Size,mode:c.Mode,verified:c.Verified,
    prepareMeanMS:mean(c.PrepareNS), finalMeanMS:mean(c.FinalNS),
    totalMeanMS:mean(total), finalP95MS:p95(c.FinalNS), totalP95MS:p95(total)};
});
const comparisons = [];
for(const trial of [0,1,2])for(const size of [50,200]){
  const get=mode=>results.find(c=>c.trial===trial&&c.size===size&&c.mode===mode);
  const rows=get('rows'), envelope=get('envelope'), staged=get('staged');
  comparisons.push({trial,size,envelopeVsRows:envelope.totalMeanMS/rows.totalMeanMS,
    stagedTotalVsRows:staged.totalMeanMS/rows.totalMeanMS,
    markerVsRows:staged.finalMeanMS/rows.totalMeanMS});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,results,comparisons,
  limitations:'Storage-only synthetic surrogate, no service identity index or policy validation; append intervals exclude encoding/digest, source lookup, learner staging and replay. Marker is not a valid admission protocol. No daemon speedup or power-loss proof.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(out,null,2));
