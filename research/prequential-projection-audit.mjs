import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {projectPrequential,prequentialFeatures} from './prequential-projection.mjs';
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const stream=fs.createReadStream(process.argv[2]),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));const it=createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
assert.equal(JSON.parse((await it.next()).value).Version,'soft-learners-v120');
const records=[],seen=new Set();let checks=0,maxError=0,poisonChecks=0,sensitive=0;
for(;;){
  const line=await it.next();if(line.done)break;const s=JSON.parse(line.value),before=JSON.stringify(s),v=projectPrequential(s);
  const id=`${s.Phase}:${s.Case}:${s.Index}:${s.Schedule}`;assert(!seen.has(id));seen.add(id);
  const eligible=[];for(let j=128;j<160;j++)if(!s.Steps[j].Missing&&s.Steps[j].Delay<=160-j)eligible.push(j);
  assert.deepEqual(v.observed.map(r=>r.origin),eligible);for(let i=0;i<32;i++)assert.deepEqual(v.issued[i],{origin:128+i,x:s.Steps[128+i].X,p:s.Steps[128+i].P[10]});
  const pool=[];for(let j=152;j<160;j++)if(!eligible.includes(j))pool.push(j);
  const features=pool.map(origin=>{
    const got=prequentialFeatures(v,origin);let positive=0,brier=0,residual=0,recent=0,recentBrier=0,weight=0,localBrier=0,localResidual=0,age=0;
    for(const j of eligible){const row=s.Steps[j],y=Number(row.Y),res=y-row.P[10];let distance=0;for(let bit=0;bit<9;bit++)distance+=Number(((row.X>>bit)&1)!==((s.Steps[origin].X>>bit)&1));const w=1/(1+distance);
      positive+=y;brier+=res**2;residual+=res;age+=160-j;weight+=w;localBrier+=w*res**2;localResidual+=w*res;if(j>=152){recent++;recentBrier+=res**2;}}
    const n=eligible.length,want=[n/32,n?positive/n:.5,n?brier/n:.25,.5+.5*(n?residual/n:0),recent/8,recent?recentBrier/recent:.25,weight/32,weight?localBrier/weight:.25,.5+.5*(weight?localResidual/weight:0),n?age/n/32:1];
    for(let k=0;k<10;k++){const e=Math.abs(got[k]-want[k]);assert(e<1e-12);maxError=Math.max(maxError,e);checks++;}
    return {origin,values:got};
  });
  const mutant=structuredClone(s);mutant.Phase=-1;mutant.Case=-1;mutant.Index=-1;mutant.Seeds=[];mutant.Teacher=null;
  for(let j=0;j<256;j++){const r=mutant.Steps[j];r.Q=NaN;if(!eligible.includes(j))r.Y=!r.Y;if(j<128||j>=160){r.X=-1;r.P=null;}}
  assert.deepEqual(projectPrequential(mutant),v);poisonChecks++;
  // Different eventual outcomes/delays for unresolved packets are indistinguishable now.
  for(let j=128;j<160;j++)if(!eligible.includes(j)){mutant.Steps[j].Missing=false;mutant.Steps[j].Delay=200-j;}
  assert.deepEqual(projectPrequential(mutant),v);poisonChecks++;
  if(pool.length&&eligible.length){const changed=structuredClone(v);changed.observed[0].y=!changed.observed[0].y;assert.notDeepEqual(prequentialFeatures(changed,pool[0]),features[0].values);sensitive++;}
  assert.equal(JSON.stringify(s),before);
  records.push({phase:s.Phase,case:s.Case,index:s.Index,schedule:s.Schedule,view:v,features});
}
assert.equal(records.length,2688);const sourceSHA256=hash.digest('hex');assert.equal(sourceSHA256,'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
const schedules=[0,1].map(schedule=>{const rs=records.filter(r=>r.schedule===schedule);assert.equal(rs.length,1344);const counts=rs.map(r=>r.view.observed.length);return {schedule,records:rs.length,issued:rs.length*32,observed:counts.reduce((s,x)=>s+x,0),minObserved:Math.min(...counts),maxObserved:Math.max(...counts),candidates:rs.reduce((s,r)=>s+r.features.length,0)};});
console.log(JSON.stringify({scope:'As-of projection and feature arithmetic only; no new critic efficacy tested',sourceSHA256,scriptSHA256:sha(import.meta.filename),projectionSHA256:sha(new URL('./prequential-projection.mjs',import.meta.url)),testSHA256:sha(new URL('./prequential-projection-test.mjs',import.meta.url)),protocolSHA256:sha('docs/experiments/mmm-prequential-projection-protocol.md'),checks,maxError,poisonChecks,sensitive,schedules,records}));
