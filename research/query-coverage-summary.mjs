import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditRegimeOutcome,maximumOrigin} from './regime-query-outcome-audit.mjs';
const [coveragePath,decisionPath,sourcePath,popPath,samplePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const pop=JSON.parse(fs.readFileSync(popPath)),sample=JSON.parse(fs.readFileSync(samplePath));
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[coveragePath,decisionPath,sourcePath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
for(const [i,v]of ['query-coverage-v1','regime-query-disjoint-v1','soft-learners-v120'].entries())assert.equal(headers[i].Version,v);
for(const [file,h]of Object.entries(headers[0].Hashes)){const p=path.resolve('internal/observationlearners',file);assert(p.startsWith(process.cwd()+path.sep));assert.equal(sha(p),h);}
const near=(a,b,t=1e-10)=>{assert(Number.isFinite(a)&&Number.isFinite(b));assert(Math.abs(a-b)<=t,`${a} != ${b}`);},mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const records=[];let fits=0,marginalChecks=0,oldChecks=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [r,d,source]=next.map(x=>JSON.parse(x.value)),p=pop.records[records.length],s=sample.records[records.length];assert(p&&s&&!r.Error&&!d.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[upper],source[upper]);assert.equal(p[lower],r[upper]);assert.equal(s[lower],r[upper]);}
  auditRegimeOutcome(d.Original,source);auditRegimeOutcome(d.Candidate,source,137);
  const old=d.Original.Decision,other=d.Candidate.Decision,pool=old.Pool??[];
  assert.deepEqual(r.Pool??[],pool);assert.deepEqual(r.Origins,old.Origins);assert.equal(r.Base.length,512);assert.equal(r.Weights.length,2);
  const weights=Array.from({length:2},()=>Array(512).fill(0));for(let j=0;j<=160;j++){weights[0][source.Steps[j].X]++;if(j>=129)weights[1][source.Steps[j].X]++;}
  assert.deepEqual(r.Weights,weights);assert.equal(weights[0].reduce((a,b)=>a+b,0),161);assert.equal(weights[1].reduce((a,b)=>a+b,0),32);
  const branches=r.Branches??[];assert.deepEqual(branches.map(b=>b.Origin),pool);assert.equal(r.Fits,pool.length?1+2*pool.length:0);fits+=r.Fits;
  if(pool.length)near(r.BaseLogEvidence,old.BaseLogEvidence);
  for(let i=0;i<branches.length;i++){
    const b=branches[i],v=old.Values[i],u=other.Values[i];assert.equal(b.Conditional.length,2);assert.equal(b.Mass.length,2);near(b.Mass[0]+b.Mass[1],1);
    for(let y=0;y<2;y++){
      assert.equal(b.Conditional[y].length,512);near(b.Mass[y],v.Mass[y]);near(b.Mass[y],u.Mass[y]);near(b.LogEvidence[y],v.LogEvidence[y]);near(Math.exp(b.LogEvidence[y]-r.BaseLogEvidence),b.Mass[y]);
      for(let j=0;j<8;j++){near(b.Conditional[y][old.Probes[j]],v.Conditional[y][j]);near(b.Conditional[y][other.Probes[j]],u.Conditional[y][j]);oldChecks+=2;}
    }
    const gainAt=[];
    for(let x=0;x<512;x++){
      const base=r.Base[x],a=b.Conditional[0][x],c=b.Conditional[1][x];assert(base>0&&base<1&&a>0&&a<1&&c>0&&c<1);
      near(b.Mass[0]*a+b.Mass[1]*c,base);marginalChecks++;
      const gain=b.Mass[0]*(a-base)**2+b.Mass[1]*(c-base)**2,reduction=base*(1-base)-b.Mass[0]*a*(1-a)-b.Mass[1]*c*(1-c);near(gain,reduction);gainAt.push(gain);
    }
    for(let a=0;a<2;a++)near(b.Gain[a],weights[a].reduce((sum,n,x)=>sum+n*gainAt[x],0)/(a===0?161:32));
    near(mean(old.Probes.map(x=>gainAt[x])),v.Gain);near(mean(other.Probes.map(x=>gainAt[x])),u.Gain);
  }
  const selected=[...old.Selected,other.Selected[3],...r.Selected];
  for(let a=0;a<2;a++)assert.equal(r.Selected[a],pool.length?maximumOrigin(pool,branches.map(b=>b.Gain[a])):-1);
  const byOrigin=rows=>new Map(rows.map(b=>[b.origin,b])),pm=byOrigin(p.branches),sm=byOrigin(s.branches);
  const actual=selected.map(j=>sm.get(j).actual),expected=selected.map(j=>sm.get(j).value),integrated=selected.map(j=>pm.get(j).population),costs=selected.map(j=>Number(j>=0));
  for(let a=0;a<5;a++){near(expected[a],s.values[a],1e-12);near(integrated[a],p.population[a],1e-12);}
  if(!pool.length){assert(selected.every(j=>j===-1));assert(actual.every(x=>x===actual[0]));}else assert(costs.slice(1).every(c=>c===1));
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,selected,actual,expected,integrated,costs,fitCount:r.Fits,fullChanged:Number(selected[5]!==selected[3]),recentChanged:Number(selected[6]!==selected[3])});
}
assert.equal(records.length,2688);assert.equal(new Set(records.map(r=>`${r.phase}:${r.case}:${r.index}:${r.schedule}`)).size,2688);
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[1],pop.artifactHashes[3]);assert.equal(hashes[1],sample.artifactHashes[4]);assert.equal(hashes[2],pop.artifactHashes[1]);assert.equal(hashes[2],sample.artifactHashes[1]);for(let i=0;i<2;i++)assert.equal(headers[i].InputSHA256,hashes[2]);
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const summarize=rs=>{const out={records:rs.length,costs:Array.from({length:7},(_,a)=>rs.reduce((s,r)=>s+r.costs[a],0)),fullChanged:rs.reduce((s,r)=>s+r.fullChanged,0),recentChanged:rs.reduce((s,r)=>s+r.recentChanged,0)};for(const field of ['actual','expected','integrated'])out[field]=Array.from({length:7},(_,a)=>mean(rs.map(r=>r[field][a])));return out;};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  const gains={};for(const field of ['actual','expected','integrated'])gains[field]=[5,6].map(a=>[1,2].map(control=>ci(rs.map(r=>r[field][control]-r[field][a]))));groups.push({phase,case:c,schedule,...summarize(rs),gains});
}
const evalGroups=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const field of ['actual','expected','integrated'])screens[field]=[0,1].map(a=>{const nonharm=evalGroups.every(g=>g.gains[field][a].every(x=>x.lower>=-.001)),transition=evalGroups.filter(g=>[19,20].includes(g.case)).every(g=>g.gains[field][a].every(x=>x.lower>0));return {arm:a+5,nonharm,transition,pass:nonharm&&transition,nonharmCells:evalGroups.filter(g=>g.gains[field][a].every(x=>x.lower>=-.001)).length,positiveCells:evalGroups.filter(g=>g.gains[field][a].every(x=>x.lower>0)).length};});
console.log(JSON.stringify({scope:'Visible empirical target coverage only, unchanged publication and model; consumed data, not confirmation',arms:['noquery','random','entropy','joint8','disjoint8','visible161','recent32'],hashes:Object.fromEntries([[coveragePath,hashes[0]],[decisionPath,hashes[1]],[sourcePath,hashes[2]],...[popPath,samplePath,'research/query-coverage-summary.mjs','research/regime-query-outcome-audit.mjs'].map(p=>[p,sha(p)])]),fits,marginalChecks,oldChecks,screens,phase1Delayed:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),groups,records}));
