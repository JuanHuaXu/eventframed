import assert from 'node:assert/strict';
import {auditEnvelope} from './regime-query-envelope-audit.mjs';

function fixture(schedule){
  const source={Phase:0,Case:0,Index:0,Schedule:schedule,Steps:Array.from({length:256},(_,i)=>({X:i,Y:false,Q:.5,Delay:0,Missing:Boolean(schedule)}))};
  const choices=schedule?[-1,152,153,154,155,156,157,158,159]:[-1];
  const r={Phase:0,Case:0,Index:0,Schedule:schedule,Choices:choices,Bundles:[],ActualFits:0};
  for(let start=0;start<choices.length;start+=4){
    const selected=Array.from({length:4},(_,a)=>choices[Math.min(start+a,choices.length-1)]);
    const natural=schedule?Array.from({length:16},(_,i)=>i-16):Array.from({length:177},(_,i)=>i-16);
    const origins=selected.map(s=>[...new Set([...natural,...(s<0?[]:[s])])].sort((a,b)=>a-b).slice(-64));
    const bundle={Phase:0,Case:0,Index:0,Schedule:schedule,Decision:{Clock:160,Selected:selected,Costs:selected.map(s=>Number(s>=0))},
      Origins:origins,Redundant:selected.map(s=>s>=0&&natural.includes(s)),LogEvidence:[0,0,0,0],
      AtPublication:Array.from({length:4},()=>Array(31).fill(.5)),Predictions:Array.from({length:4},()=>Array(31).fill(.5)),
      ActualFits:new Set(origins.map(a=>a.join(','))).size};
    r.Bundles.push(bundle);r.ActualFits+=bundle.ActualFits;
  }
  return {r,source};
}
for(const schedule of [0,1]){const {r,source}=fixture(schedule);assert.equal(auditEnvelope(r,source).branches.length,schedule?9:1);}
const {r,source}=fixture(1);
const corruptions=[
  x=>x.Choices.pop(),
  x=>x.Bundles[0].Decision.Selected[1]=160,
  x=>x.Bundles[0].Decision.Costs[1]=0,
  x=>x.Bundles[0].Origins[1].push(155),
  x=>x.Bundles[0].Redundant[1]=true,
  x=>x.Bundles[0].Predictions[1][10]=.6,
  x=>x.Bundles[0].AtPublication[1][0]=NaN,
  x=>x.ActualFits++,
  x=>x.Bundles[2].Predictions[1][0]=.4,
];
for(const corrupt of corruptions){const x=structuredClone(r);corrupt(x);assert.throws(()=>auditEnvelope(x,source));}
console.log('PASS: two schedule fixtures; nine enumeration, support, cost, redundancy, probability, aging, fit and padding corruptions rejected');
