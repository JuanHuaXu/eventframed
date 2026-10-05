// Isolated mechanical clone; subsequent semantic patches remain separately
// reviewable. All original source and checkpoint artifacts stay unchanged.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const parent='research/checkpoint-2026-10-04-tree-v60/manifest.json';
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await fileHash(parent),'0bc6cdb1eb7f5eb0c5c1e92e1f7f799237e075d116276fde7a29c76b6f963240');
const checkpoint=JSON.parse(fs.readFileSync(parent));
for(const[p,h]of Object.entries(checkpoint.copies))assert.equal(await fileHash(path.dirname(parent)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(checkpoint.generatedCompilerCopies))assert.equal(await fileHash(path.dirname(parent)+'/'+p),h,p);
for(const[p,h]of Object.entries(checkpoint.trackedHashes))assert.equal(await fileHash(p),h,p);
const sources={},outputs={},hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function clone(src,dest,selector=false){
 assert(!fs.existsSync(dest));const raw=fs.readFileSync(src);sources[src]=hash(raw);let s=raw.toString();
 if(selector){
  const rename={researchretention:'researchwindowbank',MaxTrials:'SelectorMaxTrials',Ticket:'SelectorTicket',Receipt:'SelectorReceipt',row:'selectorRow',member:'selectorMember',New:'newSelector',finite:'selectorFinite',snapshot:'selectorSnapshot',unchanged:'selectorUnchanged',closeValue:'selectorCloseValue'};
  for(const[a,b]of Object.entries(rename))s=s.replace(new RegExp('\\b'+a+'\\b','g'),b);
 }else{
  s=s.replaceAll('package researchlocal_test','package researchwindowbank_test').replaceAll('package researchlocal','package researchwindowbank');
  s=s.replaceAll('"github.com/JuanHuaXu/eventframed/internal/researchlocal"','"github.com/JuanHuaXu/eventframed/internal/researchwindowbank"');
  if(src.endsWith('inference.go')){
   s=s.replace('"sort"','"sort"\n parent "github.com/JuanHuaXu/eventframed/internal/researchlocal"');
   assert(s.includes('type Config struct{ Depth, Window int }'));s=s.replace('type Config struct{ Depth, Window int }','type Config = parent.Config');
  }
 }
 fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,s,{flag:'wx',mode:0o600});outputs[dest]=hash(s);
}
for(const file of ['inference.go','individual.go','lifecycle.go','model_test.go','reference_test.go'])clone('internal/researchlocal/'+file,'internal/researchwindowbank/'+file);
clone('internal/researchretention/selector.go','internal/researchwindowbank/selector.go',true);
clone('internal/researchretention/selector_test.go','internal/researchwindowbank/selector_test.go',true);
fs.writeFileSync('research/retention-v62-generation.json',JSON.stringify({parent,parentCopiesVerified:Object.keys(checkpoint.copies).length,sources,outputs,
 stage:'mechanical clone only; shared tables and integrated law/lifecycle are subsequent manual patches'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({sources:Object.keys(sources).length,outputs:Object.keys(outputs).length,parentCopiesVerified:Object.keys(checkpoint.copies).length}));
