import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
const [massPath,shrinkPath,rawPath,sourcePath,popPath,samplePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const load=p=>JSON.parse(fs.readFileSync(p)),mass=load(massPath),shrink=load(shrinkPath),pop=load(popPath),sample=load(samplePath),near=(a,b)=>assert(Math.abs(a-b)<1e-12),mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
for(const [path,obj]of [[massPath,mass],[shrinkPath,shrink]]){
  assert(fs.readFileSync(path).equals(fs.readFileSync(path.replace(/\.json$/,'-replay.json'))));assert.equal(obj.records.length,2688);assert.equal(obj.groups.length,84);
  for(const [file,h]of Object.entries(obj.hashes))assert.equal(sha(file),h);
}
let identities=0,bounds=0,lossChecks=0;
for(const r of mass.records){
  const truth=b=>b.losses[0]+b.q*(b.losses[1]-b.losses[0]),model=b=>b.losses[0]+b.p*(b.losses[1]-b.losses[0]);
  const byOrigin=new Map(r.branches.map(b=>[b.origin,b]));
  for(const b of r.branches){near(b.truth,truth(b));near(b.modeled,model(b));near(b.distortion,model(b)-truth(b));identities++;}
  for(let a=0;a<11;a++){near(r.losses[a],truth(byOrigin.get(r.selected[a])));assert.equal(r.costs[a],Number(r.selected[a]>=0));lossChecks++;}
  for(const [ai,bi,abstain,k]of [[5,6,false,0],[8,9,true,1]]){
    const actions=abstain||r.branches.length===1?r.branches:r.branches.slice(1),a=byOrigin.get(r.selected[ai]),b=byOrigin.get(r.selected[bi]);
    near(truth(a),Math.min(...actions.map(truth)));near(model(b),Math.min(...actions.map(model)));
    const regret=truth(b)-truth(a),bound=Math.abs(model(a)-truth(a))+Math.abs(model(b)-truth(b));near(r.regret[k].value,regret);near(r.regret[k].bound,bound);assert(regret>=-1e-12&&regret<=bound+1e-12);bounds++;
  }
}
function input(p){return createInterface({input:fs.createReadStream(p),crlfDelay:Infinity})[Symbol.asyncIterator]();}
const streams=[rawPath,sourcePath].map(input);for(const it of streams)await it.next();let index=0;const train=[];
for(;;){
  const next=[];for(const it of streams)next.push(await it.next());if(next.some(x=>x.done)){assert(next.every(x=>x.done));break;}
  const [raw,source]=next.map(x=>JSON.parse(x.value)),r=shrink.records[index],p=pop.records[index],s=sample.records[index++];
  for(const [upper,lower]of [['Phase','phase'],['Case','case'],['Index','index'],['Schedule','schedule']]){assert.equal(r[lower],source[upper]);assert.equal(r[lower],raw.Original[upper]);assert.equal(r[lower],p[lower]);assert.equal(r[lower],s[lower]);}
  const values=raw.Original.Decision.Values??[];
  if(r.phase===0&&r.schedule===1)for(const v of values)train.push({p:v.Mass[1],y:Number(source.Steps[v.Origin].Y),q:source.Steps[v.Origin].Q,w:1/values.length});
  const expected=[];
  for(const mode of [0,1,2,3]){
    const scores=values.map(v=>{
      const a=mode===0?shrink.labelModel.alpha:mode===1?shrink.teacherModel.alpha:0;
      const probability=mode===3?source.Steps[v.Origin].Q:a===1?v.Mass[1]:.5+a*(v.Mass[1]-.5);
      return mean(v.Conditional[0].map((x,j)=>{const y=v.Conditional[1][j],m=(1-probability)*x+probability*y;return (1-probability)*x*x+probability*y*y-m*m;}));
    });
    if(mode===0){assert.equal(scores.length,r.labelScores.length);scores.forEach((x,j)=>near(x,r.labelScores[j]));}
    expected.push(values.length?Math.min(...values.filter((_,j)=>Math.max(...scores)-scores[j]<=1e-10).map(v=>v.Origin)):-1);
  }
  assert.deepEqual(r.selected.slice(0,4),raw.Original.Decision.Selected);assert.deepEqual(r.selected.slice(4),expected);
  for(let a=0;a<8;a++){
    const j=r.selected[a],pb=p.branches.find(b=>b.origin===j),sb=s.branches.find(b=>b.origin===j);
    near(r.actual[a],sb.actual);near(r.expected[a],sb.value);near(r.integrated[a],pb.population);assert.equal(r.costs[a],Number(j>=0));lossChecks+=3;
  }
}
assert.equal(index,2688);assert.equal(train.length,4640);
for(const [target,m]of [['y',shrink.labelModel],['q',shrink.teacherModel]]){
  const numerator=train.reduce((s,r)=>s+r.w*(r.p-.5)*(r[target]-.5),0),denominator=train.reduce((s,r)=>s+r.w*(r.p-.5)**2,0);
  near(numerator,m.numerator);near(denominator,m.denominator);near(m.alpha,Math.max(0,Math.min(1,numerator/denominator)));
  const loss=a=>train.reduce((s,r)=>s+r.w*(.5+a*(r.p-.5)-r[target])**2,0);
  for(let i=0;i<=1000;i++)assert(loss(m.alpha)<=loss(i/1000)+1e-10);
}
for(const g of shrink.groups){
  const rs=shrink.records.filter(r=>r.phase===g.phase&&r.case===g.case&&r.schedule===g.schedule);assert.equal(rs.length,32);
  for(const field of ['actual','expected','integrated'])for(let a=0;a<8;a++)near(g[field][a],mean(rs.map(r=>r[field][a])));
  for(const field of ['actual','expected','integrated'])for(let a=4;a<8;a++)for(let c=1;c<=2;c++){
    const xs=rs.map(r=>r[field][c]-r[field][a]),m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32),stored=g.gains[field][a-4][c-1];near(stored.mean,m);near(stored.lower,m-3.5*se);near(stored.upper,m+3.5*se);
  }
}
for(const field of ['actual','expected','integrated'])for(let a=0;a<4;a++){
  const gs=shrink.groups.filter(g=>g.phase===1&&g.schedule===1),nonharm=gs.filter(g=>g.gains[field][a].every(x=>x.lower>=-.001)).length,positive=gs.filter(g=>g.gains[field][a].every(x=>x.lower>0)).length,transition=gs.filter(g=>[19,20].includes(g.case)).every(g=>g.gains[field][a].every(x=>x.lower>0));
  assert.deepEqual(shrink.screens[field][a],{arm:a+4,nonharm:nonharm===21,transition,pass:nonharm===21&&transition,nonharmCells:nonharm,positiveCells:positive});
}
console.log(JSON.stringify({scope:'Independent linear-risk, regret, second-moment selector, original-source calibration and saved-loss/screen audit; not efficacy',identities,bounds,lossChecks,records:index,trainingRows:train.length,replaysExact:true,hashes:Object.fromEntries([massPath,shrinkPath,rawPath,sourcePath,popPath,samplePath,'research/query-weight-audit.mjs'].map(p=>[p,sha(p)]))}));
