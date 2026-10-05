import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hashes={},hash=b=>crypto.createHash('sha256').update(b).digest('hex');
for(const name of ['inference.go','lifecycle.go']){
 const src='internal/researchlocalref/'+name,dest='internal/researchwindowbankref/'+name;assert(!fs.existsSync(dest));
 const raw=fs.readFileSync(src),out=raw.toString().replaceAll('researchlocalref','researchwindowbankref')
  .replaceAll('"github.com/JuanHuaXu/eventframed/internal/researchlocal"','"github.com/JuanHuaXu/eventframed/internal/researchwindowbank"');
 fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,out,{flag:'wx',mode:0o600});hashes[src]={before:hash(raw),dest,after:hash(out)};
}
fs.writeFileSync('research/retention-v62-reference-generation.json',JSON.stringify(hashes,null,2)+'\n',{flag:'wx',mode:0o600});
