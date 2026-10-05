import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

const bytes=readFileSync(process.argv[3]),s=JSON.parse(bytes),arms=s.arms;
for(const [path,want] of Object.entries(s.hashes))assert.equal(createHash('sha256').update(readFileSync(path)).digest('hex'),want);
const stream=createReadStream(process.argv[2]),hash=createHash('sha256');stream.on('data',b=>hash.update(b));
const sigmoid=z=>z>=0?1/(1+Math.exp(-z)):Math.exp(z)/(1+Math.exp(z));
function features(step,full){
  const u=arms.map(k=>{const p=Math.max(1e-12,Math.min(1-1e-12,step.P[k]));return Math.log(p)-Math.log1p(-p);});
  return {offset:u[0],x:full?[1,u[0],...u.slice(1).map(v=>v-u[0])]:[1,u[0]]};
}
const linear=(f,beta)=>f.offset+f.x.reduce((v,x,i)=>v+x*beta[i],0);
let header,index=0,fits=0,forecasts=0,maxGradient=0,maxObjectiveError=0,maxScoreError=0,maxAccuracyError=0,maxLogLossError=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),record=s.records[index++];assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[record.phase,record.case,record.index,record.schedule]);
  for(let a=0;a<4;a++){
    assert.equal(record.fits[a].length,8);const scores=[0,0],accuracy=[0,0],logs=[0,0];
    for(let f=0;f<8;f++){
      const fit=record.fits[a][f];assert.equal(fit.clock,f*32);assert.equal(fit.beta.length,a<2?2:9);assert(fit.beta.every(Number.isFinite));
      const origins=[];for(let j=0;j<fit.clock;j++)if(!r.Steps[j].Missing&&j+r.Steps[j].Delay<=fit.clock)origins.push(j);
      const selected=origins.slice(-(a%2===0?64:32));assert.deepEqual(selected,fit.origins);
      const g=fit.beta.slice();let objective=fit.beta.reduce((v,x)=>v+x*x/2,0);
      for(const j of selected){
        const row=r.Steps[j],feature=features(row,a>=2),z=linear(feature,fit.beta),p=sigmoid(z),y=Number(row.Y);
        for(let k=0;k<g.length;k++)g[k]+=feature.x[k]*(p-y);
        objective+=Math.log1p(Math.exp(-Math.abs(z)))+((z>=0&&!row.Y)||(z<0&&row.Y)?Math.abs(z):0);
      }
      const gradient=Math.max(...g.map(Math.abs));maxGradient=Math.max(maxGradient,gradient);assert(gradient<=1.001e-8);
      maxObjectiveError=Math.max(maxObjectiveError,Math.abs(objective-fit.objective));assert(Math.abs(objective-fit.objective)<=1e-9*(1+objective));fits++;
      for(let t=fit.clock;t<fit.clock+32;t++){
        const row=r.Steps[t],q=row.Q,p=Math.max(1e-12,Math.min(1-1e-12,sigmoid(linear(features(row,a>=2),fit.beta)))),loss=(p-q)**2+q*(1-q),acc=p>=.5?q:1-q,ll=-q*Math.log(p)-(1-q)*Math.log1p(-p);
        scores[0]+=loss/256;accuracy[0]+=acc/256;logs[0]+=ll/256;if(t>=192){scores[1]+=loss/64;accuracy[1]+=acc/64;logs[1]+=ll/64;}forecasts++;
      }
    }
    for(let segment=0;segment<2;segment++){
      maxScoreError=Math.max(maxScoreError,Math.abs(scores[segment]-record.scores[a][segment]));
      maxAccuracyError=Math.max(maxAccuracyError,Math.abs(accuracy[segment]-record.accuracy[a][segment]));
      maxLogLossError=Math.max(maxLogLossError,Math.abs(logs[segment]-record.logLoss[a][segment]));
    }
  }
  assert(maxScoreError<1e-12&&maxAccuracyError<1e-12&&maxLogLossError<1e-10);
}
assert.equal(index,2688);assert.equal(fits,86016);assert.equal(forecasts,2752512);assert.equal(hash.digest('hex'),s.artifactSHA256);
console.log(JSON.stringify({scope:'Independent as-of, stationarity and forecast reconstruction; no fitter import',artifactSHA256:s.artifactSHA256,diagnosticSHA256:createHash('sha256').update(bytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),fits,forecasts,maxGradient,maxObjectiveError,maxScoreError,maxAccuracyError,maxLogLossError},null,2));
