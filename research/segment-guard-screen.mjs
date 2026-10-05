import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {testDelayedShare} from './delayed-fixed-share.mjs';
import {testPointwiseGuard} from './pointwise-brier-guard.mjs';
import {guardedHead} from './guarded-head.mjs';
export {guardedHead} from './guarded-head.mjs';
function tests(){
  testDelayedShare();testPointwiseGuard();
  const rows=Array.from({length:96},(_,i)=>({b:.2,c:.8,y:i%3===0,delay:i%8,missing:i%7===0}));
  const a=guardedHead(rows),poison=rows.map((r,i)=>({...r,y:i>=40||r.missing||i+r.delay>40?!r.y:r.y})),b=guardedHead(poison);
  for(const k of ['point','local'])assert.deepEqual(a[k].slice(0,41),b[k].slice(0,41));
  const same=guardedHead(rows.map(r=>({...r,c:r.b})));assert(same.point.every(p=>p===.2));assert(same.local.every(p=>Math.abs(p-.2)<1e-15));
  console.log('matched head composition/as-of tests PASS');
}
tests();
const[source,output]=process.argv.slice(2);
if(source){
  assert(output);const results=[];let header=true;
  for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
    const s=JSON.parse(line);if(header){assert.equal(s.Cohort,'spike-independent-v1');assert.equal(s.Hazard,.01);assert.equal(s.GenericMass,.95);header=false;continue;}
    assert.equal(s.Steps.length,256);
    const rows=arm=>s.Steps.map(v=>({b:v.P[12],c:v.P[arm],y:v.Y,delay:v.Delay,missing:v.Missing}));
    const seg=guardedHead(rows(10)),stat=guardedHead(rows(13)),expected=Array(7).fill(0),realized=Array(7).fill(0),blocks=Array.from({length:8},()=>({expected:Array(7).fill(0),realized:Array(7).fill(0)}));
    const forecasts=Array.from({length:7},()=>[]),localLoss=[0,0];
    for(let i=0;i<256;i++){
      const v=s.Steps[i],q=v.Q,y=Number(v.Y),base=v.P[12],ps=[seg.point[i],stat.point[i],base,seg.local[i],stat.local[i],v.P[10],v.P[13]];
      if(i%32===0)localLoss.fill(0);
      for(let a=0;a<7;a++){
        const p=ps[a];assert(Number.isFinite(p)&&p>=0&&p<=1);forecasts[a].push(p);
        const e=(p-q)**2+q*(1-q),l=(p-y)**2;expected[a]+=e/256;realized[a]+=l/256;blocks[Math.floor(i/32)].expected[a]+=e/32;blocks[Math.floor(i/32)].realized[a]+=l/32;
        if(a<2)assert((p-base)*(p+base-2*q)<=.01+1e-12);
        if(a===3||a===4){localLoss[a-3]+=(p-base)*(p+base-2*y);assert(localLoss[a-3]<=.01*(i%32+1)+1e-12);}
      }
    }
    results.push({key:[s.Phase,s.Case,s.Index,s.Schedule].join(':'),expected,realized,blocks,forecasts});
  }
  assert.equal(results.length,672);assert.equal(new Set(results.map(r=>r.key)).size,672);
  const names=['segmentPoint','staticPoint','Markov','segmentLocal','staticLocal','segmentRaw','staticRaw'];
  const summary={cohort:'consumed-segment-guard-v1',records:672,arms:names.map((arm,a)=>({arm,expected:results.reduce((s,r)=>s+r.expected[a]/672,0),terminal:results.reduce((s,r)=>s+(r.blocks[6].expected[a]+r.blocks[7].expected[a])/1344,0),windowHarms:results.reduce((s,r)=>s+r.blocks.filter(b=>b.expected[a]-b.expected[2]>.01+1e-12).length,0)})),limitations:'Post-hoc matched segment/static screen on consumed data. All source model cost remains charged. No new calibration, recovery-speed or real-agent validation.'};
  fs.writeFileSync(output,JSON.stringify({summary,results},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
}
