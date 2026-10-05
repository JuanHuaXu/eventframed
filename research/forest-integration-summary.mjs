import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2); assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(rows.length,320);assert.equal(header.PerCell,16);assert.deepEqual(header.Seeds,[2026092221,2026092222]);
for(const [name,digest] of Object.entries(header.Hashes)) {assert.equal(hash(header.Sources[name]),digest);assert.equal(hash(fs.readFileSync(name)),digest);}
assert.equal(new Set(rows.map(r=>r.TrainSeed)).size,320);
const cases=['stable_uniform','bit_uniform','xor2_uniform','majority3_uniform','mux_uniform','parity4_uniform','bit_dependent','xor2_dependent','stable_dependent','null_uniform'];
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['design','confirmation']) for(const name of cases) {
 const rs=rows.filter(r=>r.Phase===phase&&r.Case===name);assert.equal(rs.length,16);
 assert.deepEqual(rs.map(r=>r.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
 for(const r of rs) {
  assert.equal(r.StreamSeed,phase==='design'?2026092221:2026092222);
  assert.equal(r.Original.Split,phase);assert.equal(r.Original.Scenario,name);
  assert.equal(r.Fits*3,r.Original.Fits);
  assert.equal(r.Control.SplitAt,r.Original.Arms[3].SplitAt);
  for(const a of ['Control','Fixed','Coupled']) {
   assert.equal(r[a].SplitAt,r.Control.SplitAt);
   for(const p of ['Full','Post']) {
    const m=r[a][p];assert.equal(m.N,p==='Full'?512:256);
    assert(Number.isFinite(m.Brier)&&m.Brier>=0&&m.Brier<=m.N);
    assert(Number.isInteger(m.Cost)&&m.Cost>=0&&m.Cost<=6*m.N);
    if(a==='Fixed')assert.equal(m.Cost,r.Control[p].Cost);
   }
  }
  assert.equal(r.Tapes.length,3);for(const h of r.Tapes)assert.match(h,/^[a-f0-9]{64}$/);
 }
 const loss=(r,a,p)=>r[a][p].Brier/r[a][p].N;
 const arms=Object.fromEntries(['Control','Fixed','Coupled'].map(a=>[a,{post:mean(rs.map(r=>loss(r,a,'Post'))),cost:mean(rs.map(r=>r[a].Full.Cost/512))}]));
 const contrasts={};
 for(const a of ['Fixed','Coupled']) {
  const post=interval(rs.map(r=>loss(r,'Control','Post')-loss(r,a,'Post'))),full=interval(rs.map(r=>loss(r,'Control','Full')-loss(r,a,'Full')));
  contrasts[a]={post,full,gain:post.mean>=.005&&post.lower>0,nonHarm:post.lower>=-.01&&full.lower>=-.01};
 }
 cells.push({phase,case:name,arms,contrasts});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,trajectories:320,cells,limits:'Exploratory paired mean +/-3.5SE; no anytime or population guarantee. Fixed-arm success cannot replace coupled primary failure.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(out,null,2));

