import assert from 'node:assert/strict';
import {noiseEnvelope} from './noise-envelope.mjs';
export const TYPES=[0,1,2,7];
const average=a=>a.reduce((s,v)=>s+v,0)/a.length;
export function acquisitionBelief(history) {
  const masses=[0,0,0,0];
  for(const model of noiseEnvelope(history)){
    const prior=.5/16+(model.mask===0||model.mask===15?.25:0);
    model.classCoefficients.forEach((c,y)=>masses[y]+=prior*average(c));
  }
  const mass=masses.reduce((s,v)=>s+v,0);assert(mass>0);
  return {mass,forecast:masses.map(v=>v/mass)};
}
export function acquisitionChoice(history,policy,random,cache=new Map()) {
  assert(history.length>=4&&history.length<10);
  const belief=h=>{const key=JSON.stringify(h);if(!cache.has(key))cache.set(key,acquisitionBelief(h));return cache.get(key);};
  const parent=belief(history),options=TYPES.map(t=>[1,t,history.filter(s=>s.action[0]===1&&s.action[1]===t).length]);
  if(policy==='random')return options[Math.floor(random()*options.length)];
  if(policy==='fixed')return options[TYPES.indexOf([0,1,2,7,0,1][history.length-4])];
  assert(['target','entropy'].includes(policy));
  const scores=options.map(action=>{
    const children=[false,true].map(outcome=>belief([...history,{action,outcome}]));
    const probabilities=children.map(b=>b.mass/parent.mass);
    assert(Math.abs(probabilities[0]+probabilities[1]-1)<1e-12);
    if(policy==='entropy')return probabilities.reduce((s,p)=>s+(p>0?p*Math.log(p):0),0);
    return children.reduce((s,b,i)=>s+probabilities[i]*(1-b.forecast.reduce((a,p)=>a+p*p,0)),0);
  });
  return options[scores.indexOf(Math.min(...scores))];
}

