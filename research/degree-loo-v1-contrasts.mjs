import assert from 'node:assert/strict';
import fs from 'node:fs';
import {createInterface} from 'node:readline';
async function* rows(p){for await(const s of createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))if(s.trim())yield JSON.parse(s);}
const dir='docs/experiments/';
const source=rows(dir+'mmm-soft-learners-v120.jsonl');
const old=rows(dir+'mmm-learned-degree-v1-forecasts.jsonl');
const next=rows(dir+'mmm-degree-loo-v1-forecasts.jsonl');
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const groups=new Map();let count=0,linearChecks=0;
for(;;) {
 const a=await source.next(),b=await old.next(),c=await next.next();
 assert.equal(a.done,b.done);assert.equal(a.done,c.done);if(a.done)break;
 const s=a.value,o=b.value,n=c.value;
 for(const k of ['Phase','Case','Index','Schedule']){assert.equal(s[k],o[k]);assert.equal(s[k],n[k]);}
 const gains=Array.from({length:4},()=>[0,0]);
 for(let t=0;t<256;t++) {
  for(let w=0;w<2;w++){assert.equal(o.P[t][w],n.P[t][w]);linearChecks++;}
  const q=s.Steps[t].Q;
  for(let k=0;k<4;k++) {
   const gain=(o.P[t][2+k%2]-q)**2-(n.P[t][2+k]-q)**2;
   gains[k][0]+=gain/256;if(t>=192)gains[k][1]+=gain/64;
  }
 }
 const key=[s.Phase,s.Case,s.Schedule].join(':');if(!groups.has(key))groups.set(key,[]);
 groups.get(key).push({index:s.Index,gains});count++;
}
assert.equal(count,2688);
const cells=[];
for(const [key,rs]of groups) {
 assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 for(let arm=0;arm<4;arm++)for(let span=0;span<2;span++) {
  const xs=rs.map(r=>r.gains[arm][span]),mean=xs.reduce((a,b)=>a+b,0)/32;
  const se=Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/(31*32));
  cells.push({key,arm,span,mean,lower:mean-3.5*se,upper:mean+3.5*se});
 }
}
const totals=[];
for(let arm=0;arm<4;arm++) {
 const cs=cells.filter(c=>c.arm===arm);
 totals.push({arm,total:cs.length,nonharm:cs.filter(c=>c.lower>=-.01).length,
 gains:cs.filter(c=>c.mean>=.005&&c.lower>0).length,meanGain:cs.reduce((s,c)=>s+c.mean,0)/cs.length});
}
console.log(JSON.stringify({count,linearChecks,totals,cells},null,2));
