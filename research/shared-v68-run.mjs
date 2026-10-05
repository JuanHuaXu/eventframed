// Exclusive research study; no production, publication or sealed data access.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import {execFileSync,spawn} from 'node:child_process';
const root='research/shared-v68-diagnostic',parentPath='research/checkpoint-2026-10-04-joint-v66-v68/manifest.json';assert(!fs.existsSync(root));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await digest(parentPath),'aaa2f3b819d752aab0c96ca71d9d3a7916ea0ea36785f63060a4a7681bcf96e0');const parent=JSON.parse(fs.readFileSync(parentPath));
for(const[p,h]of Object.entries(parent.copies))assert.equal(await digest(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await digest(path.dirname(parentPath)+'/'+p),h,p);
const packages=['./internal/researchdispersion','./internal/researchsharedsequence','./internal/researchsharedsequenceref'];
const raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:128*1024*1024}),closure=JSON.parse(raw),compiler=new Set(),generated={};
for(const row of closure){if(!row.Dir?.startsWith(process.cwd()+'/'))continue;for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);if(p.startsWith(process.cwd()+'/'))compiler.add(path.relative(process.cwd(),p));else{assert(row.ImportPath.endsWith('.test'));generated[p]=await digest(p)}}}
const extras=['go.mod','go.sum','research/shared-v68-run.mjs','research/shared-v68-run-gen.mjs','research/shared-v68-run-generation.json','research/shared-v68-readback.mjs','research/shared-v68-fixture-gen.mjs','research/shared-v68-fixture-generation.json','research/shared-v68-boundary-preflight.mjs','research/shared-v68-boundary-preflight/command.json','research/shared-v68-boundary-preflight/command.log','docs/experiments/mmm-shared-v68-protocol.md'];
const files=Object.fromEntries([...new Set([...compiler,...extras])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await digest(p),h,p);
const data={};for(const p of ['research/tree-v60-diagnostic/diagnostic.jsonl','research/joint-v66-diagnostic/diagnostic.jsonl','research/joint-v66-diagnostic/freeze.json'])data[p]=await digest(p);
fs.mkdirSync(root,{mode:0o700});const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('compiler-closure.json',closure);for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p,b=fs.readFileSync(p);assert.equal(hash(b),h);fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.writeFileSync(out,b,{flag:'wx',mode:0o600})}
const generatedCompilerCopies={};for(const[p,h]of Object.entries(generated)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(p,root+'/'+out,fs.constants.COPYFILE_EXCL);fs.chmodSync(root+'/'+out,0o600);generatedCompilerCopies[out]=h}
save('freeze.json',{time:new Date().toISOString(),files,compilerFiles:[...compiler].sort(),compilerClosureSHA256:hash(raw),generatedCompilerCopies,protectedFiles:parent.trackedHashes,data,toolchain:JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOFLAGS','GOTOOLCHAIN'],{encoding:'utf8'})),node:process.version,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},parent:{path:parentPath,sha256:await digest(parentPath),copiesVerified:Object.keys(parent.copies).length},scope:'full fourteen-arm consumed-cohort shared-context diagnostic; no fresh confirmation',goals:Array(7).fill('OPEN'),productionChanged:false});
async function unchanged(){for(const[p,h]of Object.entries({...files,...parent.trackedHashes,...data}))assert.equal(await digest(p),h,p)}
const checks=[];
async function run(name,exe,args,extra={}){
 await unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600),env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];Object.assign(env,{EVENTFRAME_SHARED_V68_FREEZE:path.resolve(root+'/freeze.json')},extra);console.log('START',name,start);let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn(exe,args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),90*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,exe,args,extra,start,end:new Date().toISOString(),exitCode:code,wallMS:performance.now()-begin,logSHA256:await digest(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();assert.equal(code,0,name);
}
try{
 await run('race-model','go',['test','-race','./internal/researchsharedsequence','./internal/researchsharedsequenceref','-count=1','-v','-timeout=10m']);
 await run('race-fixture','go',['test','-race','./internal/researchdispersion','-run','^TestSharedV68FutureAndCorruptions$','-count=1','-v','-timeout=20m']);
 await run('vet','go',['vet',...packages]);
 await run('allocation','go',['test','./internal/researchsharedsequence','-run','^TestConstructorAllocation$','-count=1','-v'],{EVENTFRAME_SHARED_V68_ALLOCATION:path.resolve(root+'/allocation.json')});
 await run('benchmark','go',['test','./internal/researchsharedsequence','-run','^$','-bench','Benchmark','-benchtime=100ms','-count=3']);
 await run('experiment','go',['test','./internal/researchdispersion','-run','^TestSharedV68Experiment$','-count=1','-v','-timeout=60m'],{EVENTFRAME_SHARED_V68_OUT:path.resolve(root+'/diagnostic.jsonl')});
 await run('audit','go',['test','./internal/researchdispersion','-run','^TestSharedV68Audit$','-count=1','-v','-timeout=80m'],{EVENTFRAME_SHARED_V68_AUDIT:path.resolve(root+'/diagnostic.jsonl')});
 await run('readback','node',['research/shared-v68-readback.mjs']);
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|jsonl|log)$/.test(p)))artifacts[p]=await digest(root+'/'+p);
 save('completed.json',{checks,artifacts,sourceUnchanged:true,allJobsTerminal:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{error:e.message,checks,allJobsTerminal:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});throw e}
