import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';
async function* records(path) {
 for await(const line of createInterface({input:createReadStream(path),crlfDelay:Infinity})) if(line.trim()) yield JSON.parse(line);
}
const source=records('docs/experiments/mmm-soft-learners-v120.jsonl')[Symbol.asyncIterator]();
const output=records('docs/experiments/mmm-spectral-regression-v1-forecasts.jsonl')[Symbol.asyncIterator]();
assert.equal((await source.next()).value.Version,'soft-learners-v120');
const pop=x=>{let n=0;for(;x;x&=x-1)n++;return n;};
const product=(x,m)=>{let v=1;for(let bit=0;bit<9;bit++)if(m&(1<<bit))v*=x&(1<<bit)?1:-1;return v;};
const choose=(n,k)=>{if(k<0||k>n)return 0;let v=1;for(let j=1;j<=k;j++)v=v*(n-j+1)/j;return v;};
// Full kernel from Hamming distance, independent of Go's feature Gram table.
function fullKernel(x,y) {
 const h=pop(x^y);let sum=0;
 for(let d=0;d<=4;d++)for(let j=0;j<=d;j++)sum+=(j%2?-1:1)*choose(h,j)*choose(9-h,d-j);
 return sum;
}
const interactions=Array.from({length:511},(_,i)=>i+1).filter(m=>pop(m)>=2&&pop(m)<=4);
function fit(samples,mode) {
 let masks=[0,...Array.from({length:9},(_,i)=>1<<i)];
 if(mode===1) {
  const tau=Math.sqrt(2*Math.log(492/.1)/samples.length);
  masks.push(...interactions.map(m=>({m,c:Math.abs(samples.reduce((s,r)=>s+r.y*product(r.x,m),0)/samples.length)}))
   .filter(r=>r.c>=tau).sort((a,b)=>b.c-a.c||a.m-b.m).slice(0,16).map(r=>r.m));
 }
 const kernel=mode===2?fullKernel:mode===0?((x,y)=>10-2*pop(x^y)):
  ((x,y)=>masks.reduce((s,m)=>s+product(x,m)*product(y,m),0));
 const n=samples.length;
 const a=samples.map((r,i)=>[...samples.map((q,j)=>kernel(r.x,q.x)+Number(i===j)),r.y]);
 // Pivoted Gaussian elimination, not Cholesky or recovered primal coefficients.
 for(let k=0;k<n;k++) {
  let pivot=k;for(let i=k+1;i<n;i++)if(Math.abs(a[i][k])>Math.abs(a[pivot][k]))pivot=i;
  [a[k],a[pivot]]=[a[pivot],a[k]];assert(Math.abs(a[k][k])>1e-10);
  for(let i=k+1;i<n;i++) {const factor=a[i][k]/a[k][k];for(let j=k+1;j<=n;j++)a[i][j]-=factor*a[k][j];a[i][k]=0;}
 }
 const alpha=Array(n).fill(0);
 for(let i=n-1;i>=0;i--) {let v=a[i][n];for(let j=i+1;j<n;j++)v-=a[i][j]*alpha[j];alpha[i]=v/a[i][i];}
 return x=>Math.max(1e-12,Math.min(1-1e-12,(1+samples.reduce((s,r,i)=>s+alpha[i]*kernel(x,r.x),0))/2));
}
let rows=0,fits=0,forecasts=0,originsChecked=0,maxError=0;
for(;;) {
 const a=await source.next(),b=await output.next();assert.equal(a.done,b.done);if(a.done)break;
 const s=a.value,r=b.value;for(const k of ['Phase','Case','Index','Schedule'])assert.equal(s[k],r[k]);rows++;
 if(s.Index!==0)continue;
 for(const clock of [0,128,224])for(const [w,cap] of [64,32].entries()) {
  let origins=Array.from({length:16},(_,i)=>i-16);
  for(let i=0;i<clock;i++)if(!s.Steps[i].Missing&&i+s.Steps[i].Delay<=clock)origins.push(i);
  origins=origins.slice(-cap);assert.deepEqual(origins,s.Fits[clock/32].Origins[w]);originsChecked++;
  const samples=origins.map(i=>i<0?{x:s.Initial[i+16].Bits,y:s.Initial[i+16].Outcome?1:-1}:{x:s.Steps[i].X,y:s.Steps[i].Y?1:-1});
  for(let mode=0;mode<3;mode++) {
   const predict=fit(samples,mode);fits++;
   for(let t=clock;t<clock+32;t++) {
    const error=Math.abs(predict(s.Steps[t].X)-r.P[t][mode*2+w]);
    assert(error<1e-10,JSON.stringify({row:rows,clock,w,mode,t,error}));maxError=Math.max(maxError,error);forecasts++;
   }
  }
 }
}
assert.equal(rows,2688);assert.equal(fits,1512);assert.equal(forecasts,48384);
console.log(JSON.stringify({rows,fits,forecasts,originsChecked,maxError,pass:true},null,2));
