import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';
import {mmdState,mmdDesignView} from './query-mmd.mjs';
import {maximumOrigin} from './regime-query-outcome-audit.mjs';
const [coveragePath,sourcePath,popPath,parentPath]=process.argv.slice(2),parent=JSON.parse(fs.readFileSync(parentPath)),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[coveragePath,sourcePath,popPath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
assert.equal(headers[0].Version,'query-coverage-v1');assert.equal(headers[1].Version,'soft-learners-v120');assert.equal(headers[2].Version,'population-query-v1');
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10),mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
function select(r,s){
  const bs=r.Branches??[],view=mmdDesignView(r.Origins,r.Pool??[],s);
  if(!bs.length)return {origins:[-1,-1],baseline:null,after:[],factors:[],information:[],scores:[]};
  const state=mmdState(view.history,view.target),appends=view.candidates.map(x=>state.append(x));
  const information=bs.map(b=>mean(view.target.slice(153,161).map(x=>b.Mass[0]*b.Mass[1]*(b.Conditional[1][x]-b.Conditional[0][x])**2)));
  const scores=information.map((v,i)=>v*appends[i].factor),factors=appends.map(v=>v.factor);
  return {origins:[maximumOrigin(r.Pool,scores),maximumOrigin(r.Pool,factors)],baseline:state.baseline,after:appends.map(v=>v.after),factors,information,scores};
}
const records=[],fixtures=[];let negativeFactors=0,zeroDenominators=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [r,source,pop]=next.map(x=>JSON.parse(x.value)),ref=parent.records[records.length];assert(ref&&!r.Error&&!pop.Error);
  for(const [up,lo]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[up],source[up]);assert.equal(r[up],pop[up]);assert.equal(r[up],ref[lo]);}
  const decision=select(r,source),bs=r.Branches??[];if(bs.length)assert.equal(maximumOrigin(r.Pool,decision.information),ref.selected[3]);
  negativeFactors+=decision.factors.filter(x=>x<0).length;zeroDenominators+=Number(decision.baseline!==null&&decision.baseline<=1e-12);
  const selected=[...ref.selected.slice(0,4),...decision.origins],map=new Map(pop.Branches.map(b=>[b.Origin,b])),values={};
  for(const [f,fn]of [['actual',b=>b.Sample[Number(b.ActualY)]],['integrated',b=>(1-b.Q)*b.Population[0]+b.Q*b.Population[1]]]){
    values[f]=selected.map(j=>{assert(map.has(j));return fn(map.get(j));});for(let a=0;a<4;a++)near(values[f][a],ref.states.C[f][a]);
  }
  if(bs.length)assert(selected.slice(1).every(j=>j>=0));else assert(selected.every(j=>j===-1));
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,selected,...values,decision});
  if(r.Phase===1&&r.Schedule===1)fixtures.push({r,source,decision});
}
assert.equal(records.length,2688);const hashes=streams.map(s=>s.hash.digest('hex'));for(let i=0;i<3;i++)assert.equal(hashes[i],parent.hashes[[coveragePath,sourcePath,popPath][i]]);assert.equal(headers[0].InputSHA256,hashes[1]);
const summarize=rs=>({records:rs.length,actual:[0,1,2,3,4,5].map(a=>mean(rs.map(r=>r.actual[a]))),integrated:[0,1,2,3,4,5].map(a=>mean(rs.map(r=>r.integrated[a]))),costs:[0,1,2,3,4,5].map(a=>rs.reduce((s,r)=>s+Number(r.selected[a]>=0),0)),changed:[4,5].map(a=>rs.filter(r=>r.selected[a]!==r.selected[3]).length)});
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};},groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  groups.push({phase,case:c,schedule,...summarize(rs),gains:Object.fromEntries(['actual','integrated'].map(f=>[f,[4,5].map(a=>[1,2].map(control=>ci(rs.map(r=>r[f][control]-r[f][a]))))]))});
}
const evalGroups=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const f of ['actual','integrated'])screens[f]=[0,1].map(a=>{
  const nonharm=evalGroups.every(g=>g.gains[f][a].every(v=>v.lower>=-.001)),transition=evalGroups.filter(g=>g.case>=19).every(g=>g.gains[f][a].every(v=>v.lower>0));return {arm:a+4,nonharm,transition,pass:nonharm&&transition};
});
console.log(JSON.stringify({scope:'Frozen R-I-inspired MMD factor on original information gain; unchanged publication, consumed data',arms:['noquery','random','entropy','joint8','mmd-joint8','mmd-only'],hashes:Object.fromEntries([[coveragePath,hashes[0]],[sourcePath,hashes[1]],[popPath,hashes[2]],...[parentPath,import.meta.filename,'research/query-mmd.mjs','research/query-mmd-test.mjs','research/regime-query-outcome-audit.mjs','docs/experiments/mmm-query-mmd-protocol.md'].map(p=>[p,sha(p)])]),negativeFactors,zeroDenominators,screens,phase0:summarize(records.filter(r=>r.phase===0&&r.schedule===1)),phase1:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),groups,records}));
for(let i=0;i<100;i++)select(fixtures[i].r,fixtures[i].source);const times=[];for(let repeat=0;repeat<3;repeat++)for(const f of fixtures){const start=performance.now(),got=select(f.r,f.source);times.push(performance.now()-start);assert.deepEqual(got,f.decision);}times.sort((a,b)=>a-b);
console.error(JSON.stringify({scope:'Full MMD setup, as-of view and both selectors with precomputed forecast arrays; excludes posterior fits/retrieval/persistence/serving',node:process.version,calls:times.length,medianMS:times[Math.ceil(.5*times.length)-1],p99MS:times[Math.ceil(.99*times.length)-1],maxMS:times.at(-1),scriptSHA256:sha(import.meta.filename)}));
