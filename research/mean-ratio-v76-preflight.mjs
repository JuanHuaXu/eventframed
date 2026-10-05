// Archive the exact isolated attempt before any compiler or test execution.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawn} from 'node:child_process';
const stage=process.argv[2];assert(/^[a-z0-9-]+$/.test(stage));const root=`research/mean-ratio-v76-${stage}`,parent='research/checkpoint-2026-10-05-mean-joint-v74-anchor-v75/manifest.json';assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await hash(parent),'f4f547b44b8f2ac25bcd58155d0882c410cbb37dcc63a1c7806cc1afa18e27b4');const checkpoint=JSON.parse(fs.readFileSync(parent));
for(const[p,h]of Object.entries(checkpoint.copies))assert.equal(await hash(path.dirname(parent)+'/saved/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(checkpoint.trackedHashes).sort());
for(const[p,h]of Object.entries(checkpoint.trackedHashes))assert.equal(await hash(p),h,p);
const v74='research/mean-anchor-v75-diagnostic',done=JSON.parse(fs.readFileSync(v74+'/completed.json'));
assert(done.allJobsTerminal&&done.checks.length===4&&done.checks.every(x=>x.exitCode===0),'no test concurrency with V75 timed screen');
const packages=['./internal/researchmeanratio','./internal/researchmeanratiocheck'],raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:64*1024*1024}),closure=JSON.parse(raw),files={};
for(const row of closure)for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);assert(fs.existsSync(p),p);files[p]=await hash(p)}
for(const p of ['go.mod','go.sum','research/mean-ratio-v76-preflight.mjs','research/mean-ratio-v76-gen.mjs','research/mean-ratio-v76-generation.json','research/mean-ratio-v76-tools-gen.mjs','research/mean-ratio-v76-tools-generation.json','docs/experiments/mmm-mean-ratio-v76-protocol.md'])files[path.resolve(p)]=await hash(p);
const v74freeze=JSON.parse(fs.readFileSync(v74+'/freeze.json'));for(const[p,h]of Object.entries({...v74freeze.files,...v74freeze.data,...v74freeze.protectedFiles}))assert.equal(await hash(p),h,p);
fs.mkdirSync(root,{mode:0o700});const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(out),h,p)}
save('freeze.json',{stage,time:new Date().toISOString(),files,protectedFiles:checkpoint.trackedHashes,parent,parentSHA256:await hash(parent),v74CompletedSHA256:await hash(v74+'/completed.json'),closure,toolchain:execFileSync('go',['version'],{encoding:'utf8'}).trim(),scope:'same full27-mean joint model, exact at-anchor conditional pair replacement with full older-pair replay; dense/full64/support/branch audits, not whole-study equivalence or a quality rescue'});
async function unchanged(){for(const[p,h]of Object.entries({...files,...checkpoint.trackedHashes}))assert.equal(await hash(p),h,p)}
const checks=[];
for(const[name,args]of [['unit',['test','-v','-count=1',...packages,'-timeout=20m']],['race',['test','-race','-v','-count=1',...packages,'-timeout=30m']],['vet',['vet',...packages]],['benchmark',['test','-run=^$','-bench=Benchmark','-benchtime=250ms','-count=2','./internal/researchmeanratio','-timeout=20m']]]){
 await unchanged();const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];if(name==='unit')env.EVENTFRAME_MEANRATIO_V76_ALLOCATION=path.resolve(root+'/allocation.json');
 const begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600);console.log('START',name,new Date().toISOString());let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});c.on('error',reject);for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',x=>resolve(x??-1))})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,exitCode:code,seconds:(performance.now()-begin)/1000,logSHA256:await hash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();console.log(JSON.stringify(row));if(code!==0)break;
}
const success=checks.length===4&&checks.every(x=>x.exitCode===0);save('completed.json',{stage,time:new Date().toISOString(),checks,allJobsTerminal:true,allChecksPass:success,sourceUnchanged:true,fullStudyEquivalenceEstablished:false,scientificQualityRescueEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});if(!success)process.exitCode=1;
