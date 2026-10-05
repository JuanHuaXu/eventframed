import assert from 'node:assert/strict';
import fs from 'node:fs';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';

async function* records(path) {
  for await(const line of createInterface({input:fs.createReadStream(path),crlfDelay:Infinity}))
    if(line.trim()) yield JSON.parse(line);
}
const transition=(w,prior)=>w.map((x,i)=>.999*x+.001*prior[i]);
function step(w,prior,p,y) {
  if(y!==undefined) {
    w=w.map((x,i)=>x*(y?p[i]:1-p[i]));
    const sum=w.reduce((a,b)=>a+b,0);assert(sum>0);
    w=w.map(x=>x/sum);
  }
  return transition(w,prior);
}
function prior(n,uniform) {return Array.from({length:n},(_,i)=>uniform?1/n:i===0?.95:.05/(n-1));}
function full(steps,ps,t,initial) {
  let w=initial.slice();
  for(let j=0;j<t;j++) {
    const s=steps[j];
    w=step(w,initial,ps[j],!s.Missing&&j+s.Delay<=t?s.Y:undefined);
  }
  return w;
}
// The checkpoint folds only the settled prefix. The unresolved suffix is
// refiltered as-of now, never multiplying a prior pass's emissions again.
function weights(steps,ps,t,state) {
  while(state.base<t) {
    const j=state.base,s=steps[j];
    if((s.Missing||j+s.Delay>t)&&j>=t-31) break;
    state.checkpoint=step(state.checkpoint,state.prior,ps[j],!s.Missing&&j+s.Delay<=t?s.Y:undefined);
    state.base++;
  }
  let w=state.checkpoint.slice();
  assert(t-state.base<=31);
  for(let j=state.base;j<t;j++) {
    const s=steps[j];
    w=step(w,state.prior,ps[j],!s.Missing&&j+s.Delay<=t?s.Y:undefined);
  }
  return w;
}
const dir='docs/experiments/';
const source=records(dir+'mmm-soft-learners-v120.jsonl');
const learned=records(dir+'mmm-learned-degree-v1-forecasts.jsonl');
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const output=process.argv[2];assert(output,'output path required');
const fd=fs.openSync(output,'wx');
let count=0,checks=0,poison=0,maxControlError=0,mixerMS=0;
const contrasts=[];
try {
for(;;) {
  const a=await source.next(),b=await learned.next();assert.equal(a.done,b.done);if(a.done) break;
  const s=a.value,r=b.value;
  for(const k of ['Phase','Case','Index','Schedule']) assert.equal(s[k],r[k]);
  const tapes=Array.from({length:6},(_,v)=>s.Steps.map((x,t)=>v<2?x.P.slice(0,4):[...x.P.slice(0,4),r.P[t][2+v%2]]));
  const states=tapes.map((_,v)=>{const p=prior(v<2?4:5,v===1||v>=4);return {prior:p,checkpoint:p.slice(),base:0};});
  const P=[],loss=Array.from({length:6},()=>[0,0]);
  const start=performance.now();
  for(let t=0;t<256;t++) {
    const pred=tapes.map((ps,v)=>{
      const w=weights(s.Steps,ps,t,states[v]);
      if(s.Index===0&&[0,32,160,255].includes(t)) {
        const reference=full(s.Steps,ps,t,states[v].prior);
        for(let k=0;k<w.length;k++) assert(Math.abs(w[k]-reference[k])<1e-13);
        checks++;
        const changed=s.Steps.map((x,j)=>({...x,Y:j>=t||x.Missing||j+x.Delay>t?!x.Y:x.Y,Q:NaN}));
        assert.deepEqual(reference,full(changed,ps,t,states[v].prior));poison++;
      }
      return w.reduce((sum,x,k)=>sum+x*ps[t][k],0);
    });
    maxControlError=Math.max(maxControlError,Math.abs(pred[0]-s.Steps[t].P[12]));
    assert(maxControlError<1e-12);
    for(let v=0;v<6;v++) {
      assert(Number.isFinite(pred[v])&&pred[v]>0&&pred[v]<1);
      const q=s.Steps[t].Q,l=(pred[v]-q)**2+q*(1-q);
      loss[v][0]+=l/256;if(t>=192) loss[v][1]+=l/64;
    }
    P.push([r.P[t][0],r.P[t][1],...pred.slice(2)]);
  }
  mixerMS+=performance.now()-start;
  fs.writeSync(fd,JSON.stringify({Phase:r.Phase,Case:r.Case,Index:r.Index,Schedule:r.Schedule,P,Fits:48,MaxResidual:r.MaxResidual,FeatureTotal:r.FeatureTotal})+'\n');
  contrasts.push({Phase:r.Phase,Case:r.Case,Index:r.Index,Schedule:r.Schedule,loss});count++;
}
} finally {fs.closeSync(fd);}
assert.equal(count,2688);assert.equal(checks,2016);assert.equal(poison,2016);
fs.writeFileSync(output+'.checks.json',JSON.stringify({count,checks,poison,maxControlError,mixerMS,contrasts},null,2)+'\n');
console.log(JSON.stringify({count,checks,poison,maxControlError,mixerMS}));
