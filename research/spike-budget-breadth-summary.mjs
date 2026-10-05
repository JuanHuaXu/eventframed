import fs from 'node:fs';
import assert from 'node:assert/strict';
const [input,pilotPath,output]=process.argv.slice(2);
assert(input&&pilotPath&&output);
const a=JSON.parse(fs.readFileSync(input,'utf8')),pilot=pilotPath==='-'?{results:[]}:JSON.parse(fs.readFileSync(pilotPath,'utf8'));
assert.equal(a.results.length,672);
const pilotMap=new Map(pilot.results.map(r=>[r.key,r]));
const groups=new Map(),trajectory=Array(8).fill(0);
for(const r of a.results){
  const [phase,scenario,index,schedule]=r.key.split(':').map(Number);
  assert(Number.isInteger(index)&&index>=0&&index<8);
  if(index===0&&pilotPath!=='-'){assert.deepEqual(r,pilotMap.get(r.key));pilotMap.delete(r.key);}
  const groupKey=[phase,scenario,schedule].join(':');
  if(!groups.has(groupKey))groups.set(groupKey,new Map());
  assert(!groups.get(groupKey).has(index));
  const delta=r.expected[0]-r.expected[1];groups.get(groupKey).set(index,delta);
  trajectory[index]+=delta/84;
}
assert.equal(pilotMap.size,0);assert.equal(groups.size,84);
// Exploratory percentile intervals. Resample trajectory indices, never frames.
let state=2463534242;
function rand(){state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;}
function interval(values){
  const means=[];
  for(let b=0;b<10000;b++){let sum=0;for(let j=0;j<values.length;j++)sum+=values[Math.floor(rand()*values.length)];means.push(sum/values.length);}
  means.sort((a,b)=>a-b);
  return [means[249],means[9749]];
}
const scenarios=[];
for(const [key,g] of groups){
  assert.equal(g.size,8);const values=Array.from({length:8},(_,i)=>g.get(i));
  scenarios.push({key,meanDelta:values.reduce((a,b)=>a+b,0)/8,pointwise95:interval(values),recordsHarmedAbovePoint01:values.filter(x=>x>.01).length});
}
const freshToThisCandidate=a.results.filter(r=>Number(r.key.split(':')[2])>0);
const subset={records:freshToThisCandidate.length,meanDelta:freshToThisCandidate.reduce((s,r)=>s+r.expected[0]-r.expected[1],0)/freshToThisCandidate.length,harmedAbovePoint01:freshToThisCandidate.filter(r=>r.expected[0]-r.expected[1]>.01).length};
const report={pooledMeanDelta:trajectory.reduce((a,b)=>a+b,0)/8,pooledTrajectory95:interval(trajectory),indices1through7:subset,scenarioMeansAbovePoint01:scenarios.filter(r=>r.meanDelta>.01).length,scenarios,pilotOverlapVerified:pilotPath!=='-',limitations:'Eight previously consumed trajectories per scenario, pointwise percentile bootstrap only; no simultaneous coverage or untouched confirmation. Pooled bootstrap clusters all scenarios/schedules by index.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
const {scenarios:details,...short}=report;console.log(JSON.stringify(short,null,2));console.log('Worst scenario means',scenarios.sort((a,b)=>b.meanDelta-a.meanDelta).slice(0,5));
