// Exclusive consumed-data computational ablation; not new quality evidence.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync} from 'node:child_process';

const root='research/memo-v49-ablation',ref='research/hybrid-v48-env-repair-diagnostic';
assert(!fs.existsSync(root),'exclusive ablation root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const prior=JSON.parse(fs.readFileSync(ref+'/completed.json'));
assert(prior.sourceUnchanged&&prior.checks.every(c=>c.code===0));
assert(JSON.parse(fs.readFileSync(ref+'/artifact-audit.json')).sourceAndArtifactHashes);
const old=JSON.parse(fs.readFileSync(ref+'/freeze.json')).files;
for(const[p,h]of Object.entries(old))assert.equal(hash(fs.readFileSync(p)),h,'immutable original '+p);
const additions=['internal/researchmoment/memo_v49.go','internal/researchmoment/memo_v49_test.go','internal/researchswitch/memo_hybrid_v49.go','internal/researchswitch/memo_study_v49_test.go','internal/researchswitch/memo_collector_v49_generated_test.go','cmd/research-memo-collector-gen/main.go','research/memo-v49-generation.json','docs/experiments/mmm-memo-v49-protocol.md','research/memo-v49-ablation.mjs','research/memo-v49-audit.mjs'];
const files=Object.fromEntries([...new Set([...Object.keys(old),...additions])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const reference={};
for(const name of ['diagnostic-fixture.jsonl','diagnostic.jsonl']){reference[name]=await digest(ref+'/'+name);assert.equal(reference[name],prior.artifacts[name]);}
fs.mkdirSync(root,{mode:0o700});
function save(name,x){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
for(const[p,h]of Object.entries(files)){const dest=root+'/source/'+p,b=fs.readFileSync(p);assert.equal(hash(b),h);fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const fd=fs.openSync(dest,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
save('freeze.json',{time:new Date().toISOString(),files,referenceRoot:ref,reference,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},qualityPolicyUnchanged:true,consumedDiagnostic:true,cache:{capacity:128,keys:'exact mean; strength/mode invocation-fixed',replacement:'FIFO',crossConstructor:false}});
const checks=[];
function run(name,args,env={}){
 unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';
 try{log=execFileSync('go',args,{encoding:'utf8',timeout:7200000,env:{...process.env,...env},maxBuffer:16<<20})}
 catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
 const fd=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(fd,log);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
 const row={name,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name);
}
try{
 run('race',['test','-race','./internal/researchmoment','./internal/researchswitch','-run','^(TestMemoV49|TestHybridV48Compact|TestHybridV48Delayed|TestHybridV48Ownership|TestMean|TestLifecycleAndPrivateIssuedLaw|TestFailureDoesNotPublish|TestInvalidContractsAndCaps)','-v','-count=1','-timeout=10m']);
 run('vet',['vet','./internal/researchmoment','./internal/researchswitch','./cmd/research-memo-collector-gen']);
 run('cost',['test','./internal/researchswitch','-run','^TestMemoV49Cost$','-v','-count=1'],{EVENTFRAME_MEMO_V49_COST:path.resolve(root+'/cost-report.json')});
 run('paired',['test','./internal/researchswitch','-run','^TestMemoV49Study$','-v','-count=1','-timeout=120m'],{EVENTFRAME_MEMO_V49_ROOT:path.resolve(root),EVENTFRAME_MEMO_V49_FIXTURE:path.resolve(ref+'/diagnostic-fixture.jsonl'),EVENTFRAME_MEMO_V49_REFERENCE:path.resolve(ref+'/diagnostic.jsonl')});
 for(const name of ['original.jsonl','memo.jsonl'])assert(fs.statSync(root+'/'+name).size>0);
 run('dense-audit',['test','./internal/researchswitch','-run','^TestHybridV48StudyAudit$','-v','-count=1','-timeout=120m'],{EVENTFRAME_HYBRID_V48_FIXTURE:path.resolve(ref+'/diagnostic-fixture.jsonl'),EVENTFRAME_HYBRID_V48_AUDIT:path.resolve(root+'/memo.jsonl')});
 unchanged();for(const[p,h]of Object.entries(reference))assert.equal(await digest(ref+'/'+p),h,'reference mutation');
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[p]=await digest(root+'/'+p);
 save('completed.json',{time:new Date().toISOString(),checks,artifacts,sourceUnchanged:true,referenceUnchanged:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false,qualityConfirmation:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),checks,error:e.message,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}
