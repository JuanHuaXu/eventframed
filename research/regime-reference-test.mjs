import assert from 'node:assert/strict';
import {regimeReference} from './regime-predictive-reference.mjs';
import {familyReference} from './family-evidence-reference.mjs';
const queries=[0,1,17,73,511];
assert.deepEqual(regimeReference([],queries).predictions,queries.map(()=>.5));
for(const n of [1,6,16,32,63]){
 const s=Array.from({length:n},(_,i)=>[(i*73+17)%512,i%3===0]),got=regimeReference(s,queries),ref=familyReference(s);
 for(const k of ['genericLog','booleanLog','logEvidence','genericMass'])assert(Math.abs(got[k]-ref[k])<1e-11);
 queries.forEach((x,i)=>{const joint=familyReference([...s,[x,true]]);assert(Math.abs(Math.exp(joint.logEvidence-ref.logEvidence)-got.predictions[i])<1e-11);});
}
console.log('PASS: empty prior, five likelihood comparisons, 25 posterior-predictive ratios through64 observations');
