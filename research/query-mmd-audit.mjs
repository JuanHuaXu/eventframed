import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [file,coveragePath,sourcePath,popPath]=process.argv.slice(2),r=JSON.parse(fs.readFileSync(file));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');for(const [p,h]of Object.entries(r.hashes))assert.equal(sha(p),h);
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10);
function input(p){const stream=fs.createReadStream(p),hash=crypto.createHash('sha256');stream.on('data',b=>hash.update(b));return {it:createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator](),hash};}
const streams=[coveragePath,sourcePath,popPath].map(input);for(const s of streams)await s.it.next();
const k=Array.from({length:512},(_,x)=>Math.exp(-x.toString(2).split('').filter(c=>c==='1').length/2));
// A single signed empirical measure replaces the implementation's three
// within/cross-kernel sums; reconstruct the whole measure after every append.
function direct(h,t){
  const weights=new Map();for(const x of h)weights.set(x,(weights.get(x)??0)+1/h.length);for(const x of t)weights.set(x,(weights.get(x)??0)-1/t.length);
  const points=[...weights];let sum=0;for(let i=0;i<points.length;i++){const [x,w]=points[i];sum+=w*w;for(let j=0;j<i;j++)sum+=2*w*points[j][1]*k[x^points[j][0]];}
  assert(sum>=-1e-12);return Math.max(0,sum);
}
let count=0,mmdChecks=0,scoreChecks=0,riskChecks=0,negative=0,zero=0;
for(;;){
  const next=[];for(const s of streams)next.push(await s.it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [c,source,pop]=next.map(x=>JSON.parse(x.value)),saved=r.records[count++];assert(saved);
  for(const [up,lo]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']])for(const x of [c,source,pop])assert.equal(x[up],saved[lo]);
  const history=c.Origins.map(j=>{assert(j>=-16&&j<160);if(j<0)return source.Initial[j+16].Bits;const s=source.Steps[j];assert(!s.Missing&&j+s.Delay<=160);return s.X;}),target=source.Steps.slice(0,161).map(s=>s.X),bs=c.Branches??[],scores=[],factors=[];
  if(bs.length){
    const base=direct(history,target);near(base,saved.decision.baseline);mmdChecks++;zero+=Number(base<=1e-12);
    for(const [i,b]of bs.entries()){
      const s=source.Steps[b.Origin];assert(s.Missing||b.Origin+s.Delay>160);
      const after=direct([...history,s.X],target),factor=base<=1e-12?1:1-.5*Math.sqrt(after)/Math.sqrt(base);near(after,saved.decision.after[i]);near(factor,saved.decision.factors[i]);mmdChecks++;negative+=Number(factor<0);
      let info=0;for(const x of target.slice(153,161))info+=(b.Mass[0]*(b.Conditional[0][x]-c.Base[x])**2+b.Mass[1]*(b.Conditional[1][x]-c.Base[x])**2)/8;
      near(info,saved.decision.information[i]);scores.push(info*factor);factors.push(factor);near(scores[i],saved.decision.scores[i]);scoreChecks++;
    }
  }else assert.equal(saved.decision.baseline,null);
  for(const [a,values]of [scores,factors].entries()){
    const largest=Math.max(...values),chosen=bs.length?Math.min(...bs.filter((_,i)=>values[i]>=largest-1e-10).map(b=>b.Origin)):-1;assert.equal(chosen,saved.selected[4+a]);
  }
  const map=new Map(pop.Branches.map(b=>[b.Origin,b]));for(let a=0;a<6;a++){
    const b=map.get(saved.selected[a]);assert(b);near(b.Sample[Number(b.ActualY)],saved.actual[a]);near(b.Population[0]+b.Q*(b.Population[1]-b.Population[0]),saved.integrated[a]);riskChecks+=2;
  }
}
assert.equal(count,2688);assert.equal(negative,r.negativeFactors);assert.equal(zero,r.zeroDenominators);
const hashes=streams.map(s=>s.hash.digest('hex'));for(let i=0;i<3;i++)assert.equal(hashes[i],r.hashes[[coveragePath,sourcePath,popPath][i]]);
const mean=xs=>xs.reduce((sum,x)=>sum+x/xs.length,0);let means=0,bounds=0;
for(const g of r.groups){
  const rows=r.records.filter(x=>x.phase===g.phase&&x.case===g.case&&x.schedule===g.schedule);assert.equal(rows.length,32);
  for(const f of ['actual','integrated']){
    for(let a=0;a<6;a++){near(mean(rows.map(x=>x[f][a])),g[f][a]);means++;}
    for(let a=4;a<6;a++)for(let ctrl=1;ctrl<=2;ctrl++){
      const xs=rows.map(x=>x[f][ctrl]-x[f][a]),m=mean(xs);let pairs=0;for(let i=0;i<32;i++)for(let j=i+1;j<32;j++)pairs+=(xs[i]-xs[j])**2;
      const se=Math.sqrt(pairs/(32*32*31)),v=g.gains[f][a-4][ctrl-1];near(v.mean,m);near(v.lower,m-3.5*se);near(v.upper,m+3.5*se);bounds++;
    }
  }
}
for(const f of ['actual','integrated'])for(const v of r.screens[f]){
  const a=v.arm-4,gs=r.groups.filter(g=>g.phase===1&&g.schedule===1),nonharm=!gs.some(g=>g.gains[f][a].some(x=>x.lower<-.001)),transition=!gs.filter(g=>g.case>=19).some(g=>g.gains[f][a].some(x=>x.lower<=0));assert.deepEqual(v,{arm:a+4,nonharm,transition,pass:nonharm&&transition});
}
console.log(JSON.stringify({status:'PASS',records:count,mmdChecks,scoreChecks,riskChecks,means,bounds,hashes:{[file]:sha(file),[import.meta.filename]:sha(import.meta.filename)},scope:'Independent signed-measure MMD, full append reconstruction, source-based scores/choices/risks and paired aggregation'}));
