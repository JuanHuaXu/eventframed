import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const input=process.argv[2],stream=fs.createReadStream(input),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));
const pairs=new Map();let header,checks=0,maxError=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
 if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
 const r=JSON.parse(line),key=[r.Phase,r.Case,r.Index].join(':');assert([0,1].includes(r.Schedule));assert.equal(r.Steps.length,256);
 const signature=crypto.createHash('sha256').update(JSON.stringify({initial:r.Initial,steps:r.Steps.map(s=>[s.X,s.Y,s.Q])})).digest('hex');
 const brier=Array.from({length:15},()=>[0,0]);
 r.Steps.forEach((s,t)=>{if(r.Schedule===0){assert.equal(s.Delay,0);assert.equal(s.Missing,false);}
  s.P.forEach((p,a)=>{assert(Number.isFinite(p)&&p>0&&p<1);const loss=(p-s.Q)**2+s.Q*(1-s.Q);brier[a][0]+=loss/256;if(t>=192)brier[a][1]+=loss/64;checks++;});
 });
 brier.forEach((row,a)=>row.forEach((v,s)=>{const error=Math.abs(v-r.Metrics[a][s].Brier);maxError=Math.max(maxError,error);assert(error<1e-12);}));
 if(!pairs.has(key))pairs.set(key,{phase:r.Phase,case:r.Case,index:r.Index,signature,schedules:[]});const pair=pairs.get(key);assert.equal(pair.signature,signature);assert(!pair.schedules[r.Schedule]);pair.schedules[r.Schedule]=brier;
}
assert.equal(pairs.size,1344);assert.equal(checks,10321920);const records=[...pairs.values()];records.forEach(r=>{assert(r.schedules[0]&&r.schedules[1]);});
const mean=a=>a.reduce((a,b)=>a+b,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let arm=0;arm<15;arm++)for(let segment=0;segment<2;segment++){
 const rs=records.filter(r=>r.phase===phase&&r.case===c);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 groups.push({phase,case:c,arm,segment,immediate:mean(rs.map(r=>r.schedules[0][arm][segment])),delayed:mean(rs.map(r=>r.schedules[1][arm][segment])),gain:interval(rs.map(r=>r.schedules[1][arm][segment]-r.schedules[0][arm][segment]))});
}
const paths=[import.meta.filename,'docs/experiments/mmm-feedback-pairs-protocol.md'];
console.log(JSON.stringify({scope:'Consumed paired complete/immediate versus delayed/missing feedback, not oracle ceiling',artifactSHA256:hash.digest('hex'),hashes:Object.fromEntries(paths.map(p=>[p,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')])),pairs:records.length,checks,maxError,records,groups},null,2));
