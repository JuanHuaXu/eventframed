import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {familyReference} from './family-evidence-reference.mjs';
const control=JSON.parse(fs.readFileSync(process.argv[5]));
const streams=[2,3,4].map(i=>fs.createReadStream(process.argv[i])),hashes=streams.map(()=>crypto.createHash('sha256'));
streams.forEach((s,i)=>s.on('data',b=>hashes[i].update(b)));
const its=streams.map(s=>createInterface({input:s,crlfDelay:Infinity})[Symbol.asyncIterator]());
const headers=[];for(const it of its)headers.push(JSON.parse((await it.next()).value));
assert.equal(headers[0].Version,'family-evidence-stream-v1');assert.equal(headers[1].Version,'soft-learners-v120');assert.equal(headers[2].Version,'refit-cadence-v1');
const records=[];let fits=0,probabilities=0,maxError=0;
for await(const line of its[0]) {
 const r=JSON.parse(line),next=await its[1].next(),nextFast=await its[2].next();assert(!next.done&&!nextFast.done&&!r.Error);
 const raw=JSON.parse(next.value),fast=JSON.parse(nextFast.value),prior=control.records[records.length];assert(prior&&!fast.Error);
 const id=x=>[x.Phase,x.Case,x.Index,x.Schedule];assert.deepEqual(id(r),id(raw));assert.deepEqual(id(r),id(fast));assert.deepEqual(id(r),[prior.phase,prior.case,prior.index,prior.schedule]);
 const cache=new Map(),brier=[],masses=[];
 for(let cadence=0;cadence<2;cadence++) {
  const out=r.Results[cadence],stride=cadence===0?32:8,base=cadence===0?raw.Steps.map(s=>[0,1,2,3,12].map(a=>s.P[a])):fast.Result.Predictions;
  assert.equal(out.Fits.length,256/stride);assert.equal(out.Predictions.length,256);
  const checked=[];
  for(let i=0;i<out.Fits.length;i++) {
   const fit=out.Fits[i];assert.equal(fit.Clock,i*stride);const available=[];
   for(let j=-16;j<fit.Clock;j++)if(j<0||(!raw.Steps[j].Missing&&j+raw.Steps[j].Delay<=fit.Clock))available.push(j);
   const windows=[];
   for(let w=0;w<2;w++) {
    const origins=available.slice(-(w===0?64:32));assert.deepEqual(fit.Origins[w],origins);
    const key=origins.join(',');let ref=cache.get(key);
    if(!ref){const samples=origins.map(j=>j<0?[raw.Initial[j+16].Bits,raw.Initial[j+16].Outcome]:[raw.Steps[j].X,raw.Steps[j].Y]);ref=familyReference(samples);cache.set(key,ref);}
    for(const [field,k] of [['GenericLog','genericLog'],['BooleanLog','booleanLog'],['LogEvidence','logEvidence'],['GenericMass','genericMass']]){const error=Math.abs(fit[field][w]-ref[k]);maxError=Math.max(maxError,error);assert(error<1e-10);}
    windows.push(ref);fits++;
   }
   checked.push(windows);
  }
  for(let w=0;w<2;w++){
   const loss=[0,0];let massSum=0;
   for(let t=0;t<256;t++){
    const p=out.Predictions[t][w],ref=checked[Math.floor(t/stride)][w];assert(Number.isFinite(p)&&p>0&&p<1);
    const expected=ref.genericMass*base[t][w*2]+(1-ref.genericMass)*base[t][w*2+1];assert(Math.abs(expected-p)<1e-10);probabilities++;
    const q=raw.Steps[t].Q,value=(p-q)**2+q*(1-q);loss[0]+=value/256;if(t>=192)loss[1]+=value/64;massSum+=ref.genericMass/256;
   }
   brier.push(loss);masses.push(massSum);
  }
 }
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,brier,masses,control:prior.brier.map(c=>c[4])});
}
assert((await its[1].next()).done);assert((await its[2].next()).done);assert.equal(records.length,2688);
assert.equal(hashes[1].digest('hex'),control.inputSHA256);assert.equal(hashes[2].digest('hex'),control.artifactSHA256);assert.equal(headers[0].InputSHA256,control.inputSHA256);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((a,x)=>a+(x-m)**2,0)/31/32);return{mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[],gates=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let segment=0;segment<2;segment++) {
 const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
 for(let arm=0;arm<4;arm++){
  const cadence=Math.floor(arm/2),meta={phase,case:c,schedule,segment,arm},gain=ci(rs.map(r=>r.control[cadence][segment]-r.brier[arm][segment]));
  groups.push({...meta,...gain,brier:mean(rs.map(r=>r.brier[arm][segment])),control:mean(rs.map(r=>r.control[cadence][segment])),genericMass:mean(rs.map(r=>r.masses[arm]))});
  gates.push({...meta,type:'nonharm',...gain,pass:gain.lower>=-.01});
  if(schedule===1&&segment===1&&[1,2,4,5,7,8,19,20].includes(c))gates.push({...meta,type:'gain',...gain,pass:gain.mean>=.005&&gain.lower>0});
 }
}
const summaries=Array.from({length:4},(_,arm)=>{const gs=gates.filter(g=>g.arm===arm);return{arm,primary:arm===0,nonharm:gs.filter(g=>g.type==='nonharm'&&g.pass).length,nonharmTotal:168,gain:gs.filter(g=>g.type==='gain'&&g.pass).length,gainTotal:16,status:gs.every(g=>g.pass)?'PASS':'FAIL'};});
console.log(JSON.stringify({scope:'Consumed same-window family re-evaluation; earlier failures retained',artifactSHA256:hashes[0].digest('hex'),inputSHA256:control.inputSHA256,cadenceArtifactSHA256:control.artifactSHA256,scriptSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),referenceSHA256:crypto.createHash('sha256').update(fs.readFileSync(new URL('./family-evidence-reference.mjs',import.meta.url))).digest('hex'),fits,probabilities,maxError,summaries,records,groups,gates}));
