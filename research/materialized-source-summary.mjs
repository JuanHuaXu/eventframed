import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input, output] = process.argv.slice(2);
const raw = fs.readFileSync(input);
const [header,...cells] = raw.toString().trim().split('\n').map(JSON.parse);
const hash = x => crypto.createHash('sha256').update(x).digest('hex');
for (const [name,source] of Object.entries(header.Sources)) {
  assert.equal(hash(source),header.Hashes[name]);
  assert.equal(hash(fs.readFileSync(name)),header.Hashes[name]);
}
assert.equal(cells.length,12);
const seen=new Set();
const mean = xs => xs.reduce((a,b)=>a+b,0)/xs.length;
const means=cells.map(c=>{
  assert.ok([0,1,2].includes(c.Trial)); assert.ok([50,200].includes(c.Size));
  assert.equal(typeof c.Candidate,'boolean');
  const key=`${c.Trial}/${c.Size}/${c.Candidate}`;
  assert.ok(!seen.has(key));seen.add(key);
  assert.equal(c.AppendNS.length,32);assert.equal(c.ReadNS.length,32*c.Size);
  assert.equal(c.Verified,c.ReadNS.length);
  for(const n of [...c.AppendNS,...c.ReadNS]) assert.ok(Number.isSafeInteger(n)&&n>0);
  return {trial:c.Trial,size:c.Size,candidate:c.Candidate,appendMS:mean(c.AppendNS)/1e6,readUS:mean(c.ReadNS)/1e3};
});
const pairs=means.filter(c=>c.candidate).map(c=>{
  const control=means.find(x=>!x.candidate&&x.trial===c.trial&&x.size===c.size);
  return {trial:c.trial,size:c.size,appendRatio:c.appendMS/control.appendMS,readRatio:c.readUS/control.readUS};
});
const result={rawHash:hash(raw),sourceCount:Object.keys(header.Sources).length,verified:cells.reduce((s,c)=>s+c.Verified,0),means,pairs,
  limits:'Admission-only isolated table. No feedback, migration, crash, concurrency, or whole-goal validity claim.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(result));
