import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';

import {runBudget as run, excess} from './spike-budget-core.mjs';

function tests() {
  let checked=0;
  for(let bits=0;bits<512;bits++) {
    const rows=Array.from({length:9},(_,i)=>({b:[.1,.8,.5][i%3],c:[.9,.2,.5][i%3],y:!!(bits&(1<<i)),delay:i%4,missing:i%5===0}));
    const out=run(rows);let actual=0;
    for(let i=0;i<9;i++) {actual+=excess(out[i].p,rows[i].b,rows[i].y);assert(actual<=.01*(i+1)+1e-12);checked++;}
    const poisoned=rows.map((r,i)=>({...r,y:i>=4||r.missing||i+r.delay>4?!r.y:r.y}));
    assert.deepEqual(out.slice(0,5),run(poisoned).slice(0,5));
  }
  const quiet=Array.from({length:32},()=>({b:.2,c:.8,y:NaN,delay:0,missing:true}));
  assert(run(quiet).every(r=>Number.isFinite(r.p)&&r.clipped));
  assert(run(quiet.map(r=>({...r,c:.2}))).every(r=>r.p===.2&&!r.clipped));
  const positive=quiet.map(r=>({...r,y:true,missing:false}));
  assert(run(positive)[1].w>run(quiet)[1].w);
  console.log(`budget self-tests PASS: ${checked} exhaustive prefix checks`);
}
async function* jsonl(path){const stream=fs.createReadStream(path),lines=readline.createInterface({input:stream,crlfDelay:Infinity});try{for await(const line of lines)if(line.trim())yield JSON.parse(line);}finally{lines.close();stream.destroy();}}
tests();
const [source,cadence,feedback,output,mode,clockArg]=process.argv.slice(2);
const startClock=clockArg===undefined?128:Number(clockArg);assert([0,128,224].includes(startClock));
assert(mode===undefined||mode==='eight');
const expectedCount=mode==='eight'?672:84;
if(source){
  assert(cadence&&feedback&&output);
  const start=performance.now(),key=r=>[r.Phase,r.Case,r.Index,r.Schedule].join(':');
  const candidates=new Map();for await(const r of jsonl(cadence)){assert(!candidates.has(key(r)));candidates.set(key(r),r);}
  assert.equal(candidates.size,expectedCount);
  const old=JSON.parse(fs.readFileSync(feedback,'utf8'));assert.equal(old.summary.startClock??128,startClock);const references=new Map(old.results.map(r=>[r.key,r]));
  const results=[];let header=true,clipped=0,maxActualExcess=-Infinity,maxBoundGap=-Infinity;
  for await(const s of jsonl(source)){
    if(header){assert.equal(s.Version,'soft-learners-v120');header=false;continue;}
    const k=key(s),r=candidates.get(k);if(!r)continue;candidates.delete(k);
    const ref=references.get(k);assert(ref);assert.equal(r.Clock??128,startClock);
    const rows=r.P.map((c,i)=>{const z=s.Steps[startClock+i];assert.equal(r.Y[i],z.Y);assert.equal(r.Q[i],z.Q);assert.deepEqual(r.Control[i],z.P);assert(c>=0&&c<=1&&z.P[12]>=0&&z.P[12]<=1);assert(Number.isInteger(z.Delay)&&z.Delay>=0);return{b:z.P[12],c,y:z.Y,delay:z.Delay,missing:z.Missing};});
    assert.equal(rows.length,32);
    const forecasts=run(rows),expected=[0,0,0,0],realized=[0,0,0,0];let actual=0;
    for(let i=0;i<32;i++){
      const f=forecasts[i],z=rows[i],q=r.Q[i];assert(Math.abs(f.proposed-ref.forecasts[i].w)<1e-12);
      clipped+=Number(f.clipped);actual+=excess(f.p,z.b,z.y);maxActualExcess=Math.max(maxActualExcess,actual);maxBoundGap=Math.max(maxBoundGap,actual-.01*(i+1));assert(actual<=f.reserved+1e-12);assert(actual<=.01*(i+1)+1e-12);
      for(const [j,p] of [f.p,z.b,ref.forecasts[i].p,(z.b+z.c)/2].entries()){expected[j]+=((p-q)**2+q*(1-q))/32;realized[j]+=(p-Number(z.y))**2/32;}
    }
    results.push({key:k,expected,realized,forecasts});
  }
  assert.equal(candidates.size,0);assert.equal(results.length,expectedCount);
  const expected=[0,0,0,0],realized=[0,0,0,0];for(const r of results)for(let j=0;j<4;j++){expected[j]+=r.expected[j]/expectedCount;realized[j]+=r.realized[j]/expectedCount;}
  const harms=results.filter(r=>r.expected[0]-r.expected[1]>.01).map(r=>({key:r.key,delta:r.expected[0]-r.expected[1]}));
  const summary={startClock,arms:['budgetMixture','Markov','feedbackMixture','fixedHalf'],expected,realized,records:expectedCount,forecasts:expectedCount*32,clipped,maxActualExcess,maxBoundGap,harms,improved:results.filter(r=>r.expected[0]<r.expected[1]).length,limitations:'Consumed-data replay. Realized prefix bound does not certify conditional expected-score non-harm or population non-inferiority.'};
  fs.writeFileSync(output,JSON.stringify({summary,processingMilliseconds:performance.now()-start,results},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
}
