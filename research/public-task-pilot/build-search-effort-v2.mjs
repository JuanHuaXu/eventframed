import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-search-effort/main.go','utf8');
for(const [a,b] of [
  ['research-search-effort NEW-output.json','research-search-effort-v2 NEW-output.json'],
  ['"cmd/research-search-effort/main.go"','"cmd/research-search-effort-v2/main.go"'],
  ['"research/public-task-pilot/SEARCH_EFFORT_PROTOCOL.md"','"research/public-task-pilot/SEARCH_EFFORT_V2_PROTOCOL.md"'],
  ['[]int{0, 100, 200, 400}','[]int{0, 800, 1600, 3200}'],
]) { if(!s.includes(a))throw Error(a);s=s.replace(a,b); }
fs.mkdirSync('cmd/research-search-effort-v2',{recursive:true});
fs.writeFileSync('cmd/research-search-effort-v2/main.go',s,{flag:'wx'});
