import fs from 'node:fs';
import cp from 'node:child_process';
const root='research/public-task-pilot/git-fusion-v1/';
const runs=[];
for(const args of [
  ['test','-race','./internal/researchfusion','./cmd/public-git-fusion','-count=1'],
  ['vet','./internal/researchfusion','./cmd/public-git-fusion'],
  ['test','./internal/researchfusion','-run','^$','-bench','^BenchmarkScores','-benchmem','-count=3']
]){
  const start=Date.now();const run=cp.spawnSync('go',args,{encoding:'utf8',timeout:180000});
  runs.push({command:['go',...args],exit:run.status,stdout:run.stdout,stderr:run.stderr,wall_ms:Date.now()-start});
  if(run.error||run.status!==0){fs.writeFileSync(root+'tests-failed.json',JSON.stringify(runs,null,2)+'\n',{flag:'wx',mode:0o600});throw run.error||Error(run.stderr+run.stdout);}
}
fs.writeFileSync(root+'tests.json',JSON.stringify(runs,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(runs,null,2));
