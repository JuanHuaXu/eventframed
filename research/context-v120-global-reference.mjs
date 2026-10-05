import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

const raw=process.argv[2],bytes=readFileSync(process.argv[3]),s=JSON.parse(bytes),arms=s.arms,prior=s.prior;
for(const [path,expected] of Object.entries(s.hashes))assert.equal(createHash('sha256').update(readFileSync(path)).digest('hex'),expected);
const stream=createReadStream(raw),hash=createHash('sha256');stream.on('data',b=>hash.update(b));
let header,index=0,forecasts=0,maxScoreError=0,maxEvidenceError=0;
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),record=s.records[index++];assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[record.phase,record.case,record.index,record.schedule]);
  const scores=[0,0];let logZ=0,accepted=0;
  for(let t=0;t<256;t++){
    // Full batch reconstruction at each clock, no cell cache, incremental
    // evidence differences, issue/deliver API, or reference component import.
    const logs=prior.map(p=>Math.log(p/prior.reduce((a,x)=>a+x,0)));accepted=0;
    for(let j=0;j<t;j++)if(!r.Steps[j].Missing&&j+r.Steps[j].Delay<=t){
      const row=r.Steps[j];for(let k=0;k<arms.length;k++)logs[k]+=row.Y?Math.log(row.P[arms[k]]):Math.log1p(-row.P[arms[k]]);accepted++;
    }
    const max=Math.max(...logs),ws=logs.map(l=>Math.exp(l-max)),z=ws.reduce((a,x)=>a+x,0);logZ=max+Math.log(z);
    const p=ws.reduce((a,w,k)=>a+w*r.Steps[t].P[arms[k]],0)/z,q=r.Steps[t].Q,loss=(p-q)**2+q*(1-q);
    scores[0]+=loss/256;if(t>=192)scores[1]+=loss/64;forecasts++;
  }
  for(let segment=0;segment<2;segment++)maxScoreError=Math.max(maxScoreError,Math.abs(scores[segment]-record.scores[0][segment]));
  maxEvidenceError=Math.max(maxEvidenceError,Math.abs(logZ-record.summaries[0].logEvidence));assert.equal(accepted,record.summaries[0].accepted);
  assert(maxScoreError<1e-12&&maxEvidenceError<1e-9);
}
assert.equal(index,2688);assert.equal(forecasts,688128);assert.equal(hash.digest('hex'),s.artifactSHA256);
console.log(JSON.stringify({scope:'Independent full global-control reconstruction, not every contextual forecast',artifactSHA256:s.artifactSHA256,diagnosticSHA256:createHash('sha256').update(bytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),records:index,forecasts,maxScoreError,maxEvidenceError},null,2));
