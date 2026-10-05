import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
const [fitPath,popPath,samplePath,replayPath]=process.argv.slice(2);
assert(fs.readFileSync(fitPath).equals(fs.readFileSync(replayPath)));
const fit=JSON.parse(fs.readFileSync(fitPath)),pop=JSON.parse(fs.readFileSync(popPath)),sample=JSON.parse(fs.readFileSync(samplePath));
assert.equal(fit.hashes[popPath],sha(popPath));assert.equal(fit.hashes[samplePath],sha(samplePath));
for(const [p,h]of Object.entries(fit.hashes))assert.equal(sha(p),h);
assert.equal(fit.records.length,2688);const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length,near=(a,b)=>assert(Math.abs(a-b)<1e-13);
let lossChecks=0;
for(let i=0;i<2688;i++){
  const r=fit.records[i],p=pop.records[i],s=sample.records[i];for(const key of ['phase','case','index','schedule']){assert.equal(r[key],p[key]);assert.equal(r[key],s[key]);}
  const byOrigin=rows=>new Map(rows.map(b=>[b.origin,b])),pm=byOrigin(p.branches),sm=byOrigin(s.branches);
  for(let a=0;a<11;a++){
    const j=r.selected[a];assert(sm.has(j)&&pm.has(j));
    near(r.actual[a],sm.get(j).actual);near(r.expected[a],sm.get(j).value);near(r.integrated[a],pm.get(j).population);lossChecks+=3;
    assert.equal(r.costs[a],j<0?0:1);
  }
  const pool=p.branches.slice(1).map(b=>b.origin),scores=r.decision.scores;assert.equal(pool.length,scores.length);
  if(pool.length){const max=Math.max(...scores),origin=Math.min(...pool.filter((_,j)=>max-scores[j]<=1e-10));assert.equal(r.selected[5],origin);assert.equal(r.selected[6],max>0?origin:-1);near(mean(scores),r.decision.poolMean);}
  assert.equal(r.costs[6],r.costs[7]);assert.equal(r.costs[6],r.costs[8]);
  assert.equal(r.selected[7],r.selected[6]<0?-1:r.selected[1]);assert.equal(r.selected[8],r.selected[6]<0?-1:r.selected[2]);
}
assert.equal(fit.groups.length,84);
for(const g of fit.groups){
  const rs=fit.records.filter(r=>r.phase===g.phase&&r.case===g.case&&r.schedule===g.schedule);assert.equal(rs.length,32);
  for(const field of ['actual','expected','integrated']){
    for(let a=0;a<11;a++)near(g[field].means[a],mean(rs.map(r=>r[field][a])));
    for(const [mode,a,controls]of [['forced',5,[1,2]],['gated',6,[7,8]]])for(let j=0;j<2;j++){
      const values=rs.map(r=>r[field][controls[j]]-r[field][a]),m=mean(values),variance=values.reduce((s,x)=>s+(x-m)*(x-m),0)/31,se=Math.sqrt(variance/32),got=g[field][mode+'Gains'][j];near(got.mean,m);near(got.lower,m-3.5*se);near(got.upper,m+3.5*se);
    }
  }
}
for(const field of ['actual','expected','integrated'])for(const mode of ['forced','gated']){
  const gs=fit.groups.filter(g=>g.phase===1&&g.schedule===1),k=mode+'Gains';
  const nonharm=gs.filter(g=>g[field][k].every(x=>x.lower>=-.001)).length,positive=gs.filter(g=>g[field][k].every(x=>x.lower>0)).length;
  const transition=gs.filter(g=>g.case===19||g.case===20).every(g=>g[field][k].every(x=>x.lower>0));
  assert.deepEqual(fit.screens[field][mode],{nonharm:nonharm===21,transition,pass:nonharm===21&&transition,nonharmCells:nonharm,positiveCells:positive});
}
const phases=[];
for(const phase of [0,1]){
  const rows=[];
  for(let i=0;i<2688;i++){
    const r=fit.records[i];if(r.phase!==phase||r.schedule!==1)continue;
    const p=pop.records[i],ys=p.branches.slice(1).map(b=>p.branches[0].population-b.population),ps=r.decision.scores;
    rows.push({ys,ps,y:mean(ys),p:mean(ps)});
  }
  assert.equal(rows.length,672);const overall=mean(rows.map(r=>r.y));
  const withinVariance=mean(rows.map(r=>mean(r.ys.map(y=>(y-r.y)**2)))),betweenVariance=mean(rows.map(r=>(r.y-overall)**2));
  const withinSSE=mean(rows.map(r=>mean(r.ys.map((y,j)=>((y-r.y)-(r.ps[j]-r.p))**2)))),betweenSSE=mean(rows.map(r=>(r.y-r.p)**2));
  const totalSSE=mean(rows.map(r=>mean(r.ys.map((y,j)=>(y-r.ps[j])**2))));near(totalSSE,withinSSE+betweenSSE);
  phases.push({phase,withinVariance,betweenVariance,withinSSE,betweenSSE,totalSSE,withinR2:1-withinSSE/withinVariance,betweenR2:1-betweenSSE/betweenVariance});
}
console.log(JSON.stringify({scope:'Independent saved-loss, selection, interval and screen checks; population oracle itself relies on separately tested Go integration',fitSHA256:sha(fitPath),populationSHA256:sha(popPath),sampleSHA256:sha(samplePath),scriptSHA256:sha(import.meta.filename),lossChecks,groups:84,replayExact:true,phases}));
