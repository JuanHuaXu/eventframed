import fs from 'node:fs';
let s=fs.readFileSync('cmd/research-task-batch-durable/main.go','utf8');
function change(a,b){if(s.split(a).length!==2)throw new Error('ambiguous anchor');s=s.replace(a,b);}
change('err = permits.Write(c, invoke)','err = researchbatch.RunAdmitted(c, permits.Write, invoke)');
change('"cmd/research-task-batch-durable/main.go"','"cmd/research-task-phase-durable/main.go", "internal/researchbatch/admission.go", "research/public-task-pilot/phase-aware-overlay-v1/queue.go.txt", "research/public-task-pilot/PHASE_QUEUE_PROTOCOL.md"');
fs.mkdirSync('cmd/research-task-phase-durable');fs.writeFileSync('cmd/research-task-phase-durable/main.go',s,{flag:'wx',mode:0o600});
