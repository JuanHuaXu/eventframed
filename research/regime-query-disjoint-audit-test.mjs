import assert from 'node:assert/strict';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';

const source={Phase:0,Case:0,Index:0,Schedule:0,Steps:Array.from({length:256},(_,i)=>({X:i,Y:false,Q:.2,Delay:0,Missing:false}))};
const record={Phase:0,Case:0,Index:0,Schedule:0,Decision:{Clock:160,Origins:Array.from({length:63},(_,i)=>97+i),Pool:[],Probes:Array.from({length:8},(_,i)=>137+i),Values:[],Selected:[-1,-1,-1,-1],Costs:[0,0,0,0]},
  Origins:Array.from({length:4},()=>Array.from({length:64},(_,i)=>97+i)),Redundant:[false,false,false,false],LogEvidence:[-1,-1,-1,-1],
  AtPublication:Array.from({length:4},()=>Array(31).fill(.4)),Predictions:Array.from({length:4},()=>Array.from({length:31},(_,i)=>.5+.99**i*(.4-.5))),ActualFits:1};
auditRegimeOutcome(record,source,137);
assert.throws(()=>auditRegimeOutcome(record,source));
assert.throws(()=>auditRegimeOutcome(record,source,145));
const wrong=structuredClone(record);wrong.Decision.Probes[0]=152;
assert.throws(()=>auditRegimeOutcome(wrong,source,137));
const old=structuredClone(record);old.Decision.Probes=Array.from({length:8},(_,i)=>153+i);
auditRegimeOutcome(old,source);
assert.throws(()=>auditRegimeOutcome(old,source,137));
console.log('PASS: old/disjoint probe contracts accepted only under their declared window; undeclared and mixed windows rejected');
