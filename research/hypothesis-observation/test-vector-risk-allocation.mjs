import assert from 'node:assert/strict';
import {allocateVectorRisk,vectorRisk} from './vector-risk-allocation.mjs';
function row(q,base){
  const mass=q.map(()=>1/q.length),joint=q.flatMap(p=>[p/q.length,(1-p)/q.length]);
  let constant=0;for(let i=0;i<base.length;i++)constant+=2*joint[i]*base[i]-mass[Math.floor(i/2)]*base[i]**2;
  return {mass,joint,constant};
}
for(const epsilon of [.01,.05,.1]){
  const base=[.5,.5],solution=allocateVectorRisk(row([.8],base),[row([.2],base)],base,epsilon);
  const expected=.2+Math.sqrt(.09+epsilon/2);
  assert(Math.abs(solution.forecast[0]-expected)<1e-7);
}
const base=[.5,.5],unconstrained=allocateVectorRisk(row([.8],base),[row([.8],base)],base);
assert(Math.abs(unconstrained.forecast[0]-.8)<1e-12);
let gridChecks=0;
for(let k=0;k<12;k++){
  const b=[.5,.5,.4,.6],objective=row([.7+k/100,.2],b),rows=[row([.2,.7],b),row([.3,.6],b)];
  const s=allocateVectorRisk(objective,rows,b,.02);let grid=Infinity;
  for(let i=0;i<=100;i++)for(let j=0;j<=100;j++){
    const p=[i/100,1-i/100,j/100,1-j/100];
    if(rows.every(r=>vectorRisk(r,p,2)<=.02))grid=Math.min(grid,vectorRisk(objective,p,2));
  }
  assert(s.primal<=grid+1e-8);assert(s.dual<=grid+1e-8);gridChecks++;
}
assert.throws(()=>allocateVectorRisk(row([.8],base),[],base));
assert.throws(()=>allocateVectorRisk(row([.8],base),[row([.2],base)],[.2,.2]));
console.log(JSON.stringify({analyticChecks:4,gridChecks,negativeTests:2}));

