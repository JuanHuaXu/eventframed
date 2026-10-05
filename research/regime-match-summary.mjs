import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {regimeReference} from './regime-predictive-reference.mjs';
const id=r=>`${r.Phase}:${r.Case}:${r.Index}:${r.Schedule}`;
async function loadSelected(path,version){const stream=fs.createReadStream(path),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));const map=new Map();let header;
 for await(const line of createInterface({input:stream,crlfDelay:Infinity})){const r=JSON.parse(line);if(!header){header=r;assert.equal(r.Version,version);continue;}if([0,19,20].includes(r.Case)){assert(!map.has(id(r)));map.set(id(r),r);}}
 assert.equal(map.size,384);return{map,hash:hash.digest('hex'),header};}
const raw=await loadSelected(process.argv[3],'soft-learners-v120'),family=await loadSelected(process.argv[4],'family-evidence-stream-v1');
const priorSummary=JSON.parse(fs.readFileSync(process.argv[5]));assert.equal(raw.hash,priorSummary.inputSHA256);assert.equal(family.hash,priorSummary.artifactSHA256);
const stream=fs.createReadStream(process.argv[2]),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));let header,checks=0,maxError=0;const records=[];
for await(const line of createInterface({input:stream,crlfDelay:Infinity})){
 const r=JSON.parse(line);if(!header){header=r;assert.equal(header.Version,'regime-match-v1');assert.equal(header.KnownCut,128);continue;}
 assert(!r.Error);const source=raw.map.get(id(r)),old=family.map.get(id(r));assert(source&&old);assert([160,192,224].includes(r.Clock));
 const eligible=[];for(let j=-16;j<r.Clock;j++)if(j<0||(!source.Steps[j].Missing&&j+source.Steps[j].Delay<=r.Clock))eligible.push(j);
 const full=eligible.slice(-64),current=full.filter(j=>j>=128);assert.deepEqual(r.Origins[0],full);assert.deepEqual(r.Origins[1]??[],current);
 assert.deepEqual(full,old.Results[0].Fits[r.Clock/32].Origins[0]);
 const brier=[],masses=[],preCounts=[];const cache=new Map();
 for(let arm=0;arm<10;arm++){
  const origins=r.Origins[arm]??[];assert.equal(new Set(origins).size,origins.length);if(arm>0)assert.equal(origins.length,current.length);
  if(arm>=2){const ranked=full.map(origin=>({origin,hash:crypto.createHash('sha256').update(`family-regime-match-v1:${r.Phase}:${r.Case}:${r.Index}:${r.Clock}:${arm-2}:${origin}`).digest('hex')})).sort((a,b)=>(a.hash<b.hash?-1:a.hash>b.hash?1:0)||a.origin-b.origin);assert.deepEqual(origins,ranked.slice(0,current.length).map(x=>x.origin).sort((a,b)=>a-b));}
  for(const j of origins)assert(full.includes(j));
  const key=origins.join(',');let ref=cache.get(key);
  if(!ref){const samples=origins.map(j=>j<0?[source.Initial[j+16].Bits,source.Initial[j+16].Outcome]:[source.Steps[j].X,source.Steps[j].Y]);ref=regimeReference(samples,source.Steps.slice(r.Clock,r.Clock+32).map(s=>s.X));cache.set(key,ref);}
  for(const [field,k] of [['GenericLog','genericLog'],['BooleanLog','booleanLog'],['LogEvidence','logEvidence'],['GenericMass','genericMass']]){const e=Math.abs(r[field][arm]-ref[k]);maxError=Math.max(maxError,e);assert(e<1e-10);}
  let loss=0;assert.equal(r.Predictions[arm].length,32);
  r.Predictions[arm].forEach((p,i)=>{assert(Number.isFinite(p)&&p>0&&p<1);assert(Math.abs(p-ref.predictions[i])<1e-10);if(arm===0)assert(Math.abs(p-old.Results[0].Predictions[r.Clock+i][0])<1e-12);const q=source.Steps[r.Clock+i].Q;loss+=((p-q)**2+q*(1-q))/32;checks++;});
  brier.push(loss);masses.push(r.GenericMass[arm]);preCounts.push(origins.filter(j=>j<128).length);
 }
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,clock:r.Clock,n:current.length,fullN:full.length,preCounts,brier,masses});
}
assert.equal(records.length,1152);assert.equal(header.InputSHA256,raw.hash);assert.equal(checks,368640);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((a,x)=>a+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(const c of [0,19,20])for(let schedule=0;schedule<2;schedule++)for(const clock of [160,192,224]){
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule&&r.clock===clock);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 groups.push({phase,case:c,schedule,clock,n:mean(rs.map(r=>r.n)),preCount:mean(rs.map(r=>r.preCounts[0])),brier:[mean(rs.map(r=>r.brier[0])),mean(rs.map(r=>r.brier[1])),mean(rs.map(r=>mean(r.brier.slice(2))))],genericMass:[mean(rs.map(r=>r.masses[0])),mean(rs.map(r=>r.masses[1])),mean(rs.map(r=>mean(r.masses.slice(2))))],randomMinusCurrent:ci(rs.map(r=>mean(r.brier.slice(2))-r.brier[1])),fullMinusCurrent:ci(rs.map(r=>r.brier[0]-r.brier[1])),fullMinusRandom:ci(rs.map(r=>r.brier[0]-mean(r.brier.slice(2))))});
}
console.log(JSON.stringify({scope:'Known-boundary sample-count-matched diagnostic; not an online detector',artifactSHA256:hash.digest('hex'),inputSHA256:raw.hash,familyArtifactSHA256:family.hash,scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),referenceSHA256:crypto.createHash('sha256').update(fs.readFileSync(new URL('./regime-predictive-reference.mjs',import.meta.url))).digest('hex'),checks,maxError,records,groups}));
