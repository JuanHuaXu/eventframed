import assert from 'node:assert/strict';
import {actualJoint} from './exact-regime.mjs';
import {solveGraph} from './global-policy.mjs';
export function brierOptimum(joint){
  assert(joint.length===4&&joint.every(v=>Number.isFinite(v)&&v>=0));
  const mass=joint.reduce((a,v)=>a+v,0),forecast=mass?joint.map(v=>v/mass):[.25,.25,.25,.25];
  const cost=joint.reduce((a,v,y)=>a+v*forecast.reduce((b,p,z)=>b+(p-Number(y===z))**2,0),0);
  return {forecast,cost};
}
export function prepare(model,worlds){
  return model.roots.map(root=>{
    const states=root.states,index=new Map(states.map((s,i)=>[s.counts.join(','),i]));
    const level=states.map(s=>s.counts.reduce((a,n)=>a+n,0));
    const children=states.map((s,i)=>level[i]===6?[]:Array.from({length:4},(_,a)=>[0,1].map(y=>{
      const c=s.counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);return j;
    })));
    const joints=new Float64Array(states.length*worlds.length*4);
    states.forEach((s,i)=>worlds.forEach((world,g)=>{
      const joint=actualJoint(root.pattern,s.counts,world.noise,world.mask);
      joint.forEach((v,y)=>joints[(i*worlds.length+g)*4+y]=v);
    }));
    return {states,level,children,joints,start:index.get('0,0,0,0,0,0,0,0'),backward:states.map((_,i)=>i).sort((a,b)=>level[b]-level[a]),worldCount:worlds.length};
  });
}

export function bestResponse(graphs,finalWeights,areaWeights){
  const final=new Float64Array(finalWeights.length),area=new Float64Array(areaWeights.length);
  let objective=0;
  for(const graph of graphs){
    const n=graph.level.length,w=graph.worldCount,cost=new Float64Array(n),forecasts=new Float64Array(n*4);
    for(let i=0;i<n;i++){
      const weights=graph.level[i]===6?finalWeights:areaWeights,factor=graph.level[i]===6?1:1/6,joint=[0,0,0,0];
      for(let g=0;g<w;g++)for(let y=0;y<4;y++)joint[y]+=weights[g]*graph.joints[(i*w+g)*4+y];
      const optimum=brierOptimum(joint);cost[i]=factor*optimum.cost;
      optimum.forecast.forEach((p,y)=>forecasts[i*4+y]=p);
    }
    const solution=solveGraph(graph,cost);objective+=solution.value;
    const occupancy=new Float64Array(n);occupancy[graph.start]=1;
    for(const i of graph.backward.slice().reverse()){
      const path=occupancy[i];if(!path)continue;
      const terminal=graph.level[i]===6,target=terminal?final:area,factor=terminal?1:1/6;
      const losses=Array.from({length:4},(_,y)=>{let sum=0;for(let z=0;z<4;z++)sum+=(forecasts[i*4+z]-Number(z===y))**2;return sum;});
      for(let g=0;g<w;g++)for(let y=0;y<4;y++)target[g]+=path*factor*graph.joints[(i*w+g)*4+y]*losses[y];
      if(!terminal)for(const child of graph.children[i][solution.actions[i]])occupancy[child]+=path;
    }
  }
  const forward=final.reduce((v,x,i)=>v+x*finalWeights[i],0)+area.reduce((v,x,i)=>v+x*areaWeights[i],0);
  assert(Math.abs(forward-objective)<1e-10);
  return {final:Array.from(final),area:Array.from(area),objective,forwardError:Math.abs(forward-objective)};
}
