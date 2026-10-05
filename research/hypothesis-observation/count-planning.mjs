import assert from 'node:assert/strict';
import {acquisitionBelief,TYPES} from './interval-acquisition.mjs';
export function countState(history) {
  assert(history.length>=4&&history.length<=10);
  let roots=0;const counts=Array(8).fill(0),used=Array(4).fill(0);
  history.forEach(({action,outcome},i)=>{
    assert(typeof outcome==='boolean'&&action.length===3);
    const [kind,t,slot]=action,j=TYPES.indexOf(t);assert(j>=0);
    if(i<4){assert.equal(kind,0);assert.equal(j,i);assert.equal(slot,0);if(outcome)roots|=1<<j;}
    else{assert.equal(kind,1);assert.equal(slot,used[j]++);counts[2*j+Number(outcome)]++;}
  });
  return {roots,counts};
}
export function canonicalHistory(roots,counts) {
  assert(Number.isInteger(roots)&&roots>=0&&roots<16);
  assert(counts.length===8&&counts.every(n=>Number.isInteger(n)&&n>=0));
  assert(counts.reduce((s,n)=>s+n,0)<=6);
  const h=TYPES.map((t,i)=>({action:[0,t,0],outcome:!!(roots&(1<<i))}));
  TYPES.forEach((t,j)=>{let slot=0;for(let y=0;y<2;y++)for(let k=0;k<counts[2*j+y];k++)h.push({action:[1,t,slot++],outcome:!!y});});
  return h;
}
export function createCountPlanner(roots) {
  const nodes=new Map(),plans=new Map(),evaluations=new Map();
  function node(counts){
    const key=counts.join(',');
    if(!nodes.has(key)){
      const b=acquisitionBelief(canonicalHistory(roots,counts));
      nodes.set(key,{...b,counts:counts.slice(),risk:1-b.forecast.reduce((s,p)=>s+p*p,0)});
    }
    return nodes.get(key);
  }
  function branches(counts,a){
    assert(Number.isInteger(a)&&a>=0&&a<4&&counts.reduce((s,n)=>s+n,0)<6);
    const parent=node(counts),children=[0,1].map(y=>{const next=counts.slice();next[2*a+y]++;const child=node(next);return {counts:next,probability:child.mass/parent.mass,node:child};});
    assert(Math.abs(children.reduce((s,c)=>s+c.probability,0)-1)<1e-12);return children;
  }
  function plan(counts,depth){
    const remaining=6-counts.reduce((s,n)=>s+n,0),d=Math.min(depth,remaining);
    assert(Number.isInteger(d)&&d>=0);
    const key=d+':'+counts.join(',');
    if(!plans.has(key)){
      if(d===0)plans.set(key,{cost:0,action:null,costs:[]});
      else{
        const costs=Array.from({length:4},(_,a)=>branches(counts,a).reduce((s,c)=>s+c.probability*(c.node.risk+plan(c.counts,d-1).cost),0));
        const action=costs.indexOf(Math.min(...costs));plans.set(key,{cost:costs[action],action,costs});
      }
    }
    return plans.get(key);
  }
  function weights(counts,policy){
    const spent=counts.reduce((s,n)=>s+n,0);
    if(policy==='random')return [.25,.25,.25,.25];
    let action;
    if(policy==='fixed')action=[0,1,2,3,0,1][spent];
    else if(policy==='entropy'){
      const costs=Array.from({length:4},(_,a)=>branches(counts,a).reduce((s,c)=>s+c.probability*Math.log(c.probability),0));
      action=costs.indexOf(Math.min(...costs));
    }else{assert(['one','two','full'].includes(policy));action=plan(counts,policy==='one'?1:policy==='two'?2:6-spent).action;}
    return Array.from({length:4},(_,a)=>Number(a===action));
  }
  function evaluate(counts,policy){
    const key=policy+':'+counts.join(',');
    if(!evaluations.has(key)){
      const current=node(counts),remaining=6-counts.reduce((s,n)=>s+n,0);
      if(!remaining)evaluations.set(key,{postSum:0,preSum:0,finalRisk:current.risk});
      else{
        const out={postSum:0,preSum:current.risk,finalRisk:0};
        weights(counts,policy).forEach((w,a)=>{if(w)for(const c of branches(counts,a)){
          const next=evaluate(c.counts,policy),p=w*c.probability;
          out.postSum+=p*(c.node.risk+next.postSum);out.preSum+=p*next.preSum;out.finalRisk+=p*next.finalRisk;
        }});
        assert(Math.abs(out.postSum-out.preSum-out.finalRisk+current.risk)<1e-12);
        evaluations.set(key,out);
      }
    }
    return evaluations.get(key);
  }
  return {node,branches,plan,weights,evaluate,nodes};
}

