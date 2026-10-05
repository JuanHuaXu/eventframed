import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import {createGzip,createGunzip} from "node:zlib";
import {pipeline} from "node:stream/promises";
import {fileURLToPath} from "node:url";

const root=path.dirname(fileURLToPath(import.meta.url)),read=n=>fs.readFileSync(path.join(root,n));
if(fs.existsSync(path.join(root,"ARCHIVES.json")))throw Error("preserve existing archives");
const run=JSON.parse(read("run-results.json"));if(!run.completed)throw Error("runner must be terminal");
const archives=[];
for(const c of run.commands.filter(c=>c.full)){
  const file=path.join(root,c.transcript),archive=c.transcript+".gz";
  const digest=crypto.createHash("sha256");for await(const b of fs.createReadStream(file))digest.update(b);if(digest.digest("hex")!==c.sha256)throw Error("raw transcript changed");
  await pipeline(fs.createReadStream(file),createGzip({level:9}),fs.createWriteStream(path.join(root,archive),{flags:"wx"}));
  const restored=crypto.createHash("sha256");let restoredBytes=0;
  for await(const b of fs.createReadStream(path.join(root,archive)).pipe(createGunzip())){restored.update(b);restoredBytes+=b.length;}
  if(restored.digest("hex")!==c.sha256||restoredBytes!==c.bytes)throw Error("archive does not restore exact transcript");
  const packed=read(archive);archives.push({original:c.transcript,originalSHA256:c.sha256,originalBytes:c.bytes,archive,archiveSHA256:crypto.createHash("sha256").update(packed).digest("hex"),archiveBytes:packed.length});
}
fs.writeFileSync(path.join(root,"ARCHIVES.json"),JSON.stringify({losslessVerified:true,archives,rawCopiesKeptLocally:true},null,2)+"\n");
fs.writeFileSync(path.join(root,"LOCAL_ONLY.json"),JSON.stringify({omittedFiles:archives.map(x=>({file:x.original,sha256:x.originalSHA256,reason:"Oversize plain transcript is preserved losslessly in "+x.archive+"; native raw copy remains local."})),privateDataUsed:false},null,2)+"\n");
console.log(JSON.stringify({verified:true,archives:archives.map(x=>({file:x.archive,bytes:x.archiveBytes})),originalBytes:archives.reduce((s,x)=>s+x.originalBytes,0),archiveBytes:archives.reduce((s,x)=>s+x.archiveBytes,0)}));
