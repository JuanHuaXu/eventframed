import fs from 'node:fs';import path from 'node:path';
const root=process.cwd(),dir=path.join(root,'research/public-task-pilot/summary-cost');fs.mkdirSync(dir,{recursive:true});
for(const [name,input] of [['control','serial-work-probe/probe.json'],['candidate','summary-backend-overlay/overlay.json']]){
 const overlay=JSON.parse(fs.readFileSync('research/public-task-pilot/'+input));
 overlay.Replace[path.join(root,'research/public-task-pilot/candidate-libravdb-v1.6.13/internal/index/hnsw/research_summary_cost_test.go')]=path.join(root,'research/public-task-pilot/summary-cost-test.go.txt');
 fs.writeFileSync(path.join(dir,name+'.json'),JSON.stringify(overlay,null,2));
}
