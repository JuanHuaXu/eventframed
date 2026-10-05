import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const paths=process.argv.slice(2,4),streams=paths.map(p=>fs.createReadStream(p)),digests=streams.map(()=>crypto.createHash('sha256'));
streams.forEach((s,i)=>s.on('data',b=>digests[i].update(b)));const readers=streams.map(s=>createInterface({input:s,crlfDelay:Infinity})[Symbol.asyncIterator]());
const header=JSON.parse((await readers[0].next()).value);assert.equal(header.version,'feedback-factorial-v1');assert.equal(JSON.parse((await readers[1].next()).value).Version,'soft-learners-v120');
const referencePath='docs/experiments/mmm-feedback-pairs-v120.json',ref=JSON.parse(fs.readFileSync(referencePath)),lookup=new Map(ref.records.map(r=>[[r.phase,r.case,r.index].join(':'),r]));assert.equal(header.inputSHA256,ref.artifactSHA256);
const arms=[0,1,2,3,12],records=[];let checks=0,maxIdentityError=0;
for await(const line of readers[0]){
 const r=JSON.parse(line);assert(!r.Error);let raw;
 do{const next=await readers[1].next();assert(!next.done);raw=JSON.parse(next.value);}while(raw.Schedule!==1);
 const key=[r.Phase,r.Case,r.Index].join(':');assert.equal(key,[raw.Phase,raw.Case,raw.Index].join(':'));const prior=lookup.get(key);assert(prior);
 const signature=crypto.createHash('sha256').update(JSON.stringify({initial:raw.Initial,steps:raw.Steps.map(s=>[s.X,s.Y,s.Q])})).digest('hex');assert.equal(signature,prior.signature);
 const scores=[arms.map(a=>prior.schedules[0][a]),null,null,arms.map(a=>prior.schedules[1][a])];
 for(let mode=1;mode<=2;mode++){
  const result=r.Results[mode-1];assert.equal((result.Queries??[]).length,0);assert.equal(result.Fits.length,8);assert.equal(result.Predictions.length,256);
  for(const f of result.Fits){const eligible=[];for(let j=-16;j<f.Clock;j++){if(j<0){eligible.push(j);continue;}const s=raw.Steps[j],delay=mode===1?0:s.Delay,missing=mode===1?s.Missing:false;if(!missing&&j+delay<=f.Clock)eligible.push(j);}
   for(let w=0;w<2;w++)assert.deepEqual(f.Origins[w],eligible.slice(-(w===0?64:32)));
  }
  const b=Array.from({length:5},()=>[0,0]);result.Predictions.forEach((ps,t)=>ps.forEach((p,a)=>{assert(Number.isFinite(p)&&p>0&&p<1);const q=raw.Steps[t].Q,l=(p-q)**2+q*(1-q);b[a][0]+=l/256;if(t>=192)b[a][1]+=l/64;checks++;}));scores[mode]=b;
 }
 for(let a=0;a<5;a++)for(let s=0;s<2;s++){const values=scores.map(v=>v[a][s]);const delay=values[2]-values[0],missing=values[1]-values[0],interaction=values[3]-values[2]-values[1]+values[0];const e=Math.abs(delay+missing+interaction-(values[3]-values[0]));maxIdentityError=Math.max(e,maxIdentityError);assert(e<1e-12);}
 records.push({phase:r.Phase,case:r.Case,index:r.Index,scores});
}
assert((await readers[1].next()).done);assert.equal(records.length,1344);assert.equal(checks,3440640);assert.equal(digests[1].copy().digest('hex'),header.inputSHA256);
const mean=a=>a.reduce((a,b)=>a+b,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let arm=0;arm<5;arm++)for(let segment=0;segment<2;segment++){
 const rs=records.filter(r=>r.phase===phase&&r.case===c);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const costs=rs.map(r=>r.scores.map(m=>m[arm][segment]));
 groups.push({phase,case:c,arm:arms[arm],segment,brier:[0,1,2,3].map(i=>mean(costs.map(v=>v[i]))),delay:interval(costs.map(v=>v[2]-v[0])),missing:interval(costs.map(v=>v[1]-v[0])),interaction:interval(costs.map(v=>v[3]-v[2]-v[1]+v[0])),total:interval(costs.map(v=>v[3]-v[0]))});
}
console.log(JSON.stringify({scope:'Consumed matched delay/missingness factorial, not new policy validation',artifactSHA256:digests[0].digest('hex'),inputSHA256:header.inputSHA256,referenceSHA256:crypto.createHash('sha256').update(fs.readFileSync(referencePath)).digest('hex'),scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),checks,maxIdentityError,records,groups},null,2));
