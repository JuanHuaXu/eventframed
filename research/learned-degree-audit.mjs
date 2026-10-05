import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';
async function* records(path) {for await(const line of createInterface({input:createReadStream(path),crlfDelay:Infinity}))if(line.trim())yield JSON.parse(line);}
const dir='docs/experiments/';
const source=records(dir+'mmm-soft-learners-v120.jsonl')[Symbol.asyncIterator]();
const output=records(dir+'mmm-learned-degree-v1-forecasts.jsonl')[Symbol.asyncIterator]();
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const pop=x=>{let n=0;for(;x;x&=x-1)n++;return n;};
const choose=(n,k)=>{if(k<0||k>n)return 0;let v=1;for(let i=1;i<=k;i++)v=v*(n-i+1)/i;return v;};
const basis=Array.from({length:10},(_,h)=>Array.from({length:4},(_,i)=>{
 const d=i+1;let v=0;for(let j=0;j<=d;j++)v+=(j%2?-1:1)*choose(h,j)*choose(9-h,d-j);return v/choose(9,d);
}));
const bounds=i=>i===4?[Math.log(.01),Math.log(4)]:[Math.log(1e-4),Math.log(16)];
function checkLogBound(v,i) {
 const [lo,hi]=bounds(i);
 // Go and V8 log(.01) differ by one ULP. Permit only a small floating-point
 // boundary envelope, record its use, and still reject substantive violations.
 const tolerance=4*Number.EPSILON*Math.max(1,Math.abs(lo),Math.abs(hi));
 assert(Number.isFinite(v)&&v>=lo-tolerance&&v<=hi+tolerance);
 return Math.max(0,lo-v,v-hi);
}
assert(checkLogBound(-4.605170185988092,4)>0);
assert.throws(()=>checkLogBound(Math.log(.01)-1e-8,4));
assert.throws(()=>checkLogBound(NaN,4));
function evaluate(samples,logs) {
 const t=logs.map(Math.exp),n=samples.length;
 const kernel=(x,y)=>1+basis[pop(x^y)].reduce((s,v,d)=>s+t[d]*v,0);
 const a=samples.map((s,i)=>[...samples.map((r,j)=>kernel(s.x,r.x)+(i===j?t[4]:0)),s.y]);
 let logdet=0,sign=1;
 for(let k=0;k<n;k++) {
  let pivot=k;for(let i=k+1;i<n;i++)if(Math.abs(a[i][k])>Math.abs(a[pivot][k]))pivot=i;
  if(pivot!==k){[a[k],a[pivot]]=[a[pivot],a[k]];sign=-sign;}
  assert(Math.abs(a[k][k])>1e-12);logdet+=Math.log(Math.abs(a[k][k]));sign*=Math.sign(a[k][k]);
  for(let i=k+1;i<n;i++) {const f=a[i][k]/a[k][k];for(let j=k+1;j<=n;j++)a[i][j]-=f*a[k][j];a[i][k]=0;}
 }
 assert.equal(sign,1);
 const alpha=Array(n).fill(0);
 for(let i=n-1;i>=0;i--){let v=a[i][n];for(let j=i+1;j<n;j++)v-=a[i][j]*alpha[j];alpha[i]=v/a[i][i];}
 const value=(logdet+samples.reduce((s,r,i)=>s+r.y*alpha[i],0))/2;
 return {value,predict:x=>Math.max(1e-12,Math.min(1-1e-12,(1+samples.reduce((s,r,i)=>s+alpha[i]*kernel(x,r.x),0))/2))};
}
let rows=0,statsChecked=0,fits=0,forecasts=0,objectives=0,derivatives=0;
let maxForecastError=0,maxObjectiveError=0,maxProjectedGradientError=0;
let boundaryRoundoffs=0,maxBoundaryRoundoff=0;
const stops=[{}, {}, {}, {}],summary=Array.from({length:4},()=>({n:0,objectiveGain:0,evaluations:0,projectedGradient:0,parameters:[0,0,0,0,0]}));
for(;;){
 const a=await source.next(),b=await output.next();assert.equal(a.done,b.done);if(a.done)break;
 const s=a.value,r=b.value;for(const key of ['Phase','Case','Index','Schedule'])assert.equal(s[key],r[key]);rows++;
 assert.equal(r.Optimization.length,8);
 for(let pub=0;pub<8;pub++)for(let arm=0;arm<4;arm++){
  const st=r.Optimization[pub][arm];statsChecked++;
  assert(st.Evaluations<=145&&st.Evaluations>=1);assert(st.Iterations<=16&&st.Iterations>=0);
  assert.equal(st.Trace.length,st.Iterations+1);assert.equal(st.Trace[0],st.Initial);
  assert(Math.abs(st.Trace.at(-1)-st.Final)<1e-11);assert(st.Final<=st.Initial);
  for(let j=1;j<st.Trace.length;j++)assert(st.Trace[j]<=st.Trace[j-1]);
  assert(['projected-gradient','iteration-cap','line-search-cap'].includes(st.Stop));
  if(st.Stop==='projected-gradient')assert(st.ProjectedGradient<=1e-6);
  if(st.Stop==='iteration-cap')assert.equal(st.Iterations,16);
  assert(Number.isFinite(st.ProjectedGradient)&&st.ProjectedGradient>=0);
  for(let j=0;j<5;j++){const error=checkLogBound(st.Logs[j],j);if(error>0)boundaryRoundoffs++;maxBoundaryRoundoff=Math.max(maxBoundaryRoundoff,error);}
  if(arm<2)assert.equal(st.Logs[4],0);
  stops[arm][st.Stop]=(stops[arm][st.Stop]??0)+1;
  const agg=summary[arm];agg.n++;agg.objectiveGain+=st.Initial-st.Final;agg.evaluations+=st.Evaluations;agg.projectedGradient+=st.ProjectedGradient;st.Logs.forEach((v,j)=>agg.parameters[j]+=Math.exp(v));
  if(s.Index!==0||![0,4,7].includes(pub))continue;
  const clock=32*pub,w=arm%2,cap=w?32:64;
  let origins=Array.from({length:16},(_,i)=>i-16);for(let i=0;i<clock;i++)if(!s.Steps[i].Missing&&i+s.Steps[i].Delay<=clock)origins.push(i);
  origins=origins.slice(-cap);assert.deepEqual(origins,s.Fits[pub].Origins[w]);
  const samples=origins.map(i=>i<0?{x:s.Initial[i+16].Bits,y:s.Initial[i+16].Outcome?1:-1}:{x:s.Steps[i].X,y:s.Steps[i].Y?1:-1});
  const fitted=evaluate(samples,st.Logs),initial=evaluate(samples,[Math.log(9),Math.log(9),Math.log(9),Math.log(9),0]);
  for(const [actual,expected]of[[fitted.value,st.Final],[initial.value,st.Initial]]){const error=Math.abs(actual-expected);assert(error<1e-8);maxObjectiveError=Math.max(maxObjectiveError,error);objectives++;}
  let pg=0;
  for(let j=0;j<(arm<2?4:5);j++){
   const plus=[...st.Logs],minus=[...st.Logs];plus[j]+=1e-5;minus[j]-=1e-5;
   const gradient=(evaluate(samples,plus).value-evaluate(samples,minus).value)/2e-5;derivatives++;
   const [lo,hi]=bounds(j);pg=Math.max(pg,Math.abs(st.Logs[j]-Math.max(lo,Math.min(hi,st.Logs[j]-gradient))));
  }
  const ge=Math.abs(pg-st.ProjectedGradient);assert(ge<2e-5);maxProjectedGradientError=Math.max(maxProjectedGradientError,ge);
  fits++;
  for(let t=clock;t<clock+32;t++){const error=Math.abs(fitted.predict(s.Steps[t].X)-r.P[t][arm+2]);assert(error<1e-10);maxForecastError=Math.max(maxForecastError,error);forecasts++;}
 }
}
assert.equal(rows,2688);assert.equal(statsChecked,86016);assert.equal(fits,1008);assert.equal(forecasts,32256);
for(const a of summary){a.objectiveGain/=a.n;a.evaluations/=a.n;a.projectedGradient/=a.n;a.parameters=a.parameters.map(v=>v/a.n);}
console.log(JSON.stringify({rows,statsChecked,fits,forecasts,objectives,derivatives,maxForecastError,maxObjectiveError,maxProjectedGradientError,boundaryRoundoffs,maxBoundaryRoundoff,stops,summary,pass:true},null,2));
