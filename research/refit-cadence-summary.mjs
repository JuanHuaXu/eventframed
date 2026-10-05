import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditCadence} from './refit-cadence-audit.mjs';
const stream=fs.createReadStream(process.argv[2]),original=fs.createReadStream(process.argv[3]);
const hash=crypto.createHash('sha256'),sourceHash=crypto.createHash('sha256');
stream.on('data',b=>hash.update(b));original.on('data',b=>sourceHash.update(b));
const lines=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),truth=createInterface({input:original,crlfDelay:Infinity})[Symbol.asyncIterator]();
const header=JSON.parse((await lines.next()).value);assert.equal(header.Version,'refit-cadence-v1');assert.equal(header.Stride,8);assert.equal(header.Records,2688);assert.equal(JSON.parse((await truth.next()).value).Version,'soft-learners-v120');
const records=[];let mutations=0;
for await(const line of lines){const next=await truth.next();assert(!next.done);const raw=JSON.parse(next.value),r=JSON.parse(line);records.push(auditCadence(r,raw));
 if(r.Phase===0&&r.Case===20&&r.Index===0&&r.Schedule===1)for(const mutate of [x=>{x.Result.Fits[1].Clock++;},x=>{x.Result.Fits[1].Origins[0][0]=255;},x=>{x.Result.Predictions[40][4]=.123456789;}]){const bad=structuredClone(r);mutate(bad);assert.throws(()=>auditCadence(bad,raw));mutations++;}
}
assert((await truth.next()).done);assert.equal(records.length,2688);assert.equal(mutations,3);assert.equal(header.InputSHA256,sourceHash.digest('hex'));
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((sum,x)=>sum+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[],interactions=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++) {
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 for(let arm=0;arm<5;arm++){
  const gain=ci(rs.map(r=>r.brier[0][arm][segment]-r.brier[1][arm][segment])),meta={phase,case:c,schedule,segment,arm};
  groups.push({...meta,brier:[0,1].map(a=>mean(rs.map(r=>r.brier[a][arm][segment]))),...gain});
  if(arm===4){gates.push({...meta,type:'nonharm',...gain,pass:gain.lower>=-.01});if(schedule===1&&segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});}
 }
 if(schedule===0) {
  const delayed=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===1);
  for(let arm=0;arm<5;arm++)interactions.push({phase,case:c,segment,arm,...ci(rs.map(r=>{const d=delayed.find(d=>d.index===r.index);assert(d);assert.equal(r.signature,d.signature);return(d.brier[0][arm][segment]-d.brier[1][arm][segment])-(r.brier[0][arm][segment]-r.brier[1][arm][segment]);}))});
 }
}
console.log(JSON.stringify({scope:'Consumed extra-compute cadence diagnostic, not equal-cost rescue',artifactSHA256:hash.digest('hex'),inputSHA256:header.InputSHA256,scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),auditSHA256:crypto.createHash('sha256').update(fs.readFileSync(new URL('./refit-cadence-audit.mjs',import.meta.url))).digest('hex'),mutations,probabilityChecks:2688*2*256*5,fitBundles:{control:2688*8,fast:2688*32},summary:{nonharm:gates.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:168,gain:gates.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:16,screen:gates.every(g=>g.pass)?'PASS':'FAIL',positiveCells:groups.filter(g=>g.lower>0).length,negativeCells:groups.filter(g=>g.upper<0).length,totalCells:groups.length},records,groups,gates,interactions}));
