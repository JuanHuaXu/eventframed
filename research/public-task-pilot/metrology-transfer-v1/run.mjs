// Prospective isolated transfer: no fitting, old artifacts never rewritten.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {audit} from './audit.mjs';
const root='research/public-task-pilot/metrology-transfer-v1/';
const old='research/public-task-pilot/ecmascript-v1/';
const model=old+'magnitude-v2-model.json';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const read=p=>JSON.parse(fs.readFileSync(p));
const save=(p,x)=>fs.writeFileSync(root+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
const prior={...read(old+'magnitude-v2-runtime-freeze.json'),...read(old+'magnitude-v2-auditor-freeze.json').files};
for(const[p,h]of Object.entries(prior))assert.equal(hash(fs.readFileSync(p)),h,'old source changed');
const prep=read(root+'prepared/freeze.json');
for(const[p,h]of Object.entries(prep.sources))assert.equal(hash(fs.readFileSync(root+p)),h,'prepared source changed');
const artifacts=read(root+'prepared/prepared.json').artifacts;
for(const[p,h]of Object.entries(artifacts))assert.equal(hash(fs.readFileSync(root+'prepared/'+p+'.json')),h,'prepared artifact changed');
const files={};
function walk(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=dir+'/'+e.name;if(e.isDirectory())walk(p);else if(e.isFile()&&p.endsWith('.go'))files[p]=hash(fs.readFileSync(p))}}
walk('internal');
for(const p of ['go.mod','go.sum','cmd/public-metrology-transfer/main.go',root+'PROTOCOL.md',root+'prepared/corpus.json',root+'prepared/queries.json',model])files[p]=hash(fs.readFileSync(p));
// Freeze verification does not make labels runtime inputs. Oracle is exclusive
// to the independent auditor inventory and read only after sealed output.
assert(!Object.keys(files).some(p=>p.endsWith('/oracle.json')||p.endsWith('/cases.mjs')));
const collector=fs.readFileSync('cmd/public-metrology-transfer/main.go','utf8');
assert(!/oracle|cases\.mjs|metadata|exec\.Command|Feedback\(/.test(collector));
assert.equal(read(model).fit_only,true);assert.equal(read(model).weight,1);
save('runtime-freeze.json',files);
const auditorFiles={};
for(const p of ['audit.mjs','run.mjs','prepare.mjs','prepare.test.mjs','cases.mjs','prepared/freeze.json','prepared/prepared.json','prepared/oracle.json'])auditorFiles[root+p]=hash(fs.readFileSync(root+p));
save('auditor-freeze.json',{time:new Date().toISOString(),files:auditorFiles,model_sha256:hash(fs.readFileSync(model)),outputs_unseen:true,no_refit:true,prior_sources_verified:Object.keys(prior).length});
function unchanged(){for(const[p,h]of Object.entries({...prior,...files,...auditorFiles}))assert.equal(hash(fs.readFileSync(p)),h,'changed '+p)}
const checks=[];
function run(name,cmd,args){unchanged();let code=0,log='';const start=performance.now();try{log=execFileSync(cmd,args,{encoding:'utf8',maxBuffer:16*1024*1024,timeout:600000})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}fs.writeFileSync(root+name+'.log',log,{flag:'wx',mode:0o600});checks.push({name,cmd,args,code,wall_ms:performance.now()-start,log_sha256:hash(log)});console.log(log);unchanged();if(code!==0)throw Error(name+' failed')}
try{
 run('prepare-test','node',['--test',root+'prepare.test.mjs']);
 run('race','go',['test','-race','./internal/researchmagnitude','./internal/researchsparse','./internal/researchfusion','./cmd/public-metrology-transfer','-count=1']);
 run('service-race','go',['test','-race','./internal/service','-run','Test(CaptureTurn|ResearchRank|ResearchFrontier)','-count=1']);
 run('vet','go',['vet','./internal/researchmagnitude','./cmd/public-metrology-transfer']);
 run('benchmark','go',['test','./internal/researchmagnitude','-run','^$','-bench','BenchmarkMagnitude200','-benchmem','-count=3']);
 run('collect','go',['run','./cmd/public-metrology-transfer',model,root+'raw.jsonl']);
 const raw=fs.readFileSync(root+'raw.jsonl','utf8').trim().split('\n').map(l=>JSON.parse(l));
 audit(raw);
 const hook=r=>r.find(v=>v.HookCalls===1),frame=r=>r.find(v=>v.Frames);
 const mutations=[
 ['footer',r=>r.at(-1).complete++],['foreign_query',r=>r[1].Case='foreign'],['duplicate',r=>r[2]=structuredClone(r[1])],
 ['journal',r=>r[1].JournalStored=false],['frame',r=>frame(r).Frames[Object.keys(frame(r).Frames)[0]][5]+='tamper'],
 ['law',r=>r[1].JournalOrder[0].law.useful+=.01],['packet',r=>r[1].Packed[0].score+=.01],
 ['lexical',r=>hook(r).Lexical[0]+=.01],['formula',r=>hook(r).HookScores[0]+=.01],
 ['sort',r=>{const v=hook(r);[v.Order[0],v.Order[1]]=[v.Order[1],v.Order[0]]}],
 ['missing',r=>r.splice(1,1)],['weight',r=>hook(r).Weight+=.01],['nomination',r=>r[1].JournalOrder[0].id='foreign'],
 ['source',r=>r[0].files[Object.keys(r[0].files)[0]]='bad'],['omitted_source',r=>delete r[0].files[Object.keys(r[0].files)[0]]],
 ['import_mode',r=>r[1].ImportMode='fictional'],['retained_count',r=>r[1].Order.pop()],
 ['warm_recall',r=>r[1].MemoAfter.misses++],['model_digest',r=>r[0].digest='wrong'],
 ['extra_row',r=>r.splice(1,0,{Type:'unaccounted'})],
 ];
 const controls=[];for(const[name,mutate]of mutations){const r=structuredClone(raw);mutate(r);assert.notDeepEqual(r,raw);assert.throws(()=>audit(r),undefined,'accepted '+name);controls.push(name)}
 save('controls.json',{time:new Date().toISOString(),normal_pass:true,rejected_controls:controls});
 run('audit','node',[root+'audit.mjs',root+'raw.jsonl',root+'results.json']);
 unchanged();
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|jsonl|log)$/.test(p)))artifacts[p]=hash(fs.readFileSync(root+p));
 save('completed.json',{time:new Date().toISOString(),checks,artifacts,controls:controls.length,source_unchanged:true,prior_sources_unchanged:true,no_refit:true,goal:'ACTIVE',all_seven_whole_goals:'OPEN',whole_goal_complete:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,whole_goal_complete:false});throw e}
