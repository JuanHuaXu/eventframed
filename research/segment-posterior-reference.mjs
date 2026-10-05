// Tiny-model mathematical QA only. This does not implement the Go learner,
// delayed serving lifecycle, truncation policy, or an experiment rescue.
import crypto from 'node:crypto';
import fs from 'node:fs';
import assert from 'node:assert/strict';

function segment(lo,hi,ys,seen,xq) {
  const joint=[],prediction=[];
  for(const split of [false,true]) {
    const counts=[[0,0],[0,0]];let likelihood=1;
    for(let j=lo;j<=hi;j++)if(seen&(1<<j)){
      const group=split?j%2:0,y=(ys>>j)&1,c=counts[group];
      likelihood*=(c[y]+.5)/(c[0]+c[1]+1);c[y]++;
    }
    const c=counts[split?xq:0];
    joint.push(likelihood/2);prediction.push((c[1]+.5)/(c[0]+c[1]+1));
  }
  const evidence=joint[0]+joint[1];
  return {evidence,p:(joint[0]*prediction[0]+joint[1]*prediction[1])/evidence};
}

function dynamic(n,ys,seen,h,xq) {
  if(!Number.isInteger(n)||n<0||n>6||!Number.isInteger(seen)||seen<0||seen>=(1<<n)||!Number.isInteger(ys)||ys<0||ys>=(1<<n)||!(h>0&&h<1)||![0,1].includes(xq))throw Error('invalid tiny-model input');
  if(n===0)return {evidence:1,p:.5};
  const f=[1];let weighted=0;
  for(let b=0;b<n;b++){
    let sum=0;
    for(let a=0;a<=b;a++){
      const s=segment(a,b,ys,seen,xq),term=f[a]*(a===0?1:h)*(1-h)**(b-a)*s.evidence;
      sum+=term;if(b===n-1)weighted+=term*s.p;
    }
    f.push(sum);
  }
  return {evidence:f[n],p:h/2+(1-h)*weighted/f[n]};
}

// Independent exact beta integral via half-integer factorial identities.
// No use of the DP's sequential predictive likelihood calculation.
const factorial=n=>{let f=1;for(let i=2;i<=n;i++)f*=i;return f;};
const betaIntegral=(fail,success)=>factorial(2*fail)*factorial(2*success)/(4**(fail+success)*factorial(fail)*factorial(success)*factorial(fail+success));
function referenceMarginal(lo,hi,ys,seen,xq,appendSuccess=false){
  const total=[0,0],groups=[[0,0],[0,0]];
  for(let j=lo;j<=hi;j++)if(seen&(1<<j)){const y=(ys>>j)&1;total[y]++;groups[j%2][y]++;}
  if(appendSuccess){total[1]++;groups[xq][1]++;}
  return (betaIntegral(...total)+betaIntegral(...groups[0])*betaIntegral(...groups[1]))/2;
}
function exhaustive(n,ys,seen,h,xq){
  if(n===0)return {evidence:1,p:.5};
  let evidence=0,weighted=0;
  for(let cuts=0;cuts<(1<<(n-1));cuts++){
    let prior=1,start=0,product=1,last=0;
    for(let j=0;j<n;j++){
      const cut=j<n-1&&Boolean(cuts&(1<<j));
      if(j<n-1)prior*=cut?h:1-h;
      if(cut||j===n-1){
        const m=referenceMarginal(start,j,ys,seen,xq);
        product*=m;
        if(j===n-1)last=referenceMarginal(start,j,ys,seen,xq,true)/m;
        start=j+1;
      }
    }
    evidence+=prior*product;weighted+=prior*product*last;
  }
  return {evidence,p:h/2+(1-h)*weighted/evidence};
}

let comparisons=0,maxEvidenceError=0,maxPredictionError=0,unseenChecks=0;
const digest=crypto.createHash('sha256');
for(let n=0;n<=6;n++)for(let ys=0;ys<(1<<n);ys++)for(let seen=0;seen<(1<<n);seen++)for(const h of [.01,.1,.5])for(const xq of [0,1]){
  const a=dynamic(n,ys,seen,h,xq),b=exhaustive(n,ys,seen,h,xq);
  maxEvidenceError=Math.max(maxEvidenceError,Math.abs(a.evidence-b.evidence));
  maxPredictionError=Math.max(maxPredictionError,Math.abs(a.p-b.p));
  assert(Math.abs(a.evidence-b.evidence)<1e-13&&Math.abs(a.p-b.p)<1e-13);
  assert(a.evidence>0&&a.evidence<=1+1e-14&&a.p>0&&a.p<1);
  if(seen===0){assert(Math.abs(a.evidence-1)<1e-14);assert(Math.abs(a.p-.5)<1e-14);}
  const changed=ys^(((1<<n)-1)^seen);
  assert.deepEqual(a,dynamic(n,changed,seen,h,xq));unseenChecks++;
  digest.update(JSON.stringify([n,ys,seen,h,xq,a]));comparisons++;
}
for(const args of [[7,0,0,.1,0],[1,0,2,.1,0],[1,0,0,0,0],[1,0,0,1,0],[1,0,0,NaN,0],[1,0,0,.1,2]])assert.throws(()=>dynamic(...args));
assert.equal(comparisons,32766);
process.stdout.write(JSON.stringify({scope:'tiny segment-posterior formulation QA; not quality validation',node:process.version,sourceSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),comparisons,unseenChecks,maxEvidenceError,maxPredictionError,resultDigest:digest.digest('hex')},null,2)+'\n');
