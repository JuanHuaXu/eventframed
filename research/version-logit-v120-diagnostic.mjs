import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';
import {fitResidualLogit} from './residual-logit.mjs';

const arms=[0,1,2,3,8,9,10,11];
function run(steps,full,cadence){
  let model;const predictions=[],fits=[];
  for(let t=0;t<steps.length;t++){
    if(t%cadence===0){
      const origins=[];for(let j=32*Math.floor(t/32);j<t;j++)if(!steps[j].Missing&&j+steps[j].Delay<=t)origins.push(j);
      const selected=origins,rows=selected.map(j=>({ps:arms.map(k=>steps[j].P[k]),y:steps[j].Y}));
      model=fitResidualLogit(rows,full);fits.push({clock:t,origins:selected,beta:model.beta,iterations:model.iterations,gradientInf:model.gradientInf,objective:model.objective});
    }
    predictions.push(model.predict(arms.map(k=>steps[t].P[k])));
  }
  return {predictions,fits};
}
const stream=createReadStream(process.argv[2]),hash=createHash('sha256');stream.on('data',b=>hash.update(b));
let header,forecasts=0,asOfChecks=0,versionChecks=0,maxIterations=0,maxGradient=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),scores=[],accuracy=[],logLoss=[],fits=[];
  for(let a=0;a<4;a++){
    const result=run(r.Steps,a>=2,a%2===0?8:16),bs=[0,0],ac=[0,0],ll=[0,0];
    for(const f of result.fits){maxIterations=Math.max(maxIterations,f.iterations);maxGradient=Math.max(maxGradient,f.gradientInf);}
    for(let t=0;t<256;t++){
      const p=result.predictions[t],q=r.Steps[t].Q,l=(p-q)**2+q*(1-q),correct=p>=.5?q:1-q,log=-q*Math.log(p)-(1-q)*Math.log1p(-p);
      assert(p>0&&p<1&&Number.isFinite(log));bs[0]+=l/256;ac[0]+=correct/256;ll[0]+=log/256;
      if(t>=192){bs[1]+=l/64;ac[1]+=correct/64;ll[1]+=log/64;}forecasts++;
    }
    if(r.Index===0)for(const t of [32,160,255]){
      const start=32*Math.floor(t/32);
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Y:j<start?!s.Y:s.Y}));
      assert.equal(run(changed,a>=2,a%2===0?8:16).predictions[t],result.predictions[t]);versionChecks++;
    }
    scores.push(bs);accuracy.push(ac);logLoss.push(ll);fits.push(result.fits);
    if(r.Index===0)for(const t of [0,32,160,255]){
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
      assert.deepEqual(run(changed,a>=2,a%2===0?8:16).predictions,result.predictions.slice(0,t+1));asOfChecks++;
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores,accuracy,logLoss,fits,controls:r.Metrics.map(m=>m.map(x=>x.Brier))});
}
assert.equal(records.length,2688);assert.equal(forecasts,2752512);assert.equal(asOfChecks,1344);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule),meta={phase,case:c,schedule,segment};
  assert.deepEqual(rs.map(r=>r.index),Array.from({length:32},(_,i)=>i));
  groups.push({...meta,brier:[0,1,2,3].map(a=>mean(rs.map(r=>r.scores[a][segment]))),accuracy:[0,1,2,3].map(a=>mean(rs.map(r=>r.accuracy[a][segment]))),logLoss:[0,1,2,3].map(a=>mean(rs.map(r=>r.logLoss[a][segment]))),controls:[0,12,8,9,10,11].map(k=>mean(rs.map(r=>r.controls[k][segment])))});
  for(let a=0;a<4;a++){
    const controls=a<2?['generic','markov']:['generic','markov','baseline_calibration','variational64','variational32','segment64','segment32'];
    for(const control of controls){
      const value=r=>control==='baseline_calibration'?r.scores[a%2][segment]:r.controls[control==='generic'?0:control==='markov'?12:control==='variational64'?8:control==='variational32'?9:control==='segment64'?10:11][segment];
      const gain=interval(rs.map(r=>value(r)-r.scores[a][segment]));
      gates.push({...meta,candidate:a,control,type:'nonharm',...gain,pass:gain.lower>=-.01});
      if(segment===1&&((c<9&&c%3!==0)||c>=19)&&['generic','markov','baseline_calibration'].includes(control))gates.push({...meta,candidate:a,control,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});
    }
  }
}
assert.equal(gates.length,3344);assert.equal(gates.filter(g=>g.type==='nonharm').length,3024);assert.equal(versionChecks,1008);
const candidates=[0,1,2,3].map(candidate=>{const gs=gates.filter(g=>g.candidate===candidate);return {candidate,nonharm:gs.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:gs.filter(g=>g.type==='nonharm').length,gain:gs.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:gs.filter(g=>g.type==='gain').length,status:gs.every(g=>g.pass)?'PASS':'FAIL'};});
const sources=['research/residual-logit.mjs','research/residual-logit-test.mjs','research/version-logit-v120-diagnostic.mjs','docs/experiments/mmm-version-logit-v120-protocol.md'];
console.log(JSON.stringify({scope:'Consumed version-local residual logit calibration; no fresh validation',artifactSHA256:hash.digest('hex'),hashes:Object.fromEntries(sources.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),arms,cadences:[8,16,8,16],forecasts,asOfChecks,versionChecks,maxIterations,maxGradient,candidates,records,groups,gates},null,2));

