import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-search-effort-v2/main.go','utf8');
for(const [a,b] of [
 ['type query struct {','type query struct {\n StoredExact bool\n StoredError string'],
 ['research-search-effort-v2 NEW-output.json','research-full-breadth NEW-output.json'],
 ['"cmd/research-search-effort-v2/main.go"','"cmd/research-full-breadth/main.go", "internal/researchindex/stored_diagnostic.go"'],
 ['"research/public-task-pilot/SEARCH_EFFORT_V2_PROTOCOL.md"','"research/public-task-pilot/FULL_BREADTH_PROTOCOL.md"'],
 ['control, controlErr :=', 'stored, storedErr := base.ResearchStoredVector(ctx,id)\n control, controlErr :='],
 ['offset < 4','offset < 2'],
 ['[]int{0, 800, 1600, 3200}[(i+offset)%4]','[]int{0, 6400}[(i+offset)%2]'],
 ['q := query{Ef: ef,','q := query{StoredExact: reflect.DeepEqual(stored,input), Ef: ef,'],
 ['start = time.Now()','if storedErr != nil { q.StoredError=storedErr.Error() }\n start = time.Now()'],
]){if(!s.includes(a))throw Error(a);s=s.replace(a,b);}
fs.mkdirSync('cmd/research-full-breadth',{recursive:true});
fs.writeFileSync('cmd/research-full-breadth/main.go',s,{flag:'wx'});
