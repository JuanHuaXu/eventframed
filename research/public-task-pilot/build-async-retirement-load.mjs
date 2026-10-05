import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-derived-nosync/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw Error(a);s=s.replace(a,b);}
change('type arm struct {','type arm struct {\n RetirementDrainNS int64');
change('researchindex.NewPartitionServing(','researchindex.NewAsyncPartitionServing(');
change('if err := s.Close(); err != nil {','cleanupStart:=time.Now()\n cleanupErr:=s.Close()\n a.RetirementDrainNS=time.Since(cleanupStart).Nanoseconds()\n if err := cleanupErr; err != nil {');
change('"cmd/research-derived-nosync/main.go"','"cmd/research-async-retirement-load/main.go", "internal/researchindex/partition_async_serving.go", "internal/researchindex/retirement_queue.go", "research/public-task-pilot/ASYNC_RETIREMENT_PROTOCOL.md"');
fs.mkdirSync('cmd/research-async-retirement-load');fs.writeFileSync('cmd/research-async-retirement-load/main.go',s,{flag:'wx'});
