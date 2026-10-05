import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

// Values are integer thousandths; squared errors and thresholds use millionths.
// BigInt avoids a floating-point or PDF-extraction ambiguity in the implication.
const abs=x=>x<0n?-x:x,square=x=>x*x;
function check(name,truth,best,learned,tau=10000n,c=2n){
  assert(truth.length===best.length&&best.length===learned.length);
  const bestRisk=best.reduce((s,b,i)=>s+square(b-truth[i]),0n),learnedRisk=learned.reduce((s,b,i)=>s+square(b-truth[i]),0n);
  assert(bestRisk<learnedRisk); // F contains precisely these two predictors.
  const B=best.reduce((s,b,i)=>abs(b-truth[i])>s?abs(b-truth[i]):s,0n);assert(B>0n);
  const points=truth.map((t,i)=>{
    const gap=abs(learned[i]-best[i]),product=(learned[i]-best[i])*(best[i]-t);
    const antecedent=gap*B>=tau+c*B*B,conclusion=product>=tau;
    return {index:i,gap:Number(gap)/1000,product:Number(product)/1e6,antecedent,conclusion};
  });
  assert(points.some(p=>p.antecedent&&!p.conclusion));
  return {name,truth:truth.map(Number),best:best.map(Number),learned:learned.map(Number),units:'input arrays are thousandths',B:Number(B)/1000,tau:Number(tau)/1e6,threshold:Number(tau+c*B*B)/Number(B)/1000,bestSquaredRisk:Number(bestRisk)/truth.length/1e6,learnedSquaredRisk:Number(learnedRisk)/truth.length/1e6,points};
}
const examples=[check('missing sign',[500n],[400n],[800n]),check('upper bias bound is not a pointwise lower bound',[500n,500n],[501n,600n],[900n,600n])];
// Bernoulli observations equal to one make the learned predictor an ERM in both
// examples; that finite sample has positive probability under the true law.
for(const e of examples){const empirical=a=>a.reduce((s,v)=>s+(1000-v)**2,0);assert(empirical(e.learned)<empirical(e.best));}
let antecedents=0,violations=0;
for(let t=0n;t<=1000n;t+=100n)for(let b=0n;b<=1000n;b+=100n)for(let h=0n;h<=1000n;h+=100n){
  const B=abs(b-t);if(B===0n||square(b-t)>=square(h-t))continue;
  if(abs(h-b)*B>=10000n+2n*B*B){antecedents++;if((h-b)*(b-t)<10000n)violations++;}
}
assert(violations>0);
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
console.log(JSON.stringify({scope:'Counterexamples to the stated Theorem2 subset implication in Tang/Sloman/Kaski2026; not a reproduction or refutation of empirical R-IDeA results',examples,grid:{antecedents,violations},scriptSHA256:sha(import.meta.filename)}));
