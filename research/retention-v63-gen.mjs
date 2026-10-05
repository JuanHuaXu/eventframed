// Mechanical isolated clone with an immutable parent checkpoint.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const parent='research/checkpoint-2026-10-04-retention-v62/manifest.json';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(hash(fs.readFileSync(parent)),'1ab64205ef92084495756178e1b215caa2528fc1ecbd9d6e80d3cff874ecae65');
const checkpoint=JSON.parse(fs.readFileSync(parent));
for(const[p,h]of Object.entries(checkpoint.copies))assert.equal(hash(fs.readFileSync(path.dirname(parent)+'/saved/'+p)),h,p);
for(const[p,h]of Object.entries(checkpoint.generatedCompilerCopies))assert.equal(hash(fs.readFileSync(path.dirname(parent)+'/'+p)),h,p);
for(const[p,h]of Object.entries(checkpoint.trackedHashes))assert.equal(hash(fs.readFileSync(p)),h,p);
const sources={},outputs={};
function clone(src,dest,transform=s=>s.replaceAll('researchwindowbank','researchwindowjournal').replaceAll('EVENTFRAME_BANK_V62','EVENTFRAME_BANK_V63')){
 assert(!fs.existsSync(dest));const raw=fs.readFileSync(src);sources[src]=hash(raw);const s=transform(raw.toString());
 fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,s,{flag:'wx',mode:0o600});outputs[dest]=hash(s);
}
for(const pkg of ['researchwindowbank','researchwindowbankref'])for(const name of fs.readdirSync('internal/'+pkg).filter(n=>n.endsWith('.go')))clone('internal/'+pkg+'/'+name,'internal/'+pkg.replace('bank','journal')+'/'+name);
clone('research/retention-v62-run.mjs','research/retention-v63-run.mjs',s=>{
 s=s.replaceAll('retention-v62','retention-v63').replaceAll('researchwindowbank','researchwindowjournal').replaceAll('EVENTFRAME_BANK_V62','EVENTFRAME_BANK_V63');
 s=s.replace('checkpoint-2026-10-04-tree-v60','checkpoint-2026-10-04-retention-v62').replace('0bc6cdb1eb7f5eb0c5c1e92e1f7f799237e075d116276fde7a29c76b6f963240','1ab64205ef92084495756178e1b215caa2528fc1ecbd9d6e80d3cff874ecae65');
 const start=s.indexOf('const extras='),end=s.indexOf('const files=',start);assert(start>=0&&end>start);
 return s.slice(0,start)+"const extras=['go.mod','go.sum','research/retention-v63-run.mjs','research/retention-v63-gen.mjs','research/retention-v63-generation.json','research/retention-v63-journal-direction.md','docs/experiments/mmm-retention-v63-protocol.md'];\n"+s.slice(end);
});
fs.writeFileSync('research/retention-v63-generation.json',JSON.stringify({parent,parentCopiesVerified:Object.keys(checkpoint.copies).length,sources,outputs,stage:'mechanical clone only; subsequent journal ownership edits separate'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({cloned:Object.keys(outputs).length,verified:Object.keys(checkpoint.copies).length}));
