// Preflight and complete paired load. No production, deployment or private data.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import os from 'node:os';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {v43TimedProcesses,v43AuditProcesses,v43NormalRunners} from './research-process-guards.mjs';

const mode=process.argv[2];assert(['preflight','load'].includes(mode));
const label=process.argv[3]??'';assert(label==='' || /^[a-z0-9-]+$/.test(label));
const suffix=label?'-'+label:'';
const preflightRoot='research/cohort-batch-v46-preflight'+suffix;
const root=mode==='preflight'?preflightRoot:'research/cohort-batch-v46'+suffix;
assert(!fs.existsSync(root),'exclusive output');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const newFiles=['cmd/research-cohort-batch-gen/main.go','research/cohort-batch-v46-generation.json','research/cohort-batch-v46-audit.mjs','research/cohort-batch-v46-run.mjs','docs/experiments/mmm-cohort-batch-v46-protocol.md'];
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(d+'/'+e.name):e.name.endsWith('.go')?[d+'/'+e.name]:[])}
const inherited=JSON.parse(fs.readFileSync('research/eager-load-v44-owner-time/freeze.json')).files;
const names=[...new Set([...Object.keys(inherited),...below('internal'),...newFiles])].sort();
const files=Object.fromEntries(names.map(p=>[p,hash(fs.readFileSync(p))]));
const idle=()=>assert.equal(v43TimedProcesses().length+v43AuditProcesses().length+v43NormalRunners().length,0,'no competing V43 experiment');
function unchanged(){for(const[p,h]of Object.entries(files)) assert.equal(hash(fs.readFileSync(p)),h,p)}
if(mode==='load'){
  const pre=JSON.parse(fs.readFileSync(preflightRoot+'/completed.json'));
  const freeze=JSON.parse(fs.readFileSync(preflightRoot+'/freeze.json'));
  assert(pre.sourceUnchanged && pre.checks.every(c=>c.code===0));assert.deepEqual(files,freeze.files,'same preflight source');
}
idle();fs.mkdirSync(root,{mode:0o700});
function save(name,value){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(value,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
for(const[p,h]of Object.entries(files)){const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const b=fs.readFileSync(p);assert.equal(hash(b),h);fs.writeFileSync(dest,b,{flag:'wx',mode:0o600})}
save('freeze.json',{time:new Date().toISOString(),files,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg(),memory:os.totalmem()},originalGatesUnchanged:true,productionUntouched:true});
const checks=[];
function run(name,command,args,env={},requiredRoots=[]){
  idle();unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';
  try{log=execFileSync(command,args,{encoding:'utf8',env:{...process.env,...env},timeout:900000,maxBuffer:64<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
  fs.writeFileSync(root+'/'+name+'.log',log,{flag:'wx',mode:0o600});
  const row={name,command,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:hash(log),requiredRoots};checks.push(row);save(name+'-command.json',row);
  console.log(name,code,row.wallMS.toFixed(1)+'ms');unchanged();assert.equal(code,0,log.slice(-12000));
  for(const test of requiredRoots){assert(log.includes('--- PASS: '+test+' ('),'root did not execute '+test);assert(!log.includes('--- SKIP: '+test+' ('),'root skipped '+test)}
}
try{
  if(mode==='preflight'){
    const roots=['TestResearchCohortSourcePreservationV46','TestResearchCohortEightJournalDurabilityV46',
      ...['Lifecycle','Cancellation','Batch','Gap'].map(n=>'TestResearchJoinedWitness'+n+'V25_CohortV46'),
      ...['Retry','Lifecycle','Interruption'].map(n=>'TestResearchBatchArchive'+n+'V34_CohortV46'),
      'TestResearchEagerTracedOwnerAndDurabilityV44_CohortV46'];
    run('candidate-race','go',['test','-race','./internal/store/libravdbstore','-run','^TestResearch.*(CohortV46|PreservationV46|DurabilityV46)$','-v','-count=1','-timeout=5m'],{EVENTFRAME_RUN_JOINED_WITNESS_V25:'1',EVENTFRAME_RUN_ARCHIVE_COST_V34:'1',EVENTFRAME_RUN_EAGER_LOAD_V44:'1'},roots);
    const old=['TestResearchEagerPredicatePreservationV44','TestResearchEagerWorkloadGenerationV44','TestResearchEagerTracedOwnerAndDurabilityV44'];
    run('original-race','go',['test','-race','./internal/store/libravdbstore','-run','^TestResearchEager.*V44$','-v','-count=1','-timeout=5m'],{EVENTFRAME_RUN_EAGER_LOAD_V44:'1'},old);
    run('validity-race','go',['test','-race','./internal/researchadmission','./internal/researchvalidity','-count=3','-timeout=2m']);
    run('vet','go',['vet','./internal/store/libravdbstore','./cmd/research-cohort-batch-gen']);
    run('checker-controls','node',['research/cohort-batch-v46-audit.mjs','--self-test']);
    run('stream-controls','node',['research/eager-load-v44-stream-test.mjs']);
    run('owner-time-controls','node',['research/eager-load-v44-owner-time-test.mjs']);
    run('process-controls','node',['research/research-process-guards-test.mjs']);
    unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,executedRequiredRoots:roots,loadedExperiment:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});
  }else{
    run('load','go',['test','./internal/store/libravdbstore','-run','^TestResearchCohortBatchLoadV46$','-v','-count=1','-timeout=10m'],{EVENTFRAME_COHORT_BATCH_V46_OUTPUT:path.resolve(root+'/raw.ndjson'),EVENTFRAME_COHORT_BATCH_V46_FREEZE:path.resolve(root+'/freeze.json')},['TestResearchCohortBatchLoadV46']);
    run('audit','node',['research/cohort-batch-v46-audit.mjs',root+'/raw.ndjson']);
    const report=JSON.parse(fs.readFileSync(root+'/audit.log'));save('audit.json',report);assert.equal(report.trials.length,32);assert.equal(report.rawSHA256,hash(fs.readFileSync(root+'/raw.ndjson')));
    unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,rawSHA256:report.rawSHA256,passedFiniteTrials:report.trials.filter(t=>t.pass).length,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});
  }
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}
