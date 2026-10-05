import assert from 'node:assert/strict';

const bitCount=Array.from({length:512},(_,x)=>x.toString(2).replaceAll('0','').length);
const logPrior=bitCount.map(k=>k*Math.log(1/3)+(9-k)*Math.log(2/3));
const half=[0],factorial=[0];
for(let n=1;n<=64;n++){half[n]=half[n-1]+Math.log(n-.5);factorial[n]=factorial[n-1]+Math.log(n);}
const beta=Array.from({length:65},(_,n)=>Array.from({length:n+1},(_,y)=>half[y]+half[n-y]-factorial[n]));
const logSum=xs=>{const m=Math.max(...xs);return m+Math.log(xs.reduce((s,x)=>s+Math.exp(x-m),0));};

// Closed beta-integral likelihood from unordered sufficient counts, independent
// of the Go fitters' product of ordered predictive factors.
export function familyReference(samples,mass=.95) {
 assert(samples.length>0&&samples.length<=64&&mass>=0&&mass<=1);
 for(const [x,y] of samples)assert(Number.isInteger(x)&&x>=0&&x<512&&typeof y==='boolean');
 const counts=new Uint8Array(512),yes=new Uint8Array(512),g=[],b=[];
 for(let mask=0;mask<512;mask++) {
  const touched=[];let agreement=0;
  for(const [x,y] of samples){const key=x&mask;if(counts[key]===0)touched.push(key);counts[key]++;if(y)yes[key]++;if(Boolean(bitCount[key]%2)===y)agreement++;}
  let logG=0;for(const key of touched){logG+=beta[counts[key]][yes[key]];counts[key]=0;yes[key]=0;}
  g.push(logPrior[mask]+logG);b.push(logPrior[mask]+beta[samples.length][agreement]);
 }
 const genericLog=logSum(g),booleanLog=logSum(b),lg=Math.log(mass)+genericLog,lb=Math.log1p(-mass)+booleanLog;
 const logEvidence=logSum([lg,lb]);
 return {genericLog,booleanLog,logEvidence,genericMass:Math.exp(lg-logEvidence)};
}
