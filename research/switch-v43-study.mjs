// Exclusive, frozen V43 study; no production or private-data access.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync} from 'node:child_process';

const stage=process.argv[2];assert(['diagnostic','normal'].includes(stage));
const root='research/switch-v43-study-'+stage;assert(!fs.existsSync(root),'exclusive study root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const original=JSON.parse(fs.readFileSync('research/moment-v41-normal/freeze.json')).files;
assert(JSON.parse(fs.readFileSync('research/moment-v41-normal/completed.json')).source_unchanged);
for(const[p,h]of Object.entries(original))assert.equal(hash(fs.readFileSync(p)),h,'immutable original learner/generator '+p);
const paths=[...new Set([...Object.keys(original),...fs.readdirSync('internal/researchswitch').filter(p=>p.endsWith('.go')).map(p=>'internal/researchswitch/'+p),'internal/researchswitchref/reference.go','internal/researchdispersion/switch_fixture_v43_test.go','docs/experiments/mmm-switch-v43-contract.md','docs/experiments/mmm-switch-v43-study-protocol.md','docs/experiments/mmm-switch-v43-span-rescue.md','research/switch-v43-study.mjs','research/switch-v43-readback.mjs'])].sort();
const files=Object.fromEntries(paths.map(p=>[p,hash(fs.readFileSync(p))]));
if(stage==='normal'){assert.deepEqual(files,JSON.parse(fs.readFileSync('research/switch-v43-study-diagnostic/freeze.json')).files);assert(JSON.parse(fs.readFileSync('research/switch-v43-study-diagnostic/completed.json')).sourceUnchanged)}
fs.mkdirSync(root,{mode:0o700});
function save(name,x){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
for(const[p,h]of Object.entries(files)){const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const b=fs.readFileSync(p);assert.equal(hash(b),h);const fd=fs.openSync(dest,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
save('freeze.json',{time:new Date().toISOString(),stage,files,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},qualityGateUnchanged:true,policyTuning:false,seedBases:{diagnostic:2026104307,design:2026104309,confirmation:2026104311}});
const checks=[];
function run(name,args,env={},timeout=3600000){unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync('go',args,{encoding:'utf8',timeout,env:{...process.env,...env},maxBuffer:16<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
const fd=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(fd,log);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
const row={name,args,env,start,end:new Date().toISOString(),code,wallMS:performance.now()-begin,logSHA256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name)}
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
try{
 if(stage==='diagnostic'){
  run('model-race',['test','-race','./internal/researchswitch','./internal/researchswitchref','-v','-count=1','-timeout=10m']);
  run('seed-race',['test','-race','./internal/researchdispersion','-run','^TestSwitchActualSeedSeparationV43$','-v','-count=1']);
  run('vet',['vet','./internal/researchswitch','./internal/researchswitchref','./internal/researchdispersion']);
 }
 for(const split of stage==='diagnostic'?['diagnostic']:['design','confirmation']){
  const fixture=path.resolve(root+'/'+split+'-fixture.jsonl'),raw=path.resolve(root+'/'+split+'.jsonl');
  run(split+'-fixture',['test','./internal/researchdispersion','-run','^TestSwitchFixtureV43$','-v','-count=1','-timeout=60m'],{EVENTFRAME_SWITCH_V43_FIXTURE:fixture,EVENTFRAME_SWITCH_V43_SPLIT:split});
  run(split,['test','./internal/researchswitch','-run','^TestSwitchStudyV43$','-v','-count=1','-timeout=60m'],{EVENTFRAME_SWITCH_V43_FIXTURE:fixture,EVENTFRAME_SWITCH_V43_OUT:raw});
  run(split+'-audit',['test','./internal/researchswitch','-run','^TestSwitchStudyAuditV43$','-v','-count=1','-timeout=60m'],{EVENTFRAME_SWITCH_V43_FIXTURE:fixture,EVENTFRAME_SWITCH_V43_AUDIT:raw});
 }
 unchanged();const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(jsonl|json|log)$/.test(p)))artifacts[p]=await fileHash(root+'/'+p);
 save('completed.json',{time:new Date().toISOString(),stage,checks,artifacts,sourceUnchanged:true,qualityPassUnproven:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),checks,error:e.message,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});throw e}
