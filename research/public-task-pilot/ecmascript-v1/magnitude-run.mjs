import fs from 'node:fs';import crypto from 'node:crypto';import {execFileSync} from 'node:child_process';
const root='research/public-task-pilot/ecmascript-v1/',hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function save(p,data){fs.writeFileSync(root+p,JSON.stringify(data,null,2)+'\n',{flag:'wx',mode:0o600})}
const files={};function scan(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=dir+'/'+e.name;if(e.isDirectory())scan(p);else if(p.endsWith('.go'))files[p]=hash(fs.readFileSync(p))}}
scan('internal');scan('cmd/public-ecma-magnitude');for(const p of ['go.mod','go.sum',root+'prepared/corpus.json',root+'prepared/queries.json',root+'MAGNITUDE_PROTOCOL.md',root+'MAGNITUDE_SETUP.md'])files[p]=hash(fs.readFileSync(p));
const prep=JSON.parse(fs.readFileSync(root+'prepared/prepared.json'));for(const [name,h]of Object.entries(prep.artifacts))if(hash(fs.readFileSync(root+'prepared/'+name+'.json'))!==h)throw Error('prepared fixture changed');
const labels=JSON.parse(fs.readFileSync(root+'prepared/oracle.json')).filter(l=>l.split==='fit');if(labels.length!==36)throw Error('fit count');save('magnitude-fit-labels.json',labels);save('magnitude-runtime-freeze.json',files);
const auditFiles={};for(const p of [root+'magnitude-run.mjs',root+'magnitude-audit.mjs',root+'magnitude-fit-labels.json',root+'prepared/oracle.json'])auditFiles[p]=hash(fs.readFileSync(p));save('magnitude-auditor-freeze.json',auditFiles);
const checks=[];function run(name,cmd,args){for(const [p,h]of Object.entries({...files,...auditFiles}))if(hash(fs.readFileSync(p))!==h)throw Error('changed '+p);const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync(cmd,args,{encoding:'utf8',maxBuffer:16*1024*1024,timeout:600000})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}fs.writeFileSync(root+'magnitude-'+name+'.log',log,{flag:'wx',mode:0o600});checks.push({name,cmd,args,start,code,wall_ms:performance.now()-begin,log_sha256:hash(log)});console.log(log);if(code!==0)throw Error(name+' failed')}
try{
 run('race','go',['test','-race','./internal/researchmagnitude','./cmd/public-ecma-magnitude','-count=1']);
 run('vet','go',['vet','./internal/researchmagnitude','./cmd/public-ecma-magnitude']);
 run('bench','go',['test','./internal/researchmagnitude','-run','^$','-bench','^BenchmarkMagnitude','-benchmem','-count=3','-benchtime=100ms']);
 run('fit','go',['run','./cmd/public-ecma-magnitude','fit','unused',root+'magnitude-fit.jsonl']);
 run('fit-audit','node',[root+'magnitude-audit.mjs','fit',root+'magnitude-fit.jsonl',root+'magnitude-model.json']);
 const modelSHA=hash(fs.readFileSync(root+'magnitude-model.json'));save('magnitude-heldout-freeze.json',{time:new Date().toISOString(),model_sha256:modelSHA,design_and_confirmation_unseen:true});
 run('heldout','go',['run','./cmd/public-ecma-magnitude','heldout',root+'magnitude-model.json',root+'magnitude-heldout.jsonl']);
 run('heldout-audit','node',[root+'magnitude-audit.mjs','heldout',root+'magnitude-heldout.jsonl',root+'magnitude-results.json']);
 if(hash(fs.readFileSync(root+'magnitude-model.json'))!==modelSHA)throw Error('model changed');for(const [p,h]of Object.entries({...files,...auditFiles}))if(hash(fs.readFileSync(p))!==h)throw Error('final source changed '+p);
 save('magnitude-completed.json',{time:new Date().toISOString(),checks,source_files:Object.keys(files).length,auditor_files:Object.keys(auditFiles).length,model_sha256:modelSHA,goal:'ACTIVE',whole_goal_complete:false});
}catch(e){save('magnitude-failure.json',{time:new Date().toISOString(),error:e.message,checks,whole_goal_complete:false});throw e}
