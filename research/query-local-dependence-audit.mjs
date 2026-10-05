import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [file,coveragePath,sourcePath,dependencePath]=process.argv.slice(2),r=JSON.parse(fs.readFileSync(file)),prior=JSON.parse(fs.readFileSync(dependencePath));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');for(const [p,h]of Object.entries(r.hashes))assert.equal(sha(p),h);
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[coveragePath,sourcePath].map(input);for(const s of streams)await s.it.next();
const fit=Array.from({length:4},()=>({rows:0,weight:0,d0:0,d1:0}));let count=0,scoreChecks=0,forecastChecks=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [c,source]=next.map(x=>JSON.parse(x.value)),saved=r.records[count],old=prior.records[count++];assert(saved&&old);
  for(const [up,lo]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(c[up],source[up]);assert.equal(c[up],saved[lo]);assert.equal(saved[lo],old[lo]);}
  const observed=[];
  for(const j of c.Origins){
    assert(j<160);if(j<0)observed.push(source.Initial[j+16].Bits);else {const s=source.Steps[j];assert(!s.Missing&&j+s.Delay<=160);observed.push(s.X);}
  }
  const cell=(p,x)=>(p*(1-p)>=.125?2:0)+(observed.includes(x)?1:0),counts=[0,0,0,0],bs=c.Branches??[],scores=[];
  for(const branch of bs){
    const p=branch.Mass[1],qy=Number(source.Steps[branch.Origin].Y),weight=1/(31*bs.length);
    let gain=0;
    for(const s of source.Steps.slice(153,161)){
      const base=c.Base[s.X],l=r.fit[cell(p,s.X)].lambda;
      gain+=l*l*((1-p)*(branch.Conditional[0][s.X]-base)**2+p*(branch.Conditional[1][s.X]-base)**2)/8;
    }
    scores.push(gain);
    for(let age=0;age<31;age++){
      const target=source.Steps[161+age],k=cell(p,target.X);counts[k]++;
      if(c.Phase!==0)continue;
      const base=.5+.99**age*(c.Base[target.X]-.5),after=.5+.99**age*(branch.Conditional[qy][target.X]-.5),delta=after-base,sign=target.Y?1:-1;
      fit[k].rows++;fit[k].weight+=weight;fit[k].d0-=weight*sign*delta/(target.Y?base:1-base);fit[k].d1-=weight*sign*delta/(target.Y?after:1-after);
    }
  }
  assert.deepEqual(counts,saved.cellCounts);scores.forEach((v,i)=>{near(v,saved.scores[i]);scoreChecks++;});
  const highest=Math.max(...scores),selected=bs.length?Math.min(...bs.filter((_,i)=>scores[i]>=highest-1e-10).map(b=>b.Origin)):-1;
  assert.equal(selected,saved.selected[4]);assert.equal(saved.selected[4],saved.selected[3]);
  for(const f of ['jointLog','targetExpectedBrier']){
    if(!bs.length){assert.equal(saved.forecast[f],null);continue;}
    for(let a=0;a<3;a++){near(saved.forecast[f][a],old.metrics[f][a]);forecastChecks++;}
  }
}
assert.equal(count,2688);for(let k=0;k<4;k++){
  assert.equal(fit[k].rows,r.fit[k].rows);for(const field of ['weight','d0','d1'])near(fit[k][field],r.fit[k][field]);
  assert(fit[k].d1<0);assert.equal(r.fit[k].lambda,1);
}
const hashes=streams.map(s=>s.hash.digest('hex'));assert.equal(hashes[0],r.hashes[coveragePath]);assert.equal(hashes[1],r.hashes[sourcePath]);
const mean=xs=>xs.reduce((sum,x)=>sum+x/xs.length,0);let means=0,bounds=0;
for(const g of r.groups){
  const rows=r.records.filter(x=>x.phase===g.phase&&x.case===g.case&&x.schedule===g.schedule);assert.equal(rows.length,32);
  for(const [type,fields,arms,controls,candidate]of [['forecast',['jointLog','targetExpectedBrier'],3,[0,1],2],['selection',['actual','integrated'],5,[1,2],4]]){
    if(type==='forecast'&&!g.schedule)continue;
    for(const f of fields){
      for(let a=0;a<arms;a++){near(mean(rows.map(x=>x[type][f][a])),g[type][f][a]);means++;}
      for(const [idx,ctrl]of controls.entries()){
        const xs=rows.map(x=>x[type][f][ctrl]-x[type][f][candidate]),m=mean(xs);let pairs=0;
        for(let i=0;i<32;i++)for(let j=i+1;j<32;j++)pairs+=(xs[i]-xs[j])**2;
        const se=Math.sqrt(pairs/(32*32*31)),v=g[type+'Gains'][f][idx];near(v.mean,m);near(v.lower,m-3.5*se);near(v.upper,m+3.5*se);bounds++;
      }
    }
  }
}
for(const [key,v]of Object.entries(r.screens)){
  const [type,f]=key.split(':'),gs=r.groups.filter(g=>g.phase===1&&g.schedule===1),nonharm=!gs.some(g=>g[type+'Gains'][f].some(x=>x.lower<-.001)),transition=!gs.filter(g=>g.case>=19).some(g=>g[type+'Gains'][f].some(x=>x.lower<=0));
  assert.deepEqual(v,{nonharm,transition,pass:nonharm&&transition});
}
console.log(JSON.stringify({status:'PASS',records:count,fit,scoreChecks,forecastChecks,means,bounds,hashes:{[file]:sha(file),[dependencePath]:sha(dependencePath),[import.meta.filename]:sha(import.meta.filename)},scope:'Independent cell assignment, phase0 derivatives and convex boundary optima, variance-based selection, original-forecast equality, paired aggregation and four screens'}));
