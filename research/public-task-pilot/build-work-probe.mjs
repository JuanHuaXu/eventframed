import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const base=process.cwd(), dir=path.join(base,'research/public-task-pilot/work-probe-overlay');
fs.mkdirSync(dir,{recursive:true});
const backend=path.join(base,'research/public-task-pilot/candidate-libravdb-v1.6.13/internal/index/hnsw');
const replace={}, hashes={};
function patch(text,needle,replacement){assert.equal(text.split(needle).length,2,needle);return text.replace(needle,replacement);}
function save(target,name,text){const p=path.join(dir,name);fs.writeFileSync(p,text);replace[path.join(backend,target)]=p;}
for(const name of ['hnsw.go','delete.go','search.go']){
 const original=fs.readFileSync(path.join(backend,name),'utf8');let s=original;
 hashes[name]=crypto.createHash('sha256').update(original).digest('hex');
 if(name==='hnsw.go')for(const [method,kind] of [['getNodeLinks','links'],['getNodeBacklinks','backlinks']]){
  const signature=`func (h *Index) ${method}(node *Node, level int) []uint32 {`;
  s=patch(s,signature,signature+`\n if node != nil { researchWorkHit("${kind}", node.Ordinal) }`);
 }
 if(name==='search.go'){
  const signature='func (h *Index) computeDistanceOptimized(query []float32, node *Node, queryState any) (float32, error) {';
  s=patch(s,signature,signature+'\n if node != nil { researchWorkHit("queryDistance",node.Ordinal) }');
 }
 if(name==='delete.go'){
  const signature='func (h *Index) computePublishedNodeDistance(leftID, rightID uint32) (float32, error) {';
  s=patch(s,signature,signature+'\n researchWorkHit("publishedPair",leftID); researchWorkHit("pairRight",rightID)');
  const loop='for i := 0; i < h.nodes.Len(); i++ {';
  s=patch(s,loop,loop+'\n researchWorkHit("entryScan",uint32(i))');
 }
 save(name,name, s);
}
save('research_work_probe.go','research_work_probe.go',`package hnsw
import("sync";"sync/atomic")
type researchWorkCounts struct {Calls map[string]int;Unique map[string]int}
var researchWorkEnabled atomic.Bool
var researchWorkMu sync.Mutex
var researchWorkCalls map[string]int
var researchWorkIDs map[string]map[uint32]bool
func researchWorkStart(){researchWorkMu.Lock();researchWorkCalls=map[string]int{};researchWorkIDs=map[string]map[uint32]bool{};researchWorkEnabled.Store(true);researchWorkMu.Unlock()}
func researchWorkHit(kind string,id uint32){if !researchWorkEnabled.Load(){return};researchWorkMu.Lock();defer researchWorkMu.Unlock();if !researchWorkEnabled.Load(){return};researchWorkCalls[kind]++;if researchWorkIDs[kind]==nil{researchWorkIDs[kind]=map[uint32]bool{}};researchWorkIDs[kind][id]=true}
func researchWorkStop()researchWorkCounts{researchWorkMu.Lock();defer researchWorkMu.Unlock();researchWorkEnabled.Store(false);out:=researchWorkCounts{researchWorkCalls,map[string]int{}};for k,v:=range researchWorkIDs{out.Unique[k]=len(v)};return out}
`);
let test=fs.readFileSync(path.join(base,'research/public-task-pilot/hnsw-touch-test.go.txt'),'utf8');
test=patch(test,'Before,After touchSnapshot}','Before,After touchSnapshot;Work researchWorkCounts}');
test=patch(test,'if op<8{e=h.Insert','researchWorkStart();if op<8{e=h.Insert');
test=patch(test,'};if e!=nil{t.Fatal(kind,id,e)};after:=', '};work:=researchWorkStop();if e!=nil{t.Fatal(kind,id,e)};after:=');
test=patch(test,'touchResult{n,kind,id,changed,before,after}','touchResult{n,kind,id,changed,before,after,work}');
save('research_touch_test.go','research_touch_test.go',test);
fs.writeFileSync(path.join(dir,'overlay.json'),JSON.stringify({Replace:replace},null,2));
fs.writeFileSync(path.join(dir,'source-hashes.json'),JSON.stringify(hashes,null,2));
console.log(path.join(dir,'overlay.json'));
