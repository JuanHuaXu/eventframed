import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-partial-run-load/main.go','utf8');
s=s.replaceAll('1024','8192').replaceAll('512','4096');
const a='[]int{800, 3200, 6400}';if(s.split(a).length!==2)throw Error('corpus anchor');s=s.replace(a,'[]int{6400}');
s=s.replace('"cmd/research-partial-run-load/main.go"','"cmd/research-long-partial-run/main.go", "research/public-task-pilot/LONG_PARTIAL_RUN_PROTOCOL.md"');
fs.mkdirSync('cmd/research-long-partial-run');fs.writeFileSync('cmd/research-long-partial-run/main.go',s,{flag:'wx'});
let v=fs.readFileSync('research/public-task-pilot/check-partial-run-load.mjs','utf8')
 .replace('partial-run-load-results.json','long-partial-run-results.json')
 .replace('d.Arms.length,6','d.Arms.length,2')
 .replace('[800,3200,6400]','[6400]')
 .replaceAll('1024','8192').replaceAll('512','4096')
 .replace('Six arms, sidecars, hashes and 9216 operations','Two arms, sidecars, hashes and 24576 operations');
fs.writeFileSync('research/public-task-pilot/check-long-partial-run.mjs',v,{flag:'wx'});
