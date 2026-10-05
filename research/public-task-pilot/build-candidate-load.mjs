import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-async-retirement-load/main.go','utf8');
const old='"cmd/research-async-retirement-load/main.go"';if(s.split(old).length!==2)throw Error('hash anchor');
s=s.replace(old,'"cmd/research-candidate-load/main.go", "research-candidate-only.mod", "research-candidate-only.sum", "research/public-task-pilot/CANDIDATE_ONLY_PROTOCOL.md", "research/public-task-pilot/candidate-only-overlay-v1/collection.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json"');
fs.mkdirSync('cmd/research-candidate-load');fs.writeFileSync('cmd/research-candidate-load/main.go',s,{flag:'wx'});
