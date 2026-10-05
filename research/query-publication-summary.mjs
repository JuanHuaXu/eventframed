import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import path from 'node:path';
import {createInterface} from 'node:readline';

const [rawPath,coveragePath,sourcePath,popPath,parentPath]=process.argv.slice(2);
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const parent=JSON.parse(fs.readFileSync(parentPath));
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
// Attach every iterator before consuming any stream.
const streams=[rawPath,coveragePath,sourcePath,popPath].map(input),headers=[];
for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
assert.equal(headers[0].Version,'query-publication-v1');
assert.equal(headers[1].Version,'query-coverage-v1');
assert.equal(headers[2].Version,'soft-learners-v120');
for(const [p,h]of Object.entries(headers[0].Hashes)){
  const file=path.resolve('internal/observationlearners',p);
  assert(file.startsWith(process.cwd()+path.sep));assert.equal(sha(file),h);
}
const near=(a,b)=>{assert(Number.isFinite(a)&&Number.isFinite(b));assert(Math.abs(a-b)<1e-10,`${a} != ${b}`);};
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const fields=['actual','expected','actualPopulation','integrated'];
const records=[];let fitCount=0,sampleChecks=0,supportChecks=0,matchedBaseline=0,matchedPaid=0;
function sample(source,p){return mean(source.Steps.slice(161,192).map((s,i)=>{const v=.5+(.99**i)*(p[s.X]-.5);return v*v-2*v*s.Q+s.Q;}));}
function risk(pair,q,y){return {actual:pair[Number(y)].Sample,expected:(1-q)*pair[0].Sample+q*pair[1].Sample,actualPopulation:pair[Number(y)].Population,integrated:(1-q)*pair[0].Population+q*pair[1].Population};}
for(;;){
  const rows=[];for(const s of streams)rows.push(await s.it.next());
  if(rows.some(r=>r.done)){assert(rows.every(r=>r.done));break;}
  const [r,c,source,pop]=rows.map(r=>JSON.parse(r.value)),previous=parent.records[records.length];
  assert(previous&&!r.Error&&!c.Error&&!pop.Error);
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){
    for(const other of [c,source,pop])assert.equal(r[upper],other[upper]);assert.equal(r[upper],previous[lower]);
  }
  const known=Array.from({length:176},(_,i)=>i-16).filter(j=>j<0||(!source.Steps[j].Missing&&j+source.Steps[j].Delay<=160));
  const origins=[known.slice(-63),known.slice(-64)];assert.deepEqual(r.BaselineOrigins,origins);supportChecks+=2;
  const pool=Array.from({length:8},(_,i)=>152+i).filter(j=>source.Steps[j].Missing||j+source.Steps[j].Delay>160);
  assert.deepEqual(c.Pool??[],pool);assert.deepEqual((r.Branches??[]).map(b=>b.Origin),pool);
  assert.equal(r.Fits,pool.length?1:2);fitCount+=r.Fits;
  for(const [i,forecast]of [r.FrozenForecast,r.RestoredForecast].entries()){
    assert.equal(forecast.length,512);assert(forecast.every(v=>v>0&&v<1));near(sample(source,forecast),r.Baseline[i].Sample);sampleChecks++;
  }
  if(pool.length)for(let x=0;x<512;x++)near(r.FrozenForecast[x],c.Base[x]);
  const actualBranches=new Map(pop.Branches.map(b=>[b.Origin,b])),local=new Map();
  const noquery=actualBranches.get(-1);assert(noquery);
  if(JSON.stringify(origins[1])===JSON.stringify(noquery.Origins)){
    near(r.Baseline[1].Sample,noquery.Sample[0]);near(r.Baseline[1].Population,noquery.Population[0]);matchedBaseline++;
  }
  for(const [i,b]of (r.Branches??[]).entries()){
    const j=b.Origin,paid=origins.map(o=>[...o,j].sort((a,b)=>a-b).slice(-64));
    assert.deepEqual(paid[0],paid[1]);assert.deepEqual(b.Origins,paid[0]);supportChecks+=2;
    assert.equal(b.Q,source.Steps[j].Q);assert.equal(b.ActualY,source.Steps[j].Y);
    assert.equal(c.Branches[i].Origin,j);
    for(let y=0;y<2;y++){near(sample(source,c.Branches[i].Conditional[y]),b.Risk[y].Sample);sampleChecks++;}
    local.set(j,risk(b.Risk,b.Q,b.ActualY));
    const actual=actualBranches.get(j);assert(actual);
    if(JSON.stringify(b.Origins)===JSON.stringify(actual.Origins)){
      const answers=actual.Redundant?[Number(b.ActualY)]:[0,1];
      for(const y of answers){near(b.Risk[y].Sample,actual.Sample[y]);near(b.Risk[y].Population,actual.Population[y]);}matchedPaid++;
    }
  }
  const states={};
  for(const state of ['A','B','C']){
    const values=previous.selected.map(j=>{
      if(state==='C'){const b=actualBranches.get(j);assert(b);return risk([0,1].map(y=>({Sample:b.Sample[y],Population:b.Population[y]})),b.Q,b.ActualY);}
      if(j<0){const b=r.Baseline[state==='A'?0:1];return risk([b,b],.5,false);}
      assert(local.has(j));return local.get(j);
    });
    states[state]=Object.fromEntries(fields.map(f=>[f,values.map(v=>v[f])]));
  }
  for(const f of ['actual','expected','integrated'])states.C[f].forEach((v,i)=>near(v,previous[f][i]));
  for(const f of fields)previous.selected.forEach((j,i)=>{if(j>=0)assert.equal(states.A[f][i],states.B[f][i]);});
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,selected:previous.selected,states});
}
assert.equal(records.length,2688);assert.equal(fitCount,4032);
assert.equal(new Set(records.map(r=>`${r.phase}:${r.case}:${r.index}:${r.schedule}`)).size,2688);
const hashes=streams.map(s=>s.hash.digest('hex'));
assert.equal(hashes[1],headers[0].CoverageSHA256);assert.equal(hashes[2],headers[0].InputSHA256);assert.equal(hashes[2],headers[1].InputSHA256);assert.equal(hashes[1],parent.hashes[coveragePath]);
const summarize=rs=>Object.fromEntries(['A','B','C'].map(s=>[s,Object.fromEntries(fields.map(f=>[f,Array.from({length:10},(_,a)=>mean(rs.map(r=>r.states[s][f][a])))]))]));
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  const gains=Object.fromEntries(['A','B','C'].map(s=>[s,Object.fromEntries(fields.map(f=>[f,Array.from({length:7},(_,i)=>[1,2].map(control=>ci(rs.map(r=>r.states[s][f][control]-r.states[s][f][i+3]))))]))]));
  groups.push({phase,case:c,schedule,means:summarize(rs),gains});
}
const evalGroups=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const s of ['A','B','C']){
  screens[s]={};for(const f of fields)screens[s][f]=Array.from({length:7},(_,a)=>{
    const nonharm=evalGroups.every(g=>g.gains[s][f][a].every(x=>x.lower>=-.001));
    const transition=evalGroups.filter(g=>[19,20].includes(g.case)).every(g=>g.gains[s][f][a].every(x=>x.lower>0));
    return {arm:a+3,nonharm,transition,pass:nonharm&&transition};
  });
}
console.log(JSON.stringify({scope:'Consumed-data publication diagnostic, not a prospective rescue. A/B/C expected risks condition on different natural evidence; actual-answer comparisons use the same realized answer. A is a weaker noquery baseline.',arms:parent.arms,hashes:Object.fromEntries([...([rawPath,coveragePath,sourcePath,popPath].map((p,i)=>[p,hashes[i]])),...[parentPath,import.meta.filename].map(p=>[p,sha(p)])]),fitCount,sampleChecks,supportChecks,matchedBaseline,matchedPaid,phase1Delayed:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),screens,groups,records}));
