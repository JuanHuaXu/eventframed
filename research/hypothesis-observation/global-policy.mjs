import assert from 'node:assert/strict';
import {actualJoint} from './exact-regime.mjs';
export function solveGraph(graph,cost){
  const value=new Float64Array(cost),actions=new Int8Array(cost.length).fill(-1);
  for(const i of graph.backward){
    if(graph.level[i]===6)continue;
    const sums=graph.children[i].map(([j,k])=>value[j]+value[k]);
    const a=sums.indexOf(Math.min(...sums));actions[i]=a;value[i]+=sums[a];
  }
  return {actions,value:value[graph.start]};
}
export function prepare(model,worlds){
  return model.roots.map(root=>{
    const states=root.states,index=new Map(states.map((s,i)=>[s.counts.join(','),i]));
    const level=states.map(s=>s.counts.reduce((a,n)=>a+n,0));
    const children=states.map((s,i)=>level[i]===6?[]:Array.from({length:4},(_,a)=>[0,1].map(y=>{
      const c=s.counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);return j;
    })));
    const costs=new Float64Array(states.length*worlds.length);
    states.forEach((s,i)=>worlds.forEach((world,g)=>{
      const joint=actualJoint(root.pattern,s.counts,world.noise,world.mask);
      costs[i*worlds.length+g]=joint.reduce((v,p,y)=>v+p*s.forecast.reduce((z,q,k)=>z+(q-Number(k===y))**2,0),0);
    }));
    return {states,level,children,costs,start:index.get('0,0,0,0,0,0,0,0'),backward:states.map((_,i)=>i).sort((a,b)=>level[b]-level[a]),worldCount:worlds.length};
  });
}
export function bestResponse(graphs,finalWeights,areaWeights){
  const final=new Float64Array(finalWeights.length),area=new Float64Array(areaWeights.length);
  let objective=0;
  for(const graph of graphs){
    const n=graph.level.length,w=graph.worldCount,cost=new Float64Array(n);
    for(let i=0;i<n;i++){
      const weights=graph.level[i]===6?finalWeights:areaWeights,factor=graph.level[i]===6?1:1/6;
      for(let g=0;g<w;g++)cost[i]+=weights[g]*factor*graph.costs[i*w+g];
    }
    const solution=solveGraph(graph,cost);objective+=solution.value;
    const occupancy=new Float64Array(n);occupancy[graph.start]=1;
    for(const i of graph.backward.slice().reverse()){
      const path=occupancy[i];if(!path)continue;
      const terminal=graph.level[i]===6,target=terminal?final:area,factor=terminal?1:1/6;
      for(let g=0;g<w;g++)target[g]+=path*factor*graph.costs[i*w+g];
      if(!terminal)for(const child of graph.children[i][solution.actions[i]])occupancy[child]+=path;
    }
  }
  const forward=final.reduce((v,x,i)=>v+x*finalWeights[i],0)+area.reduce((v,x,i)=>v+x*areaWeights[i],0);
  assert(Math.abs(forward-objective)<1e-10);
  return {final:Array.from(final),area:Array.from(area),objective,forwardError:Math.abs(forward-objective)};
}
