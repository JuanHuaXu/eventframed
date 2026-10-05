import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';
async function* records(path){for await(const l of createInterface({input:createReadStream(path),crlfDelay:Infinity}))if(l.trim())yield JSON.parse(l);}
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

const source=records('docs/experiments/mmm-soft-learners-v120.jsonl')[Symbol.asyncIterator]();
const output=records('docs/experiments/mmm-degree-bfgs-v1.jsonl')[Symbol.asyncIterator]();
assert.equal((await source.next()).value.Version,'soft-learners-v120');
let rows=0,fits=0,forecasts=0,objectives=0,derivatives=0,maxForecastError=0,maxObjectiveError=0,maxPGError=0;
for(;;){
 const item=await source.next();if(item.done)break;rows++;
 const s=item.value;if(s.Index!==0)continue;
 for(const clock of [0,128,224])for(const [w,cap] of [64,32].entries())for(const noise of [false,true]){
  const item=await output.next();assert(!item.done);const r=item.value,st=r.Stats;
  for(const key of ['Phase','Case','Index','Schedule'])assert.equal(s[key],r[key]);
  assert.equal(r.Clock,clock);assert.equal(r.Window,w);assert.equal(r.LearnNoise,noise);
  let origins=Array.from({length:16},(_,i)=>i-16);
  for(let i=0;i<clock;i++)if(!s.Steps[i].Missing&&i+s.Steps[i].Delay<=clock)origins.push(i);
  origins=origins.slice(-cap);assert.deepEqual(origins,r.Origins);assert.deepEqual(origins,s.Fits[clock/32].Origins[w]);
  const samples=origins.map(i=>i<0?{x:s.Initial[i+16].Bits,y:s.Initial[i+16].Outcome?1:-1}:{x:s.Steps[i].X,y:s.Steps[i].Y?1:-1});
  st.Logs.forEach(checkLogBound);
  const fitted=evaluate(samples,st.Logs),old=evaluate(samples,r.OldStats.Logs),initial=evaluate(samples,[Math.log(9),Math.log(9),Math.log(9),Math.log(9),0]);
  for(const [a,b]of[[fitted.value,st.Final],[old.value,r.OldStats.Final],[initial.value,st.Initial]]){
   const error=Math.abs(a-b);assert(error<1e-8);maxObjectiveError=Math.max(maxObjectiveError,error);objectives++;
  }
  let pg=0;
  for(let j=0;j<(noise?5:4);j++){
   const p=[...st.Logs],m=[...st.Logs];p[j]+=1e-5;m[j]-=1e-5;
   const g=(evaluate(samples,p).value-evaluate(samples,m).value)/2e-5;derivatives++;
   const [lo,hi]=bounds(j);pg=Math.max(pg,Math.abs(st.Logs[j]-Math.max(lo,Math.min(hi,st.Logs[j]-g))));
  }
  const error=Math.abs(pg-st.ProjectedGradient);assert(error<2e-5);maxPGError=Math.max(maxPGError,error);
  if(st.Stop==='projected-gradient')assert(pg<=1.2e-6);
  for(let j=0;j<32;j++){
   assert.equal(r.Q[j],s.Steps[clock+j].Q);assert.equal(r.Y[j],s.Steps[clock+j].Y);
   for(const [actual,want]of[[fitted.predict(s.Steps[clock+j].X),r.Predictions[j]],[old.predict(s.Steps[clock+j].X),r.OldPredictions[j]]]){
    const error=Math.abs(actual-want);assert(error<1e-10);maxForecastError=Math.max(maxForecastError,error);forecasts++;
   }
  }
  fits++;
 }
}
assert((await output.next()).done);assert.equal(rows,2688);assert.equal(fits,1008);assert.equal(forecasts,64512);
console.log(JSON.stringify({rows,fits,forecasts,objectives,derivatives,maxForecastError,maxObjectiveError,maxPGError,pass:true},null,2));

