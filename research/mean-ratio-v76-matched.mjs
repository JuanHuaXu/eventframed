// Freeze matched same-row-position public benchmarks before dispatch.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawn} from 'node:child_process';
const root='research/mean-ratio-v76-matched';assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const f74=JSON.parse(fs.readFileSync('research/mean-anchor-v75-diagnostic/freeze.json')),f75=JSON.parse(fs.readFileSync('research/mean-ratio-v76-initial/freeze.json')),c75=JSON.parse(fs.readFileSync('research/mean-ratio-v76-initial/completed.json'));
assert(c75.allJobsTerminal&&c75.allChecksPass);for(const[p,h]of Object.entries({...f74.files,...f74.data,...f74.protectedFiles,...f75.files}))assert.equal(await hash(p),h,p);
const files={},closure=JSON.parse(execFileSync('go',['run','./cmd/research-go-list-closure','./internal/researchmeanratioperfcheck'],{maxBuffer:64*1024*1024}));
for(const row of closure)for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);files[p]=await hash(p)}
for(const p of ['go.mod','go.sum','research/mean-ratio-v76-matched.mjs'])files[path.resolve(p)]=await hash(p);
fs.mkdirSync(root,{mode:0o700});const save=(n,x)=>fs.writeFileSync(root+'/'+n,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(out),h,p)}
save('freeze.json',{time:new Date().toISOString(),files,protectedFiles:f74.protectedFiles,closure,v74CompletedSHA256:await hash('research/mean-anchor-v75-diagnostic/completed.json'),v75CompletedSHA256:await hash('research/mean-ratio-v76-initial/completed.json'),scope:'same150-member/16-round observations, same oldest/latest row, same joint model via public APIs; benchmark excludes setup, not full cohort or loaded serving'});
const checks=[];async function unchanged(){for(const[p,h]of Object.entries({...files,...f74.protectedFiles}))assert.equal(await hash(p),h,p)}
for(const[name,args]of [['unit',['test','-v','-run=^TestMatched','-count=1','./internal/researchmeanratioperfcheck']],['race',['test','-race','-v','-run=^TestMatched','-count=1','./internal/researchmeanratioperfcheck']],['vet',['vet','./internal/researchmeanratioperfcheck']],['benchmark',['test','-run=^$','-bench=BenchmarkMatched','-benchtime=100ms','-count=2','./internal/researchmeanratioperfcheck','-timeout=10m']]]){
 await unchanged();console.log('START',name,new Date().toISOString());const fd=fs.openSync(root+'/'+name+'.log','wx',0o600),begin=performance.now(),env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});c.on('error',reject);for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',x=>resolve(x??-1))})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const x={name,args,exitCode:code,seconds:(performance.now()-begin)/1000,logSHA256:await hash(root+'/'+name+'.log')};checks.push(x);save(name+'-command.json',x);await unchanged();if(code!==0)break;
}
const pass=checks.length===4&&checks.every(x=>x.exitCode===0);save('completed.json',{time:new Date().toISOString(),checks,allJobsTerminal:true,allChecksPass:pass,sourceUnchanged:true,wholeCohortEquivalenceEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});if(!pass)process.exitCode=1;
