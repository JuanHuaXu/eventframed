import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert.equal(rows.length,12);
for(const [p,h] of Object.entries(header.Hashes)){assert.equal(hash(header.Sources[p]),h);assert.equal(hash(fs.readFileSync(p)),h);}
assert.equal(new Set(rows.map(r=>[r.Trial,r.Size,r.Measured].join('/'))).size,12);
const fields=['PreflightNS','BeginNS','PrepareNS','RowsNS','CommitNS'];
const cells=rows.map(r=>{
 assert([0,1,2].includes(r.Trial)&&[50,200].includes(r.Size));assert.equal(r.Samples.length,32);assert.equal(r.Verified,r.Size*32);
 const sum=k=>r.Samples.reduce((s,p)=>s+p[k],0),total=sum('TotalNS');
 for(const p of r.Samples){assert(Number.isSafeInteger(p.TotalNS)&&p.TotalNS>0);for(const f of fields)assert(Number.isSafeInteger(p[f])&&p[f]>=0);assert(fields.reduce((s,f)=>s+p[f],0)<=p.TotalNS);}
 return {trial:r.Trial,size:r.Size,measured:r.Measured,meanMS:total/32/1e6,phases:Object.fromEntries(fields.map(f=>[f,{meanMS:sum(f)/32/1e6,fraction:sum(f)/total}])),cleanupAndGapsFraction:1-fields.reduce((s,f)=>s+sum(f),0)/total};
});
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,verifiedOriginals:rows.reduce((s,r)=>s+r.Verified,0),cells,limits:'Rotated isolated timing attribution. No serving latency, concurrency, model freshness or production speedup claim. Payload generation and owner stages are excluded.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(out));
