import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {prepareLoggedGain} from './logged-gain.mjs';
import {LoggedGainEBCS} from './logged-gain-eb.mjs';
import {preparePairedAudit,identicalAudit} from './paired-audit.mjs';

const [file]=process.argv.slice(2),projection=JSON.parse(fs.readFileSync(file));
const train=projection.records.filter(r=>r.phase===0&&r.schedule===1),test=projection.records.filter(r=>r.phase===1&&r.schedule===1);
assert.equal(train.length,672);assert.equal(test.length,672);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
function coin(rep,r,candidate,design){return crypto.createHash('sha256').update(['mmm-paired-audit-v1',rep,r.phase,r.case,r.index,candidate,design].join(':')).digest().readUInt32LE(0)%2;}
function reader(r,allowed){return i=>{assert(allowed.includes(i),'unobserved action loss');return r.losses[i];};}
const runs=[];let expectationChecks=0,identityChecks=0,accessChecks=0;
for(let rep=0;rep<64;rep++)for(const candidate of [0,2]){
  const sums=[.5,.5],counts=[1,1];let pairSum=0,pairCount=1,trainSingleCost=0,trainPairCost=0;
  for(const r of train){
    if(r.origins[candidate]===r.origins[1]){assert.equal(r.losses[candidate],r.losses[1]);continue;}
    const chosen=coin(rep,r,candidate,'single'),ids=[candidate,1],single=reader(r,[ids[chosen]]);
    sums[chosen]+=single(ids[chosen]);counts[chosen]++;trainSingleCost++;
    assert.throws(()=>single(ids[1-chosen]));accessChecks++;
    if(coin(rep,r,candidate,'paired')){const paired=reader(r,ids);pairSum+=paired(1)-paired(candidate);pairCount++;trainPairCost+=2;}
    else {const skipped=reader(r,[]);assert.throws(()=>skipped(candidate));accessChecks++;}
  }
  const models=sums.map((x,i)=>x/counts[i]),pairModel=pairSum/pairCount;
  const states=[];
  for(const design of ['single','paired'])for(const method of ['HT','DR'])states.push({design,method,cs:new LoggedGainEBCS(.05/8),cost:0,everMiss:false,positiveAt:[]});
  let truthSum=0,nonidentical=0;
  for(let i=0;i<test.length;i++){
    const r=test[i],same=r.origins[candidate]===r.origins[1],chosen=coin(rep,r,candidate,'single'),included=Boolean(coin(rep,r,candidate,'paired'));
    if(same){assert.equal(r.losses[candidate],r.losses[1]);identityChecks++;}else nonidentical++;
    const gain=r.losses[1]-r.losses[candidate];truthSum+=gain;
    for(const state of states){
      let interval;
      if(same)interval=state.cs.add(i,identicalAudit,false,undefined);
      else if(state.design==='single'){
        const prepared=prepareLoggedGain([.5,.5],[1,0],[0,1],state.method==='DR'?models:[0,0]),ids=[candidate,1],observed=reader(r,[ids[chosen]]);
        interval=state.cs.add(i,prepared,chosen,observed(ids[chosen]));state.cost++;
        const expectation=.5*prepared.observe(0,r.losses[candidate])+.5*prepared.observe(1,r.losses[1]);assert(Math.abs(expectation-gain)<1e-12);expectationChecks++;
      }else{
        const prepared=preparePairedAudit(.5,state.method==='DR'?pairModel:0),observed=reader(r,included?[candidate,1]:[]);
        interval=state.cs.add(i,prepared,included,included?observed(1)-observed(candidate):undefined);if(included)state.cost+=2;
        const expectation=.5*prepared.observe(true,gain)+.5*prepared.observe(false,undefined);assert(Math.abs(expectation-gain)<1e-12);expectationChecks++;
      }
      const truth=truthSum/(i+1);state.everMiss ||= truth<interval.lower-1e-12||truth>interval.upper+1e-12;
      if(interval.lower>0)state.positiveAt.push(i+1);
    }
  }
  for(const state of states)runs.push({rep,candidate:candidate===0?'random':'joint8',design:state.design,method:state.method,models,pairModel,trainingCounts:counts.map(n=>n-1),pairTrainingCount:pairCount-1,trainingCost:state.design==='single'?trainSingleCost:trainPairCost,evaluationCost:state.cost,expectedTrainingCost:trainSingleCost,expectedEvaluationCost:nonidentical,...state.cs.interval(),truth:truthSum/test.length,everMiss:state.everMiss,positiveAt:state.positiveAt});
}
const summaries=[];
for(const candidate of ['random','joint8'])for(const design of ['single','paired'])for(const method of ['HT','DR']){
 const rs=runs.filter(r=>r.candidate===candidate&&r.design===design&&r.method===method);
 summaries.push({candidate,design,method,truth:rs[0].truth,rmse:Math.sqrt(mean(rs.map(r=>(r.mean-r.truth)**2))),meanWidth:mean(rs.map(r=>r.upper-r.lower)),meanTrainingCost:mean(rs.map(r=>r.trainingCost)),meanEvaluationCost:mean(rs.map(r=>r.evaluationCost)),evaluationCostRange:[Math.min(...rs.map(r=>r.evaluationCost)),Math.max(...rs.map(r=>r.evaluationCost))],expectedTrainingCost:rs[0].expectedTrainingCost,expectedEvaluationCost:rs[0].expectedEvaluationCost,anytimeMisses:rs.filter(r=>r.everMiss).length,everPositive:rs.filter(r=>r.positiveAt.length).length});
}
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
console.log(JSON.stringify({scope:'Masked consumed-data paired-versus-single measurement at equal expected query cost; no new policy or prospective causal evidence',hashes:Object.fromEntries([file,import.meta.filename,'research/paired-audit.mjs','research/logged-gain.mjs','research/logged-gain-eb.mjs','docs/experiments/mmm-paired-audit-v1-protocol.md'].map(p=>[p,sha(p)])),assignments:64,training:train.length,evaluation:test.length,alphaPerCell:.05/8,expectationChecks,identityChecks,accessChecks,summaries,runs}));
