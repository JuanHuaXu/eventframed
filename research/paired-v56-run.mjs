// Generated mechanically by paired-v56-run-gen.mjs.
// Serialized isolated research run. No network, private corpus or native store.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync,spawn} from 'node:child_process';
const root='research/paired-v56-diagnostic';
assert(!fs.existsSync(root),'exclusive output root');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.mkdirSync(root,{mode:0o700});
function save(name,value){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(value,null,2)+'\n');fs.fsyncSync(fd);}finally{fs.closeSync(fd);}}
const packages=['./internal/researchpairedmemo','./internal/researchpairedref','./internal/researchdispersion','./cmd/research-go-list-closure'];
// Standard Go decoder handles streamed package JSON. Test variants' GoFiles
// are the compiled inputs; dependency TestGoFiles are not automatically built.
const begin=new Date().toISOString();
const closureRaw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:128*1024*1024});
const closure=JSON.parse(closureRaw);
save('compiler-closure.json',closure);
const repo=process.cwd(),compiled=new Set();
for(const p of closure){if(!p.Dir?.startsWith(repo+'/'))continue;for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const name of p[key]??[]){const absolute=path.resolve(p.Dir,name);if(absolute.startsWith(repo+'/'))compiled.add(path.relative(repo,absolute));}}
const extras=['go.mod','go.sum','research/paired-v56-fixture-gen.mjs','research/paired-v56-fixture-generation.json','research/paired-v56-module-gen.mjs','research/paired-v56-module-generation.json','research/paired-v56-unit-preflight.mjs','research/paired-v56-equivalence.mjs','research/paired-v56-run.mjs','research/paired-v56-readback.mjs','research/paired-v56-preflight.md','docs/experiments/mmm-paired-v56-protocol.md','research/noise-v54-direction.md','research/noise-v54-identities.mjs','research/noise-v54-identities.json'];
const files=Object.fromEntries([...new Set([...compiled,...extras])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean);
const protectedFiles=Object.fromEntries(tracked.map(p=>[p,hash(fs.readFileSync(p))]));
const envJSON=JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOTOOLCHAIN','GOTOOLDIR','GOFLAGS','GOMOD'],{encoding:'utf8'}));
const tools={};for(const name of ['compile','link']){const p=path.join(envJSON.GOTOOLDIR,name);tools[p]=hash(fs.readFileSync(p));}
for(const[p,h]of Object.entries(files)){const b=fs.readFileSync(p);assert.equal(hash(b),h);const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,b,{flag:'wx',mode:0o600});}
save('freeze.json',{time:new Date().toISOString(),compilerClosureStarted:begin,compilerClosureSHA256:hash(closureRaw),compilerFiles:[...compiled].sort(),files,protectedFiles,toolchain:envJSON,tools,node:process.version,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},seedBases:{diagnostic:2026105407,design:2026105409,confirmation:2026105411},scope:'20regimes x2geometries x3delays x7arms, n1 diagnostic, no adoption',wholeGoals:'OPEN',productionChanged:false});
const freeze=path.resolve(root+'/freeze.json');
function unchanged(){for(const[p,h]of Object.entries({...files,...protectedFiles}))assert.equal(hash(fs.readFileSync(p)),h,'unchanged '+p);}
const checks=[];
async function run(name,args,extra={}){
 unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600);const h=crypto.createHash('sha256');let code;
 const env={...process.env};for(const key of Object.keys(env))if(key.startsWith('EVENTFRAME_'))delete env[key];Object.assign(env,{EVENTFRAME_PAIRED_V56_FREEZE:freeze},extra);
 console.log('START',name,start);
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),60*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);h.update(b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})});}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,env:extra,start,end:new Date().toISOString(),exitCode:code,wallMS:performance.now()-begin,logSHA256:h.digest('hex')};checks.push(row);save(name+'-command.json',row);unchanged();assert.equal(code,0,name);
}
try{
 await run('race-model',['test','-race','./internal/researchpairedmemo','./internal/researchpairedref','-count=1','-v','-timeout=15m']);
 await run('race-fixture',['test','-race','./internal/researchdispersion','-run','^TestPairedV56(FutureAndCorruptions|SeedSeparation|ControlMetricGuard)$','-count=1','-v','-timeout=20m']);
 await run('closure-helper',['test','./cmd/research-go-list-closure','-count=1','-v']);
 await run('vet',['vet',...packages]);
 await run('allocation',['test','./internal/researchdispersion','-run','^TestPairedV56Allocation$','-count=1','-v'],{EVENTFRAME_PAIRED_V56_ALLOCATION:path.resolve(root+'/allocation.json')});
 await run('experiment',['test','./internal/researchdispersion','-run','^TestPairedV56Experiment$','-count=1','-v','-timeout=30m'],{EVENTFRAME_PAIRED_V56_OUT:path.resolve(root+'/diagnostic.jsonl')});
 await run('audit',['test','./internal/researchdispersion','-run','^TestPairedV56Audit$','-count=1','-v','-timeout=45m'],{EVENTFRAME_PAIRED_V56_AUDIT:path.resolve(root+'/diagnostic.jsonl')});
 unchanged();const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|jsonl|log)$/.test(p))){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(root+'/'+p))h.update(b);artifacts[p]=h.digest('hex')};save('completed.json',{time:new Date().toISOString(),checks,artifacts,sourceUnchanged:true,wholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,wholeGoals:'OPEN',goal:'ACTIVE'});throw e}
