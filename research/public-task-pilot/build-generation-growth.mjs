import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-generation-wide/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw new Error(`ambiguous anchor:${a}`);s=s.replace(a,b);}
change('Reads: make([]sample, 256), Writes: make([]sample, 128)','Reads: make([]sample, 1024), Writes: make([]sample, 512)');
change('i < 256','i < 1024');change('i < 128','i < 512');
change('[]int{200, 800}','[]int{800, 3200, 6400}');change('[]int{20, 5}','[]int{5}');
change('60*time.Second','120*time.Second');
change('"cmd/research-generation-wide/main.go"','"cmd/research-generation-growth/main.go", "research/public-task-pilot/GROWTH_PROTOCOL.md"');
fs.mkdirSync('cmd/research-generation-growth');fs.writeFileSync('cmd/research-generation-growth/main.go',s,{flag:'wx',mode:0o600});
