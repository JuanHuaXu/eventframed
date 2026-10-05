import assert from 'node:assert/strict';
import {acquisitionBelief,TYPES} from './interval-acquisition.mjs';
const risk=b=>1-b.forecast.reduce((s,p)=>s+p*p,0);
export function lookaheadChoice(history,cache=new Map(),requestedDepth=2) {
  assert([1,2].includes(requestedDepth));
  assert(history.length>=4&&history.length<10);
  const belief=h=>{const key=JSON.stringify(h);if(!cache.has(key))cache.set(key,acquisitionBelief(h));return cache.get(key);};
  function solve(h,depth){
    if(depth===0)return {cost:0};
    const parent=belief(h);
    const options=TYPES.map(t=>[1,t,h.filter(s=>s.action[0]===1&&s.action[1]===t).length]);
    const costs=options.map(action=>{
      const children=[false,true].map(outcome=>[...h,{action,outcome}]);
      const probabilities=children.map(child=>belief(child).mass/parent.mass);
      assert(Math.abs(probabilities[0]+probabilities[1]-1)<1e-12);
      return children.reduce((s,child,i)=>s+probabilities[i]*(risk(belief(child))+solve(child,depth-1).cost),0);
    });
    const best=costs.indexOf(Math.min(...costs));return {action:options[best],cost:costs[best]};
  }
  return solve(history,Math.min(requestedDepth,10-history.length)).action;
}

