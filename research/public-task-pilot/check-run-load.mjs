import fs from 'node:fs';
import {execFileSync} from 'node:child_process';
const path=process.argv[2]??'research/public-task-pilot/run-load-results.json';
process.stdout.write(execFileSync(process.execPath,['research/public-task-pilot/check-partition-load.mjs',path],{encoding:'utf8'}));
const out=JSON.parse(fs.readFileSync(path));
for(const a of out.Arms){
 let consolidations=0;
 for(const b of a.Builds){
  if(!['flush','consolidate'].includes(b.Mode))throw Error('unknown mode');
  if(!b.Error){
   if(b.Mode==='flush'&&(b.BeforeRuns!==1||b.AfterRuns!==2))throw Error('flush layout');
   if(b.Mode==='consolidate'){consolidations++;if(b.BeforeRuns!==2||b.AfterRuns!==1)throw Error('consolidation layout');}
  }
 }
 if(consolidations<2)throw Error('insufficient completed consolidation cycles');
 console.log(JSON.stringify({n:a.N,repeat:a.Repeat,consolidations}));
}
