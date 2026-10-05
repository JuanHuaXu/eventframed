// Technical correctness only. Offline-audit concurrency is not a cost result.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';

const label=process.argv[2];assert(/^[a-z0-9-]+$/.test(label??''));
const live=execFileSync('pgrep',['-fl','TestMomentStudyAuditV41'],{encoding:'utf8'});
assert(live.includes('go test'),'only while independent V41 audit is live; never timed collection');
const root='research/switch-v43-'+label;fs.mkdirSync(root,{recursive:true});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const prior=JSON.parse(fs.readFileSync('research/moment-v41-normal/freeze.json')).files;
for(const[p,h]of Object.entries(prior))assert.equal(hash(fs.readFileSync(p)),h,'V41 frozen source '+p);
const paths=[...new Set([...Object.keys(prior),...fs.readdirSync('internal/researchswitch').filter(p=>p.endsWith('.go')).map(p=>'internal/researchswitch/'+p),'internal/researchswitchref/reference.go','docs/experiments/mmm-switch-v43-contract.md','docs/experiments/mmm-switch-v43-cost-contract.md','research/switch-v43-preflight.mjs'])];
const files=Object.fromEntries(paths.map(p=>[p,hash(fs.readFileSync(p))]));
function save(name,x){const f=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(f,JSON.stringify(x,null,2)+'\n');fs.fsyncSync(f)}finally{fs.closeSync(f)}}
function unchanged(){for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p)}
save('freeze.json',{time:new Date().toISOString(),files,offlineAudit:live,qualityOrPerformanceResult:false});
const checks=[];
function run(name,args){unchanged();const start=new Date().toISOString(),begin=performance.now();let code=0,log='';try{log=execFileSync('go',args,{encoding:'utf8',timeout:300000,maxBuffer:8<<20})}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
const f=fs.openSync(root+'/'+name+'.log','wx',0o600);try{fs.writeFileSync(f,log);fs.fsyncSync(f)}finally{fs.closeSync(f)}
const row={name,args,start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,logSHA256:hash(log)};checks.push(row);save(name+'-command.json',row);console.log(log);unchanged();assert.equal(code,0,name)}
try{run('race',['test','-race','./internal/researchswitch','./internal/researchswitchref','-v','-count=1']);run('vet',['vet','./internal/researchswitch','./internal/researchswitchref']);unchanged();save('completed.json',{time:new Date().toISOString(),checks,sourceUnchanged:true,qualityOrPerformanceResult:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE'})}
catch(e){save('failure.json',{time:new Date().toISOString(),checks,error:e.message,qualityOrPerformanceResult:false});throw e}
