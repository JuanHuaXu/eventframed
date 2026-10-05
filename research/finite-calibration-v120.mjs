import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';
import {fitFiniteCalibration} from './finite-calibration.mjs';
const controlPaths=['docs/experiments/mmm-version-logit-v120-diagnostic.json','docs/experiments/mmm-version-logit-v120-cadence-control.json'];
const controls=controlPaths.map(p=>JSON.parse(readFileSync(p)));
function run(steps,arm){
  let model;const predictions=[],fits=[];
  for(let t=0;t<steps.length;t++){
    if(t%8===0){
      const origins=[];
      if(arm<2)for(let j=arm===0?32*Math.floor(t/32):0;j<t;j++)if(!steps[j].Missing&&j+steps[j].Delay<=t)origins.push(j);
      const selected=origins.slice(-64);model=fitFiniteCalibration(selected.map(j=>({p:steps[j].P[0],y:steps[j].Y})));
      assert(Math.abs(model.weights.reduce((s,x)=>s+x,0)-1)<1e-12&&Number.isFinite(model.logEvidence));
      fits.push({clock:t,origins:selected,weights:model.weights,logEvidence:model.logEvidence});
    }
    predictions.push(model.predict(steps[t].P[0]));
  }
  return {predictions,fits};
}
const stream=createReadStream(process.argv[2]),hash=createHash('sha256');stream.on('data',b=>hash.update(b));
let header,index=0,forecasts=0,asOf=0,isolation=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),scores=[],accuracy=[],logLoss=[],fits=[];
  for(const control of controls)assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[control.records[index].phase,control.records[index].case,control.records[index].index,control.records[index].schedule]);
  for(let arm=0;arm<3;arm++){
    const result=run(r.Steps,arm),bs=[0,0],ac=[0,0],ll=[0,0];
    for(let t=0;t<256;t++){
      const p=result.predictions[t],q=r.Steps[t].Q,l=(p-q)**2+q*(1-q),a=p>=.5?q:1-q,v=-q*Math.log(p)-(1-q)*Math.log1p(-p);
      bs[0]+=l/256;ac[0]+=a/256;ll[0]+=v/256;if(t>=192){bs[1]+=l/64;ac[1]+=a/64;ll[1]+=v/64;}forecasts++;
    }
    scores.push(bs);accuracy.push(ac);logLoss.push(ll);fits.push(result.fits);
    if(r.Index===0)for(const t of [0,32,160,255]){
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Q:1-s.Q,Y:j>=t||s.Missing||j+s.Delay>t?!s.Y:s.Y}));
      assert.deepEqual(run(changed,arm).predictions,result.predictions.slice(0,t+1));asOf++;
    }
    if(arm===0&&r.Index===0)for(const t of [32,160,255]){
      const changed=r.Steps.slice(0,t+1).map((s,j)=>({...s,Y:j<32*Math.floor(t/32)?!s.Y:s.Y}));
      assert.equal(run(changed,arm).predictions[t],result.predictions[t]);isolation++;
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,scores,accuracy,logLoss,fits,baseline:r.Metrics[0].map(m=>m.Brier),markov:r.Metrics[12].map(m=>m.Brier),map:controls.map(c=>c.records[index].scores[0])});index++;
}
const digest=hash.digest('hex');assert.equal(index,2688);assert.equal(forecasts,2064384);assert.equal(asOf,1008);assert.equal(isolation,252);
controls.forEach(c=>assert.equal(c.artifactSHA256,digest));
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const gates=[],groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule),meta={phase,case:c,schedule,segment};
  groups.push({...meta,brier:[0,1,2].map(a=>mean(rs.map(r=>r.scores[a][segment]))),baseline:mean(rs.map(r=>r.baseline[segment])),markov:mean(rs.map(r=>r.markov[segment]))});
  for(let arm=0;arm<2;arm++)for(const control of ['baseline','markov','map','prior']){
    const value=r=>control==='map'?r.map[arm][segment]:control==='prior'?r.scores[2][segment]:r[control][segment];
    const gain=interval(rs.map(r=>value(r)-r.scores[arm][segment]));
    gates.push({...meta,arm,control,type:'nonharm',...gain,pass:gain.lower>=-.01});
    if(control!=='map'&&segment===1&&((c<9&&c%3!==0)||c>=19))gates.push({...meta,arm,control,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});
  }
}
assert.equal(gates.length,1536);
const candidates=[0,1].map(arm=>{const gs=gates.filter(g=>g.arm===arm);return {arm,nonharm:gs.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:672,gains:gs.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:96,status:gs.every(g=>g.pass)?'PASS':'FAIL'};});
const paths=['research/finite-calibration.mjs','research/finite-calibration-test.mjs','research/finite-calibration-v120.mjs','docs/experiments/mmm-finite-calibration-v120-protocol.md',...controlPaths];
console.log(JSON.stringify({scope:'Consumed exact finite calibration posterior; no fresh validation',artifactSHA256:digest,hashes:Object.fromEntries(paths.map(p=>[p,createHash('sha256').update(readFileSync(p)).digest('hex')])),forecasts,asOf,isolation,candidates,records,groups,gates}));
