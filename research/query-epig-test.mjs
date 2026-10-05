import assert from 'node:assert/strict';
import {binaryEntropy,predictiveInformation} from './query-epig.mjs';
const close=(a,b)=>assert(Math.abs(a-b)<1e-12),kl=(p,q)=>(p===0?0:p*Math.log(p/q))+(p===1?0:(1-p)*Math.log((1-p)/(1-q)));
assert.equal(binaryEntropy(0),0);assert.equal(binaryEntropy(1),0);close(binaryEntropy(.5),Math.log(2));
let checks=0;
for(const mass of [.1,.3,.5,.7,.9])for(const a of [0,.1,.5,.9,1])for(const b of [0,.1,.5,.9,1]){
  const p=(1-mass)*a+mass*b,gain=predictiveInformation([p],[[a],[b]],[1-mass,mass],[1]);
  if(p>0&&p<1)close(gain,(1-mass)*kl(a,p)+mass*kl(b,p));else close(gain,0);
  close(gain,predictiveInformation([p],[[b],[a]],[mass,1-mass],[1]));assert(gain<=binaryEntropy(p)+1e-12);checks++;
}
close(predictiveInformation([.5],[[0],[1]],[.5,.5],[1]),Math.log(2));
close(predictiveInformation([.2],[[.2],[.2]],[.5,.5],[1]),0);
const base=[.5,.25],conditional=[[.1,.1],[.9,.4]],mass=[.5,.5],weights=[2,1],before=structuredClone({base,conditional,mass,weights});
close(predictiveInformation(base,conditional,mass,weights),predictiveInformation([.5,.5,.25],[[.1,.1,.1],[.9,.9,.4]],mass,[1,1,1]));assert.deepEqual({base,conditional,mass,weights},before);
assert.throws(()=>binaryEntropy(NaN));assert.throws(()=>predictiveInformation([.5],[[.1],[.9]],[.5,.5],[0]));assert.throws(()=>predictiveInformation([.5],[[.1],[.9]],[.2,.2],[1]));assert.throws(()=>predictiveInformation([.7],[[.1],[.9]],[.5,.5],[1]));
console.log(`PASS: ${checks} entropy/KL, relabeling and upper-bound fixtures; endpoints, zero information, multiplicity, ownership and invalid inputs`);
