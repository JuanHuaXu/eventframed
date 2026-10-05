import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...cells]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind,'header');assert.equal(cells.length,12);
for(const [name,digest]of Object.entries(header.Hashes)){
  assert.equal(hash(header.Sources[name]),digest);assert.equal(hash(fs.readFileSync(name)),digest);
}
assert.equal(new Set(cells.map(c=>[c.Trial,c.Size,c.Envelope].join(':'))).size,12);
const results=cells.map(c=>{
  assert([0,1,2].includes(c.Trial)&&[50,200].includes(c.Size)&&typeof c.Envelope==='boolean');
  assert.equal(c.AppendNS.length,32);assert(c.AppendNS.every(n=>Number.isSafeInteger(n)&&n>0));
  assert.equal(c.Verified,c.Size*32);assert(Number.isSafeInteger(c.LookupNS)&&c.LookupNS>0);
  const sorted=[...c.AppendNS].sort((a,b)=>a-b);
  return {trial:c.Trial,size:c.Size,envelope:c.Envelope,verified:c.Verified,
    appendMeanMS:c.AppendNS.reduce((a,b)=>a+b,0)/32/1e6,
    appendP95MS:sorted[Math.ceil(.95*32)-1]/1e6,lookupMeanUS:c.LookupNS/c.Verified/1e3};
});
const pairs=[];
for(const trial of [0,1,2])for(const size of [50,200]){
  const base=results.find(c=>c.trial===trial&&c.size===size&&!c.envelope);
  const candidate=results.find(c=>c.trial===trial&&c.size===size&&c.envelope);
  pairs.push({trial,size,appendRatio:candidate.appendMeanMS/base.appendMeanMS,lookupRatio:candidate.lookupMeanUS/base.lookupMeanUS});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,
  verified:results.reduce((n,c)=>n+c.verified,0),results,pairs,
  limits:'Storage prototype only; no learner, feedback, service guard or complete crash/replay contract. Read paths differ in defensive JSON validation. No serving promotion.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(out,null,2));
