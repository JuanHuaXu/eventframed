import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const root=process.cwd();
const dest=path.join(root,'research/public-task-pilot/norm-baseline');
fs.mkdirSync(dest);
const Replace={}, hashes={};
for(const name of ['layered_graph.go','layered_search.go','private_insert.go','private_connect.go','private_delete.go']){
 const source=path.join(root,'internal/researchindex',name),target=path.join(dest,name+'.txt');
 const bytes=fs.readFileSync(source);fs.writeFileSync(target,bytes,{flag:'wx'});Replace[source]=target;hashes[source]=crypto.createHash('sha256').update(bytes).digest('hex');
}
fs.writeFileSync(path.join(dest,'overlay.json'),JSON.stringify({Replace},null,2),{flag:'wx'});
fs.writeFileSync(path.join(dest,'hashes.json'),JSON.stringify(hashes,null,2),{flag:'wx'});
