import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/public-task-pilot/candidate-libravdb-v1.6.13/';
const paths=['internal/index/interfaces.go','internal/index/hnsw/hnsw.go','internal/index/hnsw/delete.go','internal/index/hnsw/reclamation.go','libravdb/tx.go'];
const files={};for(const path of paths){const content=fs.readFileSync(root+path,'utf8');files[path]={sha256:crypto.createHash('sha256').update(content).digest('hex')};}
const interfaces=fs.readFileSync(root+paths[0],'utf8');
assert(interfaces.includes('Commit must not\n// allocate or search'));
assert(!/func \(w \*hnswWrapper\) PrepareMutations/.test(interfaces));
assert(/func \(w \*flatWrapper\) PrepareMutations/.test(interfaces));
assert(fs.readFileSync(root+paths[2],'utf8').includes('atomic.StoreUint32(&links[i], lastVal)'));
const upstream={};
for(const [key,url]of Object.entries({head:'https://api.github.com/repos/xDarkicex/libravdb/commits/master',pulls:'https://api.github.com/repos/xDarkicex/libravdb/pulls?state=all&per_page=100',issues:'https://api.github.com/repos/xDarkicex/libravdb/issues?state=all&per_page=100'})){
 const r=await fetch(url,{headers:{Accept:'application/vnd.github+json'}});
 assert(r.ok,`${url}: ${r.status}`);upstream[key]={url,link:r.headers.get('link'),data:await r.json()};
}
const path=process.argv[2];if(!path)throw Error('new output required');
fs.writeFileSync(path,JSON.stringify({files,upstream},null,2),{flag:'wx',mode:0o600});
console.log(JSON.stringify({head:upstream.head.data.sha,pulls:upstream.pulls.data.map(x=>({number:x.number,title:x.title})),issues:upstream.issues.data.map(x=>({number:x.number,title:x.title})),pagination:[upstream.pulls.link,upstream.issues.link]}));
