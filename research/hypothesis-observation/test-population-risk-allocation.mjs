import assert from 'node:assert/strict';
import {allocateRisk} from './population-risk-allocation.mjs';
for(const limit of [.01,.09,.25,.81]){
  const s=allocateRisk({a:[1],b:[-2]},[{a:[1],b:[0]}],limit);
  assert(Math.abs(s.lambda[0]-Math.sqrt(limit))<1e-7);
}
const linear=allocateRisk({a:[1],b:[-2]},[{a:[0],b:[1]}],.3);
assert(Math.abs(linear.lambda[0]-.3)<1e-7);
const free=allocateRisk({a:[1],b:[-2]},[{a:[0],b:[0]}]);assert.equal(free.lambda[0],1);
let gridChecks=0;
for(let k=1;k<=20;k++){
  const objective={a:[1+k/20,.5],b:[-1,-.4-k/40]},rows=[{a:[.1,.2],b:[.03,-.02]},{a:[.15,.1],b:[-.01,.01]}];
  const s=allocateRisk(objective,rows,.04);let grid=Infinity;
  for(let i=0;i<=100;i++)for(let j=0;j<=100;j++){
    const x=i/100,y=j/100;
    if(rows.every(r=>r.a[0]*x*x+r.a[1]*y*y+r.b[0]*x+r.b[1]*y<=.04)){
      grid=Math.min(grid,objective.a[0]*x*x+objective.a[1]*y*y+objective.b[0]*x+objective.b[1]*y);
    }
  }
  assert(s.primal<=grid+1e-8);assert(s.dual<=grid+1e-8);gridChecks++;
}
assert.throws(()=>allocateRisk({a:[-1],b:[0]},[{a:[1],b:[0]}]));
console.log(JSON.stringify({analyticChecks:6,gridChecks,negativeTests:1}));

