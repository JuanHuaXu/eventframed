import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {performance} from 'node:perf_hooks';
import {runBudget,excess} from './spike-budget-core.mjs';
import {delayedShare,testDelayedShare} from './delayed-fixed-share.mjs';

const key=r=>[r.Phase,r.Case,r.Index,r.Schedule].join(':');
async function* lines(path){
  const stream=fs.createReadStream(path),reader=readline.createInterface({input:stream,crlfDelay:Infinity});
  try{for await(const line of reader)if(line.trim())yield JSON.parse(line);}finally{reader.close();stream.destroy();}
}
async function digest(path){const h=createHash('sha256');for await(const b of fs.createReadStream(path))h.update(b);return h.digest('hex');}

export function independentArms(rows){
  const staticGlobal=runBudget(rows),shareWeights=delayedShare(rows);
  const shareGlobal=runBudget(rows,shareWeights),reset=[],staticLocal=[],shareLocal=[];
  for(let c=0;c<rows.length;c+=32){
    const block=rows.slice(c,c+32);
    reset.push(...runBudget(block));
    staticLocal.push(...runBudget(block,staticGlobal.slice(c,c+32).map(f=>f.proposed)));
    shareLocal.push(...runBudget(block,shareWeights.slice(c,c+32)));
  }
  return [shareLocal,reset,rows.map(r=>({p:r.b})),staticGlobal,staticLocal,shareGlobal,rows.map(r=>({p:r.c}))];
}

function tests(){
  testDelayedShare();
  const rows=Array.from({length:256},(_,i)=>({b:.3,c:.7,y:i<128,delay:i%32,missing:i%5===0}));
  const arms=independentArms(rows);
  const poison=rows.map((r,i)=>({...r,y:i>=128||r.missing||i+r.delay>128?!r.y:r.y}));
  const other=independentArms(poison);
  for(let a=0;a<arms.length;a++)assert.deepEqual(arms[a].slice(0,129),other[a].slice(0,129));
  for(const a of [0,1,4])for(let c=0;c<256;c+=32){let actual=0;for(let j=0;j<32;j++){actual+=excess(arms[a][c+j].p,rows[c+j].b,rows[c+j].y);assert(actual<=.01*(j+1)+1e-12);}}
  assert.deepEqual(arms[3],runBudget(rows));
  assert.deepEqual(arms[5],runBudget(rows,delayedShare(rows)));
}
tests();
const [source,compact,output]=process.argv.slice(2);
if(source){
  assert(compact&&output);assert(!fs.existsSync(output));const start=performance.now();
  const candidates=new Map();
  for await(const r of lines(compact)){assert(!candidates.has(key(r)));candidates.set(key(r),r);}
  assert.equal(candidates.size,672);
  const results=[];let header,totalFits=0,cappedFits=0;
  for await(const src of lines(source)){
    if(!header){header=src;assert.equal(src.Cohort,'spike-independent-v1');assert.equal(src.TransferBase,3000011000);assert.equal(src.BooleanBase,3004011000);assert.equal(src.Version,'soft-learners-v120');continue;}
    const n=results.length;assert.deepEqual([src.Phase,src.Case,src.Index,src.Schedule],[Math.floor(n/336),Math.floor(n/16)%21,Math.floor(n/2)%8,n%2]);
    const k=key(src),r=candidates.get(k);assert(r);candidates.delete(k);assert.equal(r.P.length,256);assert.equal(src.Steps.length,256);
    for(const counts of [r.FitCounts,r.Capped])assert(counts.length===8&&counts.every(x=>Number.isInteger(x)&&x>=0));
    totalFits+=r.FitCounts.reduce((a,b)=>a+b,0);cappedFits+=r.Capped.reduce((a,b)=>a+b,0);
    const rows=r.P.map((c,i)=>{const s=src.Steps[i];assert([c,s.P[12],s.Q].every(p=>Number.isFinite(p)&&p>=0&&p<=1));assert(typeof s.Y==='boolean'&&typeof s.Missing==='boolean');assert(Number.isInteger(s.Delay)&&s.Delay>=0&&s.Delay<=31);return{b:s.P[12],c,y:s.Y,delay:s.Delay,missing:s.Missing};});
    const arms=independentArms(rows),expected=Array(7).fill(0),realized=Array(7).fill(0),blocks=Array.from({length:8},()=>({expected:Array(7).fill(0),realized:Array(7).fill(0)}));
    for(let a=0;a<7;a++){
      let actual=0,local=0;
      for(let i=0;i<256;i++){
        const p=arms[a][i].p,q=src.Steps[i].Q,y=Number(rows[i].y),e=(p-q)**2+q*(1-q),l=(p-y)**2;
        expected[a]+=e/256;realized[a]+=l/256;blocks[Math.floor(i/32)].expected[a]+=e/32;blocks[Math.floor(i/32)].realized[a]+=l/32;
        if(i%32===0)local=0;const delta=excess(p,rows[i].b,y);actual+=delta;local+=delta;
        if([0,1,3,4,5].includes(a))assert(actual<=.01*(i+1)+1e-12);
        if([0,1,4].includes(a))assert(local<=.01*(i%32+1)+1e-12);
      }
    }
    results.push({key:k,change:src.Change,expected,realized,blocks,forecasts:arms[0],armPredictions:arms.map(a=>a.map(f=>f.p))});
  }
  assert.equal(results.length,672);assert.equal(candidates.size,0);
  const summary={cohort:header.Cohort,arms:['shareLocal','reset32','Markov','staticGlobal','staticLocal','shareGlobal','challenger'],records:672,forecasts:172032,totalFits,cappedFits,
    expected:Array.from({length:7},(_,a)=>results.reduce((s,r)=>s+r.expected[a]/672,0)),
    realized:Array.from({length:7},(_,a)=>results.reduce((s,r)=>s+r.realized[a]/672,0)),
    wholeHarms:results.filter(r=>r.expected[0]-r.expected[2]>.01).length,
    windowHarms:results.reduce((s,r)=>s+r.blocks.filter(b=>b.expected[0]-b.expected[2]>.01).length,0),
    limitations:'New synthetic cohort, not real-task validation. Expected harms are distinct from realized ledger protection. Repeated scenarios and feedback schedules are paired; uncertainty requires clustered analysis.'};
  const hashes={source:await digest(source),compact:await digest(compact)};
  fs.writeFileSync(output,JSON.stringify({summary,hashes,processingMilliseconds:performance.now()-start,results},null,2)+'\n',{flag:'wx'});
  console.log(JSON.stringify(summary,null,2));
}
