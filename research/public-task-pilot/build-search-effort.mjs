import fs from 'node:fs';
let s = fs.readFileSync('cmd/research-static-recall/main.go', 'utf8');
function replace(a,b) { if(!s.includes(a)) throw Error(`missing ${a}`); s=s.replace(a,b); }
replace('type query struct {', 'type query struct {\n Ef int\n DefaultAgrees bool');
replace('research-static-recall NEW-output.json', 'research-search-effort NEW-output.json');
replace('"cmd/research-static-recall/main.go", "research/public-task-pilot/STATIC_RECALL_PROTOCOL.md",', '"cmd/research-search-effort/main.go", "internal/researchindex/nomination_diagnostic.go", "research/public-task-pilot/SEARCH_EFFORT_PROTOCOL.md",');
replace('q := query{ID: id, OwnedExact:', 'control, controlErr := base.Search(ctx, v, input, 10)\n check(controlErr)\n for offset := 0; offset < 4; offset++ {\n ef := []int{0,100,200,400}[(i+offset)%4]\n q := query{Ef:ef, ID: id, OwnedExact:');
replace('base.Search(ctx, v, input, 10)\n\t\t\tq.NS', 'base.ResearchNomination(ctx, input, 10, ef)\n\t\t\tq.NS');
replace('a.Queries = append(a.Queries, q)', 'if ef == 0 { q.DefaultAgrees = reflect.DeepEqual(control, q.Candidates) }\n a.Queries = append(a.Queries, q)\n }');
fs.mkdirSync('cmd/research-search-effort', {recursive:true});
fs.writeFileSync('cmd/research-search-effort/main.go',s,{flag:'wx'});
