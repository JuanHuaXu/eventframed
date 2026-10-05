import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const input=process.argv[2],truth=process.argv[3],hash=crypto.createHash('sha256'),sourceHash=crypto.createHash('sha256');
const stream=fs.createReadStream(input),original=fs.createReadStream(truth);stream.on('data',b=>hash.update(b));original.on('data',b=>sourceHash.update(b));
// Attach both iterators before the first await: the hash data listeners put
// streams into flowing mode, so a later line reader could lose initial chunks.
const inputLines=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
const iterator=createInterface({input:original,crlfDelay:Infinity})[Symbol.asyncIterator]();
assert.equal(JSON.parse((await iterator.next()).value).Version,'soft-learners-v120');
let header,aligned=false,baseline,checks=0,maxError=0;const records=[];
for await(const line of inputLines){
 if(!header){header=JSON.parse(line);assert(['acquisition-train-consumed-v1','acquisition-train-boundary-v1'].includes(header.version));aligned=header.version==='acquisition-train-boundary-v1';if(aligned)baseline=JSON.parse(fs.readFileSync('docs/experiments/mmm-acquisition-train-v1-summary.json'));continue;}
 const r=JSON.parse(line),next=await iterator.next();assert(!next.done);const raw=JSON.parse(next.value);assert(!r.Error);
 assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[raw.Phase,raw.Case,raw.Index,raw.Schedule]);
 const counts=[];
 for(let policy=0;policy<4;policy++){
  const result=r.Results[policy],queries=result.Queries??[],paid=new Map();counts.push(queries.length);
  for(const q of queries){assert.equal(q.Reveal,q.Clock+1);const boundary=aligned&&q.Clock<224&&q.Clock%32===31;const end=q.Clock+(boundary?1:0);assert(q.Clock>0&&q.Clock<=248&&(boundary||q.Clock%8===0));if(aligned)assert(q.Clock%32!==0);assert(q.Origin>=end-8&&q.Origin<end);assert(!paid.has(q.Origin));paid.set(q.Origin,q.Reveal);const s=raw.Steps[q.Origin];assert(s.Missing||q.Origin+s.Delay>q.Clock);if(boundary)assert(result.Fits.find(f=>f.Clock===q.Reveal).Origins.every(os=>os.includes(q.Origin)));}
  if(policy===0||r.Schedule===0)assert.equal(queries.length,0);
  assert.equal(result.Fits.length,8);
  for(const fit of result.Fits){assert(fit.Clock%32===0);let eligible=[];for(let j=-16;j<fit.Clock;j++){if(j<0||(!raw.Steps[j].Missing&&j+raw.Steps[j].Delay<=fit.Clock)||(paid.has(j)&&paid.get(j)<=fit.Clock))eligible.push(j);}
   for(let w=0;w<2;w++)assert.deepEqual(fit.Origins[w],eligible.slice(-(w===0?64:32)));
  }
  assert.equal(result.Predictions.length,256);
  const metrics=['Brier','Accuracy','LogLoss'].map(()=>Array.from({length:5},()=>[0,0]));
  result.Predictions.forEach((ps,t)=>ps.forEach((p,a)=>{
   assert(Number.isFinite(p)&&p>0&&p<1);const q=raw.Steps[t].Q;
   const values=[(p-q)**2+q*(1-q),p>=.5?q:1-q,-q*Math.log(p)-(1-q)*Math.log1p(-p)];
   values.forEach((v,m)=>{metrics[m][a][0]+=v/256;if(t>=192)metrics[m][a][1]+=v/64;});
   if(policy===0)assert(Math.abs(p-raw.Steps[t].P[[0,1,2,3,12][a]])<1e-12);checks++;
  }));
  ['Brier','Accuracy','LogLoss'].forEach((key,m)=>metrics[m].forEach((row,a)=>row.forEach((v,s)=>{const error=Math.abs(v-r[key][policy][a][s]);maxError=Math.max(maxError,error);assert(error<1e-12);}))); 
 }
 assert.equal(counts[1],counts[2]);assert.equal(counts[2],counts[3]);
 let old;
 if(aligned){old=baseline.records[records.length];assert.deepEqual([old.phase,old.case,old.index,old.schedule],[r.Phase,r.Case,r.Index,r.Schedule]);assert.deepEqual(counts,old.queries);assert.deepEqual(r.Brier[0],old.brier[0]);}
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,brier:r.Brier,accuracy:r.Accuracy,logLoss:r.LogLoss,queries:counts,...(aligned?{original:old.brier}: {})});
}
assert((await iterator.next()).done);assert.equal(records.length,2688);assert.equal(header.records,2688);assert.equal(header.inputSHA256,sourceHash.digest('hex'));assert.equal(checks,13762560);
if(aligned)assert.equal(baseline.inputSHA256,header.inputSHA256);
const mean=a=>a.reduce((a,b)=>a+b,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const meta={phase,case:c,schedule,segment};groups.push({...meta,brier:[0,1,2,3].map(p=>[0,1,2,3,4].map(a=>mean(rs.map(r=>r.brier[p][a][segment])))),queries:[0,1,2,3].map(p=>mean(rs.map(r=>r.queries[p])))});
 for(const control of aligned?[0,1,2,3]:[0,1,2]){const gain=interval(rs.map(r=>(control===3?r.original[3][4][segment]:r.brier[control][4][segment])-r.brier[3][4][segment]));gates.push({...meta,control,type:'nonharm',...gain,pass:gain.lower>=-.01});
  if(schedule===1&&segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,control,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});}
}
console.log(JSON.stringify({scope:'Consumed acquisition-to-training quality; no fresh validation',aligned,artifactSHA256:hash.digest('hex'),inputSHA256:header.inputSHA256,scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),checks,maxError,summary:{nonharm:gates.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:aligned?672:504,gains:gates.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:aligned?64:48,status:gates.every(g=>g.pass)?'PASS':'FAIL'},records,groups,gates}));
