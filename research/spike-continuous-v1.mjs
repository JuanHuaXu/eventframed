import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {runBudget,excess} from './spike-budget-core.mjs';
import {delayedShare,testDelayedShare} from './delayed-fixed-share.mjs';

const key=r=>[r.Phase,r.Case,r.Index,r.Schedule].join(':');
async function* jsonl(path){const stream=fs.createReadStream(path),lines=readline.createInterface({input:stream,crlfDelay:Infinity});try{for await(const line of lines)if(line.trim())yield JSON.parse(line);}finally{lines.close();stream.destroy();}}
function tests(){
  const rows=Array.from({length:256},(_,i)=>({b:.2,c:.8,y:i<160,delay:i%32,missing:i%5===0}));
  const out=runBudget(rows);let actual=0;
  for(let i=0;i<256;i++){actual+=excess(out[i].p,rows[i].b,rows[i].y);assert(actual<=.01*(i+1)+1e-12);}
  const poison=rows.map((r,i)=>({...r,y:i>=128||r.missing||i+r.delay>128?!r.y:r.y}));
  assert.deepEqual(out.slice(0,129),runBudget(poison).slice(0,129));
  assert.notDeepEqual(out.slice(129),runBudget(poison).slice(129));
  assert.notEqual(out[128].proposed,runBudget(rows.slice(128))[0].proposed);
  const missing=rows.map(r=>({...r,missing:true,y:NaN}));
  assert(runBudget(missing).every(r=>Number.isFinite(r.p)&&r.proposed===.5));
  // A global budget can spend earlier gains on later harm. This negative
  // control prevents treating prefix protection as a subwindow guarantee.
  const banked=Array.from({length:256},(_,i)=>({b:.5,c:i<224?0:1,y:false,delay:0,missing:false}));
  const bankedOut=runBudget(banked);
  const lateHarm=bankedOut.slice(224).reduce((s,f,i)=>s+excess(f.p,.5,banked[224+i].y),0)/32;
  assert(lateHarm>.1);
  assert(bankedOut.reduce((s,f,i)=>s+excess(f.p,.5,banked[i].y),0)<=2.56+1e-12);
  const local=[];
  for(let c=0;c<256;c+=32)local.push(...runBudget(banked.slice(c,c+32),bankedOut.slice(c,c+32).map(f=>f.proposed)));
  assert.deepEqual(local.map(f=>f.proposed),bankedOut.map(f=>f.proposed));
  for(let c=0;c<256;c+=32){let loss=0;for(let j=0;j<32;j++){loss+=excess(local[c+j].p,.5,false);assert(loss<=.01*(j+1)+1e-12);}}
  assert(local.slice(224).reduce((s,f)=>s+excess(f.p,.5,false),0)/32<=.01+1e-12);
  assert.throws(()=>runBudget(rows,[.5]));
  assert.throws(()=>runBudget(rows,Array(256).fill(NaN)));
  assert.throws(()=>runBudget(rows,Array(256).fill(1.1)));
  console.log('continuous prefix, delayed-label poisoning, reset negative control PASS');
}
tests();
testDelayedShare();
const [source,compact,dir,output,policy,ledgerMode]=process.argv.slice(2);
assert(policy===undefined||policy==='static'||policy==='fixed-share');
assert(ledgerMode===undefined||ledgerMode==='block32');
if(source){
  assert(compact&&dir&&output);const start=performance.now();
  const candidates=new Map();for await(const r of jsonl(compact)){assert(!candidates.has(key(r)));candidates.set(key(r),r);}assert.equal(candidates.size,672);
  const old=new Map();for(const [clock,name]of [[0,'early'],[128,'breadth'],[224,'late']])old.set(clock,new Map(JSON.parse(fs.readFileSync(`${dir}/mmm-spike-budget-${name}.json`,'utf8')).results.map(r=>[r.key,r])));
  const results=[];let header=true,totalFits=0,cappedFits=0,clipped=0,maxBoundGap=-Infinity;
  for await(const src of jsonl(source)){
    if(header){assert.equal(src.Version,'soft-learners-v120');header=false;continue;}
    const k=key(src),r=candidates.get(k);if(!r)continue;candidates.delete(k);assert.equal(r.P.length,256);
    totalFits+=r.FitCounts.reduce((a,b)=>a+b,0);cappedFits+=r.Capped.reduce((a,b)=>a+b,0);
    const rows=r.P.map((c,i)=>{const s=src.Steps[i];assert(c>=0&&c<=1&&s.P[12]>=0&&s.P[12]<=1);assert(Number.isInteger(s.Delay)&&s.Delay>=0);return{b:s.P[12],c,y:s.Y,delay:s.Delay,missing:s.Missing};});
    let forecasts=runBudget(rows,policy==='fixed-share'?delayedShare(rows):undefined);
    if(ledgerMode==='block32'){
      const weights=forecasts.map(f=>f.proposed);forecasts=[];
      for(let c=0;c<256;c+=32)forecasts.push(...runBudget(rows.slice(c,c+32),weights.slice(c,c+32)));
    }
    const reset=[];
    for(let c=0;c<256;c+=32){const block=runBudget(rows.slice(c,c+32));if(old.has(c))assert.deepEqual(block,old.get(c).get(k).forecasts);reset.push(...block);}
    const expected=Array(6).fill(0),realized=Array(6).fill(0),blocks=Array.from({length:8},()=>({expected:Array(6).fill(0),realized:Array(6).fill(0)}));let actual=0,ledgerActual=0;
    for(let i=0;i<256;i++){
      const f=forecasts[i],s=rows[i],q=src.Steps[i].Q;assert(q>=0&&q<=1);
      if(ledgerMode==='block32'&&i%32===0)ledgerActual=0;
      clipped+=Number(f.clipped);actual+=excess(f.p,s.b,s.y);ledgerActual+=excess(f.p,s.b,s.y);maxBoundGap=Math.max(maxBoundGap,actual-.01*(i+1));assert(ledgerActual<=f.reserved+1e-12);assert(actual<=.01*(i+1)+1e-12);
      if(ledgerMode==='block32')assert(ledgerActual<=.01*(i%32+1)+1e-12);
      const ps=[f.p,reset[i].p,s.b,(1-f.proposed)*s.b+f.proposed*s.c,(s.b+s.c)/2,s.c];
      for(let j=0;j<6;j++){const e=(ps[j]-q)**2+q*(1-q),l=(ps[j]-Number(s.y))**2;expected[j]+=e/256;realized[j]+=l/256;blocks[Math.floor(i/32)].expected[j]+=e/32;blocks[Math.floor(i/32)].realized[j]+=l/32;}
    }
    results.push({key:k,expected,realized,blocks,forecasts});
  }
  assert.equal(candidates.size,0);assert.equal(results.length,672);
  const mean=Array(6).fill(0),blocks=Array.from({length:8},()=>({expected:Array(6).fill(0),realized:Array(6).fill(0),harms:0}));
  for(const r of results){for(let j=0;j<6;j++)mean[j]+=r.expected[j]/672;for(let b=0;b<8;b++){for(let j=0;j<6;j++){blocks[b].expected[j]+=r.blocks[b].expected[j]/672;blocks[b].realized[j]+=r.blocks[b].realized[j]/672;}blocks[b].harms+=Number(r.blocks[b].expected[0]-r.blocks[b].expected[2]>.01);}}
  const summary={policy:policy??'static',ledgerMode:ledgerMode??'global',arms:[policy==='fixed-share'?'switchingGuard':'continuousGuard','reset32Guard','Markov',policy==='fixed-share'?'switchingUnguarded':'continuousUnguarded','fixedHalf','challenger'],records:672,forecasts:172032,totalFits,cappedFits,clipped,maxBoundGap,expected:mean,wholeStreamHarms:results.filter(r=>r.expected[0]-r.expected[2]>.01).length,blocks,limitations:'Consumed eight-index experiment. Prefix realized bound follows the declared ledger; it is not an arbitrary subwindow or conditional expected-score guarantee. Cached work and boundary refits remain charged.'};
  fs.writeFileSync(output,JSON.stringify({summary,processingMilliseconds:performance.now()-start,results},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
}
