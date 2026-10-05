import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {performance} from 'node:perf_hooks';
import {amplifiedJoint} from './query-amplification.mjs';
import {fitDependence} from './query-dependence.mjs';
import {maximumOrigin} from './regime-query-outcome-audit.mjs';
const [coveragePath,sourcePath,popPath,parentPath]=process.argv.slice(2),parent=JSON.parse(fs.readFileSync(parentPath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
async function read(paths,visit){
  const streams=paths.map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
  assert.equal(headers[0].Version,'query-coverage-v1');assert.equal(headers[1].Version,'soft-learners-v120');if(paths.length===3)assert.equal(headers[2].Version,'population-query-v1');
  let count=0;for(;;){const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}const rs=next.map(x=>JSON.parse(x.value));for(const r of rs)assert(!r.Error);for(const key of ['Phase','Case','Index','Schedule'])for(const r of rs)assert.equal(r[key],rs[0][key]);visit(...rs);count++;}
  assert.equal(count,2688);const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(headers[0].InputSHA256,hashes[1]);return hashes;
}
const rows=[];let histories=0;
const first=await read([coveragePath,sourcePath],(r,s)=>{
  if(r.Phase!==0||!(r.Branches??[]).length)return;
  histories++;const weight=1/(31*r.Branches.length);
  for(const b of r.Branches)for(let age=0;age<31;age++){
    const target=s.Steps[161+age],f=b.Conditional.map(a=>.5+.99**age*(a[target.X]-.5)),yq=Number(s.Steps[b.Origin].Y),upper=amplifiedJoint(b.Mass[1],f[0],f[1],1);
    rows.push({base:f[yq],delta:upper.conditional[yq]-f[yq],y:Number(target.Y),weight});
  }
});
assert.equal(rows.length,143840);assert.equal(histories,672);const fit=fitDependence(rows),alpha=fit.lambda,mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
function select(r,s){
  const bs=r.Branches??[],scores=bs.map(b=>mean(s.Steps.slice(153,161).map(t=>{
    const j=amplifiedJoint(b.Mass[1],b.Conditional[0][t.X],b.Conditional[1][t.X],alpha),delta=j.conditional[1]-j.conditional[0];return b.Mass[0]*b.Mass[1]*delta*delta;
  })));
  return {scores,origin:bs.length?maximumOrigin(bs.map(b=>b.Origin),scores):-1};
}
const records=[],fixtures=[],nll=(p,y)=>y?-Math.log(p):-Math.log1p(-p);let pairs=0,minCell=1,minCap=2,maxCap=1;
const second=await read([coveragePath,sourcePath,popPath],(r,s,pop)=>{
  const ref=parent.records[records.length];assert(ref);for(const [u,l]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']])assert.equal(r[u],ref[l]);
  const bs=r.Branches??[],choice=select(r,s),forecast={jointLog:[0,0,0],targetExpectedBrier:[0,0,0]};
  for(const b of bs)for(let age=0;age<31;age++){
    const target=s.Steps[161+age],yq=Number(s.Steps[b.Origin].Y),p=b.Mass[1],f=b.Conditional.map(a=>.5+.99**age*(a[target.X]-.5)),base=.5+.99**age*(r.Base[target.X]-.5),amp=amplifiedJoint(p,f[0],f[1],alpha);near(amp.base,base);pairs++;
    minCell=Math.min(minCell,...amp.joint);minCap=Math.min(minCap,amp.cap);maxCap=Math.max(maxCap,amp.cap);
    for(const [a,q]of [base,f[yq],amp.conditional[yq]].entries()){
      const weight=1/(31*bs.length);forecast.jointLog[a]+=weight*(nll(p,yq)+nll(q,Number(target.Y)));forecast.targetExpectedBrier[a]+=weight*((q-target.Q)**2+target.Q*(1-target.Q));
    }
  }
  if(!bs.length)for(const f of Object.keys(forecast))forecast[f]=null;
  const map=new Map(pop.Branches.map(b=>[b.Origin,b])),selected=[...ref.selected.slice(0,4),choice.origin],selection={};
  for(const [field,fn]of [['actual',b=>b.Sample[Number(b.ActualY)]],['integrated',b=>(1-b.Q)*b.Population[0]+b.Q*b.Population[1]]]){
    selection[field]=selected.map(j=>{assert(map.has(j));return fn(map.get(j));});for(let a=0;a<4;a++)near(selection[field][a],ref.states.C[field][a]);
  }
  if(bs.length)assert(selected.slice(1).every(j=>j>=0));else assert(selected.every(j=>j===-1));
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,forecast,selection,selected,scores:choice.scores,pairs:bs.length*31});
  if(r.Phase===1&&r.Schedule===1)fixtures.push({r,s,expected:choice});
});
assert.deepEqual(first,second.slice(0,2));for(let i=0;i<3;i++)assert.equal(second[i],parent.hashes[[coveragePath,sourcePath,popPath][i]]);assert.equal(pairs,287587);
const summarize=rs=>({records:rs.length,forecast:Object.fromEntries(['jointLog','targetExpectedBrier'].map(f=>[f,rs.some(r=>r.forecast[f])?[0,1,2].map(a=>mean(rs.filter(r=>r.forecast[f]).map(r=>r.forecast[f][a]))):null])),selection:Object.fromEntries(['actual','integrated'].map(f=>[f,[0,1,2,3,4].map(a=>mean(rs.map(r=>r.selection[f][a])))])),costs:[0,1,2,3,4].map(a=>rs.reduce((sum,r)=>sum+Number(r.selected[a]>=0),0))});
const ci=xs=>{const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/(xs.length-1)/xs.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};},groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  groups.push({phase,case:c,schedule,...summarize(rs),forecastGains:schedule===1?Object.fromEntries(['jointLog','targetExpectedBrier'].map(f=>[f,[0,1].map(a=>ci(rs.map(r=>r.forecast[f][a]-r.forecast[f][2])))])):null,selectionGains:Object.fromEntries(['actual','integrated'].map(f=>[f,[1,2].map(a=>ci(rs.map(r=>r.selection[f][a]-r.selection[f][4])))]))});
}
const evalGroups=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const [type,fields]of [['forecast',['jointLog','targetExpectedBrier']],['selection',['actual','integrated']]])for(const f of fields){const key=type+'Gains',nonharm=evalGroups.every(g=>g[key][f].every(v=>v.lower>=-.001)),transition=evalGroups.filter(g=>g.case>=19).every(g=>g[key][f].every(v=>v.lower>0));screens[type+':'+f]={nonharm,transition,pass:nonharm&&transition};}
console.log(JSON.stringify({scope:'Bounded amplification: separate frozen-support forecast diagnostic and unchanged-publication selection; consumed data',hashes:Object.fromEntries([[coveragePath,second[0]],[sourcePath,second[1]],[popPath,second[2]],...[parentPath,import.meta.filename,'research/query-amplification.mjs','research/query-amplification-test.mjs','research/query-dependence.mjs','research/regime-query-outcome-audit.mjs','docs/experiments/mmm-query-amplification-protocol.md'].map(p=>[p,sha(p)])]),training:{histories,pairs:rows.length,weight:rows.reduce((s,r)=>s+r.weight,0)},fit,pairs,minCell,minCap,maxCap,screens,phase0:summarize(records.filter(r=>r.phase===0&&r.schedule===1)),phase1:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),selectionChanged:records.filter(r=>r.phase===1&&r.schedule===1&&r.selected[3]!==r.selected[4]).length,groups,records}));
for(let i=0;i<100;i++)select(fixtures[i].r,fixtures[i].s);const times=[];for(let repeat=0;repeat<3;repeat++)for(const f of fixtures){const start=performance.now(),got=select(f.r,f.s);times.push(performance.now()-start);assert.deepEqual(got,f.expected);}times.sort((a,b)=>a-b);
console.error(JSON.stringify({scope:'Selection only with precomputed laws, excluding posterior fits/retrieval/persistence/serving',calls:times.length,node:process.version,medianMS:times[Math.ceil(.5*times.length)-1],p99MS:times[Math.ceil(.99*times.length)-1],maxMS:times.at(-1),scriptSHA256:sha(import.meta.filename)}));
