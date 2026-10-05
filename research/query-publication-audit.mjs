import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const [file]=process.argv.slice(2),r=JSON.parse(fs.readFileSync(file));
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for(const [p,h]of Object.entries(r.hashes))assert.equal(sha(p),h);
const close=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-12);
const average=a=>a.reduce((sum,x)=>sum+x/a.length,0);
let means=0,bounds=0,screens=0;
assert.equal(r.records.length,2688);assert.equal(r.groups.length,84);
for(const g of r.groups){
  const rows=r.records.filter(x=>x.phase===g.phase&&x.case===g.case&&x.schedule===g.schedule);
  assert.equal(rows.length,32);assert.equal(new Set(rows.map(x=>x.index)).size,32);
  for(const s of ['A','B','C'])for(const f of ['actual','expected','actualPopulation','integrated']){
    for(let a=0;a<10;a++){close(average(rows.map(x=>x.states[s][f][a])),g.means[s][f][a]);means++;}
    for(let a=3;a<10;a++)for(let ctrl=1;ctrl<=2;ctrl++){
      const xs=rows.map(x=>x.states[s][f][ctrl]-x.states[s][f][a]),m=average(xs);
      // Pairwise squared differences independently recover sample variance.
      let pairwise=0;for(let i=0;i<32;i++)for(let j=i+1;j<32;j++)pairwise+=(xs[i]-xs[j])**2;
      const se=Math.sqrt(pairwise/(32*31*32)),saved=g.gains[s][f][a-3][ctrl-1];
      close(saved.mean,m);close(saved.lower,m-3.5*se);close(saved.upper,m+3.5*se);bounds++;
    }
  }
}
const evalGroups=r.groups.filter(g=>g.phase===1&&g.schedule===1);
for(const s of ['A','B','C'])for(const f of ['actual','expected','actualPopulation','integrated']){
  const rows=r.records.filter(x=>x.phase===1&&x.schedule===1);
  for(let a=0;a<10;a++){close(average(rows.map(x=>x.states[s][f][a])),r.phase1Delayed[s][f][a]);means++;}
  for(const saved of r.screens[s][f]){
    const a=saved.arm-3;
    const nonharm=!evalGroups.some(g=>g.gains[s][f][a].some(x=>x.lower<-.001));
    const transition=!evalGroups.filter(g=>g.case>=19).some(g=>g.gains[s][f][a].some(x=>x.lower<=0));
    assert.equal(saved.nonharm,nonharm);assert.equal(saved.transition,transition);assert.equal(saved.pass,nonharm&&transition);screens++;
  }
}
console.log(JSON.stringify({summarySHA256:sha(file),auditSHA256:sha(import.meta.filename),means,bounds,screens,status:'PASS',scope:'Independent aggregation, pairwise-variance intervals, frozen screens and input hashes; forecast truth checked separately by the source scorer and Go contracts.'}));
