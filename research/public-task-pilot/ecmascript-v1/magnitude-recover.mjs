// Reaudit sealed FIT, never resample it. Freeze the supplementary procedure
// before fitting and before the first held-out retrieval.
import fs from 'node:fs';import crypto from 'node:crypto';import {execFileSync} from 'node:child_process';import assert from 'node:assert/strict';import {audit} from './magnitude-supplement.mjs';
const root='research/public-task-pilot/ecmascript-v1/',hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const read=p=>JSON.parse(fs.readFileSync(root+p)),save=(p,x)=>fs.writeFileSync(root+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
const failure=read('magnitude-failure.json');assert.equal(failure.checks.length,5);assert.equal(failure.checks.at(-1).code,1);for(const c of failure.checks)assert.equal(hash(fs.readFileSync(root+'magnitude-'+c.name+'.log')),c.log_sha256);
const files={...read('magnitude-runtime-freeze.json'),...read('magnitude-auditor-freeze.json')};
for(const p of ['magnitude-recover.mjs','magnitude-supplement.mjs','MAGNITUDE_AUDIT_RECOVERY.md','magnitude-fit.jsonl','magnitude-failure.json'])files[root+p]=hash(fs.readFileSync(root+p));
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,'changed '+p)}
unchanged();save('magnitude-supplement-freeze.json',{time:new Date().toISOString(),files,fit_recollected:false,original_audit_pass:false,design_and_confirmation_unseen:true});
const raw=fs.readFileSync(root+'magnitude-fit.jsonl','utf8').trim().split('\n').map(l=>JSON.parse(l));audit(raw,'fit');
const mutations=[
 ['footer',r=>r.at(-1).complete++],['foreign_query',r=>r[1].Case='foreign'],['duplicate_case',r=>r[2]=structuredClone(r[1])],['journal',r=>r[1].JournalStored=false],
 ['frame',r=>r.find(v=>v.Frames).Frames[Object.keys(r.find(v=>v.Frames).Frames)[0]][5]+='tamper'],
 ['law',r=>r[1].Order[0].law.useful+=.01],['packet_score',r=>r[1].Packed[0].score+=.01],
 ['lexical',r=>r.find(v=>v.HookCalls===1).Lexical[0]+=.01],['formula',r=>r.find(v=>v.HookCalls===1).HookScores[0]+=.01],
 ['sort',r=>{const v=r.find(v=>v.HookCalls===1);[v.Order[0],v.Order[1]]=[v.Order[1],v.Order[0]]}],
 ['missing',r=>r.splice(1,1)],['weight',r=>r.find(v=>v.HookCalls===1).Weight+=.01],['foreign_candidate',r=>r[1].Order[0].id='foreign'],
 ['source_freeze',r=>r[0].files[Object.keys(r[0].files)[0]]='bad'],
];const controls=[];for(const[name,mutate]of mutations){const r=structuredClone(raw);mutate(r);assert.notDeepEqual(r,raw,'identity mutation '+name);assert.throws(()=>audit(r,'fit'),undefined,'accepted '+name);controls.push(name)}
save('magnitude-supplement-controls.json',{time:new Date().toISOString(),accepted_normal:true,rejected_controls:controls,original_failure_preserved:true});
const checks=[];function run(name,cmd,args){unchanged();let code=0,log='';const begin=performance.now();try{log=execFileSync(cmd,args,{encoding:'utf8',maxBuffer:16*1024*1024,timeout:600000})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}fs.writeFileSync(root+'magnitude-supplement-'+name+'.log',log,{flag:'wx',mode:0o600});checks.push({name,cmd,args,code,wall_ms:performance.now()-begin,log_sha256:hash(log)});console.log(log);if(code!==0)throw Error(name+' failed')}
try{
 run('fit-audit','node',[root+'magnitude-supplement.mjs','fit',root+'magnitude-fit.jsonl',root+'magnitude-model.json']);
 const modelSHA=hash(fs.readFileSync(root+'magnitude-model.json'));save('magnitude-heldout-freeze.json',{time:new Date().toISOString(),model_sha256:modelSHA,supplement_freeze_sha256:hash(fs.readFileSync(root+'magnitude-supplement-freeze.json')),design_and_confirmation_unseen:true});
 run('heldout','go',['run','./cmd/public-ecma-magnitude','heldout',root+'magnitude-model.json',root+'magnitude-heldout.jsonl']);
 run('heldout-audit','node',[root+'magnitude-supplement.mjs','heldout',root+'magnitude-heldout.jsonl',root+'magnitude-results.json']);
 const held=fs.readFileSync(root+'magnitude-heldout.jsonl','utf8').trim().split('\n').map(l=>JSON.parse(l));assert.equal(held[0].model_sha256,modelSHA);assert.equal(hash(fs.readFileSync(root+'magnitude-model.json')),modelSHA);unchanged();
 save('magnitude-supplement-completed.json',{time:new Date().toISOString(),checks,controls:controls.length,source_unchanged:true,fit_recollected:false,original_failure_preserved:true,fit_sha256:hash(fs.readFileSync(root+'magnitude-fit.jsonl')),heldout_sha256:hash(fs.readFileSync(root+'magnitude-heldout.jsonl')),model_sha256:modelSHA,goal:'ACTIVE',whole_goal_complete:false});
}catch(e){save('magnitude-supplement-failure.json',{time:new Date().toISOString(),error:e.message,checks,whole_goal_complete:false});throw e}
