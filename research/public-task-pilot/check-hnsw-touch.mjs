import fs from 'node:fs';
import assert from 'node:assert/strict';
const path=process.argv[2]??'research/public-task-pilot/hnsw-touch-results-v2.json';
const rows=JSON.parse(fs.readFileSync(path));assert.equal(rows.length,32);
let maxOut=0,maxBack=0,maxCombined=0;
for(const r of rows){
 const ids=new Set([...Object.keys(r.Before.Nodes),...Object.keys(r.After.Nodes)]);
 const changed=[...ids].filter(id=>JSON.stringify(r.Before.Nodes[id])!==JSON.stringify(r.After.Nodes[id])).length;
 assert.equal(changed,r.Changed);
 for(const n of Object.values(r.After.Nodes)){const o=n.Links.flat().length,b=n.Backlinks.flat().length;maxOut=Math.max(maxOut,o);maxBack=Math.max(maxBack,b);maxCombined=Math.max(maxCombined,o+b);}
 console.log(JSON.stringify({n:r.N,kind:r.Kind,id:r.ID,changed,globalChanged:r.Before.Global!==r.After.Global}));
}
console.log(JSON.stringify({maxOut,maxBack,maxCombined}));
