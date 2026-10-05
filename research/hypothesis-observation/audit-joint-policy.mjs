import assert from 'node:assert/strict';
import {actualJoint} from './exact-regime.mjs';
import {auditLikelihood} from './noisy-audit.mjs';
export {bestResponse} from './joint-policy.mjs';
export function prepare(model,worlds){
  const graphs=[];
  for(let pattern=0;pattern<16;pattern++){
    const roots=[0,1].map(a=>model.roots.find(r=>r.pattern===pattern&&r.audit===a));assert(roots.every(Boolean));
    const states=[{counts:Array(8).fill(0),audit:null}];
    for(let a=0;a<2;a++)for(const s of roots[a].states)states.push({...s,audit:a});
    const index=new Map(states.map((s,i)=>[s.audit+':'+s.counts.join(','),i]));
    const level=states.map(s=>s.audit===null?0:1+s.counts.reduce((a,v)=>a+v,0));
    const auditChildren=[0,1].map(a=>index.get(a+':0,0,0,0,0,0,0,0'));assert(auditChildren.every(Number.isInteger));
    const children=states.map((s,i)=>{
      if(i===0)return Array.from({length:4},()=>auditChildren.slice());
      if(level[i]===6)return [];
      return Array.from({length:4},(_,a)=>[0,1].map(y=>{const c=s.counts.slice();c[2*a+y]++;const j=index.get(s.audit+':'+c.join(','));assert(j!==undefined);return j;}));
    });
    const joints=new Float64Array(states.length*worlds.length*4);
    states.forEach((s,i)=>worlds.forEach((world,g)=>{
      const factor=s.audit===null?1:auditLikelihood(world.mask,s.audit),joint=actualJoint(pattern,s.counts,world.noise,world.mask);
      joint.forEach((v,y)=>joints[(i*worlds.length+g)*4+y]=v*factor);
    }));
    assert.equal(states.length,2575);
    graphs.push({pattern,states,level,children,joints,start:0,backward:states.map((_,i)=>i).sort((a,b)=>level[b]-level[a]),worldCount:worlds.length});
  }return graphs;
}
