import fs from 'node:fs';import assert from 'node:assert/strict';
const sum=a=>a.reduce((s,x)=>s+x,0),reports=[];
for(let rep=0;rep<2;rep++){
 const read=name=>JSON.parse(fs.readFileSync(`research/public-task-pilot/lazy-repair-cost/${name}${rep}.json`));
 const a=read('control'),b=read('candidate');assert.equal(a.length,2);assert.equal(b.length,2);
 for(let i=0;i<2;i++){
  const c=a[i],s=b[i];assert.equal(c.N,s.N);assert.equal(c.InitialHash,s.InitialHash);assert.equal(c.FinalHash,s.FinalHash);
  for(const x of [c,s]){assert.equal(x.InsertNS.length,32);assert.equal(x.DeleteNS.length,8);assert(x.SeedNS>0);assert([...x.InsertNS,...x.DeleteNS].every(n=>n>0));}
  const ratios={seed:s.SeedNS/c.SeedNS,insert:sum(s.InsertNS)/sum(c.InsertNS),delete:sum(s.DeleteNS)/sum(c.DeleteNS)};
  reports.push({rep,n:c.N,ratios,controlDeleteMS:sum(c.DeleteNS)/1e6,candidateDeleteMS:sum(s.DeleteNS)/1e6,pass:ratios.seed<=1.05&&ratios.insert<=1.05&&ratios.delete<=0.9});
 }
}
const result={pass:reports.every(r=>r.pass),reports};fs.writeFileSync('research/public-task-pilot/lazy-repair-cost/comparison.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
