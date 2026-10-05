import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';
import {predictiveInformation} from './query-epig.mjs';
import {maximumOrigin} from './regime-query-outcome-audit.mjs';
const [rawPath,decisionPath,parentPath,popPath,samplePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),load=p=>JSON.parse(fs.readFileSync(p));
const parent=load(parentPath),pop=load(popPath),sample=load(samplePath);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[rawPath,decisionPath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));assert.equal(headers[0].Version,'query-coverage-v1');assert.equal(headers[1].Version,'regime-query-disjoint-v1');
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length,near=(a,b)=>assert(Math.abs(a-b)<1e-10),kl=(p,q)=>p*Math.log(p/q)+(1-p)*Math.log((1-p)/(1-q));
const choose=(r,weights)=>{
  const branches=r.Branches??[],pool=r.Pool??[];
  const scores=weights.map(w=>branches.map(b=>predictiveInformation(r.Base,b.Conditional,b.Mass,w)));
  return {scores,origins:scores.map(s=>pool.length?maximumOrigin(pool,s):-1)};
};
const records=[],fixtures=[];let klChecks=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [r,d]=next.map(x=>JSON.parse(x.value)),i=records.length,previous=parent.records[i],p=pop.records[i],s=sample.records[i];assert(previous&&!r.Error&&!d.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[upper],previous[lower]);assert.equal(r[upper],d.Original[upper]);assert.equal(r[upper],p[lower]);assert.equal(r[upper],s[lower]);}
  const w8=Array(512).fill(0);for(const x of d.Original.Decision.Probes)w8[x]++;
  const weights=[w8,...r.Weights],decision=choose(r,weights);
  for(let j=0;j<(r.Branches??[]).length;j++){
    const b=r.Branches[j],local=r.Base.map((base,x)=>{klChecks++;return b.Mass[0]*kl(b.Conditional[0][x],base)+b.Mass[1]*kl(b.Conditional[1][x],base);});
    weights.forEach((w,a)=>near(decision.scores[a][j],w.reduce((sum,n,x)=>sum+n*local[x],0)/w.reduce((sum,n)=>sum+n,0)));
  }
  const selected=[...previous.selected,...decision.origins],pb=new Map(p.branches.map(b=>[b.origin,b])),sb=new Map(s.branches.map(b=>[b.origin,b]));
  const actual=selected.map(j=>sb.get(j).actual),expected=selected.map(j=>sb.get(j).value),integrated=selected.map(j=>pb.get(j).population),costs=selected.map(j=>Number(j>=0));
  for(const [key,values]of [['actual',actual],['expected',expected],['integrated',integrated],['costs',costs]])assert.deepEqual(values.slice(0,7),previous[key]);
  if(r.Schedule===0)assert(selected.every(j=>j===-1));else assert(costs.slice(1).every(c=>c===1));
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,selected,actual,expected,integrated,costs,scores:decision.scores});
  if(r.Phase===1&&r.Schedule===1)fixtures.push({r,weights,decision});
}
assert.equal(records.length,2688);assert.equal(fixtures.length,672);const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],parent.hashes[rawPath]);assert.equal(hashes[1],parent.hashes[decisionPath]);assert.equal(sha(popPath),parent.hashes[popPath]);assert.equal(sha(samplePath),parent.hashes[samplePath]);
const summarize=rs=>{const out={records:rs.length,costs:Array.from({length:10},(_,a)=>rs.reduce((s,r)=>s+r.costs[a],0))};for(const field of ['actual','expected','integrated'])out[field]=Array.from({length:10},(_,a)=>mean(rs.map(r=>r[field][a])));return out;};
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};},groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  const gains={};for(const field of ['actual','expected','integrated'])gains[field]=[7,8,9].map(a=>[1,2].map(control=>ci(rs.map(r=>r[field][control]-r[field][a]))));groups.push({phase,case:c,schedule,...summarize(rs),gains});
}
const evalGroups=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const field of ['actual','expected','integrated'])screens[field]=[0,1,2].map(a=>{const nonharm=evalGroups.every(g=>g.gains[field][a].every(x=>x.lower>=-.001)),transition=evalGroups.filter(g=>[19,20].includes(g.case)).every(g=>g.gains[field][a].every(x=>x.lower>0));return {arm:a+7,nonharm,transition,pass:nonharm&&transition,nonharmCells:evalGroups.filter(g=>g.gains[field][a].every(x=>x.lower>=-.001)).length,positiveCells:evalGroups.filter(g=>g.gains[field][a].every(x=>x.lower>0)).length};});
console.log(JSON.stringify({scope:'EPIG-style Bernoulli historical-query adaptation, unchanged posterior publication; consumed data, not confirmation',arms:[...parent.arms,'epig8','epig161','epig32'],hashes:Object.fromEntries([[rawPath,hashes[0]],[decisionPath,hashes[1]],...[parentPath,popPath,samplePath,'research/query-epig.mjs','research/query-epig-test.mjs','research/query-epig-experiment.mjs','research/regime-query-outcome-audit.mjs','docs/experiments/mmm-query-epig-protocol.md'].map(p=>[p,sha(p)])]),klChecks,screens,phase1Delayed:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),groups,records}));
for(let i=0;i<100;i++)choose(fixtures[i].r,fixtures[i].weights);const times=[];
for(let repeat=0;repeat<3;repeat++)for(const f of fixtures){const t=performance.now(),d=choose(f.r,f.weights);times.push(performance.now()-t);assert.deepEqual(d,f.decision);}
times.sort((a,b)=>a-b);const q=p=>times[Math.ceil(p*times.length)-1];
console.error(JSON.stringify({scope:'All three target scores together, with full precomputed conditional arrays; excludes posterior fits, independent KL audit, retrieval, I/O and serving',node:process.version,calls:times.length,selectionMS:{median:q(.5),p95:q(.95),p99:q(.99),max:times.at(-1)},scriptSHA256:sha(import.meta.filename)}));
