import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {prepareLoggedGain} from './logged-gain.mjs';
import {LoggedGainEBCS as LoggedGainCS} from './logged-gain-eb.mjs';

const [path]=process.argv.slice(2),input=JSON.parse(fs.readFileSync(path));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const train=input.records.filter(r=>r.phase===0&&r.schedule===1),test=input.records.filter(r=>r.phase===1&&r.schedule===1);
assert.equal(train.length,672);assert.equal(test.length,672);
function draw(rep,r){
  for(let counter=0;;counter++){
    const word=crypto.createHash('sha256').update(`mmm-logged-gain-v1:${rep}:${r.phase}:${r.case}:${r.index}:${counter}`).digest().readUInt32LE(0);
    if(word<4294967295)return word%3;
  }
}
function view(r,selected,models){
  const origins=[...new Set(r.origins)],members=origins.map(o=>r.origins.map((x,i)=>x===o?i:-1).filter(i=>i>=0));
  const p=members.map(a=>a.length/3),m=members.map(a=>mean(a.map(i=>models[i])));
  return {p,m,action:origins.indexOf(r.origins[selected]),target:i=>origins.map(o=>Number(o===r.origins[i])),members};
}
const runs=[];let expectationChecks=0;
for(let rep=0;rep<64;rep++){
  const sums=[.5,.5,.5],counts=[1,1,1];
  for(const r of train){
    const a=draw(rep,r),loss=r.losses[a];
    r.origins.forEach((origin,i)=>{if(origin===r.origins[a]){sums[i]+=loss;counts[i]++;}});
  }
  const models=sums.map((x,i)=>x/counts[i]),states=[];
  for(const candidate of [0,2])for(const method of ['IPS','DR'])states.push({candidate,method,cs:new LoggedGainCS(.05/4),truth:0,everMiss:false,positiveAt:[],errors:[]});
  for(let i=0;i<test.length;i++){
    const r=test[i],selected=draw(rep,r),loss=r.losses[selected];
    // Only this chosen scalar loss crosses the logged-estimator boundary.
    for(const state of states){
      const v=view(r,selected,state.method==='DR'?models:[0,0,0]);
      const prepared=prepareLoggedGain(v.p,v.target(state.candidate),v.target(1),v.m);
      const interval=state.cs.add(i,prepared,v.action,loss);
      // Full-table data below belong to the simulator audit, not the estimator.
      const g=r.losses[1]-r.losses[state.candidate];state.truth+=g;
      const truth=state.truth/(i+1);state.everMiss ||= truth<interval.lower-1e-12||truth>interval.upper+1e-12;
      if(interval.lower>0)state.positiveAt.push(i+1);
      state.errors.push(interval.mean-truth);
      let conditional=0;
      for(let a=0;a<v.p.length;a++)conditional+=v.p[a]*prepared.observe(a,r.losses[v.members[a][0]]);
      assert(Math.abs(conditional-g)<1e-12);expectationChecks++;
    }
  }
  for(const s of states)runs.push({rep,candidate:s.candidate===0?'random':'joint8',control:'entropy',method:s.method,models,trainingCounts:counts.map(x=>x-1),...s.cs.interval(),truth:s.truth/test.length,everMiss:s.everMiss,positiveAt:s.positiveAt,finalError:s.errors.at(-1)});
}
const summaries=[];
for(const candidate of ['random','joint8'])for(const method of ['IPS','DR']){
  const rs=runs.filter(r=>r.candidate===candidate&&r.method===method);
  summaries.push({candidate,method,replicates:rs.length,truth:rs[0].truth,meanEstimate:mean(rs.map(r=>r.mean)),rmse:Math.sqrt(mean(rs.map(r=>r.finalError**2))),meanWidth:mean(rs.map(r=>r.upper-r.lower)),anytimeMisses:rs.filter(r=>r.everMiss).length,positiveFinal:rs.filter(r=>r.lower>0).length,everPositive:rs.filter(r=>r.positiveAt.length).length});
}
const byCase=[];
for(let c=0;c<21;c++){
 const rs=test.filter(r=>r.case===c);assert.equal(rs.length,32);
 byCase.push({case:c,n:32,random:mean(rs.map(r=>r.losses[0])),entropy:mean(rs.map(r=>r.losses[1])),joint8:mean(rs.map(r=>r.losses[2])),jointGain:mean(rs.map(r=>r.losses[1]-r.losses[2]))});
}
const original=JSON.parse(fs.readFileSync('docs/experiments/mmm-logged-gain-v1.json'));
assert.deepEqual(byCase,original.byCase);assert.equal(runs.length,original.runs.length);
for(let i=0;i<runs.length;i++)for(const field of ['rep','candidate','control','method','models','trainingCounts','n','mean','truth','finalError'])assert.deepEqual(runs[i][field],original.runs[i][field]);
console.log(JSON.stringify({pointControl:'Exact original assignments/regressions/point estimates/truths',scope:'Consumed fixed-potential-outcome masked logging simulation; not prospective agent logs, new policy efficacy, or 64 independent datasets',hashes:Object.fromEntries([path,import.meta.filename,'research/logged-gain.mjs','research/logged-gain-eb.mjs','docs/experiments/mmm-logged-gain-eb-v1-protocol.md'].map(p=>[p,sha(p)])),records:input.records.length,training:train.length,evaluation:test.length,loggingReplicates:64,paidQueriesPerReplicate:train.length+test.length,alphaPerComparisonMethod:.05/4,expectationChecks,identicalJointEntropy:test.filter(r=>r.origins[1]===r.origins[2]).length,fullInformation:{random:mean(test.map(r=>r.losses[0])),entropy:mean(test.map(r=>r.losses[1])),joint8:mean(test.map(r=>r.losses[2]))},summaries,byCase,runs}));

