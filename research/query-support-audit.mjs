import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [populationPath,decisionPath,sourcePath,referencePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),reference=JSON.parse(fs.readFileSync(referencePath));
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[populationPath,decisionPath,sourcePath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
['population-query-v1','regime-query-disjoint-v1','soft-learners-v120'].forEach((v,i)=>assert.equal(headers[i].Version,v));
const keep=(xs,n)=>[...new Set(xs)].sort((a,b)=>a-b).slice(-n),same=(a,b)=>JSON.stringify(a)===JSON.stringify(b),mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
assert.deepEqual(keep([1,2,3],3),[1,2,3]);assert.deepEqual(keep([1,2,3,4],3),[2,3,4]);assert.deepEqual(keep([3,2,3],3),[2,3]);
const records=[];let supportChecks=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [r,d,source]=next.map(x=>JSON.parse(x.value));assert(!r.Error&&!d.Error);
  for(const key of ['Phase','Case','Index','Schedule']){assert.equal(r[key],source[key]);assert.equal(r[key],d.Original[key]);}
  const known=[],natural=[];
  for(let j=-16;j<161;j++){
    if(j<0){known.push(j);natural.push(j);continue;}
    if(!source.Steps[j].Missing&&j+source.Steps[j].Delay<=161)natural.push(j);
    if(j<160&&!source.Steps[j].Missing&&j+source.Steps[j].Delay<=160)known.push(j);
  }
  const decision=keep(known,63),base=keep(natural,64),ds=new Set(decision),ks=new Set(known),bs=new Set(base),ns=new Set(natural);
  assert.deepEqual(decision,d.Original.Decision.Origins);assert.deepEqual(base,d.Original.Origins[0]);assert.deepEqual(base,r.Branches[0].Origins);supportChecks+=3;
  const restored=base.filter(j=>!ds.has(j)&&ks.has(j)),newEligible=base.filter(j=>!ks.has(j)),retired=decision.filter(j=>!bs.has(j));
  assert.equal(base.length-decision.length,restored.length+newEligible.length-retired.length);
  let paidUnchanged=0,paidRedundant=0;
  const pool=d.Original.Decision.Pool??[];assert.deepEqual(r.Branches.map(b=>b.Origin),[-1,...pool]);
  for(const branch of r.Branches.slice(1)){
    const actual=keep([...natural,branch.Origin],64),virtual=keep([...decision,branch.Origin],64);
    assert.deepEqual(branch.Origins,actual);assert.equal(branch.Redundant,ns.has(branch.Origin));supportChecks++;
    paidUnchanged+=Number(same(actual,virtual));paidRedundant+=Number(branch.Redundant);
  }
  for(const record of [d.Original,d.Candidate])for(let a=0;a<4;a++){
    const j=record.Decision.Selected[a];assert.deepEqual(record.Origins[a],keep([...natural,...(j<0?[]:[j])],64));supportChecks++;
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,decisionSize:decision.length,baselineSize:base.length,baselineChanged:Number(!same(base,decision)),restored:restored.length,newEligible:newEligible.length,retired:retired.length,paidBranches:pool.length,paidUnchanged,paidChanged:pool.length-paidUnchanged,paidRedundant});
}
assert.equal(records.length,2688);assert.equal(new Set(records.map(r=>`${r.phase}:${r.case}:${r.index}:${r.schedule}`)).size,2688);
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],reference.artifactHashes[0]);assert.equal(hashes[1],reference.artifactHashes[3]);assert.equal(hashes[2],reference.artifactHashes[1]);for(let i=0;i<2;i++)assert.equal(headers[i].InputSHA256,hashes[2]);
const summarize=rs=>({records:rs.length,baselineChanged:rs.reduce((s,r)=>s+r.baselineChanged,0),withRestoredKnown:rs.filter(r=>r.restored>0).length,withNewEligible:rs.filter(r=>r.newEligible>0).length,meanRestored:mean(rs.map(r=>r.restored)),meanNewEligible:mean(rs.map(r=>r.newEligible)),meanRetired:mean(rs.map(r=>r.retired)),paidBranches:rs.reduce((s,r)=>s+r.paidBranches,0),paidUnchanged:rs.reduce((s,r)=>s+r.paidUnchanged,0),paidChanged:rs.reduce((s,r)=>s+r.paidChanged,0),paidRedundant:rs.reduce((s,r)=>s+r.paidRedundant,0)}),groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);groups.push({phase,case:c,schedule,...summarize(rs)});}
console.log(JSON.stringify({scope:'Read-only support mismatch accounting; not causal attribution or a weakened-baseline rescue',supportChecks,hashes:Object.fromEntries([[populationPath,hashes[0]],[decisionPath,hashes[1]],[sourcePath,hashes[2]],...[referencePath,'research/query-support-audit.mjs','docs/experiments/mmm-query-support-protocol.md'].map(p=>[p,sha(p)])]),delayed:summarize(records.filter(r=>r.schedule===1)),complete:summarize(records.filter(r=>r.schedule===0)),groups,records}));
