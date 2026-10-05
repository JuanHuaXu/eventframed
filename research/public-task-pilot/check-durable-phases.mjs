import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync('research/public-task-pilot/durable-phase-results.json'));
assert.equal(d.Arms.length,8);
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
const groups=new Map();let calls=0;
for(const a of d.Arms) {
 assert(a.Lease);assert.equal(a.Mode,'in-window');assert.equal(a.Samples.length,32);assert.equal(a.Writes.length,16);
 assert.equal((a.ReopenErrors??[]).length,0);
 for(const [kind,rows]of [['read',a.Samples],['write',a.Writes]])for(const r of rows) {
  calls++;
  const phases=[...(r.Phases??[])].sort((a,b)=>a.StartNS-b.StartNS);
  let last=r.WaitNS,spent=0;
  if(!r.Entered)assert.equal(phases.length,0);
  const sums={search:0,journal_commit:0,event_put:0};
  for(const p of phases){
   assert(p.Name in sums);assert(p.NS>=0);assert(p.StartNS>=last);assert(p.StartNS+p.NS<=r.NS);
   last=p.StartNS+p.NS;spent+=p.NS;sums[p.Name]+=p.NS;
  }
  assert(spent+r.WaitNS<=r.NS);
  const key=[a.ReadIntervalMS,a.Enabled,kind,r.Error?'error':'success'].join('/');
  const g=groups.get(key)??{n:0,total:0,wait:0,search:0,journal_commit:0,event_put:0,remainder:0,notEntered:0};
  g.n++;g.total+=r.NS;g.wait+=r.WaitNS;g.remainder+=r.NS-r.WaitNS-spent;g.notEntered+=!r.Entered;
  for(const k of Object.keys(sums))g[k]+=sums[k];groups.set(key,g);
 }
}
assert.equal(calls,384);
for(const [key,g]of groups){const meanMS={};for(const k of ['total','wait','search','journal_commit','event_put','remainder'])meanMS[k]=g[k]/g.n/1e6;console.log(JSON.stringify({key,n:g.n,notEntered:g.notEntered,meanMS}));}
console.log('PASS:384 operations, source hashes, nonoverlapping phase bounds and reopen checks');
