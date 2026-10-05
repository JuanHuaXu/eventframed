import assert from 'node:assert/strict';
import {acquisitionBelief,acquisitionChoice,TYPES} from './interval-acquisition.mjs';
import {lookaheadChoice} from './interval-lookahead.mjs';
const risk=h=>{const b=acquisitionBelief(h);return {mass:b.mass,risk:1-b.forecast.reduce((s,p)=>s+p*p,0)};};
let enumerations=0,oneStep=0;
for(const bits of [0,1,3,5,7,10,12,15])for(const length of [4,6,9]){
  const h=TYPES.map((t,i)=>({action:[0,t,0],outcome:!!(bits&(1<<i))}));
  while(h.length<length){const t=[0,7,1,2,0][h.length-4];h.push({action:[1,t,h.filter(s=>s.action[0]===1&&s.action[1]===t).length],outcome:!!((bits+h.length)%2)});}
  const before=JSON.stringify(h);
  assert.deepEqual(lookaheadChoice(h,new Map(),1),acquisitionChoice(h,'target',()=>{throw Error('Unexpected RNG');}));oneStep++;
  const selected=lookaheadChoice(h),parent=risk(h),scores=[];
  for(const t of TYPES){
    const action=[1,t,h.filter(s=>s.action[0]===1&&s.action[1]===t).length];let cost=0;
    for(const outcome of [false,true]){
      const child=[...h,{action,outcome}],b=risk(child);cost+=b.mass/parent.mass*b.risk;
      if(h.length<9){
        let best=Infinity;
        for(const next of TYPES){
          const a=[1,next,child.filter(s=>s.action[0]===1&&s.action[1]===next).length];let second=0;
          for(const y of [false,true]){const g=risk([...child,{action:a,outcome:y}]);second+=g.mass/parent.mass*g.risk;}
          best=Math.min(best,second);
        }
        cost+=best;
      }
    }
    scores.push(cost);
  }
  assert(scores[TYPES.indexOf(selected[1])]<=Math.min(...scores)+1e-12);enumerations++;
  assert.equal(JSON.stringify(h),before);
}
assert.throws(()=>lookaheadChoice([],new Map()));
console.log(JSON.stringify({twoStepEnumerations:enumerations,oneStepParity:oneStep,negativeTests:1}));

