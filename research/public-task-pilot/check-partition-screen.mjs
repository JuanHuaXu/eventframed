import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';
const d=JSON.parse(fs.readFileSync(process.argv[2]??'research/public-task-pilot/partition-screen-results.json'));
assert.equal(d.N,6400);assert.equal(d.Dimension,768);assert.equal(d.Arms.length,4);
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
const vector=id=>Array.from({length:24},(_,b)=>[...crypto.createHash('sha256').update(`${id}/block-${b}`).digest()].map(x=>x-127.5)).flat();
const seeds=Array.from({length:6400},(_,i)=>vector(`seed-${i}`));
const norm=a=>Math.sqrt(a.reduce((s,x)=>s+x*x,0));
const norms=seeds.map(norm);
const score=(q,qn,i)=>{let dot=0;for(let j=0;j<768;j++)dot+=q[j]*seeds[i][j];return dot/(qn*norms[i]);};
const sort=(a,b)=>b.Score-a.Score||(a.ID<b.ID?-1:a.ID>b.ID?1:0);
const owner=(id,n)=>Number(crypto.createHash('sha256').update(id).digest().readBigUInt64LE()%BigInt(n));
const oracles=Array.from({length:128},(_,i)=>{
  const q=vector(`probe-${i}`),qn=norm(q);
  return seeds.map((_,j)=>({ID:`seed-${j}`,Score:score(q,qn,j)})).sort(sort).slice(0,10);
});
const scores=[];
for(const [index,a]of d.Arms.entries()){
  assert.equal(a.Repeat,Math.floor(index/2));assert.equal(a.Partitions,index%2?8:1);
  assert.equal(a.Queries.length,1152);assert.equal(a.BuildNS.length,a.Partitions);
  const sizes=Array(a.Partitions).fill(0);for(let i=0;i<6400;i++)sizes[owner(`seed-${i}`,a.Partitions)]++;
  assert.deepEqual(a.Sizes,sizes);
  let selfMiss=0,hits=0,maxNS=0;
  for(const [i,q]of a.Queries.entries()){
    const id=i<1024?`seed-${i}`:`probe-${i-1024}`;assert.equal(q.ID,id);assert.equal(q.Error,'');
    assert.equal(q.Local.length,a.Partitions);assert(q.NS>=0);maxNS=Math.max(maxNS,q.NS);
    const input=vector(id),qn=norm(input);
    for(const [p,cs]of q.Local.entries()){
      assert.equal(cs.length,10);
      for(const c of cs){assert(/^seed-\d+$/.test(c.ID));const j=Number(c.ID.slice(5));assert(j<6400);assert.equal(owner(c.ID,a.Partitions),p);assert(Math.abs(c.Score-score(input,qn,j))<1e-12);}
    }
    const all=q.Local.flat();assert.equal(new Set(all.map(c=>c.ID)).size,all.length);
    assert.deepEqual(q.Merged,[...all].sort(sort).slice(0,10));
    if(i<1024){if(!q.Merged.some(c=>c.ID===id))selfMiss++;}
    else{
      const oracle=oracles[i-1024];assert.deepEqual(q.Oracle.map(c=>c.ID),oracle.map(c=>c.ID));
      q.Oracle.forEach((c,j)=>assert(Math.abs(c.Score-oracle[j].Score)<1e-12));
      const ids=new Set(oracle.map(c=>c.ID));hits+=q.Merged.filter(c=>ids.has(c.ID)).length;
    }
  }
  const recall=hits/1280;
  const passed=a.Partitions===8&&selfMiss===0&&recall>=.99&&recall>=scores[index-1].recall-.005&&maxNS<1e8&&a.ReplacementNS<320e6;
  const s={repeat:a.Repeat,partitions:a.Partitions,selfMiss,recall,maxMS:maxNS/1e6,replacementMS:a.ReplacementNS/1e6,screenPassed:passed};scores.push(s);console.log(JSON.stringify(s));
}
