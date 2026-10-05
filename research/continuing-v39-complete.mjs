// Resume verification, NOT outcome resampling or a post-hoc scoring change.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import {execFileSync} from 'node:child_process';
const prior=path.resolve('research/continuing-v39-recovery'),dir=path.resolve('research/continuing-v39-completion');fs.mkdirSync(dir,{recursive:true});
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const sealed=JSON.parse(fs.readFileSync(path.join(prior,'freeze.json'))),failure=JSON.parse(fs.readFileSync(path.join(prior,'failure.json')));
if(failure.checks.length!==5||failure.checks.at(-1).name!=='design-audit'||failure.checks.at(-1).code!==1)throw Error('unexpected terminal attempt');
for(const r of failure.checks){if(hash(fs.readFileSync(path.join(prior,r.name+'.log')))!==r.log_sha256)throw Error('old log changed');if(r.name!=='design-audit'&&r.code!==0)throw Error('bad prior step')}
const files={...sealed.files,'research/continuing-v39-complete.mjs':hash(fs.readFileSync('research/continuing-v39-complete.mjs')),'docs/experiments/mmm-continuing-v39-verification-recovery.md':hash(fs.readFileSync('docs/experiments/mmm-continuing-v39-verification-recovery.md'))};
function unchanged(){for(const[p,h]of Object.entries(files))if(hash(fs.readFileSync(p))!==h)throw Error('source changed '+p)}
unchanged();const rawSHA=await fileHash(path.join(prior,'design.jsonl')),benchmarkSHA=await fileHash(path.join(prior,'benchmarks.log'));
fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),files,prior_freeze_sha256:await fileHash(path.join(prior,'freeze.json')),prior_failure_sha256:await fileHash(path.join(prior,'failure.json')),sealed_design_sha256:rawSHA,benchmark_sha256:benchmarkSHA,verification_timeout_minutes:35,learner_elapsed_gate_ns:400000000},null,2)+'\n',{flag:'wx',mode:0o600});
fs.symlinkSync(path.join(prior,'design.jsonl'),path.join(dir,'design.jsonl'));
const checks=[];
function run(name,args,env={}){unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,...env},timeout:2160000,maxBuffer:16*1024*1024})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}fs.writeFileSync(path.join(dir,name+'-command.log'),log,{flag:'wx',mode:0o600});const row={name,args,env,start,end:new Date().toISOString(),code,wall_ms:performance.now()-begin,log_sha256:hash(log)};checks.push(row);fs.writeFileSync(path.join(dir,name+'-command.json'),JSON.stringify(row,null,2)+'\n',{flag:'wx',mode:0o600});console.log(log);unchanged();if(code!==0)throw Error(name+' failed; artifacts retained')}
try{
 run('design-audit',['test','./internal/researchdispersion','-run','^TestStudyAuditV39$','-count=1','-v','-timeout=35m'],{EVENTFRAME_CONTINUING_V39_AUDIT:path.join(dir,'design.jsonl'),EVENTFRAME_CONTINUING_V39_BENCHMARK:path.join(prior,'benchmarks.log')});
 run('confirmation',['test','./internal/researchdispersion','-run','^TestExperimentV39$','-count=1','-v','-timeout=12m'],{EVENTFRAME_CONTINUING_V39_OUT:path.join(dir,'confirmation.jsonl'),EVENTFRAME_CONTINUING_V39_SPLIT:'confirmation'});
 run('confirmation-audit',['test','./internal/researchdispersion','-run','^TestStudyAuditV39$','-count=1','-v','-timeout=35m'],{EVENTFRAME_CONTINUING_V39_AUDIT:path.join(dir,'confirmation.jsonl'),EVENTFRAME_CONTINUING_V39_BENCHMARK:path.join(prior,'benchmarks.log')});
 if(await fileHash(path.join(prior,'design.jsonl'))!==rawSHA||await fileHash(path.join(prior,'benchmarks.log'))!==benchmarkSHA)throw Error('sealed evidence changed');
 const artifacts={};for(const name of fs.readdirSync(dir).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[name]=await fileHash(path.join(dir,name));
 fs.writeFileSync(path.join(dir,'completed.json'),JSON.stringify({time:new Date().toISOString(),checks,artifacts,source_unchanged:true,original_design_recollected:false,original_attempts_preserved:true,all_seven_goals:'OPEN',whole_goal_complete:false},null,2)+'\n',{flag:'wx',mode:0o600});console.log('Both frozen outcome audits complete; reports retain any adoption failures.');
}catch(e){fs.writeFileSync(path.join(dir,'failure.json'),JSON.stringify({time:new Date().toISOString(),error:e.message,checks,whole_goal_complete:false},null,2)+'\n',{flag:'wx',mode:0o600});throw e}
