import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {switchBudgetCosts} from './switch-budget-oracle.mjs';
const arms=[0,1,2,3,10,11],budgets=[0,1,2,4,8];
const input=process.argv[2],digest=crypto.createHash('sha256'),stream=fs.createReadStream(input);stream.on('data',b=>digest.update(b));
const records=[];let header;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
 if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue}
 const r=JSON.parse(line);assert.equal(r.Steps.length,256);
 const scores=[];
 for(const start of [0,192]){
  const steps=r.Steps.slice(start),n=steps.length;
  const loss=(p,q)=>(p-q)**2+q*(1-q);
  const costs=steps.map(s=>arms.map(i=>loss(s.P[i],s.Q))),opt=switchBudgetCosts(costs,8).map(v=>v/n);
  const markov=steps.reduce((v,s)=>v+loss(s.P[12],s.Q),0)/n;
  const unrestricted=costs.reduce((v,row)=>v+Math.min(...row),0)/n;
  const hull=steps.reduce((v,s)=>v+loss(Math.max(Math.min(...arms.map(i=>s.P[i])),Math.min(Math.max(...arms.map(i=>s.P[i])),s.Q)),s.Q),0)/n;
  const fixed=Math.min(...arms.map((_,i)=>costs.reduce((v,row)=>v+row[i],0)/n));assert(Math.abs(fixed-opt[0])<1e-12);
  for(let i=1;i<opt.length;i++)assert(opt[i]<=opt[i-1]+1e-12);
  assert(hull<=unrestricted+1e-12&&unrestricted<=opt[8]+1e-12);
  scores.push({markov,hull,unrestricted,oracle:budgets.map(b=>opt[b])});
 }
 records.push({phase:r.Phase,case:r.Case,schedule:r.Schedule,index:r.Index,scores});
}
assert.equal(records.length,2688);
const mean=a=>a.reduce((n,v)=>n+v,0)/a.length;
const interval=a=>{assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((n,v)=>n+(v-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se}};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 groups.push({phase,case:c,schedule,segment,markov:mean(rs.map(r=>r.scores[segment].markov)),hull:mean(rs.map(r=>r.scores[segment].hull)),unrestricted:mean(rs.map(r=>r.scores[segment].unrestricted)),oracle:budgets.map((_,i)=>mean(rs.map(r=>r.scores[segment].oracle[i]))),gains:budgets.map((_,i)=>interval(rs.map(r=>r.scores[segment].markov-r.scores[segment].oracle[i])))});
}
const hashes={};for(const file of [import.meta.filename,'research/switch-budget-oracle.mjs','docs/experiments/mmm-switch-budget-v120-protocol.md'])hashes[file]=crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
console.log(JSON.stringify({scope:'CONSUMED HINDSIGHT DIAGNOSTIC; hidden Q and future costs used; not admissible predictions',artifactSHA256:digest.digest('hex'),hashes,arms,budgets,records:records.length,groups},null,2));
