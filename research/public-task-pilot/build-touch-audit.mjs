import fs from 'node:fs';
const base=process.cwd();
const target=base+'/research/public-task-pilot/candidate-libravdb-v1.6.13/internal/index/hnsw/research_touch_test.go';
const source=base+'/research/public-task-pilot/hnsw-touch-test.go.txt';
fs.writeFileSync('research/public-task-pilot/hnsw-touch-overlay.json',JSON.stringify({Replace:{[target]:source}},null,2),{flag:'wx'});
