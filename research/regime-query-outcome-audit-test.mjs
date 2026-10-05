import assert from 'node:assert/strict';
import {auditRegimeOutcome, randomOrigin, maximumOrigin} from './regime-query-outcome-audit.mjs';

// Synthetic arithmetic fixture, not a fitted-model or efficacy reference.
function fixture(schedule) {
  const source = {Phase:0,Case:0,Index:0,Schedule:schedule,
    Initial:Array.from({length:16},()=>({Bits:0,Outcome:false})),
    Steps:Array.from({length:256},(_,i)=>({X:i%512,Y:false,Q:.3,Delay:schedule?20:0,Missing:false}))};
  const available=t=>{const a=Array.from({length:16},(_,i)=>i-16);for(let j=0;j<t;j++)if(j+source.Steps[j].Delay<=t)a.push(j);return a;};
  const pool=schedule?Array.from({length:8},(_,i)=>152+i):[];
  const selected=[-1,...(schedule?[randomOrigin(source,pool),152,152]:[-1,-1,-1])];
  const values=pool.map(Origin=>({Origin,Mass:[.5,.5],LogEvidence:[-10-Math.log(2),-10-Math.log(2)],Base:Array(8).fill(.4),Conditional:[Array(8).fill(.3),Array(8).fill(.5)],Gain:.01}));
  const record={Phase:0,Case:0,Index:0,Schedule:schedule,Decision:{Clock:160,Origins:available(160).slice(-63),Pool:pool,Probes:source.Steps.slice(153,161).map(s=>s.X),Values:values,BaseLogEvidence:schedule?-10:0,Selected:selected,Costs:schedule?[0,1,1,1]:[0,0,0,0]},
    Origins:selected.map(j=>[...new Set([...available(161),...(j<0?[]:[j])])].sort((a,b)=>a-b).slice(-64)),
    Redundant:[false,false,false,false],AtPublication:Array.from({length:4},()=>Array(31).fill(.4)),
    Predictions:Array.from({length:4},()=>Array.from({length:31},(_,i)=>.5+.99**i*(.4-.5))),LogEvidence:[-9,-9,-9,-9]};
  record.ActualFits=new Set(record.Origins.map(a=>a.join(','))).size;
  return {source,record};
}
for(const schedule of [0,1]) {const {source,record}=fixture(schedule);auditRegimeOutcome(record,source);}
const {source,record}=fixture(1);
for(const mutate of [
  r=>r.Decision.Costs[1]=0,
  r=>r.Decision.Selected[2]=159,
  r=>r.Decision.Values[0].Gain+=.01,
  r=>r.Decision.Values[0].Mass[0]+=.01,
  r=>r.Origins[3][0]=-9,
  r=>r.Redundant[3]=true,
  r=>r.Predictions[3][0]+=.01,
  r=>r.ActualFits++
]) {const bad=structuredClone(record);mutate(bad);assert.throws(()=>auditRegimeOutcome(bad,source));}
assert.equal(maximumOrigin([9,3,1],[.5,.5,.1]),3);
assert.equal(maximumOrigin([1,3,9],[.1,.5,.5]),3);
assert.throws(()=>maximumOrigin([],[]));
assert.throws(()=>maximumOrigin([1],[NaN]));
console.log('PASS: complete/delayed arithmetic fixtures, eight rejected mutations, deterministic ties and invalid selectors');
