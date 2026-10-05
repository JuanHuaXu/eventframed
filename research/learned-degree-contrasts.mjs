import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';
async function* records(path) {
 for await(const line of createInterface({input:createReadStream(path),crlfDelay:Infinity})) if(line.trim())yield JSON.parse(line);
}
const dir='docs/experiments/';
const source=records(dir+'mmm-soft-learners-v120.jsonl')[Symbol.asyncIterator]();
const old=records(dir+'mmm-degree-prior-v1-forecasts.jsonl')[Symbol.asyncIterator]();
const next=records(dir+'mmm-learned-degree-v1-forecasts.jsonl')[Symbol.asyncIterator]();
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const cells=new Map();let recordsChecked=0,linearChecks=0;
for(;;) {
 const a=await source.next(),b=await old.next(),c=await next.next();
 assert.equal(a.done,b.done);assert.equal(a.done,c.done);if(a.done)break;
 const s=a.value,o=b.value,r=c.value;
 for(const k of ['Phase','Case','Index','Schedule']) {assert.equal(s[k],o[k]);assert.equal(s[k],r[k]);}
 recordsChecked++;
 const differences=Array.from({length:10},()=>[0,0]);
 for(let t=0;t<256;t++) {
  const q=s.Steps[t].Q;
  for(let w=0;w<2;w++) {
   assert.equal(r.P[t][w],o.P[t][w]);linearChecks++;
   const pairs=[[r.P[t][2+w],o.P[t][2+w]],[r.P[t][4+w],o.P[t][2+w]],[r.P[t][2+w],o.P[t][4+w]],[r.P[t][4+w],o.P[t][4+w]],[r.P[t][4+w],r.P[t][2+w]]];
   for(let j=0;j<5;j++) {
    const [candidate,control]=pairs[j];assert(Number.isFinite(candidate)&&Number.isFinite(control));
    const delta=(candidate-q)**2-(control-q)**2;
    differences[2*j+w][0]+=delta/256;if(t>=192)differences[2*j+w][1]+=delta/64;
   }
  }
 }
 const key=[s.Phase,s.Case,s.Schedule].join(':');if(!cells.has(key))cells.set(key,[]);cells.get(key).push({index:s.Index,differences});
}
assert.equal(recordsChecked,2688);
const recovery=new Set([1,2,4,5,7,8,19,20]);
const gates=[];
for(const [key,rs] of cells) {
 assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 const [phase,scenario,schedule]=key.split(':').map(Number);
 for(let contrast=0;contrast<10;contrast++)for(let span=0;span<2;span++) {
  const xs=rs.map(r=>r.differences[contrast][span]);
  const mean=xs.reduce((s,x)=>s+x,0)/32;
  const radius=3.5*Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/(31*32));
  gates.push({phase,case:scenario,schedule,contrast,span,type:'nonharm',mean,lower:mean-radius,upper:mean+radius,pass:mean+radius<=.01});
  if(span===1&&recovery.has(scenario))gates.push({phase,case:scenario,schedule,contrast,span,type:'gain',mean:-mean,lower:-mean-radius,upper:-mean+radius,pass:-mean>=.005&&-mean-radius>0});
 }
}
const counts=[];
for(let contrast=0;contrast<10;contrast++)for(const type of ['nonharm','gain']) {
 const subset=gates.filter(g=>g.contrast===contrast&&g.type===type);
 counts.push({contrast,type,passed:subset.filter(g=>g.pass).length,total:subset.length});
}
console.log(JSON.stringify({recordsChecked,linearChecks,names:['learned-fixed64-v-degree64','learned-fixed32-v-degree32','learned-noise64-v-degree64','learned-noise32-v-degree32','learned-fixed64-v-scalar64','learned-fixed32-v-scalar32','learned-noise64-v-scalar64','learned-noise32-v-scalar32','learned-noise64-v-learned-fixed64','learned-noise32-v-learned-fixed32'],counts,gates},null,2));

