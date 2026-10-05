import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditBurstRecord} from './query-burst-audit.mjs';

const stream=fs.createReadStream(process.argv[2]),original=fs.createReadStream(process.argv[3]);
const hash=crypto.createHash('sha256'),sourceHash=crypto.createHash('sha256');
stream.on('data',b=>hash.update(b));original.on('data',b=>sourceHash.update(b));
// Attach both consumers before awaiting either flowing stream.
const lines=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
const truth=createInterface({input:original,crlfDelay:Infinity})[Symbol.asyncIterator]();
const header=JSON.parse((await lines.next()).value);
assert.equal(header.Version,'query-burst-consumed-v1');
assert.equal(JSON.parse((await truth.next()).value).Version,'soft-learners-v120');
const records=[];let mutationChecks=0;
for await(const line of lines) {
 const next=await truth.next();assert(!next.done);
 const r=JSON.parse(line),raw=JSON.parse(next.value);
 records.push(auditBurstRecord(r,raw));
 // Prove the audit rejects altered artifacts, not just valid outputs.
 if(r.Phase===0&&r.Case===20&&r.Index===0&&r.Schedule===1) {
  for(const mutate of [
   x=>{x.Results[6].Queries[0].Reveal++;},
   x=>{x.Results[6].Arrivals[0].Surprise=!x.Results[6].Arrivals[0].Surprise;},
   x=>{x.Results[6].Fits[1].Origins[0][0]=255;},
   x=>{x.Results[0].Predictions[0][4]=.123456789;},
  ]) {const bad=structuredClone(r);mutate(bad);assert.throws(()=>auditBurstRecord(bad,raw));mutationChecks++;}
 }
}
assert((await truth.next()).done);assert.equal(records.length,2688);assert.equal(mutationChecks,4);
assert.equal(header.InputSHA256,sourceHash.digest('hex'));
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((a,x)=>a+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++) {
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const meta={phase,case:c,schedule,segment};
 groups.push({...meta,brier:Array.from({length:7},(_,a)=>mean(rs.map(r=>r.brier[a][4][segment]))),diagnostics:Array.from({length:7},(_,a)=>({...Object.fromEntries(['queries','triggerClocks','paidTriggers','lateMixerOmissions','emptyClocks','burstQueries','fallbackQueries'].map(k=>[k,mean(rs.map(r=>r.diagnostics[a][k]))])),training64:Object.fromEntries(['added','advanced','alreadyNatural','noFit','waitSum'].map(k=>[k,mean(rs.map(r=>r.diagnostics[a].training64[k]))]))}))});
 for(const control of [0,1,2,3]) {
  const gain=ci(rs.map(r=>r.brier[control][4][segment]-r.brier[6][4][segment]));
  const unequal=control===0?0:rs.filter(r=>r.diagnostics[control].queries!==r.diagnostics[6].queries).length;
  gates.push({...meta,control,type:'nonharm',...gain,unequal,pass:gain.lower>=-.01&&unequal===0});
  if(schedule===1&&segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,control,type:'gain',...gain,unequal,pass:gain.mean>=.005&&gain.lower>0&&unequal===0});
 }
}
const count=type=>({passed:gates.filter(g=>g.type===type&&g.pass).length,total:gates.filter(g=>g.type===type).length});
console.log(JSON.stringify({scope:'Consumed exploratory query timing; not fresh confirmation',artifactSHA256:hash.digest('hex'),inputSHA256:header.InputSHA256,scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),auditSHA256:crypto.createHash('sha256').update(fs.readFileSync(new URL('./query-burst-audit.mjs',import.meta.url))).digest('hex'),mutationChecks,probabilityChecks:records.length*7*256*5,summary:{nonharm:count('nonharm'),gain:count('gain'),status:gates.every(g=>g.pass)?'PASS':'FAIL'},records,groups,gates}));
