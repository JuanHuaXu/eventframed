import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';
import {createContextRouter} from './context-expert-router.mjs';

const arms=[0,1,2,3,8,9,10,11],prior=arms.map((_,i)=>i===0?.95:.05/7);
function run(steps,contextual){
  const model=createContextRouter(9,prior,contextual),seen=new Set(),predictions=[],mass=[];
  for(let t=0;t<steps.length;t++){
    for(let j=0;j<t;j++)if(!seen.has(j)&&!steps[j].Missing&&j+steps[j].Delay<=t){model.deliver(j,steps[j].Y);seen.add(j);}
    model.expireBefore(Math.max(0,t-31));predictions.push(model.issue(steps[t].X,arms.map(i=>steps[t].P[i])).p);
    mass.push(1-model.summary().globalMass);
  }
  return {predictions,mass,summary:model.summary()};
}
const hash=createHash('sha256'),stream=createReadStream(process.argv[2]);stream.on('data',b=>hash.update(b));
let header,forecasts=0,asOfChecks=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),scores=[],mass=[],summaries=[];
  for(const contextual of [false,true]){
    const result=run(r.Steps,contextual),score=[0,0],ms=[0,0];
    for(let t=0;t<256;t++){
      const p=result.predictions[t],q=r.Steps[t].Q,l=(p-q)**2+q*(1-q);assert(p>0&&p<1);
      score[0]+=l/256;ms[0]+=result.mass[t]/256;if(t>=192){score[1]+=l/64;ms[1]+=result.mass[t]/64;}forecasts++;
    }
    scores.push(score);mass.push(ms);summaries.push(result.summary);
    if(r.Index===0)for(const t of [0,32,160,255]){
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
      assert.deepEqual(run(changed,contextual).predictions,result.predictions.slice(0,t+1));asOfChecks++;
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores,mass,summaries,markov:r.Metrics[12].map(m=>m.Brier)});
}
assert.equal(records.length,2688);assert.equal(forecasts,1376256);assert.equal(asOfChecks,672);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule),meta={phase,case:c,schedule,segment};
  assert.deepEqual(rs.map(r=>r.index),Array.from({length:32},(_,i)=>i));
  groups.push({...meta,brier:[mean(rs.map(r=>r.markov[segment])),...([0,1].map(a=>mean(rs.map(r=>r.scores[a][segment]))))],contextMass:mean(rs.map(r=>r.mass[1][segment]))});
  const globalLoss=interval(rs.map(r=>r.scores[0][segment]-r.markov[segment]));
  gates.push({...meta,type:'global_nonharm',...globalLoss,pass:globalLoss.upper<=.01});
  for(const control of ['global','markov']){
    const gain=interval(rs.map(r=>(control==='global'?r.scores[0][segment]:r.markov[segment])-r.scores[1][segment]));
    gates.push({...meta,type:'context_nonharm',control,...gain,pass:gain.lower>=-.01});
    if(segment===1&&((c<9&&c%3!==0)||c>=19))gates.push({...meta,type:'context_gain',control,...gain,pass:gain.mean>=.005&&gain.lower>0});
  }
}
assert.equal(gates.length,568);
const counts=Object.fromEntries(['global_nonharm','context_nonharm','context_gain'].map(type=>{const gs=gates.filter(g=>g.type===type);return[type,{passed:gs.filter(g=>g.pass).length,total:gs.length}];}));
const sources=['research/context-expert-router.mjs','research/context-expert-router-test.mjs','research/context-v120-diagnostic.mjs','docs/experiments/mmm-context-v120-protocol.md'];
console.log(JSON.stringify({scope:'Consumed finite contextual-expert model; not fresh confirmation',artifactSHA256:hash.digest('hex'),hashes:Object.fromEntries(sources.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),arms,prior,forecasts,asOfChecks,counts,status:gates.filter(g=>g.type!=='global_nonharm').every(g=>g.pass)?'CONSUMED_CONTEXT_PASS':'CONTEXT_FAIL',records,groups,gates},null,2));
