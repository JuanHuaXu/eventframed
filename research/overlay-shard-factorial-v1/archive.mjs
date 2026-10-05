import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {pipeline} from 'node:stream/promises';
import {createGzip,createGunzip} from 'node:zlib';
import {fileURLToPath} from 'node:url';
const root=path.dirname(fileURLToPath(import.meta.url)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function digest(stream){const h=crypto.createHash('sha256');let bytes=0;for await(const b of stream){h.update(b);bytes+=b.length;}return{sha256:h.digest('hex'),bytes};}
const run=JSON.parse(fs.readFileSync(path.join(root,'run-results.json')));if(!run.completed)throw Error('all commands must be terminal before archival');
if(fs.existsSync(path.join(root,'ARCHIVES.json')))throw Error('preserve archive manifest');
const archives=[];
for(const c of run.commands.filter(x=>x.kind==='quality')){
 const file=path.join(root,c.transcript),archive=file+'.gz';
 await pipeline(fs.createReadStream(file),createGzip(),fs.createWriteStream(archive,{flags:'wx'}));
 const restored=await digest(fs.createReadStream(archive).pipe(createGunzip()));if(restored.sha256!==c.sha256||restored.bytes!==c.bytes)throw Error('lossless archive drift');
 const stored=await digest(fs.createReadStream(archive));archives.push({original:c.transcript,originalSHA256:c.sha256,originalBytes:c.bytes,archive:path.basename(archive),archiveSHA256:stored.sha256,archiveBytes:stored.bytes,losslessVerified:true});console.log(JSON.stringify(archives.at(-1)));
}
fs.writeFileSync(path.join(root,'ARCHIVES.json'),JSON.stringify({losslessVerified:true,archives},null,2)+'\n');
fs.writeFileSync(path.join(root,'LOCAL_ONLY.json'),JSON.stringify({files:archives.map(x=>x.original),reason:'Oversized public numerical traces remain local; exact byte-equivalent gzip archives preserve every pass and failure.',privateDataUsed:false},null,2)+'\n');
const m=JSON.parse(fs.readFileSync(path.join(root,'manifest.json'))),sourceArchives=[];
for(const arm of['rebuild','overlay','shard','both'])for(const file of['go.mod','internal/store/libravdbstore/store.go',...(arm==='shard'||arm==='both'?['internal/store/libravdbstore/sharded_topology_test.go']:[])]){
 const b=fs.readFileSync(path.join(m.paths[arm],file));if(hash(b)!==m.sources[arm][file])throw Error('frozen source drift');const dest=arm+'-'+file.replaceAll('/','-')+'.txt';fs.writeFileSync(path.join(root,dest),b);sourceArchives.push({arm,file,archive:dest,sha256:hash(b)});
}
fs.writeFileSync(path.join(root,'SOURCE_ARCHIVE.json'),JSON.stringify({sourceArchives,parent:'index-overlay-v6',dependencyCopiedUnchanged:true,runtimeAlgorithmChanges:false},null,2)+'\n');
