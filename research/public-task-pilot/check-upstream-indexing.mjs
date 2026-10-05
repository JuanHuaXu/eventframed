import fs from 'node:fs';
import crypto from 'node:crypto';
const sha='18515d4ce0620b24fd2512b0f9023b8b1f4d8355';
const dir='research/public-task-pilot/upstream-indexing-check';fs.mkdirSync(dir,{mode:0o700});
const urls={
 pulls:'https://api.github.com/repos/xDarkicex/libravdb/pulls?state=all&per_page=100',
 issues:'https://api.github.com/repos/xDarkicex/libravdb/issues?state=all&per_page=100',
 tx:`https://raw.githubusercontent.com/xDarkicex/libravdb/${sha}/libravdb/tx.go`,
 index:`https://raw.githubusercontent.com/xDarkicex/libravdb/${sha}/internal/index/interfaces.go`,
};
const entries={};
for(const [name,url] of Object.entries(urls)) {
 const r=await fetch(url);if(!r.ok)throw new Error(`${r.status} ${url}`);
 const body=await r.text();fs.writeFileSync(`${dir}/${name}.txt`,body,{flag:'wx',mode:0o600});
 entries[name]={url,hash:crypto.createHash('sha256').update(body).digest('hex'),pagination:r.headers.get('link')};
 if(name==='pulls'||name==='issues') console.log(name,JSON.stringify(JSON.parse(body).map(x=>({number:x.number,title:x.title,state:x.state,merged:x.merged_at}))));
 if(name==='tx')console.log('transaction evidence',body.includes('collection.getAllVectors(ctx)'),body.includes('state.buildIndexes(ctx, rebuildNames'));
 if(name==='index')console.log('PrepareMutations declarations',body.split('\n').filter(x=>x.startsWith('func ')&&x.includes('PrepareMutations')));
}
fs.writeFileSync(`${dir}/manifest.json`,JSON.stringify({checkedAt:new Date().toISOString(),sha,entries},null,2)+'\n',{flag:'wx',mode:0o600});
