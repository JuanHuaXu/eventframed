import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [file,coveragePath,sourcePath,popPath]=process.argv.slice(2),r=JSON.parse(fs.readFileSync(file));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');for(const [p,h]of Object.entries(r.hashes))assert.equal(sha(p),h);
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[coveragePath,sourcePath,popPath].map(input);for(const s of streams)await s.it.next();
// Derive the bound from the joint table's Frechet interval, independently of
// the implementation's conditional-probability inequalities.
function law(p,f0,f1,alpha){
  const base=(1-p)*f0+p*f1,c=p*(1-p)*(f1-f0),lo=Math.max(0,p+base-1),hi=Math.min(p,base);
  const limit=c>0?(hi-p*base)/c:c<0?(lo-p*base)/c:Infinity;
  const cap=Math.min(2,1+.99*(Math.max(1,limit)-1)),stretch=1+alpha*(cap-1),p11=p*base+stretch*c;
  const cells=[1-p-base+p11,base-p11,p-p11,p11];assert(cells.every(v=>v>0));
  return {cap,cells,conditional:[cells[1]/(1-p),cells[3]/p]};
}
const alpha=r.fit.lambda,derivative=[0,0,0];let count=0,pairs=0,training=0,scoreChecks=0,forecastChecks=0,riskChecks=0,minCell=1;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [c,source,pop]=next.map(x=>JSON.parse(x.value)),saved=r.records[count++];assert(saved);
  for(const [up,lo]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']])for(const x of [c,source,pop])assert.equal(x[up],saved[lo]);
  const bs=c.Branches??[],scores=[],forecast={jointLog:[0,0,0],targetExpectedBrier:[0,0,0]};
  for(const b of bs){
    const p=b.Mass[1],qy=Number(source.Steps[b.Origin].Y);let gain=0;
    for(const target of source.Steps.slice(153,161)){
      const f=law(p,b.Conditional[0][target.X],b.Conditional[1][target.X],alpha).conditional;
      gain+=p*(1-p)*(f[1]-f[0])**2/8;
    }
    scores.push(gain);
    for(let age=0;age<31;age++){
      const target=source.Steps[161+age],f=b.Conditional.map(a=>.5+.99**age*(a[target.X]-.5)),base=.5+.99**age*(c.Base[target.X]-.5),upper=law(p,...f,1),fit=law(p,...f,alpha),weight=1/(31*bs.length),ty=Number(target.Y);
      minCell=Math.min(minCell,...fit.cells);pairs++;
      if(c.Phase===0){
        training++;const delta=upper.conditional[qy]-f[qy];
        for(const [i,a]of [0,1,alpha].entries()){const q=f[qy]+a*delta;derivative[i]+=weight*delta*(q-ty)/(q*(1-q));}
      }
      for(const [a,q]of [base,f[qy],fit.conditional[qy]].entries()){
        const qmass=qy?p:1-p;forecast.jointLog[a]-=weight*Math.log(qmass*(ty?q:1-q));forecast.targetExpectedBrier[a]+=weight*(q*q-2*q*target.Q+target.Q);
      }
    }
  }
  scores.forEach((v,i)=>{near(v,saved.scores[i]);scoreChecks++;});
  const max=Math.max(...scores),selected=bs.length?Math.min(...bs.filter((_,i)=>scores[i]>=max-1e-10).map(b=>b.Origin)):-1;assert.equal(selected,saved.selected[4]);
  for(const f of Object.keys(forecast))if(bs.length)for(let a=0;a<3;a++){near(forecast[f][a],saved.forecast[f][a]);forecastChecks++;}else assert.equal(saved.forecast[f],null);
  const map=new Map(pop.Branches.map(b=>[b.Origin,b]));for(let a=0;a<5;a++){
    const b=map.get(saved.selected[a]);assert(b);near(b.Sample[Number(b.ActualY)],saved.selection.actual[a]);near(b.Population[0]+b.Q*(b.Population[1]-b.Population[0]),saved.selection.integrated[a]);riskChecks+=2;
  }
}
assert.equal(count,2688);assert.equal(pairs,287587);assert.equal(training,143840);near(derivative[0],r.fit.d0);near(derivative[1],r.fit.d1);near(derivative[2],0);near(minCell,r.minCell);assert(alpha>0&&alpha<1&&derivative[0]<0&&derivative[1]>0);
const hashes=streams.map(s=>s.hash.digest('hex'));for(let i=0;i<3;i++)assert.equal(hashes[i],r.hashes[[coveragePath,sourcePath,popPath][i]]);
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
  const [type,f]=key.split(':'),gs=r.groups.filter(g=>g.phase===1&&g.schedule===1),nonharm=!gs.some(g=>g[type+'Gains'][f].some(x=>x.lower<-.001)),transition=!gs.filter(g=>g.case>=19).some(g=>g[type+'Gains'][f].some(x=>x.lower<=0));assert.deepEqual(v,{nonharm,transition,pass:nonharm&&transition});
}
console.log(JSON.stringify({status:'PASS',records:count,pairs,training,scoreChecks,forecastChecks,riskChecks,means,bounds,derivative,minCell,hashes:{[file]:sha(file),[import.meta.filename]:sha(import.meta.filename)},scope:'Independent Frechet joint bounds, raw-source scores and forecast losses, phase0 stationary-point/convex optimum, paired aggregation and four screens'}));
