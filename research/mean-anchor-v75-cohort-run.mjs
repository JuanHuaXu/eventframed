// Complete native-policy equivalent-computation screen; no production access.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import {execFileSync,spawn} from 'node:child_process';
const root='research/mean-anchor-v75-diagnostic',parentPath='research/checkpoint-2026-10-05-mean-joint-v74-anchor-v75/manifest.json';assert(!fs.existsSync(root));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await digest(parentPath),'f4f547b44b8f2ac25bcd58155d0882c410cbb37dcc63a1c7806cc1afa18e27b4');const parent=JSON.parse(fs.readFileSync(parentPath));
for(const[p,h]of Object.entries(parent.copies))assert.equal(await digest(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await digest(path.dirname(parentPath)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await digest(p),h,p);
for(const d of ['research/mean-anchor-v75-initial','research/mean-anchor-v75-matched']){
 const f=JSON.parse(fs.readFileSync(d+'/freeze.json')),c=JSON.parse(fs.readFileSync(d+'/completed.json'));assert(c.allJobsTerminal&&c.allChecksPass&&c.sourceUnchanged);for(const[p,h]of Object.entries(f.files))assert.equal(await digest(p),h,p);
}
const packages=['./internal/researchdispersion','./internal/researchmeananchor','./internal/researchmeananchorcheck','./internal/researchmeanjointref'],raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:128*1024*1024}),closure=JSON.parse(raw),compiler=new Set(),generated={};
for(const row of closure){if(!row.Dir?.startsWith(process.cwd()+'/'))continue;for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);if(p.startsWith(process.cwd()+'/'))compiler.add(path.relative(process.cwd(),p));else{assert(row.ImportPath.endsWith('.test'));generated[p]=await digest(p)}}}
const extras=['go.mod','go.sum','research/mean-anchor-v75-cohort-run.mjs','research/mean-anchor-v75-cohort-gen.mjs','research/mean-anchor-v75-cohort-generation.json','research/mean-anchor-v75-cohort-tools-gen.mjs','research/mean-anchor-v75-cohort-tools-generation.json','research/mean-anchor-v75-cohort-readback.mjs','research/mean-anchor-v75-cohort-cost-audit.mjs','research/mean-anchor-v75-comparison.mjs','docs/experiments/mmm-mean-anchor-v75-cohort-protocol.md'];
const files=Object.fromEntries([...new Set([...compiler,...extras])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const data={};for(const p of ['research/mean-joint-v74-diagnostic/diagnostic.jsonl','research/mean-joint-v74-diagnostic/freeze.json','research/mean-joint-v74-diagnostic/readback.json','research/mean-joint-v74-diagnostic/completed.json','research/mean-anchor-v75-initial/completed.json','research/mean-anchor-v75-matched/completed.json'])data[p]=await digest(p);
fs.mkdirSync(root,{mode:0o700});const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('compiler-closure.json',closure);for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p,b=fs.readFileSync(p);assert.equal(hash(b),h);fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.writeFileSync(out,b,{flag:'wx',mode:0o600})}
const generatedCompilerCopies={};for(const[p,h]of Object.entries(generated)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(p,root+'/'+out,fs.constants.COPYFILE_EXCL);generatedCompilerCopies[out]=h}
save('freeze.json',{time:new Date().toISOString(),files,compilerFiles:[...compiler].sort(),generatedCompilerCopies,protectedFiles:parent.trackedHashes,data,toolchain:JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOFLAGS','GOTOOLCHAIN'],{encoding:'utf8'})),node:process.version,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},parent:{path:parentPath,sha256:await digest(parentPath),copiesVerified:Object.keys(parent.copies).length},scope:'complete native-policy30-arm consumed screen; all prior model tests and matched fixture verified; compare every old control exactly and mean arm at frozen scalar/discrete rules; no fresh all-arm dense reference replay',goals:Array(7).fill('OPEN'),productionChanged:false});
async function unchanged(){for(const[p,h]of Object.entries({...files,...parent.trackedHashes,...data}))assert.equal(await digest(p),h,p)}
const checks=[];
async function run(name,exe,args,extra={}){
 await unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600),env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];Object.assign(env,{EVENTFRAME_MEANANCHOR_V75_FREEZE:path.resolve(root+'/freeze.json')},extra);console.log('START',name,start);let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn(exe,args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),90*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,exe,args,extra,start,end:new Date().toISOString(),exitCode:code,wallMS:performance.now()-begin,logSHA256:await digest(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();assert.equal(code,0,name);
}
try{
 await run('vet','go',['vet',...packages]);
 await run('allocation','go',['test','./internal/researchmeananchor','-run','^TestConstructorAllocation$','-count=1','-v'],{EVENTFRAME_MEANANCHOR_V75_ALLOCATION:path.resolve(root+'/allocation.json')});
 await run('experiment','go',['test','./internal/researchdispersion','-run','^TestMeanAnchorV75Experiment$','-count=1','-v','-timeout=60m'],{EVENTFRAME_MEANANCHOR_V75_OUT:path.resolve(root+'/diagnostic.jsonl')});
 await run('readback','node',['research/mean-anchor-v75-cohort-readback.mjs']);
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|jsonl|log)$/.test(p)))artifacts[p]=await digest(root+'/'+p);
 save('completed.json',{checks,artifacts,sourceUnchanged:true,allJobsTerminal:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{error:e.message,checks,allJobsTerminal:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});throw e}
