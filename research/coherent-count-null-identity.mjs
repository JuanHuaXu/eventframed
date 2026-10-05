import assert from 'node:assert/strict';
import fs from 'node:fs';
const [output]=process.argv.slice(2);assert(output);
// Conditional on m inputs falling in a fixed context, null labels have
// K~Binomial(m,1/2). Symmetric smoothing a gives p=(K+a/2)/(m+a).
// E[(p-1/2)^2] = m/[4(m+a)^2]. This is excess null Brier risk.
let checks=0,maxError=0;
for(let d=0;d<=9;d++)for(let m=0;m<=64;m++){
 const a=2*2**(-d);let probability=2**(-m),literal=0;
 for(let k=0;k<=m;k++){
  literal+=probability*((k+a/2)/(m+a)-.5)**2;
  if(k<m)probability*= (m-k)/(k+1);
 }
 const formula=m/(4*(m+a)**2),old=m/(4*(m+2)**2);
 maxError=Math.max(maxError,Math.abs(literal-formula));
 assert(Math.abs(literal-formula)<1e-13);
 if(m>0&&d>0)assert(formula>old);else assert.equal(formula,old);
 checks++;
}
const result={checks,maxError,formula:'excess null Brier = m / [4 (m+a)^2]',
 conclusion:'For every fixed nonempty observed context with depth>0, coherent uniform-prior mass2 has greater conditional expected null risk than per-context Beta(1,1). No runtime labels or observations are used by this evaluator.',
 caveat:'Fixed-context, independent null labels conditional on sample count. Not an adaptive selection or arbitrary-data theorem.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(result);
