import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...cells]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind,'header');assert.equal(cells.length,36);
for(const [path,digest]of Object.entries(header.Hashes)){
  assert.equal(hash(header.Sources[path]),digest);assert.equal(hash(fs.readFileSync(path)),digest);
}
assert.equal(new Set(cells.map(c=>[c.Trial,c.Mode,c.Size,c.Layout].join(':'))).size,36);
const results=cells.map(c=>{
  assert([0,1,2].includes(c.Trial)&&[50,200].includes(c.Size));
  assert(['indexed','cached','uncached'].includes(c.Mode)&&['grouped','scattered'].includes(c.Layout));
  assert.equal(c.ReadNS.length,32);assert(c.ReadNS.every(n=>Number.isSafeInteger(n)&&n>0));
  assert.equal(c.Verified,c.Size*32);assert(c.MaxBodyBytes<=8*1024*1024);
  if(c.Mode==='indexed'){assert.equal(c.Bodies,0);assert.equal(c.Slices,0);}
  if(c.Mode==='uncached'){assert.equal(c.Bodies,0);assert.equal(c.Slices,c.Verified);}
  if(c.Mode==='cached'){assert.equal(c.Bodies,c.Layout==='grouped'?32:1024);assert.equal(c.Slices,0);}
  const sorted=[...c.ReadNS].sort((a,b)=>a-b);
  return {trial:c.Trial,size:c.Size,mode:c.Mode,layout:c.Layout,verified:c.Verified,
    meanMS:c.ReadNS.reduce((a,b)=>a+b,0)/32/1e6,p95MS:sorted[Math.ceil(.95*32)-1]/1e6,
    bodies:c.Bodies,slices:c.Slices,maxBodyBytes:c.MaxBodyBytes};
});
const comparisons=[];
for(const trial of [0,1,2])for(const size of [50,200])for(const layout of ['grouped','scattered']){
  const get=mode=>results.find(c=>c.trial===trial&&c.size===size&&c.layout===layout&&c.mode===mode);
  comparisons.push({trial,size,layout,cachedVsIndexed:get('cached').meanMS/get('indexed').meanMS,
    uncachedVsIndexed:get('uncached').meanMS/get('indexed').meanMS,
    cachedVsUncached:get('cached').meanMS/get('uncached').meanMS});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,
  verified:results.reduce((n,c)=>n+c.verified,0),results,comparisons,
  limits:'Synthetic consumed read microbenchmark, fixed layout order, no confidence bounds; preparation/append excluded; no daemon/learner/crash/serving proof.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(out,null,2));
