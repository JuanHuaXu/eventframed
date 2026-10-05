import assert from 'node:assert/strict';
import fs from 'node:fs';
let checks=0,maxError=0;
for(let pi=0;pi<=10;pi++)for(let li=0;li<=10;li++)for(let ai=0;ai<=10;ai++)for(const y of [0,1]){
  const p=pi/10,l=li/10,a=ai/10,next=p+a*(l-p);
  const actual=(y-p)**2-(y-next)**2;
  const derived=2*a*(y-p)*(l-p)-a*a*(l-p)**2;
  const error=Math.abs(actual-derived);assert(error<1e-14);maxError=Math.max(maxError,error);checks++;
}
const y=1,other=.9,pool=.2,local=.6,oldWeight=.01,birthWeight=.1;
const pooled=(1-oldWeight)*other+oldWeight*pool;
const replaced=(1-oldWeight)*other+oldWeight*local;
const born=(1-birthWeight)*other+birthWeight*local;
assert((y-local)**2<(y-pool)**2);
assert((y-replaced)**2<(y-pooled)**2);
assert((y-born)**2>(y-replaced)**2);
const alpha=(birthWeight-oldWeight)/(1-oldWeight);
assert(Math.abs(replaced+alpha*(local-replaced)-born)<1e-14);
const out={checks,maxError,counterexample:{y,other,pool,local,oldWeight,birthWeight,pooled,replaced,born,
  pooledLoss:(y-pooled)**2,replacedLoss:(y-replaced)**2,bornLoss:(y-born)**2},
  conclusion:'Local beating pooled does not imply that increasing local weight improves the whole mixture. This finite algebra check is not a learned policy or empirical guarantee.'};
if(process.argv[2])fs.writeFileSync(process.argv[2],JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(out,null,2));
