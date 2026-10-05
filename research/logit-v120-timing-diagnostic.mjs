import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

// Baseline-only calibration diagnosis on all cases. Teacher probabilities
// evaluate old/future risk AFTER fitting; they never change a coefficient.
const bytes=readFileSync(process.argv[3]),s=JSON.parse(bytes),rows=[];
const stream=createReadStream(process.argv[2]),hash=createHash('sha256');stream.on('data',b=>hash.update(b));
const clamp=p=>Math.max(1e-12,Math.min(1-1e-12,p));
const sig=z=>z>=0?1/(1+Math.exp(-z)):Math.exp(z)/(1+Math.exp(z));
function prediction(p,b){p=clamp(p);return clamp(sig(b[0]+(1+b[1])*(Math.log(p)-Math.log1p(-p))));}
const loss=(p,q)=>(p-q)**2+q*(1-q);
let header,index=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),record=s.records[index++];assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[record.phase,record.case,record.index,record.schedule]);
  for(let a=0;a<2;a++)for(const fit of record.fits[a].slice(1)){
    assert.equal(fit.beta.length,2);assert(fit.origins.length>0);
    let trainObservedGain=0,trainExpectedGain=0,futureExpectedGain=0,trainRaw=0,futureRaw=0;
    for(const j of fit.origins){
      const step=r.Steps[j],p=prediction(step.P[0],fit.beta),base=step.P[0];
      trainObservedGain+=((base-Number(step.Y))**2-(p-Number(step.Y))**2)/fit.origins.length;
      trainExpectedGain+=(loss(base,step.Q)-loss(p,step.Q))/fit.origins.length;trainRaw+=loss(base,step.Q)/fit.origins.length;
    }
    for(let t=fit.clock;t<fit.clock+32;t++){
      const step=r.Steps[t],p=prediction(step.P[0],fit.beta),base=step.P[0];
      futureExpectedGain+=(loss(base,step.Q)-loss(p,step.Q))/32;futureRaw+=loss(base,step.Q)/32;
    }
    rows.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,arm:a,clock:fit.clock,trainObservedGain,trainExpectedGain,futureExpectedGain,trainRaw,futureRaw,slope:1+fit.beta[1],intercept:fit.beta[0],minOrigin:fit.origins[0]});
  }
}
assert.equal(index,2688);assert.equal(rows.length,37632);assert.equal(hash.digest('hex'),s.artifactSHA256);
const mean=a=>a.reduce((v,x)=>v+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((v,x)=>v+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let arm=0;arm<2;arm++)for(let clock=32;clock<256;clock+=32){
  const rs=rows.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule&&r.arm===arm&&r.clock===clock);assert.equal(rs.length,32);
  groups.push({phase,case:c,schedule,arm,clock,metrics:Object.fromEntries(['trainObservedGain','trainExpectedGain','futureExpectedGain','trainRaw','futureRaw','slope','intercept'].map(k=>[k,interval(rs.map(r=>r[k]))])),negativeSlopes:rs.filter(r=>r.slope<0).length,minimumTrainingOrigin:Math.min(...rs.map(r=>r.minOrigin))});
}
console.log(JSON.stringify({scope:'Consumed baseline-calibration timing diagnosis; teacher q only evaluates, never fits',artifactSHA256:s.artifactSHA256,diagnosticSHA256:createHash('sha256').update(bytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),rows:rows.length,groups},null,2));
