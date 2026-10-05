import assert from 'node:assert/strict';
import {createReadStream,readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';
const raw=process.argv[2],bytes=readFileSync(process.argv[3]),s=JSON.parse(bytes);
const stream=createReadStream(raw),hash=createHash('sha256');stream.on('data',b=>hash.update(b));
let header,index=0;const rows=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),v=s.records[index++];
  assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[v.phase,v.case,v.index,v.schedule]);
  rows.push({...v,control:r.Metrics[12].map(x=>x.Brier)});
}
assert.equal(index,2688);assert.equal(hash.digest('hex'),s.artifactSHA256);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){assert.equal(a.length,32);const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++)for(let candidate=0;candidate<2;candidate++){
  const rs=rows.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);
  const gain=interval(rs.map(r=>r.control[segment]-r.scores[candidate][segment]));
  groups.push({phase,case:c,schedule,segment,candidate,gain,nonharm:gain.lower>=-.01,meaningfulGain:gain.mean>=.005&&gain.lower>0});
}
const candidates=[0,1].map(candidate=>{const gs=groups.filter(g=>g.candidate===candidate);return {candidate,nonharm:gs.filter(g=>g.nonharm).length,meaningfulGain:gs.filter(g=>g.meaningfulGain).length,total:gs.length};});
console.log(JSON.stringify({scope:'Consumed paired Squint comparison against archived Markov12; no fresh validation',artifactSHA256:s.artifactSHA256,diagnosticSHA256:createHash('sha256').update(bytes).digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),candidates,groups},null,2));
