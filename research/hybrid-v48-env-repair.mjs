// Exclusive isolated study. Never launches a production service or reads chats.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync} from 'node:child_process';

const stage=process.argv[2];assert(['diagnostic','normal'].includes(stage));
const root='research/hybrid-v48-env-repair-'+stage;
assert(!fs.existsSync(root),'exclusive study root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const old=JSON.parse(fs.readFileSync('research/specialist-v47-pair/freeze.json')).files;
for(const[p,h]of Object.entries(old))assert.equal(hash(fs.readFileSync(p)),h,'immutable original '+p);
const additions=[...fs.readdirSync('internal/researchswitch').filter(p=>p.endsWith('.go')).map(p=>'internal/researchswitch/'+p),...fs.readdirSync('internal/researchhybridref').filter(p=>p.endsWith('.go')).map(p=>'internal/researchhybridref/'+p),'internal/researchdispersion/hybrid_fixture_v48_test.go','cmd/research-compact-mixer-gen/main.go','research/compact-v48-generation.json','research/compact-v48-rejected-generator.go.txt','research/compact-v48-rejected-generated.go.txt','research/compact-v48-rejected-generation.json','research/compact-v48-rejected-build.md','research/hybrid-v48-before-extreme-reference.go.txt','docs/experiments/mmm-hybrid-v48-protocol.md','research/hybrid-v48-study.mjs','research/hybrid-v48-readback.mjs','research/hybrid-v48-artifact-audit.mjs','research/hybrid-v48-env-repair.mjs','research/hybrid-v48-env-repair-readback.mjs','research/hybrid-v48-env-repair-artifact-audit.mjs','research/hybrid-v48-env-repair.md'];
const files=Object.fromEntries([...new Set([...Object.keys(old),...additions])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
if(stage==='normal'){
 assert.deepEqual(files,JSON.parse(fs.readFileSync('research/hybrid-v48-env-repair-diagnostic/freeze.json')).files);
 assert(JSON.parse(fs.readFileSync('research/hybrid-v48-env-repair-diagnostic/completed.json')).sourceUnchanged);
}
fs.mkdirSync(root,{mode:0o700});
function save(name,x){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
for(const[p,h]of Object.entries(files)){
 const dest=root+'/source/'+p,b=fs.readFileSync(p);assert.equal(hash(b),h);
 fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});
 const fd=fs.openSync(dest,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
}
save('freeze.json',{time:new Date().toISOString(),stage,files,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},qualityGateUnchanged:true,policyTuning:false,seedBases:{diagnostic:2026104807,design:2026104809,confirmation:2026104811}});
const checks=[];
function run(name,args,env={}){
 unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';
 try{log=execFileSync('go',args,{encoding:'utf8',timeout:7200000,env:{...process.env,...env},maxBuffer:16<<20})}
 catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
 const fd=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(fd,log);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
 const row={name,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:hash(log)};
 checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name);
}
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
try{
 if(stage==='diagnostic'){
  run('model-race',['test','-race','./internal/researchswitch','./internal/researchswitchref','./internal/researchhybridref','-v','-count=1','-timeout=10m']);
  run('seed-race',['test','-race','./internal/researchdispersion','-run','^TestSpecialistActualSeedSeparationV48$','-v','-count=1','-timeout=10m']);
  run('vet',['vet','./internal/researchswitch','./internal/researchswitchref','./internal/researchhybridref','./internal/researchdispersion']);
 }
 // Cost screen is repeated unchanged in each stage; FAIL remains scientific data.
 run('cost',['test','./internal/researchswitch','-run','^TestHybridV48CostScreen$','-v','-count=1'],{EVENTFRAME_HYBRID_V48_COST:path.resolve(root+'/cost-report.json')});
 for(const split of stage==='diagnostic'?['diagnostic']:['design','confirmation']){
  const fixture=path.resolve(root+'/'+split+'-fixture.jsonl'),raw=path.resolve(root+'/'+split+'.jsonl');
  run(split+'-fixture',['test','./internal/researchdispersion','-run','^TestSpecialistFixtureV48$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:fixture,EVENTFRAME_HYBRID_V48_SPLIT:split});
  assert(fs.statSync(fixture).size>0,'fixture command produced no data');
  run(split,['test','./internal/researchswitch','-run','^TestHybridV48Study$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:fixture,EVENTFRAME_HYBRID_V48_OUT:raw});
  assert(fs.statSync(raw).size>0,'collector produced no data');
  run(split+'-audit',['test','./internal/researchswitch','-run','^TestHybridV48StudyAudit$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:fixture,EVENTFRAME_HYBRID_V48_AUDIT:raw});
 }
 unchanged();const artifacts={};
 for(const p of fs.readdirSync(root).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[p]=await fileHash(root+'/'+p);
 save('completed.json',{time:new Date().toISOString(),stage,checks,artifacts,sourceUnchanged:true,qualityPassUnproven:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),checks,error:e.message,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}

