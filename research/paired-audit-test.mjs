import assert from 'node:assert/strict';
import {preparePairedAudit,identicalAudit} from './paired-audit.mjs';
import {LoggedGainEBCS} from './logged-gain-eb.mjs';
let expectations=0,endpoints=0;
for(const q of [.5,.6,.75,1])for(const m of [-1,-.3,0,.2,1])for(const d of [-1,-.7,0,.4,1]){
  const a=preparePairedAudit(q,m),included=a.observe(true,d);
  assert(Math.abs(q*included+(q===1?0:(1-q)*a.observe(false,undefined))-d)<1e-12);expectations++;
  assert(included>=a.lower-1e-12&&included<=a.upper+1e-12&&a.lower>=-3&&a.upper<=3);endpoints++;
  assert.throws(()=>a.observe(false,d));
}
for(const [q,m]of [[0,0],[.49,0],[NaN,0],[.5,2],[.5,NaN]])assert.throws(()=>preparePairedAudit(q,m));
const a=preparePairedAudit(.5,.2);assert.throws(()=>a.observe(true,undefined));assert.throws(()=>a.observe(true,2));assert.throws(()=>a.observe(0,0));
const cs=new LoggedGainEBCS();assert.equal(cs.add(0,identicalAudit,false,undefined).lower,0);
const before=cs.interval();assert.throws(()=>cs.add(1,identicalAudit,false,0));assert.deepEqual(cs.interval(),before);
console.log(JSON.stringify({status:'PASS',expectations,endpoints,scope:'Exact inclusion expectation, support, skipped outcome and exact-identity contracts; logged probabilities and branch isolation remain external obligations'}));
