import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditEnvelope} from './regime-query-envelope-audit.mjs';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function input(p){const s=fs.createReadStream(p),hash=crypto.createHash('sha256');s.on('data',b=>hash.update(b));return {it:createInterface({input:s,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=process.argv.slice(2).map(input);assert.equal(streams.length,4);const headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
['population-query-v1','soft-learners-v120','regime-query-counterfactual-v1','regime-query-disjoint-v1'].forEach((v,i)=>assert.equal(headers[i].Version,v));
for(const [file,hash]of Object.entries(headers[0].Hashes)){const p=path.resolve('internal/observationlearners',file);assert(p.startsWith(process.cwd()+path.sep));assert.equal(sha(p),hash);}
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length,records=[],ids=new Set();let fits=0,checks=0,changed=0;
for(;;){
  const lines=[];for(const s of streams)lines.push(await s.it.next());if(lines.some(x=>x.done)){assert(lines.every(x=>x.done));break;}
  const [r,source,prior,disjoint]=lines.map(x=>JSON.parse(x.value));assert(!r.Error&&!prior.Error&&!disjoint.Error);
  for(const k of ['Phase','Case','Index','Schedule'])assert.equal(r[k],source[k]);const id=`${r.Phase}:${r.Case}:${r.Index}:${r.Schedule}`;assert(!ids.has(id));ids.add(id);
  const original=auditEnvelope(prior.Original,source),flipped=structuredClone(source);
  for(let j=152;j<160;j++)if(flipped.Steps[j].Missing||j+flipped.Steps[j].Delay>161)flipped.Steps[j].Y=!flipped.Steps[j].Y;
  auditEnvelope(prior.Flipped,flipped);auditRegimeOutcome(disjoint.Original,source);auditRegimeOutcome(disjoint.Candidate,source,137);
  assert.deepEqual(r.Branches.map(b=>b.Origin),prior.Original.Choices);let expectedFits=1;
  const branches=r.Branches.map((b,i)=>{
    const o=original.branches[i],j=b.Origin;assert.deepEqual(b.Origins,o.support);assert.equal(b.Redundant,o.redundant);assert.equal(b.ActualY,j>=0?source.Steps[j].Y:false);assert.equal(b.Q,j>=0?source.Steps[j].Q:.5);
    for(const field of ['Population','Sample','LogEvidence','AtPublication','Predictions'])assert.equal(b[field].length,2);
    if(j>=0&&!b.Redundant)expectedFits+=2;
    for(let y=0;y<2;y++){
      const bundle=((y===Number(b.ActualY)||j<0||b.Redundant)?prior.Original:prior.Flipped).Bundles[Math.floor(i/4)],a=i%4;
      assert.deepEqual(b.AtPublication[y],bundle.AtPublication[a]);assert.deepEqual(b.Predictions[y],bundle.Predictions[a]);assert.equal(b.LogEvidence[y],bundle.LogEvidence[a]);
      const sample=mean(b.Predictions[y].map((p,k)=>{const q=source.Steps[161+k].Q;checks++;return (p-q)**2+q*(1-q);}));assert(Math.abs(sample-b.Sample[y])<1e-12);assert(Number.isFinite(b.Population[y])&&b.Population[y]>=0&&b.Population[y]<=1);
    }
    if(j<0||b.Redundant)assert.equal(b.Population[0],b.Population[1]);
    return {origin:j,population:(1-b.Q)*b.Population[0]+b.Q*b.Population[1],sample:(1-b.Q)*b.Sample[0]+b.Q*b.Sample[1]};
  });
  assert.equal(r.ActualFits,expectedFits);fits+=r.ActualFits;
  const paid=branches.slice(1),best=field=>paid.length?paid.reduce((a,b)=>b[field]<a[field]?b:a):branches[0];
  const pbest=best('population'),sbest=best('sample');if(pbest.origin!==sbest.origin)changed++;
  const select=[...disjoint.Original.Decision.Selected,disjoint.Candidate.Decision.Selected[3]];
  const control=select.map(j=>{const b=branches.find(b=>b.origin===j);assert(b);return b;});
  const abstaining=pbest.population<branches[0].population?pbest:branches[0];
  const values=field=>[...control.map(b=>b[field]),paid.length?mean(paid.map(b=>b[field])):branches[0][field],pbest[field],abstaining[field],sbest[field]];
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,population:values('population'),sample:values('sample'),branches,pbest:pbest.origin,sbest:sbest.origin,anyGain:Number(pbest.population<branches[0].population-1e-12),allHarm:Number(paid.length>0&&pbest.population>branches[0].population+1e-12)});
}
assert.equal(records.length,2688);const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[1],'5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');for(const i of [0,2,3])assert.equal(headers[i].InputSHA256,hashes[1]);
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);groups.push({phase,case:c,schedule,population:Array.from({length:9},(_,a)=>mean(rs.map(r=>r.population[a]))),bestGains:Array.from({length:6},(_,a)=>ci(rs.map(r=>r.population[a]-r.population[6]))),anyGain:rs.reduce((s,r)=>s+r.anyGain,0),allHarm:rs.reduce((s,r)=>s+r.allHarm,0)});
}
const delayed=records.filter(r=>r.schedule===1);
console.log(JSON.stringify({scope:'Known-generator input-population oracle, still conditional on future natural evidence; not an online policy',arms:['noquery','random','entropy','joint','disjoint','uniform','population-best-paid','population-best-with-abstention','sample-best-paid'],artifactHashes:hashes,scriptSHA256:sha(import.meta.filename),envelopeAuditSHA256:sha(new URL('./regime-query-envelope-audit.mjs',import.meta.url)),outcomeAuditSHA256:sha(new URL('./regime-query-outcome-audit.mjs',import.meta.url)),checks,fits,changed,delayed:{records:delayed.length,population:Array.from({length:9},(_,a)=>mean(delayed.map(r=>r.population[a]))),anyGain:delayed.reduce((s,r)=>s+r.anyGain,0),allHarm:delayed.reduce((s,r)=>s+r.allHarm,0)},groups,records}));
