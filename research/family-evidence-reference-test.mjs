import assert from 'node:assert/strict';
import {familyReference} from './family-evidence-reference.mjs';
for(const x of [0,1,511])for(const y of [false,true])for(const mass of [0,.5,.95,1]){
 const r=familyReference([[x,y]],mass);assert(Math.abs(r.logEvidence-Math.log(.5))<1e-12);assert(Math.abs(r.genericMass-mass)<1e-12);
}
// Repeating one input reduces both model families to the same beta rate.
for(const n of [2,16,32,64]){
 const samples=Array.from({length:n},(_,i)=>[73,i%3===0]);
 const r=familyReference(samples),s=familyReference([...samples].reverse());
 assert.deepEqual(r,s);assert(Math.abs(r.genericLog-r.booleanLog)<1e-12);assert(Math.abs(r.genericMass-.95)<1e-12);
}
assert.throws(()=>familyReference([]));
const factorial=n=>{let p=1;for(let i=2;i<=n;i++)p*=i;return p;};
const integral=(a,b)=>factorial(2*a)*factorial(2*b)/(4**(a+b)*factorial(a)*factorial(b)*factorial(a+b));
for(let n=2;n<=6;n++){
 const samples=Array.from({length:n},(_,i)=>[(i*73+17)%512,i%3===0]);let zg=0,zb=0;
 for(let mask=0;mask<512;mask++){
  const cells=new Map();let k=0;for(let x=mask;x;x&=x-1)k++;
  let agreement=0;
  for(const [x,y] of samples){const key=x&mask,pair=cells.get(key)??[0,0];pair[Number(y)]++;cells.set(key,pair);let odd=false;for(let v=key;v;v&=v-1)odd=!odd;if(odd===y)agreement++;}
  let g=1;for(const [no,yes] of cells.values())g*=integral(no,yes);
  const prior=(1/3)**k*(2/3)**(9-k);zg+=prior*g;zb+=prior*integral(n-agreement,agreement);
 }
 const got=familyReference(samples);assert(Math.abs(Math.exp(got.genericLog)-zg)<1e-12);assert(Math.abs(Math.exp(got.booleanLog)-zb)<1e-12);assert(Math.abs(Math.exp(got.logEvidence)-(.95*zg+.05*zb))<1e-12);
}
console.log('PASS: 24 single-observation identities, four repeated-input cases, five exhaustive-mask factorial references, empty input rejection');
