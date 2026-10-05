import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

import {createIntervalSurrogate} from './interval-surrogate.mjs';
// Consumed normalized interval-surrogate prototype, not a claim that the
// delayed adaptation inherits Squint-CE's immediate-feedback theorem.
const variants=[[0,1,2,3],[0,1,2,3,8,9,10,11]],rates=[.5,.25,.125,.0625];
function run(steps,arms){
  const prior=arms.map((_,i)=>i===0?.95:.05/(arms.length-1));
  // The declared horizon stays256 even in prefix tests; it is not future data.
  const model=createIntervalSurrogate(256,prior),seen=new Set(),predictions=[];
  for(let t=0;t<steps.length;t++){
    for(let j=0;j<t;j++)if(!seen.has(j)&&!steps[j].Missing&&j+steps[j].Delay<=t){
      model.deliver(j,steps[j].Y);seen.add(j);
    }
    model.expireBefore(Math.max(0,t-31));
    predictions.push(model.issue(arms.map(k=>steps[t].P[k])).p);
  }
  return {predictions};
}
const raw=process.argv[2],comparisonBytes=readFileSync(process.argv[3]),comparison=JSON.parse(comparisonBytes);
const stream=createReadStream(raw),digest=createHash('sha256');stream.on('data',b=>digest.update(b));
let header,forecastChecks=0,asOfChecks=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),scores=[];
  for(const arms of variants){
    const result=run(r.Steps,arms),metric=[0,0];
    for(let t=0;t<256;t++){const q=r.Steps[t].Q,p=result.predictions[t],l=(p-q)**2+q*(1-q);metric[0]+=l/256;if(t>=192)metric[1]+=l/64;forecastChecks++;}
    scores.push(metric);
    if(r.Index===0)for(const t of [0,32,160,255]){
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
      assert.deepEqual(run(changed,arms).predictions,result.predictions.slice(0,t+1));asOfChecks++;
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores});
}
assert.equal(records.length,2688);assert.equal(forecastChecks,1376256);assert.equal(asOfChecks,672);
const artifactSHA256=digest.digest('hex');assert.equal(artifactSHA256,comparison.artifactSHA256);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  const ref=comparison.groups.find(g=>g.phase===phase&&g.scenario===c&&g.schedule===schedule&&g.segment===segment);assert(ref);
  groups.push({phase,case:c,schedule,segment,meanBrier:variants.map((_,v)=>mean(rs.map(r=>r.scores[v][segment]))),incumbentMeans:ref.meanBrier,fourToEightGain:interval(rs.map(r=>r.scores[0][segment]-r.scores[1][segment]))});
}
console.log(JSON.stringify({scope:'Consumed normalized interval-surrogate diagnostic; no fresh or delayed regret guarantee',artifactSHA256,comparisonSHA256:createHash('sha256').update(comparisonBytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),variants,rates,prior:'generic64=.95, remainder uniform',componentSHA256:createHash('sha256').update(readFileSync(new URL('./interval-surrogate.mjs',import.meta.url))).digest('hex'),forecastChecks,asOfChecks,records,groups},null,2));

