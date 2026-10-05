import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);
const raw=fs.readFileSync(input);
const [header,...cells]=raw.toString().trim().split('\n').map(JSON.parse);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
for(const [name,source] of Object.entries(header.Sources)) {
  assert.equal(hash(source),header.Hashes[name]);assert.equal(hash(fs.readFileSync(name)),header.Hashes[name]);
}
assert.equal(cells.length,24);const seen=new Set();
const means=cells.map(c=>{
  assert.ok([0,1,2].includes(c.Trial));assert.ok([50,200].includes(c.Size));
  assert.ok(['indexed','materialized'].includes(c.Mode));assert.ok(['grouped','scattered'].includes(c.Layout));
  const key=`${c.Trial}/${c.Size}/${c.Mode}/${c.Layout}`;assert.ok(!seen.has(key));seen.add(key);
  assert.equal(c.ReadNS.length,32);assert.equal(c.Verified,32*c.Size);
  for(const n of c.ReadNS)assert.ok(Number.isSafeInteger(n)&&n>0);
  return {trial:c.Trial,size:c.Size,mode:c.Mode,layout:c.Layout,meanMS:c.ReadNS.reduce((a,b)=>a+b,0)/32/1e6};
});
const pairs=means.filter(c=>c.mode==='materialized').map(c=>{
 const control=means.find(x=>x.mode==='indexed'&&x.trial===c.trial&&x.size===c.size&&x.layout===c.layout);
 return {...c,controlMS:control.meanMS,ratio:c.meanMS/control.meanMS};
});
const result={rawHash:hash(raw),capturedSources:Object.keys(header.Sources).length,verified:cells.reduce((a,c)=>a+c.Verified,0),means,pairs,
limits:'Batch-read microbenchmark. No write rescue, concurrent serving tails, migration, or whole-goal completion.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(result));
