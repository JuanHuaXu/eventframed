// Freeze and measure the full proposed mean-grid storage layout, not a learner.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawnSync} from 'node:child_process';
const root='research/mean-layout-v73-initial',parent='research/checkpoint-2026-10-04-dynvariance-v71/manifest.json';
assert(!fs.existsSync(root));const hash=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
assert.equal(hash(parent),'9d6cb6dcb7fc8281b407ebc9cb5035c27a2e8b29bcf9230013517e2da9a2b297');
const protectedFiles=JSON.parse(fs.readFileSync(parent)).trackedHashes;
for(const[p,h]of Object.entries(protectedFiles))assert.equal(hash(p),h,p);
const raw=execFileSync('go',['run','./cmd/research-go-list-closure','./internal/researchmeanlayout'],{maxBuffer:64*1024*1024}),closure=JSON.parse(raw),files={};
for(const row of closure)for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);assert(fs.existsSync(p),p);files[p]=hash(p)}
for(const p of ['go.mod','go.sum','research/mean-layout-v73-run.mjs','research/mean-joint-v73-preflight.mjs','research/mean-joint-v73-preflight/results.json'])files[path.resolve(p)]=hash(p);
fs.mkdirSync(root,{mode:0o700});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(hash(out),h,p)}
const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('freeze.json',{time:new Date().toISOString(),files,protectedFiles,closure,parent,parentSHA256:hash(parent),goVersion:execFileSync('go',['version'],{encoding:'utf8'}).trim(),scope:'all27 means, three families, three noise states,64-trial full planned journal; initial priors/allocation only, NOT posteriors, delayed replay or stream quality'});
const unchanged=()=>{for(const[p,h]of Object.entries({...files,...protectedFiles}))assert.equal(hash(p),h,p)};
const jobs=[['unit',['test','./internal/researchmeanlayout','-v','-count=1']],['race',['test','-race','./internal/researchmeanlayout','-count=1']],['vet',['vet','./internal/researchmeanlayout']]],checks=[];
for(const[name,args]of jobs){
 unchanged();const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];if(name==='unit')env.EVENTFRAME_MEAN_LAYOUT_V73_ALLOCATION=path.resolve(root+'/allocation.json');
 const begin=performance.now(),p=spawnSync('go',args,{env,encoding:'utf8',maxBuffer:64*1024*1024}),log=(p.stdout??'')+(p.stderr??'');fs.writeFileSync(root+'/'+name+'.log',log,{flag:'wx',mode:0o600});
 checks.push({name,args,exitCode:p.status,signal:p.signal,error:p.error?.message??null,seconds:(performance.now()-begin)/1000,logSHA256:hash(root+'/'+name+'.log')});console.log(JSON.stringify(checks.at(-1)));unchanged();
}
const passed=checks.every(x=>x.exitCode===0);save('completed.json',{checks,allJobsTerminal:checks.every(x=>x.exitCode!==null),allChecksPass:passed,sourceUnchanged:true,allocationSHA256:fs.existsSync(root+'/allocation.json')?hash(root+'/allocation.json'):null,completeLearnerImplemented:false,scientificQualityRescueEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});if(!passed)process.exitCode=1;
