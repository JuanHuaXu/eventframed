import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const streams=process.argv.slice(2,4).map(p=>fs.createReadStream(p)),hashes=streams.map(()=>crypto.createHash('sha256'));
streams.forEach((s,i)=>s.on('data',b=>hashes[i].update(b)));
const iterators=streams.map(s=>createInterface({input:s,crlfDelay:Infinity})[Symbol.asyncIterator]());
const header=JSON.parse((await iterators[0].next()).value),rawHeader=JSON.parse((await iterators[1].next()).value);
assert(['acquisition-train-consumed-v1','acquisition-train-boundary-v1'].includes(header.version));assert.equal(rawHeader.Version,'soft-learners-v120');
const empty=()=>({queries:0,noFit:0,missingAdded:0,advanced:0,alreadyAvailable:0,fitWaitSum:0,fitWaitCount:0});
const global=new Map(),cells=new Map();let records=0;
for await(const line of iterators[0]){
 const r=JSON.parse(line),next=await iterators[1].next();assert(!next.done);const raw=JSON.parse(next.value);assert(!r.Error);
 assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[raw.Phase,raw.Case,raw.Index,raw.Schedule]);
 for(let policy=1;policy<4;policy++)for(let window=0;window<2;window++){
  const result=r.Results[policy];
  for(const q of result.Queries??[]){
   const fit=result.Fits.find(f=>f.Origins[window].includes(q.Origin));
   const s=raw.Steps[q.Origin],arrival=s.Missing?Infinity:q.Origin+s.Delay;
   let category='noFit',wait=null;
   if(fit){assert(q.Reveal<=fit.Clock);assert(q.Origin<fit.Clock);wait=fit.Clock-q.Reveal;category=s.Missing?'missingAdded':arrival>fit.Clock?'advanced':'alreadyAvailable';}
   const keys=[[global,`${policy}:${window}:all`],[global,`${policy}:${window}:${q.Clock%32}`],[cells,`${r.Phase}:${r.Case}:${r.Schedule}:${policy}:${window}`]];
   for(const [map,key]of keys){if(!map.has(key))map.set(key,empty());const x=map.get(key);x.queries++;x[category]++;if(wait!==null){x.fitWaitSum+=wait;x.fitWaitCount++;}}
  }
 }
 records++;
}
assert((await iterators[1].next()).done);assert.equal(records,2688);assert.equal(hashes[1].copy().digest('hex'),header.inputSHA256);
function finish(map){return [...map].map(([key,x])=>{assert.equal(x.queries,x.noFit+x.missingAdded+x.advanced+x.alreadyAvailable);return{key,...x,meanFitWait:x.fitWaitCount?x.fitWaitSum/x.fitWaitCount:null,trainingLeadFraction:(x.missingAdded+x.advanced)/x.queries};});}
console.log(JSON.stringify({scope:'Consumed first-fit training lead, not learning gain',hashes:hashes.map(h=>h.digest('hex')),sourceSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),protocolSHA256:crypto.createHash('sha256').update(fs.readFileSync('docs/experiments/mmm-training-lead-protocol.md')).digest('hex'),records,global:finish(global),cells:finish(cells)},null,2));
