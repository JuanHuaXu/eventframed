import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [resultPath,rawPath,coveragePath,sourcePath]=process.argv.slice(2),r=JSON.parse(fs.readFileSync(resultPath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for(const [p,h]of Object.entries(r.hashes))assert.equal(sha(p),h);
const mean=xs=>xs.reduce((s,x)=>s+x/xs.length,0),near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-11);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[rawPath,coveragePath,sourcePath].map(input),headers=[];for(const s of streams)headers.push(JSON.parse((await s.it.next()).value));
let count=0,lossChecks=0,choiceChecks=0,jensenChecks=0,maxJensenError=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [raw,c,source]=next.map(x=>JSON.parse(x.value)),saved=r.records[count++];assert(saved);
  for(const [up,lo]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']])for(const x of [raw,c,source])assert.equal(x[up],saved[lo]);
  const paid=(raw.Branches??[]).map((b,i)=>{
    const losses=b.Risk.map(x=>x.Population),p=c.Branches[i].Mass[1],q=b.Q;
    const truth=losses[0]+q*(losses[1]-losses[0]);
    near(truth,saved.paid[i].truth);near(p,saved.paid[i].p);near(q,saved.paid[i].q);
    const variance=mean(source.Steps.slice(161,192).map((step,age)=>{
      const delta=c.Branches[i].Conditional[1][step.X]-c.Branches[i].Conditional[0][step.X];
      return p*(1-p)*(.99**(2*age))*delta*delta;
    }));
    const modeledSample=(1-p)*b.Risk[0].Sample+p*b.Risk[1].Sample;
    const error=Math.abs(modeledSample-raw.Baseline[0].Sample-variance);
    maxJensenError=Math.max(maxJensenError,error);near(modeledSample-raw.Baseline[0].Sample,variance);jensenChecks++;
    assert((1-p)*losses[0]+p*losses[1]>=raw.Baseline[0].Population-1e-11);
    return {origin:b.Origin,losses,p,q,truth,actual:losses[Number(b.ActualY)],sample:b.Risk[0].Sample+q*(b.Risk[1].Sample-b.Risk[0].Sample)};
  });
  for(const [si,state]of ['A','B'].entries()){
    const base=raw.Baseline[si],actions=[{origin:-1,losses:[base.Population,base.Population],p:.5,q:.5,truth:base.Population,actual:base.Population,sample:base.Sample},...paid],s=saved.states[state];
    for(let a=0;a<16;a++){
      const b=actions.find(b=>b.origin===s.selected[a]);assert(b);
      near(b.truth,s.losses[a]);near(b.actual,s.actual[a]);near(b.sample,s.sample[a]);assert.equal(s.costs[a],Number(b.origin>=0));lossChecks+=3;
    }
    for(let a=10;a<16;a++){
      const rows=a>=13||!paid.length?actions:paid,field=['q','p','neutral'][(a-10)%3];
      const ranked=rows.map(b=>({origin:b.origin,value:b.losses[0]+(field==='neutral'?.5:b[field])*(b.losses[1]-b.losses[0])})).sort((a,b)=>a.value-b.value||a.origin-b.origin);
      const selected=ranked.find(b=>b.origin===s.selected[a]);assert(selected);near(selected.value,ranked[0].value);choiceChecks++;
    }
  }
}
assert.equal(count,2688);assert.equal(jensenChecks,9277);
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],r.hashes[rawPath]);assert.equal(hashes[1],r.hashes[coveragePath]);assert.equal(hashes[2],headers[0].InputSHA256);
let means=0,contrasts=0;
for(const g of r.groups){
  const rows=r.records.filter(x=>x.phase===g.phase&&x.case===g.case&&x.schedule===g.schedule);assert.equal(rows.length,32);
  for(const s of ['A','B']){
    for(const f of ['losses','actual','sample'])for(let a=0;a<16;a++){near(mean(rows.map(x=>x.states[s][f][a])),g.states[s][f][a]);means++;}
    for(const [name,left,right]of [['modelVsEntropy',2,11],['modelRegret',11,10]]){
      const xs=rows.map(x=>x.states[s].losses[left]-x.states[s].losses[right]),m=mean(xs);
      let pairs=0;for(let i=0;i<32;i++)for(let j=i+1;j<32;j++)pairs+=(xs[i]-xs[j])**2;
      const se=Math.sqrt(pairs/(32*32*31)),v=g.contrasts[s][name];near(v.mean,m);near(v.lower,m-3.5*se);near(v.upper,m+3.5*se);contrasts++;
    }
  }
}
console.log(JSON.stringify({status:'PASS',lossChecks,choiceChecks,jensenChecks,maxJensenError,means,contrasts,hashes:{[resultPath]:sha(resultPath),[sourcePath]:hashes[2],[import.meta.filename]:sha(import.meta.filename)},scope:'Source-based risks, independent minimum comparisons and pairwise variance aggregation. Jensen identity was investigated after observing universal A abstention; not a predeclared efficacy test.'}));
