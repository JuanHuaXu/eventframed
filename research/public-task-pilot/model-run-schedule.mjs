import fs from 'node:fs';
import assert from 'node:assert/strict';

// Discrete append-only accounting, deliberately without a fitted timing model.
// Physical runs must fit before publication: an eighth tail is not free merely
// because a later merge reduces the count. Base graphs remain untouched.
function simulate(batches,slots,maxMerge){
 const runs=[];let accepted=0,rewritten=0,maxBuilt=0;const events=[];
 for(let batch=1;batch<=batches;batch++){
  if(runs.length===slots)return {accepted,rewritten,maxBuilt,runs,stop:'no flush slot',batch,events};
  runs.unshift(32);accepted+=32;events.push({kind:'flush',batch,size:32});
  while(runs.length>=2&&runs[0]===runs[1]){
   const size=runs[0]+runs[1];
   if(size>maxMerge)return {accepted,rewritten,maxBuilt,runs,stop:'merge size cap',batch,events};
   runs.splice(0,2,size);rewritten+=size;maxBuilt=Math.max(maxBuilt,size);events.push({kind:'merge',batch,size});
  }
  assert.equal(runs.reduce((a,b)=>a+b,0),accepted);
 }
 return {accepted,rewritten,maxBuilt,runs,stop:null,events};
}
const ideal=simulate(128,8,Infinity);assert.equal(ideal.accepted,4096);assert.equal(ideal.rewritten,28672);assert.equal(ideal.maxBuilt,4096);
const bounded=simulate(128,7,Infinity);assert.equal(bounded.stop,'no flush slot');assert.equal(bounded.accepted,4064);
const capped=simulate(128,7,512);assert.equal(capped.stop,'merge size cap');assert.equal(capped.batch,32);
const scenarios=[];
for(const slots of [7,8])for(const cap of [512,1024,2048,null])scenarios.push({slots,maxMerge:cap,...simulate(128,slots,cap??Infinity)});
const output={model:'append-only binary tiering; no latency prediction',batchSize:32,batches:128,baseGraphs:8,pendingDeltaSlot:1,growingTailRewriteEntries:32*((128*129/2)-1),scenarios};
const path=process.argv[2];if(!path)throw Error('new output path required');fs.writeFileSync(path,JSON.stringify(output,null,2),{flag:'wx',mode:0o600});
for(const {events,...s}of scenarios)console.log(JSON.stringify(s));
