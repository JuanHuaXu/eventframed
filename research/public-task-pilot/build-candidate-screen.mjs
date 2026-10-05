import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-partition-screen/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw Error(a);s=s.replace(a,b);}
change('type layout struct {','type layout struct {\n views []ri.View');
change('(*ri.HNSWBase, int64) {','(*ri.HNSWBase, int64,ri.View) {');
change('return h, time.Since(start).Nanoseconds()','return h, time.Since(start).Nanoseconds(),v');
change('h, ns := build(ctx, rs, filepath.Join(root, fmt.Sprintf("r%d-p%d-base%d"','h, ns, view := build(ctx, rs, filepath.Join(root, fmt.Sprintf("r%d-p%d-base%d"');
change('l.indexes = append(l.indexes, h)','l.indexes = append(l.indexes, h)\n l.views=append(l.views,view)');
change('h.ResearchNomination(ctx, q, 10, 0)','h.Search(ctx, l.views[p], q, 10)');
change('h, ns := build(ctx, rs, filepath.Join(root, fmt.Sprintf("r%d-p%d-replacement"','h, ns, _ := build(ctx, rs, filepath.Join(root, fmt.Sprintf("r%d-p%d-replacement"');
change('"cmd/research-partition-screen/main.go"','"cmd/research-candidate-screen/main.go", "research-candidate-only.mod", "research-candidate-only.sum", "research/public-task-pilot/CANDIDATE_ONLY_PROTOCOL.md", "research/public-task-pilot/candidate-only-overlay-v1/collection.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json"');
fs.mkdirSync('cmd/research-candidate-screen');fs.writeFileSync('cmd/research-candidate-screen/main.go',s,{flag:'wx'});
