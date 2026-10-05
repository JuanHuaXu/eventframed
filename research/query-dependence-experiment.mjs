import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {binaryJoint,fitDependence} from './query-dependence.mjs';
const [coveragePath,sourcePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
async function read(visit){
  const streams=[coveragePath,sourcePath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
  assert.equal(headers[0].Version,'query-coverage-v1');assert.equal(headers[1].Version,'soft-learners-v120');
  let count=0;
  for(;;){
    const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
    const [r,source]=next.map(x=>JSON.parse(x.value));assert(!r.Error);
    for(const key of ['Phase','Case','Index','Schedule'])assert.equal(r[key],source[key]);
    visit(r,source);count++;
  }
  assert.equal(count,2688);const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(headers[0].InputSHA256,hashes[1]);return hashes;
}
function pairs(r,source,visit){
  const branches=r.Branches??[];
  for(const b of branches){
    const query=source.Steps[b.Origin],yq=Number(query.Y),p=b.Mass[1];
    for(let age=0;age<31;age++){
      const target=source.Steps[161+age],decay=.99**age;
      const base=.5+decay*(r.Base[target.X]-.5),f=b.Conditional.map(a=>.5+decay*(a[target.X]-.5));
      near(base,(1-p)*f[0]+p*f[1]);
      visit({base,f,p,yq,yt:Number(target.Y),qt:target.Q,qq:query.Q,weight:1/(branches.length*31)});
    }
  }
}
const rows=[];let trainHistories=0;
const firstHashes=await read((r,s)=>{
  if(r.Phase!==0||!(r.Branches??[]).length)return;
  trainHistories++;
  pairs(r,s,v=>rows.push({base:v.base,delta:v.f[v.yq]-v.base,y:v.yt,weight:v.weight}));
});
assert.equal(trainHistories,672);assert.equal(rows.length,143840);
const fit=fitDependence(rows),lambdas=[0,1,fit.lambda];
const fields=['jointLog','conditionalLog','brier','targetExpectedLog','targetExpectedBrier','teacherJointLog'];
const records=[];let pairCount=0;
const nll=(p,y)=>y?-Math.log(p):-Math.log1p(-p),cross=(p,q)=>-q*Math.log(p)-(1-q)*Math.log1p(-p);
const secondHashes=await read((r,source)=>{
  const metrics=Object.fromEntries(fields.map(f=>[f,[0,0,0]]));let count=0;
  pairs(r,source,v=>{
    count++;
    for(const [a,lambda]of lambdas.entries()){
      const j=binaryJoint(v.p,v.f[0],v.f[1],lambda),forecast=j.conditional[v.yq];
      const values={jointLog:-Math.log(j.joint[2*v.yq+v.yt]),conditionalLog:nll(forecast,v.yt),brier:(forecast-v.yt)**2,targetExpectedLog:cross(forecast,v.qt),targetExpectedBrier:(forecast-v.qt)**2+v.qt*(1-v.qt),teacherJointLog:cross(v.p,v.qq)+(1-v.qq)*cross(j.conditional[0],v.qt)+v.qq*cross(j.conditional[1],v.qt)};
      near(values.jointLog,values.conditionalLog+nll(v.p,v.yq));
      for(const field of fields)metrics[field][a]+=v.weight*values[field];
    }
  });
  pairCount+=count;
  if(count===0)for(const field of fields)metrics[field]=null;
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,pairs:count,metrics});
});
assert.deepEqual(firstHashes,secondHashes);assert.equal(pairCount,287587);
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const summarize=rs=>{
  const active=rs.filter(r=>r.pairs>0);
  return {records:rs.length,active:active.length,metrics:Object.fromEntries(fields.map(f=>[f,active.length?lambdas.map((_,a)=>mean(active.map(r=>r.metrics[f][a]))):null]))};
};
const ci=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++){
  const rs=records.filter(r=>r.phase===phase&&r.case===c&&r.schedule===schedule);assert.equal(rs.length,32);
  groups.push({phase,case:c,schedule,...summarize(rs),gains:schedule===1?Object.fromEntries(fields.map(f=>[f,[0,1].map(control=>ci(rs.map(r=>r.metrics[f][control]-r.metrics[f][2])))])):null});
}
const evalGroups=groups.filter(g=>g.phase===1&&g.schedule===1),screens={};
for(const f of ['jointLog','targetExpectedBrier']){
  const nonharm=evalGroups.every(g=>g.gains[f].every(x=>x.lower>=-.001)),transition=evalGroups.filter(g=>g.case>=19).every(g=>g.gains[f].every(x=>x.lower>0));
  screens[f]={nonharm,transition,pass:nonharm&&transition};
}
console.log(JSON.stringify({scope:'Pairwise Bernoulli dependence shrinkage; frozen-support consumed-data transfer screen, not a full multivariate process or ranking rescue',hashes:Object.fromEntries([[coveragePath,firstHashes[0]],[sourcePath,firstHashes[1]],...[import.meta.filename,'research/query-dependence.mjs','research/query-dependence-test.mjs','docs/experiments/mmm-query-dependence-protocol.md'].map(p=>[p,sha(p)])]),fit,training:{histories:trainHistories,pairs:rows.length,weight:rows.reduce((s,r)=>s+r.weight,0),supervision:'actual phase0 query and target labels including originally missing query outcomes; no teacher q or phase1 labels'},lambdas,pairCount,screens,phase0:summarize(records.filter(r=>r.phase===0&&r.schedule===1)),phase1:summarize(records.filter(r=>r.phase===1&&r.schedule===1)),groups,records}));
