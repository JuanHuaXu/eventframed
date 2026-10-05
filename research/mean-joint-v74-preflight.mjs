// Immutable attempt archives before unit/race/vet; no empirical cohort access.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawn} from 'node:child_process';
const stage=process.argv[2];assert(/^[a-z0-9-]+$/.test(stage));const root=`research/mean-joint-v74-${stage}`,parent='research/checkpoint-2026-10-04-dynvarcache-v72-mean-v73/manifest.json';assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await hash(parent),'d4045ac945e8044155f54fb0871d8b2fffb30fa8a4806d431beaf028ea22cc54');const checkpoint=JSON.parse(fs.readFileSync(parent));
for(const[p,h]of Object.entries(checkpoint.copies))assert.equal(await hash(path.dirname(parent)+'/saved/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(checkpoint.trackedHashes).sort());
for(const[p,h]of Object.entries(checkpoint.trackedHashes))assert.equal(await hash(p),h,p);
const packages=['./internal/researchmeanjoint','./internal/researchmeanjointref'],raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:64*1024*1024}),closure=JSON.parse(raw),files={};
for(const row of closure)for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);assert(fs.existsSync(p),p);files[p]=await hash(p)}
for(const p of ['go.mod','go.sum','research/mean-joint-v74-preflight.mjs','research/mean-joint-v74-ref-gen.mjs','research/mean-joint-v74-ref-generation.json','research/mean-joint-v74-bench-gen.mjs','research/mean-joint-v74-bench-generation.json'])files[path.resolve(p)]=await hash(p);
fs.mkdirSync(root,{mode:0o700});const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(out),h,p)}
save('freeze.json',{stage,time:new Date().toISOString(),files,protectedFiles:checkpoint.trackedHashes,parent,parentSHA256:await hash(parent),closure,toolchain:execFileSync('go',['version'],{encoding:'utf8'}).trim(),scope:'full27 means with likelihood/predictive/observation integration; independent dense reference and public lifecycle; no stream quality or scientific adoption yet'});
async function unchanged(){for(const[p,h]of Object.entries({...files,...checkpoint.trackedHashes}))assert.equal(await hash(p),h,p)}
const checks=[];
for(const[name,args]of [['unit',['test','-v','-count=1',...packages,'-timeout=20m']],['race',['test','-race','-v','-count=1',...packages,'-timeout=30m']],['vet',['vet',...packages]]]){
 await unchanged();const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];if(name==='unit')env.EVENTFRAME_MEANJOINT_V74_ALLOCATION=path.resolve(root+'/allocation.json');
 const begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600);console.log('START',name,new Date().toISOString());let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});c.on('error',reject);for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',x=>resolve(x??-1))})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,exitCode:code,seconds:(performance.now()-begin)/1000,logSHA256:await hash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();console.log(JSON.stringify(row));if(code!==0)break;
}
const success=checks.length===3&&checks.every(x=>x.exitCode===0);save('completed.json',{stage,time:new Date().toISOString(),checks,allJobsTerminal:true,allChecksPass:success,sourceUnchanged:true,streamQualityExperimentRun:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});if(!success)process.exitCode=1;
