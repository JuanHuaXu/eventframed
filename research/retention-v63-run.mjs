// Exclusive frozen integration run. All prior models/data stay untouched.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import os from 'node:os';
import {execFileSync,spawn} from 'node:child_process';
const root='research/retention-v63-integration',parentPath='research/checkpoint-2026-10-04-retention-v62/manifest.json';assert(!fs.existsSync(root));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await fileHash(parentPath),'1ab64205ef92084495756178e1b215caa2528fc1ecbd9d6e80d3cff874ecae65');
const parent=JSON.parse(fs.readFileSync(parentPath));
for(const[p,h]of Object.entries(parent.copies))assert.equal(await fileHash(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await fileHash(path.dirname(parentPath)+'/'+p),h,p);
const packages=['./internal/researchwindowjournal','./internal/researchwindowjournalref'];
const raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:64*1024*1024});
const closure=JSON.parse(raw),compiler=new Set(),generated={};
for(const row of closure){if(!row.Dir?.startsWith(process.cwd()+'/'))continue;for(const key of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[key]??[]){const absolute=path.resolve(row.Dir,n);if(absolute.startsWith(process.cwd()+'/'))compiler.add(path.relative(process.cwd(),absolute));else{assert(row.ImportPath.endsWith('.test'));generated[absolute]=await fileHash(absolute)}}}
const extras=['go.mod','go.sum','research/retention-v63-run.mjs','research/retention-v63-gen.mjs','research/retention-v63-generation.json','research/retention-v63-journal-direction.md','docs/experiments/mmm-retention-v63-protocol.md'];
const files=Object.fromEntries([...new Set([...compiler,...extras])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await fileHash(p),h,p);
fs.mkdirSync(root,{mode:0o700});
const save=(name,x)=>fs.writeFileSync(root+'/'+name,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('compiler-closure.json',closure);
for(const[p,h]of Object.entries(files)){const b=fs.readFileSync(p);assert.equal(hash(b),h);const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,b,{flag:'wx',mode:0o600})}
const generatedCopies={};for(const[p,h]of Object.entries(generated)){const name='generated/'+h+'-test-main.go',dest=root+'/'+name;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(p,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);generatedCopies[name]=h}
save('freeze.json',{time:new Date().toISOString(),files,compilerFiles:[...compiler].sort(),compilerClosureSHA256:hash(raw),generatedCompilerFiles:generated,generatedCompilerCopies:generatedCopies,
 protectedFiles:parent.trackedHashes,toolchain:JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOFLAGS','GOTOOLCHAIN'],{encoding:'utf8'})),node:process.version,
 host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},parent:{path:parentPath,sha256:await fileHash(parentPath),copiesVerified:Object.keys(parent.copies).length},
 scope:'integration/component ownership,actual joint law,atomic updates,allocation and core fixed-nomination loop; NOT adaptive acquisition or full scientific experiment',goals:Array(7).fill('OPEN'),productionChanged:false});
async function unchanged(){for(const[p,h]of Object.entries({...files,...parent.trackedHashes}))assert.equal(await fileHash(p),h,p)}
const checks=[];
async function run(name,args,extra={}){
 await unchanged();const begin=performance.now(),start=new Date().toISOString(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600);
 const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];Object.assign(env,extra);let code;
 console.log('START',name,start);
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),15*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const stream of[c.stdout,c.stderr])stream.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,extra,start,end:new Date().toISOString(),wallMS:performance.now()-begin,exitCode:code,logSHA256:await fileHash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();assert.equal(code,0,name);
}
try{
 await run('race',['test','-race',...packages,'-count=1','-v','-timeout=10m']);
 await run('vet',['vet',...packages]);
 await run('allocation',['test','./internal/researchwindowjournal','-run','^TestBankAllocation$','-count=1','-v'],{EVENTFRAME_BANK_V63_ALLOCATION:path.resolve(root+'/allocation.json')});
 await run('benchmark',['test','./internal/researchwindowjournal','-run','^$','-bench','BenchmarkBank','-benchtime=100ms','-count=3']);
 const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|log)$/.test(p)))artifacts[p]=await fileHash(root+'/'+p);
 save('completed.json',{checks,artifacts,sourceUnchanged:true,goals:Array(7).fill('OPEN'),goal:'ACTIVE',allJobsTerminal:true,notFullScientificValidation:true,productionChanged:false});
}catch(e){save('failure.json',{error:e.message,checks,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});throw e}
