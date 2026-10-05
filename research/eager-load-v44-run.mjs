// Complete offered mixed load, exclusive output, no daemon/deploy/private data.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import{execFileSync}from'node:child_process';
import{v43TimedProcesses,v43AuditProcesses,v43NormalRunners}from'./research-process-guards.mjs';
const source=process.argv[2];assert(/^research\/eager-load-v44-preflight-[a-z0-9-]+$/.test(source??''),'explicit passed preflight');
const completion=JSON.parse(fs.readFileSync(source+'/completed.json'));assert(completion.sourceUnchanged&&completion.checks.every(c=>c.code===0));
const pre=JSON.parse(fs.readFileSync(source+'/freeze.json')),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const root=process.argv[3]??'research/eager-load-v44';assert(/^research\/eager-load-v44(?:-[a-z0-9-]+)?$/.test(root),'explicit isolated output root');assert(!fs.existsSync(root),'exclusive normal root');
function idle(){assert.equal(v43TimedProcesses().length+v43AuditProcesses().length+v43NormalRunners().length,0,'wait for V43 runner and independent audits to finish')}
function unchanged(){for(const[p,h]of Object.entries(pre.files))assert.equal(hash(fs.readFileSync(p)),h,p)}
idle();unchanged();fs.mkdirSync(root,{mode:0o700});
function save(name,value){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(value,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
for(const[p,h]of Object.entries(pre.files)){const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const b=fs.readFileSync(p);assert.equal(hash(b),h);fs.writeFileSync(dest,b,{flag:'wx',mode:0o600})}
save('freeze.json',{time:new Date().toISOString(),files:pre.files,preflight:source,preflightCompletedSHA256:hash(fs.readFileSync(source+'/completed.json')),host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg(),memory:os.totalmem()},originalGatesUnchanged:true});
const checks=[];
function run(name,command,args,env={},timeout=600000){
 idle();unchanged();const begin=performance.now(),start=new Date().toISOString();let code=0,log='';try{log=execFileSync(command,args,{encoding:'utf8',env:{...process.env,...env},timeout,maxBuffer:32<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
 fs.writeFileSync(root+'/'+name+'.log',log,{flag:'wx',mode:0o600});const row={name,command,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name);
}
try{
 run('load','go',['test','./internal/store/libravdbstore','-run','^TestResearchEagerLoadV44$','-v','-count=1','-timeout=5m'],{EVENTFRAME_EAGER_LOAD_V44_OUTPUT:path.resolve(root+'/raw.ndjson')},310000);
 run('audit','node',['research/eager-load-v44-audit.mjs',root+'/raw.ndjson']);
 const report=JSON.parse(fs.readFileSync(root+'/audit.log'));save('audit.json',report);assert.equal(report.rawSHA256,hash(fs.readFileSync(root+'/raw.ndjson')));
 assert.equal(report.trials.length,16);unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,rawSHA256:report.rawSHA256,passedFiniteTrials:report.trials.filter(t=>t.pass).length,qualityOrWholeGoalPassUnproven:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}
