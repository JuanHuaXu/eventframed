// Isolated correctness checks; their durations are not serving benchmarks.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import{execFileSync}from'node:child_process';
import{v43TimedProcesses}from'./research-process-guards.mjs';
const label=process.argv[2];assert(/^[a-z0-9-]+$/.test(label??''),'new explicit artifact label');
const root='research/eager-load-v44-preflight-'+label;assert(!fs.existsSync(root),'exclusive preflight root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function live(){return v43TimedProcesses()}
assert.equal(live().length,0,'do not contaminate timed V43 candidate/control loops');
fs.mkdirSync(root,{mode:0o700});
function save(name,value){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(value,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
const checks=[];let files;
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(d+'/'+e.name):e.name.endsWith('.go')?[d+'/'+e.name]:[])}
const audits=['durable-witness-v23-audit','batch-archive-v33-audit','archive-cost-v34-audit','metadata-projection-v35-audit','metadata-core-v36-audit','eager-load-v44-audit','eager-load-v44-preflight','eager-load-v44-run','eager-load-v44-stream','eager-load-v44-stream-test','eager-load-v44-owner-time-test'].map(p=>'research/'+p+'.mjs');
const inventory=()=>[...below('internal'),'cmd/research-eager-load-gen/main.go','go.mod','go.sum',...audits,'research/research-process-guards.mjs','research/research-process-guards-test.mjs','docs/experiments/mmm-eager-load-v44-protocol.md'].sort();
function unchanged(){if(!files)return;assert.deepEqual(inventory(),Object.keys(files));for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
function run(name,command,args,env={},requiredRoots=[]){
 assert.equal(live().length,0,'V43 timed collection resumed before '+name);unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';
 try{log=execFileSync(command,args,{encoding:'utf8',env:{...process.env,...env},timeout:600000,maxBuffer:16<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
 const fd=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(fd,log);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
 const row={name,command,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:hash(log),requiredRoots};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name);
 for(const test of requiredRoots){assert(log.includes('--- PASS: '+test+' ('),'root did not execute '+test);assert(!log.includes('--- SKIP: '+test+' ('),'root skipped '+test)}
}
try{
 if(!fs.existsSync('internal/store/libravdbstore/research_eager_load_v44_generated_test.go'))run('generator','go',['run','./cmd/research-eager-load-gen']);
 files=Object.fromEntries(inventory().map(p=>[p,hash(fs.readFileSync(p))]));
 for(const[p,h]of Object.entries(files)){const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const b=fs.readFileSync(p);assert.equal(hash(b),h);fs.writeFileSync(dest,b,{flag:'wx',mode:0o600})}
 save('freeze.json',{time:new Date().toISOString(),files,loadedExperiment:false,noTimedV43Process:live(),possibleConcurrentOfflineAudit:true});
 const v44=['TestResearchEagerPredicatePreservationV44','TestResearchEagerWorkloadGenerationV44','TestResearchEagerTracedOwnerAndDurabilityV44'];
 run('v44-race','go',['test','-race','./internal/store/libravdbstore','-run','^TestResearchEager.*V44$','-count=1','-v','-timeout=5m'],{EVENTFRAME_RUN_EAGER_LOAD_V44:'1'},v44);
 const old=['TestResearchEagerCoreWireControlsV42','TestResearchEagerCoreOwnerBoundaryV42','TestResearchEagerCoreMutationV42','TestResearchEagerCoreGapAndCancellationV42','TestResearchWarmGenerationV37','TestResearchWarmControlsV37','TestResearchWarmCancellationV37','TestResearchBatchArchiveRetryV33','TestResearchBatchArchiveLifecycleV33','TestResearchBatchArchiveInterruptionV33'];
 run('adjacent-race','go',['test','-race','./internal/store/libravdbstore','-run','^TestResearchEagerCore.*V42$|^TestResearchWarm(Generation|Controls|Cancellation)V37$|^TestResearchBatchArchive(Retry|Lifecycle|Interruption)V33$','-count=1','-v','-timeout=5m'],{EVENTFRAME_RUN_EAGER_CORE_V42:'1',EVENTFRAME_RUN_WARM_SEARCH_V37:'1',EVENTFRAME_RUN_BATCH_ARCHIVE_V33:'1'},old);
 run('validity-race','go',['test','-race','./internal/researchadmission','./internal/researchvalidity','-count=3','-timeout=2m']);
 run('vet','go',['vet','./internal/store/libravdbstore','./cmd/research-eager-load-gen']);
 run('auditor-selftest','node',['research/eager-load-v44-audit.mjs','--self-test']);
 run('stream-controls','node',['research/eager-load-v44-stream-test.mjs']);
 run('owner-time-controls','node',['research/eager-load-v44-owner-time-test.mjs']);
 run('process-guards','node',['research/research-process-guards-test.mjs']);
 unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,executedRequiredRoots:[...v44,...old],loadedExperiment:false,qualityOrLatencyAdoption:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,loadedExperiment:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}
