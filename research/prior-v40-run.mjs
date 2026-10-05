// Research-only staged protocol. Existing artifacts never overwritten.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const stage=process.argv[2];assert(['diagnostic','normal'].includes(stage),'choose diagnostic or normal');
const root='research/prior-v40-'+stage;fs.mkdirSync(root,{recursive:true});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const prior=JSON.parse(fs.readFileSync('research/continuing-v39-completion/freeze.json')).files;
for(const[p,h]of Object.entries(prior))assert.equal(hash(fs.readFileSync(p)),h,'prior source changed '+p);
const files={...prior};for(const dir of ['internal/researchprior','internal/researchpriorref','internal/researchdispersion'])for(const p of fs.readdirSync(dir).filter(p=>p.endsWith('.go')))files[dir+'/'+p]=hash(fs.readFileSync(dir+'/'+p));
for(const p of ['docs/experiments/mmm-prior-v40-protocol.md','research/prior-v40-run.mjs'])files[p]=hash(fs.readFileSync(p));
const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
if(stage==='normal'){const old=JSON.parse(fs.readFileSync('research/prior-v40-diagnostic/freeze.json'));assert.deepEqual(files,old.files,'normal source/model changed after diagnostic');assert(JSON.parse(fs.readFileSync('research/prior-v40-diagnostic/completed.json')).source_unchanged)}
save('freeze.json',{time:new Date().toISOString(),files,stage,candidate_tuning:false,learner_elapsed_gate_ns:400000000,constructor_gate_bytes:8<<20});
const checks=[];function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,'changed '+p)}
function run(name,args,env={},timeout=2100000){unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,...env},timeout,maxBuffer:16*1024*1024})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}fs.writeFileSync(root+'/'+name+'.log',log,{flag:'wx',mode:0o600});const row={name,args,env,start,end:new Date().toISOString(),code,wall_ms:performance.now()-begin,log_sha256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();if(code!==0)throw Error(name+' failed')}
try{
 if(stage==='diagnostic'){
  run('model-race',['test','-race','./internal/researchprior','./internal/researchpriorref','-v','-count=1']);
  run('integration-race',['test','-race','./internal/researchdispersion','-run','^TestPrior(AuditorNegative|FuturePrefix)V40$','-v','-count=1','-timeout=10m']);
  run('vet',['vet','./internal/researchprior','./internal/researchpriorref','./internal/researchdispersion']);
  run('benchmarks',['test','./internal/researchprior','-run','^$','-bench','^BenchmarkPrior','-benchmem','-benchtime=50ms','-count=3','-timeout=5m']);
 }
 const benchmark=path.resolve('research/prior-v40-diagnostic/benchmarks.log');
 for(const split of stage==='diagnostic'?['diagnostic']:['design','confirmation']){
  const raw=path.resolve(root+'/'+split+'.jsonl');
  run(split,['test','./internal/researchdispersion','-run','^TestPriorExperimentV40$','-count=1','-v','-timeout=35m'],{EVENTFRAME_PRIOR_V40_OUT:raw,EVENTFRAME_PRIOR_V40_SPLIT:split});
  run(split+'-audit',['test','./internal/researchdispersion','-run','^TestPriorStudyAuditV40$','-count=1','-v','-timeout=35m'],{EVENTFRAME_PRIOR_V40_AUDIT:raw,EVENTFRAME_PRIOR_V40_BENCHMARK:benchmark});
 }
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[p]=await fileHash(root+'/'+p);
 save('completed.json',{time:new Date().toISOString(),checks,artifacts,source_unchanged:true,stage,normal_cohorts_consumed:stage==='normal',whole_goal_complete:false,goal:'ACTIVE',all_seven_whole_goals:'OPEN'});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,whole_goal_complete:false});throw e}
