import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
const base=process.cwd(),dir=path.join(base,'research/public-task-pilot/serial-work-probe');
fs.mkdirSync(dir,{recursive:true});
const backend=path.join(base,'research/public-task-pilot/candidate-libravdb-v1.6.13/internal/index/hnsw');
function patch(s,a,b){assert.equal(s.split(a).length,2,a);return s.replace(a,b);}
for(const instrumented of [false,true]){
 const name=instrumented?'probe':'control';
 let s=fs.readFileSync(instrumented?'research/public-task-pilot/work-probe-overlay/research_touch_test.go':'research/public-task-pilot/hnsw-touch-test.go.txt','utf8');
 s=patch(s,'if e=h.BatchInsert(ctx,entries);e!=nil{t.Fatal(e)}','for _,entry:=range entries {if e=h.Insert(ctx,entry);e!=nil{t.Fatal(e)}}');
 s=patch(s,'e=h.Delete(ctx,id)','if op==15 { id=strings.Clone(h.ordinalToID.Get(h.getEntryPoint().Ordinal)) };e=h.Delete(ctx,id)');
 const target=path.join(dir,name+'_test.go.txt');fs.writeFileSync(target,s);
 const overlay=instrumented?JSON.parse(fs.readFileSync('research/public-task-pilot/work-probe-overlay/overlay.json')):{Replace:{}};
 overlay.Replace[path.join(backend,'research_touch_test.go')]=target;
 fs.writeFileSync(path.join(dir,name+'.json'),JSON.stringify(overlay,null,2));
}
