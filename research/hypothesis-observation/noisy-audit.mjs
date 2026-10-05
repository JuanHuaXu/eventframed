import assert from 'node:assert/strict';
import {noiseEnvelope} from './noise-envelope.mjs';
import {canonicalHistory} from './count-planning.mjs';
import {acquisitionBelief} from './interval-acquisition.mjs';
import {tieChoice} from './tie-policy.mjs';
export function auditLikelihood(mask,audit,reliability=.8){
  assert(Number.isInteger(mask)&&mask>=0&&mask<16&&[0,1].includes(audit)&&reliability>=.5&&reliability<1);
  return Number(!!(mask&1))===audit?reliability:1-reliability;
}
export function auditBelief(pattern,counts,audit,reliability=.8){
  const joint=[0,0,0,0];
  for(const m of noiseEnvelope(canonicalHistory(pattern,counts))){
    const prior=.5/16+(m.mask===0||m.mask===15?.25:0),weight=prior*auditLikelihood(m.mask,audit,reliability);
    m.classCoefficients.forEach((c,y)=>joint[y]+=weight*c.reduce((a,v)=>a+v,0)/c.length);
  }
  const mass=joint.reduce((a,v)=>a+v,0);assert(mass>0);
  return {mass,forecast:joint.map(v=>v/mass)};
}
export function compileAudit(pattern,audit){
  const map=new Map();
  function state(counts){
    const key=counts.join(',');if(map.has(key))return map.get(key);
    const b=auditBelief(pattern,counts,audit),step=counts.reduce((a,v)=>a+v,0),s={counts,...b,actions:null};map.set(key,s);
    if(step<5){
      const target=[],entropy=[];
      for(let a=0;a<4;a++){
        const children=[0,1].map(y=>{const c=counts.slice();c[2*a+y]++;return state(c);}),p=children.map(c=>c.mass/s.mass);
        assert(Math.abs(p[0]+p[1]-1)<1e-12);
        target.push(children.reduce((v,c,y)=>v+p[y]*(1-c.forecast.reduce((v,q)=>v+q*q,0)),0));
        entropy.push(p.reduce((v,q)=>v+q*Math.log(q),0));
      }
      s.actions={target:tieChoice(target),entropy:tieChoice(entropy)};
    }
    return s;
  }
  const zero=Array(8).fill(0);state(zero);assert.equal(map.size,1287);
  return {pattern,audit,preAudit:acquisitionBelief(canonicalHistory(pattern,zero)).forecast,states:[...map.values()]};
}
export function auditPaths(root,policy){
  assert(['target','entropy','random'].includes(policy));
  const index=new Map(root.states.map((s,i)=>[s.counts.join(','),i])),layers=[new Map([[index.get('0,0,0,0,0,0,0,0'),1]])];
  for(let step=0;step<5;step++){
    const next=new Map();for(const[i,w]of layers[step]){
      const s=root.states[i];for(let a=0;a<4;a++){
        const weight=policy==='random'?.25:Number(s.actions[policy]===a);if(!weight)continue;
        for(let y=0;y<2;y++){const c=s.counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);next.set(j,(next.get(j)||0)+w*weight);}
      }
    }layers.push(next);
  }return layers;
}
