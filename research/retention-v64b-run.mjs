// Exclusive serialized diagnostic; original and production paths unchanged.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import {execFileSync,spawn} from 'node:child_process';
const root='research/retention-v64b-diagnostic',parentPath='research/checkpoint-2026-10-04-retention-v62/manifest.json';assert(!fs.existsSync(root));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await fileHash(parentPath),'1ab64205ef92084495756178e1b215caa2528fc1ecbd9d6e80d3cff874ecae65');const parent=JSON.parse(fs.readFileSync(parentPath));
for(const[p,h]of Object.entries(parent.copies))assert.equal(await fileHash(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await fileHash(path.dirname(parentPath)+'/'+p),h,p);
const prior=JSON.parse(fs.readFileSync('research/tree-v60-diagnostic/freeze.json'));
for(const[p,h]of Object.entries(prior.files))assert.equal(await fileHash(p),h,p);
const integration=JSON.parse(fs.readFileSync('research/retention-v63-integration/completed.json'));assert(integration.checks.every(x=>x.exitCode===0));assert(integration.allJobsTerminal);
const packages=['./internal/researchdispersion','./internal/researchwindowjournal','./internal/researchwindowjournalref'];
const raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:128*1024*1024}),closure=JSON.parse(raw),compiler=new Set(),generated={};
for(const row of closure){if(!row.Dir?.startsWith(process.cwd()+'/'))continue;for(const key of['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const p=path.resolve(row.Dir,n);if(p.startsWith(process.cwd()+'/'))compiler.add(path.relative(process.cwd(),p));else{assert(row.ImportPath.endsWith('.test'));generated[p]=await fileHash(p)}}}
const extras=['go.mod','go.sum','research/retention-v64b-run.mjs','research/retention-v64b-readback.mjs','research/retention-v64-gen.mjs','research/retention-v64-generation.json','research/retention-v64-preflight-repair.mjs','research/retention-v64-preflight-failure.json','docs/experiments/mmm-retention-v64-protocol.md'];
const files=Object.fromEntries([...new Set([...compiler,...extras])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await fileHash(p),h,p);
const data={};for(const p of ['research/tree-v60-diagnostic/diagnostic.jsonl','research/tree-v60-diagnostic/freeze.json','research/retention-v63-integration/completed.json','research/retention-v63-integration/allocation.json'])data[p]=await fileHash(p);
fs.mkdirSync(root,{mode:0o700});const save=(name,x)=>fs.writeFileSync(root+'/'+name,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('compiler-closure.json',closure);for(const[p,h]of Object.entries(files)){const b=fs.readFileSync(p);assert.equal(hash(b),h);const out=root+'/source/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.writeFileSync(out,b,{flag:'wx',mode:0o600})}
const generatedCopies={};for(const[p,h]of Object.entries(generated)){const name='generated/'+h+'-test-main.go',out=root+'/'+name;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);fs.chmodSync(out,0o600);generatedCopies[name]=h}
save('freeze.json',{time:new Date().toISOString(),files,compilerFiles:[...compiler].sort(),compilerClosureSHA256:hash(raw),generatedCompilerCopies:generatedCopies,protectedFiles:parent.trackedHashes,data,toolchain:JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOFLAGS','GOTOOLCHAIN'],{encoding:'utf8'})),host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},node:process.version,scope:'all40 consumed worlds x3delays x5arms; no fresh confirmation or goal7 claim',goals:Array(7).fill('OPEN'),parent:{path:parentPath,sha256:await fileHash(parentPath)}});
const checks=[];async function unchanged(){for(const[p,h]of Object.entries({...files,...parent.trackedHashes,...data}))assert.equal(await fileHash(p),h,p)}
async function run(name,args,extra={}){
 await unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600);let code;
 const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];Object.assign(env,{EVENTFRAME_RETENTION_V64_FREEZE:path.resolve(root+'/freeze.json')},extra);console.log('START',name,start);
 try {code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),60*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,extra,start,end:new Date().toISOString(),exitCode:code,wallMS:performance.now()-begin,logSHA256:await fileHash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();assert.equal(code,0,name);
}
try {
 await run('race-fixture',['test','-race','./internal/researchdispersion','-run','^TestRetentionV64FutureAndCorruptions$','-count=1','-v','-timeout=20m']);
 await run('vet',['vet',...packages]);
 await run('experiment',['test','./internal/researchdispersion','-run','^TestRetentionV64Experiment$','-count=1','-v','-timeout=30m'],{EVENTFRAME_RETENTION_V64_OUT:path.resolve(root+'/diagnostic.jsonl')});
 await run('audit',['test','./internal/researchdispersion','-run','^TestRetentionV64Audit$','-count=1','-v','-timeout=50m'],{EVENTFRAME_RETENTION_V64_AUDIT:path.resolve(root+'/diagnostic.jsonl')});
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|jsonl|log)$/.test(p)))artifacts[p]=await fileHash(root+'/'+p);
 save('completed.json',{checks,artifacts,sourceUnchanged:true,allJobsTerminal:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{error:e.message,checks,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});throw e}
