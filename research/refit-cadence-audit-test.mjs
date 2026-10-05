import assert from 'node:assert/strict';
import {auditCadence} from './refit-cadence-audit.mjs';
for(const missing of [false,true])for(const delay of [0,31]) {
 const raw={Phase:0,Case:0,Index:0,Schedule:1,Initial:[],Steps:Array.from({length:256},()=>({X:0,Y:false,Q:.5,P:Array(15).fill(.5),Missing:missing,Delay:delay}))};
 const result={Predictions:Array.from({length:256},()=>Array(5).fill(.5)),Queries:[],Fits:[]};
 for(let clock=0;clock<256;clock+=8){const eligible=[];for(let j=-16;j<clock;j++)if(j<0||(!missing&&j+delay<=clock))eligible.push(j);result.Fits.push({Clock:clock,Origins:[eligible.slice(-64),eligible.slice(-32)]});}
 const r={Phase:0,Case:0,Index:0,Schedule:1,Result:result};
 const checked=auditCadence(r,raw);
 for(const cadence of checked.brier)for(const arm of cadence)assert.deepEqual(arm,[.25,.25]);
 for(const mutate of [x=>x.Result.Fits[1].Origins[0].push(8),x=>x.Result.Fits[1].Clock++,x=>x.Result.Predictions[32][4]=.9]){const bad=structuredClone(r);mutate(bad);assert.throws(()=>auditCadence(bad,raw));}
}
console.log('PASS: four immediate/missing/delayed constant-forecast tapes, exact Brier and 12 rejected mutations');
