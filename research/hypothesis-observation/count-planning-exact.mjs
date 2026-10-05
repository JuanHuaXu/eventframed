import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {createCountPlanner} from './count-planning.mjs';
const policies=['fixed','random','entropy','one','two','full'],roots=[],totals=Object.fromEntries(policies.map(p=>[p,{postSum:0,preSum:0,finalRisk:0}]));
let rootMass=0,bellmanChecks=0,stateCount=0;
for(let pattern=0;pattern<16;pattern++){
  const planner=createCountPlanner(pattern),zero=Array(8).fill(0),initial=planner.node(zero),optimum=planner.plan(zero,6);
  const values=Object.fromEntries(policies.map(p=>[p,planner.evaluate(zero,p)]));
  assert(Math.abs(optimum.cost-values.full.postSum)<1e-12);
  policies.forEach(p=>{assert(values.full.postSum<=values[p].postSum+1e-12);for(const k of ['postSum','preSum','finalRisk'])totals[p][k]+=initial.mass*values[p][k];});
  assert.equal(planner.nodes.size,3003);rootMass+=initial.mass;stateCount+=planner.nodes.size;
  const states=Array.from(planner.nodes.values()).map(n=>{
    const remaining=6-n.counts.reduce((s,v)=>s+v,0),plan=planner.plan(n.counts,remaining);
    plan.costs.forEach(c=>{assert(plan.cost<=c+1e-12);bellmanChecks++;});
    const actions=remaining?Object.fromEntries(policies.filter(p=>p!=='random').map(p=>[p,planner.weights(n.counts,p).indexOf(1)])):null;
    return {counts:n.counts,mass:n.mass,forecast:n.forecast,fullCost:plan.cost,actions};
  });
  roots.push({pattern,mass:initial.mass,initialRisk:initial.risk,values,states});
}
assert(Math.abs(rootMass-1)<1e-12);assert.equal(stateCount,48048);
for(const p of policies)totals[p].areaRisk=totals[p].preSum/6;
const paths=['count-planning.mjs','test-count-planning.mjs','count-planning-exact.mjs','interval-acquisition.mjs','noise-envelope.mjs','COUNT_PLANNING_PROTOCOL.md'].map(p=>'research/hypothesis-observation/'+p);
console.log(JSON.stringify({scope:'Exact six-query Bayesian planning under declared prior; not a rollout success screen or real-world guarantee',hashes:Object.fromEntries(paths.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),rootMass,stateCount,bellmanChecks,totals,roots}));

