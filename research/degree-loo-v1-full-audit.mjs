import assert from 'node:assert/strict';
import fs from 'node:fs';
import {createInterface} from 'node:readline';
async function* rows(p){for await(const s of createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))if(s.trim())yield JSON.parse(s);}
const dir='docs/experiments/';
const source=rows(dir+'mmm-soft-learners-v120.jsonl');
const predictions=rows(process.argv[2]);
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const pop=x=>{let n=0;for(;x;x&=x-1)n++;return n;};
const choose=(n,k)=>{if(k<0||k>n)return 0;let r=1;for(let i=1;i<=k;i++)r=r*(n-i+1)/i;return r;};
const basis=Array.from({length:10},(_,h)=>Array.from({length:4},(_,j)=>{
 const d=j+1;let v=0;for(let i=0;i<=d;i++)v+=(i%2?-1:1)*choose(h,i)*choose(9-h,d-i);return v/choose(9,d);
}));
// Gauss-Jordan with partial pivoting, independent of the Go Cholesky path.
function evaluate(s,u,clipped) {
 const n=s.length,k=(x,y)=>1+basis[pop(x^y)].reduce((sum,g,d)=>sum+Math.exp(u[d])*g,0);
 const a=s.map((r,i)=>[...s.map((v,j)=>k(r.x,v.x)+(i===j?1:0)),...Array.from({length:n},(_,j)=>Number(i===j))]);
 for(let c=0;c<n;c++) {
  let p=c;for(let i=c+1;i<n;i++)if(Math.abs(a[i][c])>Math.abs(a[p][c]))p=i;
  [a[c],a[p]]=[a[p],a[c]];const v=a[c][c];assert(Math.abs(v)>1e-10);
  for(let j=0;j<2*n;j++)a[c][j]/=v;
  for(let i=0;i<n;i++)if(i!==c){const f=a[i][c];for(let j=0;j<2*n;j++)a[i][j]-=f*a[c][j];}
 }
 const inv=a.map(r=>r.slice(n)),alpha=inv.map(r=>r.reduce((sum,v,j)=>sum+v*s[j].y,0));
 let value=0;
 for(let i=0;i<n;i++) {
  const residual=alpha[i]/inv[i][i];
  if(!clipped)value+=residual**2/(4*n);
  else {const p=Math.max(1e-12,Math.min(1-1e-12,(1+s[i].y-residual)/2));value+=(p-(1+s[i].y)/2)**2/n;}
 }
 return {value,predict:x=>Math.max(1e-12,Math.min(1-1e-12,(1+s.reduce((sum,r,i)=>sum+alpha[i]*k(r.x,x),0))/2))};
}
let records=0,stats=0,fits=0,forecasts=0,derivatives=0,maxError=0,maxObjectiveError=0,maxPGError=0;
const stops=Array.from({length:4},()=>({})),totals=Array.from({length:4},()=>({fits:0,evaluations:0,iterations:0,initial:0,final:0}));
for await(const s of source) {
 const {value:r,done}=await predictions.next();assert(!done);
 for(const key of ['Phase','Case','Index','Schedule'])assert.equal(r[key],s[key]);
 assert.equal(r.Fits,48);assert.equal(r.P.length,256);records++;
 for(let pub=0;pub<8;pub++)for(let arm=0;arm<4;arm++) {
  const st=r.Optimization[pub][arm];stats++;
  assert(st.Evaluations>=1&&st.Evaluations<=1345&&st.Iterations>=0&&st.Iterations<=64);
  assert.equal(st.Trace.length,st.Iterations+1);assert.equal(st.Trace[0],st.Initial);assert.equal(st.Trace.at(-1),st.Final);
  assert(st.Final<=st.Initial&&Number.isFinite(st.Final));assert.equal(st.Logs[4],0);
  for(let i=1;i<st.Trace.length;i++)assert(st.Trace[i]<=st.Trace[i-1]+1e-13);
  assert(['iteration-cap','line-search-cap','projected-gradient'].includes(st.Stop));
  if(st.Stop==='projected-gradient')assert(st.ProjectedGradient<=1e-6);
  for(let j=0;j<4;j++)assert(Number.isFinite(st.Logs[j])&&st.Logs[j]>=Math.log(1e-4)-1e-14&&st.Logs[j]<=Math.log(16)+1e-14);
  stops[arm][st.Stop]=(stops[arm][st.Stop]??0)+1;
  const total=totals[arm];total.fits++;total.evaluations+=st.Evaluations;total.iterations+=st.Iterations;total.initial+=st.Initial;total.final+=st.Final;
  if(s.Index!==0||![0,4,7].includes(pub))continue;
  const clock=pub*32,cap=arm%2?32:64;
  let origins=Array.from({length:16},(_,i)=>i-16);
  for(let i=0;i<clock;i++)if(!s.Steps[i].Missing&&i+s.Steps[i].Delay<=clock)origins.push(i);
  origins=origins.slice(-cap);assert.deepEqual(origins,s.Fits[pub].Origins[arm%2]);
  const samples=origins.map(i=>i<0?{x:s.Initial[i+16].Bits,y:s.Initial[i+16].Outcome?1:-1}:{x:s.Steps[i].X,y:s.Steps[i].Y?1:-1});
  const clipped=arm>=2,fit=evaluate(samples,st.Logs,clipped),initial=evaluate(samples,[Math.log(9),Math.log(9),Math.log(9),Math.log(9),0],clipped);
  for(const [a,b]of [[fit.value,st.Final],[initial.value,st.Initial]]) {const error=Math.abs(a-b);assert(error<1e-9);maxObjectiveError=Math.max(maxObjectiveError,error);}
  let pg=0;
  for(let j=0;j<4;j++) {
   const a=st.Logs.slice(),b=st.Logs.slice();a[j]+=1e-5;b[j]-=1e-5;
   const g=(evaluate(samples,a,clipped).value-evaluate(samples,b,clipped).value)/2e-5;
   pg=Math.max(pg,Math.abs(st.Logs[j]-Math.max(Math.log(1e-4),Math.min(Math.log(16),st.Logs[j]-g))));derivatives++;
  }
  const error=Math.abs(pg-st.ProjectedGradient);assert(error<2e-5);maxPGError=Math.max(maxPGError,error);
  for(let t=clock;t<clock+32;t++){const e=Math.abs(fit.predict(s.Steps[t].X)-r.P[t][arm+2]);assert(e<1e-10);maxError=Math.max(maxError,e);forecasts++;}
  fits++;
 }
}
assert((await predictions.next()).done);assert.equal(records,2688);assert.equal(stats,86016);assert.equal(fits,1008);assert.equal(forecasts,32256);
for(const t of totals)for(const k of ['evaluations','iterations','initial','final'])t[k]/=t.fits;
console.log(JSON.stringify({records,stats,fits,forecasts,derivatives,maxError,maxObjectiveError,maxPGError,stops,totals,pass:true},null,2));
