// Full unchanged quality cohorts. Audits follow serial timed collection.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import{spawn}from'node:child_process';
const root='research/memo-v49-normal',oldRoot='research/memo-v49-ablation';assert(!fs.existsSync(root),'exclusive normal root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const prior=JSON.parse(fs.readFileSync(oldRoot+'/completed.json')),ablation=JSON.parse(fs.readFileSync(oldRoot+'/readback.json'));assert(prior.sourceUnchanged&&prior.checks.every(c=>c.code===0));assert(ablation.exactAllFieldsExceptCost&&ablation.total.memo.loopPass&&ablation.screen.memo.allocationPass);
const old=JSON.parse(fs.readFileSync(oldRoot+'/freeze.json')).files;for(const[p,h]of Object.entries(old))assert.equal(hash(fs.readFileSync(p)),h,'unchanged candidate '+p);
const packages=['internal/researchswitch','internal/researchmoment','internal/researchdispersion','internal/researchhybridref','internal/researchswitchref'];
const additions=[...packages.flatMap(dir=>fs.readdirSync(dir).filter(p=>p.endsWith('.go')).map(p=>dir+'/'+p)),'cmd/research-memo-normal-gen/main.go','research/memo-v49-normal-generation.json','research/memo-v49-normal-readback-generated.mjs','research/memo-v49-normal.mjs','research/memo-v49-normal-artifact-audit.mjs','docs/experiments/mmm-memo-v49-normal-protocol.md'];
const files=Object.fromEntries([...new Set([...Object.keys(old),...additions])].sort().map(p=>[p,hash(fs.readFileSync(p))]));fs.mkdirSync(root,{mode:0o700});
function save(name,x){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
for(const[p,h]of Object.entries(files)){const dest=root+'/source/'+p,b=fs.readFileSync(p);assert.equal(hash(b),h);fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const fd=fs.openSync(dest,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
save('freeze.json',{time:new Date().toISOString(),stage:'normal',files,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},qualityGateUnchanged:true,policyTuning:false,seedBases:{diagnostic:2026104807,design:2026104809,confirmation:2026104811},auditWorkers:4,timedCollectionSerialized:true});
const checks=[];
async function run(name,args,env={}){
 unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600),log=crypto.createHash('sha256');
 console.log('START '+name+' '+start);let code=-1;
 try{code=await new Promise((resolve,reject)=>{const p=spawn('go',args,{env:{...process.env,...env},stdio:['ignore','pipe','pipe']});const timeout=setTimeout(()=>p.kill('SIGTERM'),3*60*60*1000);p.on('error',e=>{clearTimeout(timeout);reject(e)});for(const s of [p.stdout,p.stderr])s.on('data',b=>{fs.writeSync(fd,b);log.update(b);process.stdout.write(b)});p.on('close',(c,signal)=>{clearTimeout(timeout);resolve(c??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:log.digest('hex')};checks.push(row);save(name+'-command.json',row);unchanged();assert.equal(code,0,name);console.log('DONE '+name);
}
try{
 await run('race',['test','-race','./internal/researchswitch','./internal/researchmoment','-run','^(TestMemoV49|TestHybridV48Compact|TestHybridV48Delayed|TestHybridV48Ownership)','-v','-count=1','-timeout=10m']);
 await run('seed-race',['test','-race','./internal/researchdispersion','-run','^TestSpecialistActualSeedSeparationV48$','-v','-count=1','-timeout=10m']);
 await run('vet',['vet',...packages.map(p=>'./'+p),'./cmd/research-memo-normal-gen']);
 await run('diagnostic-parallel-audit',['test','./internal/researchswitch','-run','^TestMemoV49ParallelAuditReference$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:path.resolve('research/hybrid-v48-env-repair-diagnostic/diagnostic-fixture.jsonl'),EVENTFRAME_HYBRID_V48_AUDIT:path.resolve(oldRoot+'/memo.jsonl')});
 await run('cost',['test','./internal/researchswitch','-run','^TestMemoV49Cost$','-v','-count=1'],{EVENTFRAME_MEMO_V49_COST:path.resolve(root+'/cost-report.json')});
 for(const split of ['design','confirmation']){
  const fixture=path.resolve(root+'/'+split+'-fixture.jsonl'),raw=path.resolve(root+'/'+split+'.jsonl');
  await run(split+'-fixture',['test','./internal/researchdispersion','-run','^TestSpecialistFixtureV48$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:fixture,EVENTFRAME_HYBRID_V48_SPLIT:split});assert(fs.statSync(fixture).size>0,'fixture SKIP/no data');
  await run(split,['test','./internal/researchswitch','-run','^TestMemoV49Normal$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:fixture,EVENTFRAME_HYBRID_V48_OUT:raw});assert(fs.statSync(raw).size>0,'collector SKIP/no data');
  await run(split+'-audit',['test','./internal/researchswitch','-run','^TestMemoV49ParallelAuditReference$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:fixture,EVENTFRAME_HYBRID_V48_AUDIT:raw});
 }
 unchanged();const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[p]=await digest(root+'/'+p);save('completed.json',{time:new Date().toISOString(),stage:'normal',checks,artifacts,sourceUnchanged:true,qualityPassUnproven:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),checks,error:e.message,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}
