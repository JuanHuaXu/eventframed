import assert from 'node:assert/strict';
import {solveGraph} from './global-policy.mjs';
let checks=0;
// Two decision levels: exhaust all64 root/contingent-child deterministic policies.
const graph={start:0,level:[4],children:[[]],backward:[]};
for(let a=0;a<4;a++){
  const pair=[];
  for(let y=0;y<2;y++){
    const j=graph.level.length;pair.push(j);graph.level.push(5);graph.children.push([]);
    for(let b=0;b<4;b++){
      const leaf=[];
      for(let z=0;z<2;z++){leaf.push(graph.level.length);graph.level.push(6);graph.children.push([]);}
      graph.children[j].push(leaf);
    }
  }
  graph.children[0].push(pair);
}
graph.backward=graph.level.map((_,i)=>i).sort((a,b)=>graph.level[b]-graph.level[a]);
for(let seed=0;seed<32;seed++){
  const cost=graph.level.map((_,i)=>((i*37+seed*17)%53-26)/17),values=[];
  for(let a=0;a<4;a++)for(let b=0;b<4;b++)for(let c=0;c<4;c++){
    const [x,y]=graph.children[0][a];
    values.push(cost[0]+cost[x]+cost[y]+graph.children[x][b].reduce((v,j)=>v+cost[j],0)+graph.children[y][c].reduce((v,j)=>v+cost[j],0));
  }
  assert(Math.abs(solveGraph(graph,cost).value-Math.min(...values))<1e-12);checks++;
}
const zero=solveGraph(graph,Array(graph.level.length).fill(0));assert.equal(zero.actions[0],0);checks++;
console.log(JSON.stringify({exhaustiveOracleChecks:checks,policiesPerNontrivialCheck:64}));
