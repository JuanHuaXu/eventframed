// Separate re-freeze after a confirmed no-op preflight fixture; no normal retry.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import {execFileSync} from 'node:child_process';import os from 'node:os';
const dir=path.resolve('research/continuing-v39-recovery');fs.mkdirSync(dir,{recursive:true});
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const original=JSON.parse(fs.readFileSync('research/continuing-v39/freeze.json'));
if(hash(fs.readFileSync('research/continuing-v39/preserved-source/continuing_audit_test.go'))!==original.files['internal/researchdispersion/continuing_audit_test.go'])throw Error('failed source archive mismatch');
for(const[p,h]of Object.entries(original.files))if(p!=='internal/researchdispersion/continuing_audit_test.go'&&hash(fs.readFileSync(p))!==h)throw Error('unrelated original source changed '+p);
const sources=[...Object.keys(original.files),'research/continuing-v39-recovery-run.mjs','docs/experiments/mmm-continuing-v39-setup.md'].sort();
const hashes=Object.fromEntries(sources.map(p=>[p,hash(fs.readFileSync(p))]));
fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),files:hashes,original_freeze_sha256:hash(fs.readFileSync('research/continuing-v39/freeze.json')),original_failure_sha256:hash(fs.readFileSync('research/continuing-v39/failure.json')),node:process.version,os:{platform:os.platform(),release:os.release(),arch:os.arch(),cpu:os.cpus()[0].model,cpus:os.cpus().length,memory:os.totalmem()}},null,2)+'\n',{flag:'wx',mode:0o600});
const checks=[];
function unchanged(){for(const[p,h]of Object.entries(hashes))if(hash(fs.readFileSync(p))!==h)throw Error('source changed '+p)}
function run(name,args,env={}){unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,...env},timeout:900000,maxBuffer:16*1024*1024})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}fs.writeFileSync(path.join(dir,name+'.log'),log,{flag:'wx',mode:0o600});const row={name,args,env,start,end:new Date().toISOString(),code,wall_ms:performance.now()-begin,log_sha256:hash(log)};checks.push(row);fs.writeFileSync(path.join(dir,name+'.json'),JSON.stringify(row,null,2)+'\n',{flag:'wx',mode:0o600});console.log(log);unchanged();if(code!==0)throw Error(name+' failed; artifacts retained')}
try{
 run('race',['test','-race','./internal/researchdispersion','-count=1','-v','-timeout=10m']);
 run('vet',['vet','./internal/researchdispersion']);
 run('benchmarks',['test','./internal/researchdispersion','-run','^$','-bench','^BenchmarkContinuing','-benchmem','-benchtime=100ms','-count=3','-timeout=3m']);
 for(const split of ['design','confirmation']){
  const raw=path.join(dir,split+'.jsonl');
  run(split,['test','./internal/researchdispersion','-run','^TestExperimentV39$','-count=1','-v','-timeout=12m'],{EVENTFRAME_CONTINUING_V39_OUT:raw,EVENTFRAME_CONTINUING_V39_SPLIT:split});
  // The frozen auditor's benchmark lookup is the only path parameter here.
  run(split+'-audit',['test','./internal/researchdispersion','-run','^TestStudyAuditV39$','-count=1','-v','-timeout=12m'],{EVENTFRAME_CONTINUING_V39_AUDIT:raw,EVENTFRAME_CONTINUING_V39_BENCHMARK:path.join(dir,'benchmarks.log')});
 }
 const artifacts={};for(const name of fs.readdirSync(dir).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[name]=await fileHash(path.join(dir,name));
 fs.writeFileSync(path.join(dir,'completed.json'),JSON.stringify({time:new Date().toISOString(),checks,artifacts,source_unchanged:true,all_seven_goals:'OPEN',whole_goal_complete:false},null,2)+'\n',{flag:'wx',mode:0o600});
 console.log('Completed both fresh splits and audits; adoption is in the audit JSON, not the test exit.');
}catch(e){fs.writeFileSync(path.join(dir,'failure.json'),JSON.stringify({time:new Date().toISOString(),error:e.message,checks,whole_goal_complete:false},null,2)+'\n',{flag:'wx',mode:0o600});throw e}
