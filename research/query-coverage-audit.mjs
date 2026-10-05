import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [summaryPath,popPath,samplePath]=process.argv.slice(2),sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),load=p=>JSON.parse(fs.readFileSync(p));
const summary=load(summaryPath),pop=load(popPath),sample=load(samplePath),near=(a,b)=>assert(Math.abs(a-b)<1e-13),mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const armCount=summary.arms.length,firstNewArm=armCount===7?5:7;
assert([7,10].includes(armCount));assert.deepEqual(summary.arms.slice(firstNewArm),armCount===7?['visible161','recent32']:['epig8','epig161','epig32']);
assert(fs.readFileSync(summaryPath).equals(fs.readFileSync(summaryPath.replace(/\.json$/,'-replay.json'))));
assert.equal(summary.hashes[popPath],sha(popPath));assert.equal(summary.hashes[samplePath],sha(samplePath));
for(const [p,h]of Object.entries(summary.hashes))assert.equal(sha(p),h);
assert.equal(summary.records.length,2688);assert.equal(summary.groups.length,84);let lossChecks=0;
for(let i=0;i<2688;i++){
  const r=summary.records[i],p=pop.records[i],s=sample.records[i];for(const key of ['phase','case','index','schedule']){assert.equal(r[key],p[key]);assert.equal(r[key],s[key]);}
  for(let a=0;a<armCount;a++){
    const j=r.selected[a],pb=p.branches.find(b=>b.origin===j),sb=s.branches.find(b=>b.origin===j);assert(pb&&sb);
    near(r.actual[a],sb.actual);near(r.expected[a],sb.value);near(r.integrated[a],pb.population);assert.equal(r.costs[a],Number(j>=0));lossChecks+=3;
  }
  if(r.schedule===0)assert(r.selected.every(j=>j===-1));else assert(r.costs.slice(1).every(c=>c===1));
}
for(const g of summary.groups){
  const rs=summary.records.filter(r=>r.phase===g.phase&&r.case===g.case&&r.schedule===g.schedule);assert.equal(rs.length,32);assert.equal(new Set(rs.map(r=>r.index)).size,32);
  for(const field of ['actual','expected','integrated']){
    for(let a=0;a<armCount;a++)near(g[field][a],mean(rs.map(r=>r[field][a])));
    for(let a=firstNewArm;a<armCount;a++)for(let control=1;control<=2;control++){
      const values=rs.map(r=>r[field][control]-r[field][a]),m=mean(values),variance=values.reduce((s,x)=>s+(x-m)*(x-m),0)/31,se=Math.sqrt(variance/32),got=g.gains[field][a-firstNewArm][control-1];near(got.mean,m);near(got.lower,m-3.5*se);near(got.upper,m+3.5*se);
    }
  }
}
for(const field of ['actual','expected','integrated'])for(let a=0;a<armCount-firstNewArm;a++){
  const gs=summary.groups.filter(g=>g.phase===1&&g.schedule===1),nonharm=gs.filter(g=>g.gains[field][a].every(x=>x.lower>=-.001)).length,positive=gs.filter(g=>g.gains[field][a].every(x=>x.lower>0)).length,transition=gs.filter(g=>[19,20].includes(g.case)).every(g=>g.gains[field][a].every(x=>x.lower>0));
  assert.deepEqual(summary.screens[field][a],{arm:a+firstNewArm,nonharm:nonharm===21,transition,pass:nonharm===21&&transition,nonharmCells:nonharm,positiveCells:positive});
}
console.log(JSON.stringify({scope:'Independent saved-outcome and screen audit; full-response and histogram checks performed by separate raw scorer',records:2688,groups:84,lossChecks,replayExact:true,hashes:Object.fromEntries([summaryPath,popPath,samplePath,'research/query-coverage-audit.mjs'].map(p=>[p,sha(p)]))}));
