import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';
const p=process.argv[2];
assert(p, 'provide result JSON');
const d=JSON.parse(fs.readFileSync(p));
const full=p.includes('full-breadth');
const levels=full?[0,6400]:p.includes('-v2-')?[0,800,1600,3200]:[0,100,200,400];
const vector=id=>Array.from({length:24},(_,b)=>[...crypto.createHash('sha256').update(`${id}/block-${b}`).digest()].map(x=>x-127.5)).flat();
const cosine=(a,b)=>a.reduce((s,x,i)=>s+x*b[i],0)/Math.sqrt(a.reduce((s,x)=>s+x*x,0)*b.reduce((s,x)=>s+x*x,0));
assert.equal(d.N,6400); assert.equal(d.Dimension,768); assert.equal(d.Arms.length,2);
for(const [path,h] of Object.entries(d.Hashes)) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path)).digest('hex'),h,path);
for(const [repeat,a] of d.Arms.entries()) {
  assert.equal(a.Repeat,repeat); assert.equal(a.Queries.length,1024*levels.length);
  for(const ef of levels) {
    const qs=a.Queries.filter(q=>q.Ef===ef);
    assert.equal(qs.length,1024);
    for(const [i,q] of qs.entries()) {
      assert.equal(q.ID,`seed-${i}`); assert(q.OwnedExact); assert.equal(q.Error,'');
      if(full){assert(q.StoredExact);assert.equal(q.StoredError,'');}
      assert.equal(q.Candidates.length,10); assert.equal(new Set(q.Candidates.map(c=>c.ID)).size,10);
      assert.equal(q.Hit,q.Candidates.some(c=>c.ID===q.ID));
      if(ef===0)assert(q.DefaultAgrees);
      const v=vector(q.ID);
      for(const c of q.Candidates) {
        assert(/^seed-\d+$/.test(c.ID)); assert(Number(c.ID.slice(5))<6400);
        assert(Math.abs(c.Score-cosine(v,vector(c.ID)))<1e-12);
      }
    }
    console.log(JSON.stringify({repeat,ef,misses:qs.filter(q=>!q.Hit).map(q=>q.ID)}));
  }
}
