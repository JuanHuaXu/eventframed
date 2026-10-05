import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
const root=process.cwd(),dir=path.join(root,'research/public-task-pilot/summary-backend-overlay');
fs.mkdirSync(dir,{recursive:true});
const backend=path.join(root,'research/public-task-pilot/candidate-libravdb-v1.6.13/internal/index/hnsw');
const overlay=JSON.parse(fs.readFileSync('research/public-task-pilot/serial-work-probe/probe.json'));
function patch(s,a,b){assert.equal(s.split(a).length,2,a);return s.replace(a,b);}
function save(name,s){const p=path.join(dir,name);fs.writeFileSync(p,s);overlay.Replace[path.join(backend,name)]=p;}
let s=fs.readFileSync(overlay.Replace[path.join(backend,'hnsw.go')],'utf8');
s=patch(s,'type Index struct {','type Index struct {\n researchEntryMu sync.Mutex\n researchEntrySummary EntrySummary');
s=patch(s,'h.nodes.Set(nodeID, node)','h.nodes.Set(nodeID, node)\n h.researchSetEntry(nodeID,node.Level)');
save('hnsw.go',s);
s=fs.readFileSync(overlay.Replace[path.join(backend,'delete.go')],'utf8');
const start=s.indexOf('\t// Fallback to scan all nodes for highest level');
const end=s.indexOf('\n\tif fallbackEntryPoint == nil {',start);
assert(start>=0&&end>start);
s=s.slice(0,start)+`\th.researchEntryMu.Lock()
 candidate, err := h.researchEntrySummary.With(deletedID,-1)
 replacement, _, ok := candidate.Best()
 h.researchEntryMu.Unlock()
 if err != nil { return err }
 if !ok { return fmt.Errorf("research summary has no replacement") }
 researchWorkHit("entrySummary",replacement)
 fallbackEntryPoint := h.nodes.Get(replacement)
`+s.slice(end);
s=patch(s,'h.nodes.Set(nodeID, nil)','h.nodes.Set(nodeID, nil)\n h.researchSetEntry(nodeID,-1)');
save('delete.go',s);
s=fs.readFileSync('internal/researchindex/entry_summary.go','utf8');
s=patch(s,'package researchindex','package hnsw');
s+=`\n// Research-only single-writer fixture hook. Bulk load/recovery/reset are not wired.
func(h *Index) researchSetEntry(id uint32,level int) {
 h.researchEntryMu.Lock();defer h.researchEntryMu.Unlock()
 next,err:=h.researchEntrySummary.With(id,level);if err!=nil {panic(err)}
 h.researchEntrySummary=next
}
`;
save('research_entry_summary.go',s);
fs.writeFileSync(path.join(dir,'overlay.json'),JSON.stringify(overlay,null,2));
