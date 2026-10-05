// Isolated technical tests, not a loaded serving-latency experiment.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';

const label=process.argv[2];assert(/^[a-z0-9-]+$/.test(label??''),'explicit new artifact label required');
const design=JSON.parse(fs.readFileSync('research/moment-v41-normal/design-command.json'));
assert.equal(design.code,0,'do not contend with V41 timed design collection');
const live=execFileSync('pgrep',['-fl','TestMomentStudyAuditV41'],{encoding:'utf8'});
assert(live.includes('go test'),'V41 independent offline audit must be live, not the confirmation collector');
const root='research/eager-core-v42-'+label;fs.mkdirSync(root,{recursive:true});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function paths(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?paths(dir+'/'+e.name):e.name.endsWith('.go')?[dir+'/'+e.name]:[])}
const inventory=()=>[...paths('internal'),...paths('cmd'),'go.mod','go.sum','research/eager-core-v42-preflight.mjs','docs/experiments/mmm-eager-core-v42-preflight.md'].sort();
const files=Object.fromEntries(inventory().map(p=>[p,hash(fs.readFileSync(p))]));
function save(name,value){const f=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(f,JSON.stringify(value,null,2)+'\n');fs.fsyncSync(f)}finally{fs.closeSync(f)}}
function unchanged(){assert.deepEqual(inventory(),Object.keys(files));for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
save('freeze.json',{time:new Date().toISOString(),files,liveOfflineAudit:live,loadedStudy:false,productionEdits:false});
const checks=[];
function run(name,args,env={}){
 unchanged();const begin=performance.now(),start=new Date().toISOString();let code=0,log='';
 try{log=execFileSync('go',args,{env:{...process.env,...env},encoding:'utf8',timeout:600000,maxBuffer:8<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
 const f=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(f,log);fs.fsyncSync(f)}finally{fs.closeSync(f)}
 const row={name,args,env,start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,logSHA256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name+' failed');
}
try{
 run('race',['test','-race','./internal/store/libravdbstore','-run','^TestResearchEagerCore.*V42$','-v','-count=1','-timeout=5m'],{EVENTFRAME_RUN_EAGER_CORE_V42:'1'});
 run('vet',['vet','./internal/store/libravdbstore']);
 unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,loadedStudy:false,qualityOrLatencyAdoption:false,goal:'ACTIVE',allSevenWholeGoals:'OPEN'});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,loadedStudy:false,qualityOrLatencyAdoption:false});throw e}
