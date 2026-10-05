// Serialized post-V66 parameter diagnostic. Existing sources remain frozen.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import {execFileSync,spawn} from 'node:child_process';
const root='research/joint-rate-v67-diagnostic';assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const prior='research/joint-v66-diagnostic',p=JSON.parse(fs.readFileSync(prior+'/freeze.json')),complete=JSON.parse(fs.readFileSync(prior+'/completed.json'));assert(complete.allJobsTerminal&&complete.checks.every(x=>x.exitCode===0));
for(const[n,h]of Object.entries({...p.files,...p.protectedFiles,...p.data}))assert.equal(await hash(n),h,n);
for(const[n,h]of Object.entries(complete.artifacts))assert.equal(await hash(prior+'/'+n),h,n);
const packages=['./internal/researchdispersion','./internal/researchjointsequence','./internal/researchjointsequenceref'];
const raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:128*1024*1024}),closure=JSON.parse(raw),compiler=new Set(),generated={};
for(const row of closure){if(!row.Dir?.startsWith(process.cwd()+'/'))continue;for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const x=path.resolve(row.Dir,n);if(x.startsWith(process.cwd()+'/'))compiler.add(path.relative(process.cwd(),x));else{assert(row.ImportPath.endsWith('.test'));generated[x]=await hash(x)}}}
const extras=['go.mod','go.sum','research/joint-rate-v67-run.mjs','research/joint-rate-v67-readback.mjs','research/joint-rate-v67-gen.mjs','research/joint-rate-v67-generation.json','docs/experiments/mmm-joint-rate-v67-protocol.md'];
const files=Object.fromEntries(await Promise.all([...new Set([...compiler,...extras])].sort().map(async n=>[n,await hash(n)])));
const data={};for(const n of ['freeze.json','completed.json','readback.json','diagnostic.jsonl'])data[prior+'/'+n]=await hash(prior+'/'+n);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(p.protectedFiles).sort());
fs.mkdirSync(root,{mode:0o700});const save=(n,x)=>fs.writeFileSync(root+'/'+n,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});save('compiler-closure.json',closure);
for(const[n,h]of Object.entries(files)){const out=root+'/source/'+n;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(n,out,fs.constants.COPYFILE_EXCL);fs.chmodSync(out,0o600);assert.equal(await hash(out),h)}
const generatedCompilerCopies={};for(const[n,h]of Object.entries(generated)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(n,root+'/'+out,fs.constants.COPYFILE_EXCL);fs.chmodSync(root+'/'+out,0o600);generatedCompilerCopies[out]=h}
save('freeze.json',{time:new Date().toISOString(),files,compilerFiles:[...compiler].sort(),compilerClosureSHA256:crypto.createHash('sha256').update(raw).digest('hex'),generatedCompilerCopies,protectedFiles:p.protectedFiles,data,toolchain:JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOFLAGS','GOTOOLCHAIN'],{encoding:'utf8'})),host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},hazards:[0,1/16,1],policies:['no_pair','uncertainty'],scope:'complete rate-transition boundary ablation; no new model, fitting, adoption or confirmation',goals:Array(7).fill('OPEN'),productionChanged:false});
async function unchanged(){for(const[n,h]of Object.entries({...files,...p.protectedFiles,...data}))assert.equal(await hash(n),h,n)}
const checks=[];
async function run(name,exe,args,extra={}){
 await unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600),env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];Object.assign(env,{EVENTFRAME_JOINT_V66_FREEZE:path.resolve(root+'/freeze.json')},extra);console.log('START',name,start);let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn(exe,args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),60*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,exe,args,extra,start,end:new Date().toISOString(),exitCode:code,wallMS:performance.now()-begin,logSHA256:await hash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();assert.equal(code,0,name);
}
try{
 await run('race','go',['test','-race','./internal/researchjointsequence','./internal/researchdispersion','-run','^Test(AtomicLifecycleAndEpochs|FaultsCapsAndPriorSupport|ExhaustiveTwoStageJoint|IndependentDelayedJointAndTower|FutureForkNonvacuous|JointV66FutureAndCorruptions)$','-count=1','-v','-timeout=20m']);
 await run('vet','go',['vet',...packages]);
 await run('experiment','go',['test','./internal/researchdispersion','-run','^TestJointRateV67Experiment$','-count=1','-v','-timeout=30m'],{EVENTFRAME_JOINT_RATE_V67_OUT:path.resolve(root+'/diagnostic.jsonl')});
 await run('audit','go',['test','./internal/researchdispersion','-run','^TestJointRateV67Audit$','-count=1','-v','-timeout=50m'],{EVENTFRAME_JOINT_RATE_V67_AUDIT:path.resolve(root+'/diagnostic.jsonl')});
 await run('readback','node',['research/joint-rate-v67-readback.mjs']);
 const artifacts={};for(const n of fs.readdirSync(root).filter(n=>/\.(json|jsonl|log)$/.test(n)))artifacts[n]=await hash(root+'/'+n);
 save('completed.json',{checks,artifacts,sourceUnchanged:true,allJobsTerminal:true,goal:'ACTIVE',goals:Array(7).fill('OPEN'),productionChanged:false});
}catch(e){save('failure.json',{error:e.message,checks,allJobsTerminal:true,goal:'ACTIVE',goals:Array(7).fill('OPEN')});throw e}
