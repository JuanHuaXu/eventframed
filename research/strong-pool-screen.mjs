import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {delayedShare, testDelayedShare} from './delayed-fixed-share.mjs';
import {strongPool, poolWeights} from './strong-brier-pool.mjs';
import './strong-brier-pool.test.mjs';

function mix(rows) {
  const tape=rows.map(r=>({...r,p:r.p??[r.b,r.c]}));
  const k=tape[0].p.length,prior=Array.from({length:k},(_,j)=>j===0?.5:.5/(k-1));
  const weights=poolWeights(tape,prior);
  return {adaptive:tape.map((r,i)=>strongPool(r.p,weights[i])),
    half:tape.map((r,i)=>r.p.reduce((v,p,j)=>v+p*weights[i][j],0))};
}
testDelayedShare();
const fixture = Array.from({length:80}, (_,i) => ({b:.2, c:.8, y:i%3===0, delay:i%7, missing:i%5===0}));
const original = mix(fixture);
const poisoned = mix(fixture.map((r,i) => ({...r, y:i>=39 || r.missing || i+r.delay>39 ? !r.y : r.y})));
for (const arm of ['adaptive','half']) assert.deepEqual(original[arm].slice(0,40), poisoned[arm].slice(0,40));
const equal = mix(fixture.map(r => ({...r,c:r.b})));
assert(equal.adaptive.every(p=>Math.abs(p-.2)<1e-12)); assert(equal.half.every(p=>p===.2));

const [source, output, timingOutput] = process.argv.slice(2);
assert(source && output && timingOutput);
let header = true, milliseconds = 0;
const results = [], names = ['fullPoolStrong','noSegmentStrong','fullPoolLinear','noSegmentLinear'];
for await (const line of readline.createInterface({input:fs.createReadStream(source), crlfDelay:Infinity})) {
  const r = JSON.parse(line);
  if (header) { assert.equal(r.Cohort,'spike-independent-v1'); header=false; continue; }
  assert.equal(r.Steps.length,256);
  // Only issue-time forecasts and explicitly scheduled outcomes reach mixing.
  const rows = arm => r.Steps.map(s=>({p:(arm===10?[12,0,1,2,3,10]:[12,0,1,2,3]).map(k=>s.P[k]), y:s.Y, delay:s.Delay, missing:s.Missing}));
  const sr=rows(10), nr=rows(13), start=performance.now();
  const segment=mix(sr), stat=mix(nr);
  milliseconds += performance.now()-start;
  const forecasts=[segment.adaptive,stat.adaptive,segment.half,stat.half];
  const expected=Array(8).fill(0), realized=Array(8).fill(0);
  const blocks=Array.from({length:8},()=>({expected:Array(8).fill(0),realized:Array(8).fill(0)}));
  for(let t=0;t<256;t++) {
    const s=r.Steps[t], ps=[...forecasts.map(f=>f[t]),s.P[0],s.P[1],s.P[12],s.P[13]];
    ps.forEach((p,a)=>{
      assert(Number.isFinite(p)&&p>=0&&p<=1);
      const e=(p-s.Q)**2+s.Q*(1-s.Q), l=(p-Number(s.Y))**2;
      expected[a]+=e/256; realized[a]+=l/256;
      blocks[Math.floor(t/32)].expected[a]+=e/32; blocks[Math.floor(t/32)].realized[a]+=l/32;
    });
  }
  for(const [a,raw] of [[4,0],[5,1],[6,12],[7,13]]) {
    assert(Math.abs(expected[a]-r.Metrics[raw][0].Brier)<1e-12);
    assert(Math.abs((blocks[6].expected[a]+blocks[7].expected[a])/2-r.Metrics[raw][1].Brier)<1e-12);
  }
  results.push({key:[r.Phase,r.Case,r.Index,r.Schedule].join(':'),expected,realized,blocks,forecasts});
}
assert.equal(results.length,672); assert.equal(new Set(results.map(r=>r.key)).size,672);
const mean=v=>v.reduce((a,b)=>a+b,0)/v.length;
const interval=v=>{const m=mean(v),se=Math.sqrt(v.reduce((s,x)=>s+(x-m)**2,0)/(v.length*(v.length-1)));return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=new Map();
for(const r of results) {const [p,c,,s]=r.key.split(':'),key=[p,c,s].join(':');if(!groups.has(key))groups.set(key,[]);groups.get(key).push(r);}
assert.equal(groups.size,84);
const gates=[];
for(const [key,rows] of groups) {
  assert.deepEqual(rows.map(r=>Number(r.key.split(':')[2])).sort((a,b)=>a-b),[0,1,2,3,4,5,6,7]);
  const c=Number(key.split(':')[1]), changing=c<9?c%3!==0:c>=19;
  for(let arm=0;arm<4;arm++) for(const period of ['whole','terminal64']) {
    const metric=r=>period==='whole'?r.expected:r.expected.map((_,a)=>(r.blocks[6].expected[a]+r.blocks[7].expected[a])/2);
    for(const control of [4,5,6,7]) {
      const v=interval(rows.map(r=>metric(r)[arm]-metric(r)[control]));
      gates.push({key,arm,period,control,type:'nonharm',...v,pass:v.upper<=.01});
    }
    if(changing&&period==='terminal64')for(const control of [4,6,7]) {
      const v=interval(rows.map(r=>metric(r)[control]-metric(r)[arm]));
      gates.push({key,arm,period,control,type:'gain',...v,pass:v.mean>=.005&&v.lower>0});
    }
  }
}
assert.equal(gates.length,4*768);
const summary=names.map((arm,a)=>({arm,expected:mean(results.map(r=>r.expected[a])),terminal:mean(results.map(r=>(r.blocks[6].expected[a]+r.blocks[7].expected[a])/2)),
  windowHarms:results.reduce((n,r)=>n+r.blocks.filter(b=>b.expected[a]-b.expected[6]>.01+1e-12).length,0),
  nonharm:gates.filter(g=>g.arm===a&&g.type==='nonharm'&&g.pass).length,
  gains:gates.filter(g=>g.arm===a&&g.type==='gain'&&g.pass).length}));
const strata=['stationary','changing'].map(regime=>{
  const rows=results.filter(r=>{const c=Number(r.key.split(':')[1]),ch=c<9?c%3!==0:c>=19;return ch===(regime==='changing');});
  return {regime,records:rows.length,expected:Array.from({length:8},(_,a)=>mean(rows.map(r=>r.expected[a])))};
});
fs.writeFileSync(output,JSON.stringify({names:[...names,'generic64','Boolean64','Markov','static64'],summary,strata,gates,results,
  limitations:'Consumed eight-index multi-head-pool ablation. Approximate +/-3.5SE screens, not original32-index confirmation or simultaneous guarantees. No runtime safety certificate.'},null,2)+'\n',{flag:'wx'});
fs.writeFileSync(timingOutput,JSON.stringify({milliseconds,microsecondsPerPoolForecast:milliseconds*1000/(672*256*2),limitations:'One-pass mixing only (strong and matched linear together), excludes input preparation, I/O, scoring, fits and serving. Includes cold/JIT effects; not a stable microbenchmark.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({summary,strata},null,2));


