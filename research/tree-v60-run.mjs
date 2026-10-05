// Exclusive, serialized research run with immutable compiler-source copies.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import os from 'node:os';import {execFileSync,spawn} from 'node:child_process';
const root='research/tree-v60-diagnostic',parentPath='research/checkpoint-2026-10-04-tree-v59/manifest.json';
assert(!fs.existsSync(root));const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
assert.equal(await fileHash(parentPath),'6bff139a1658a268b7286dde9b27eef0ae5e8f416de7982857d1323bfb57780c');
const parent=JSON.parse(fs.readFileSync(parentPath));for(const[p,h]of Object.entries(parent.copies))assert.equal(await fileHash(path.dirname(parentPath)+'/saved/'+p),h,p);
fs.mkdirSync(root,{mode:0o700});
function save(name,value){const fd=fs.openSync(root+'/'+name,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(value,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}}
const packages=['./internal/researchlocal','./internal/researchlocalref','./internal/researchdispersion','./cmd/research-go-list-closure'];
const raw=execFileSync('go',['run','./cmd/research-go-list-closure',...packages],{maxBuffer:128*1024*1024}),closure=JSON.parse(raw);save('compiler-closure.json',closure);
const compiled=new Set(),repo=process.cwd();for(const p of closure){if(!p.Dir?.startsWith(repo+'/'))continue;for(const key of['GoFiles','CgoFiles','EmbedFiles'])for(const n of p[key]??[]){const absolute=path.resolve(p.Dir,n);if(absolute.startsWith(repo+'/'))compiled.add(path.relative(repo,absolute))}}
const extras=['go.mod','go.sum','research/tree-v60-run.mjs','research/tree-v60-readback.mjs','research/tree-v60-comparison.mjs','research/tree-v60-gen.mjs','research/tree-v60-generation.json','research/tree-v60-direction.md','research/tree-v60-identities.mjs','research/tree-v60-identities.json','research/tree-v60-beta-identities.mjs','research/tree-v60-beta-identities.json','research/tree-v60-beta-gen.mjs','research/tree-v60-beta-generation.json','research/tree-v60-preflight.md','docs/experiments/mmm-tree-v60-protocol.md'];
const files=Object.fromEntries([...new Set([...compiled,...extras])].sort().map(p=>[p,hash(fs.readFileSync(p))]));
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean),protectedFiles=Object.fromEntries(tracked.map(p=>[p,hash(fs.readFileSync(p))]));assert.deepEqual(protectedFiles,parent.trackedHashes);
const toolchain=JSON.parse(execFileSync('go',['env','-json','GOVERSION','GOOS','GOARCH','GOTOOLCHAIN','GOTOOLDIR','GOFLAGS','GOMOD'],{encoding:'utf8'}));const tools={};for(const n of['compile','link']){const p=path.join(toolchain.GOTOOLDIR,n);tools[p]=hash(fs.readFileSync(p))}
for(const[p,h]of Object.entries(files)){const b=fs.readFileSync(p);assert.equal(hash(b),h);const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,b,{flag:'wx',mode:0o600})}
save('freeze.json',{time:new Date().toISOString(),compilerClosureSHA256:hash(raw),compilerFiles:[...compiled].sort(),files,protectedFiles,tools,toolchain,node:process.version,host:{cpu:os.cpus()[0]?.model,logical:os.cpus().length,load:os.loadavg()},parent:{path:parentPath,sha256:await fileHash(parentPath),copiesVerified:Object.keys(parent.copies).length},seedBases:{diagnostic:2026105407,design:2026105409,confirmation:2026105411},model:{depth:7,window:600,identityWeight:.8,terminalIndependentPrior:.5,localGrid:21,localAtoms:22,localGridPrior:'beta-binomial(20,b,1-b)'},scope:'All40 consumed worlds x3delays x8arms; n1, no adoption',wholeGoals:'OPEN',productionChanged:false});
const freeze=path.resolve(root+'/freeze.json'),checks=[];
function unchanged(){for(const[p,h]of Object.entries({...files,...protectedFiles}))assert.equal(hash(fs.readFileSync(p)),h,p)}
async function run(name,args,extra={}) {unchanged();const start=new Date().toISOString(),begin=performance.now(),fd=fs.openSync(root+'/'+name+'.log','wx',0o600),h=crypto.createHash('sha256');let code;const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];Object.assign(env,{EVENTFRAME_PAIRED_V60_FREEZE:freeze},extra);console.log('START',name,start);
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});const timer=setTimeout(()=>c.kill('SIGTERM'),60*60*1000);c.on('error',e=>{clearTimeout(timer);reject(e)});for(const s of[c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);h.update(b);process.stdout.write(b)});c.on('close',code=>{clearTimeout(timer);resolve(code??-1)})})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,extra,start,end:new Date().toISOString(),exitCode:code,wallMS:performance.now()-begin,logSHA256:h.digest('hex')};checks.push(row);save(name+'-command.json',row);unchanged();assert.equal(code,0,name)}
try {
 await run('race-model',['test','-race','./internal/researchlocal','./internal/researchlocalref','-count=1','-v','-timeout=15m']);
 await run('race-fixture',['test','-race','./internal/researchdispersion','-run','^TestPairedV60(FutureAndCorruptions|SeedSeparation|ControlMetricGuard)$','-count=1','-v','-timeout=30m']);
 await run('closure-helper',['test','./cmd/research-go-list-closure','-count=1','-v']);
 await run('vet',['vet',...packages]);
 await run('microbench',['test','./internal/researchlocal','-run','^$','-bench','Benchmark','-benchtime=100ms','-count=3']);
 await run('allocation',['test','./internal/researchdispersion','-run','^TestPairedV60Allocation$','-count=1','-v'],{EVENTFRAME_PAIRED_V60_ALLOCATION:path.resolve(root+'/allocation.json')});
 await run('experiment',['test','./internal/researchdispersion','-run','^TestPairedV60Experiment$','-count=1','-v','-timeout=30m'],{EVENTFRAME_PAIRED_V60_OUT:path.resolve(root+'/diagnostic.jsonl')});
 await run('audit',['test','./internal/researchdispersion','-run','^TestPairedV60Audit$','-count=1','-v','-timeout=50m'],{EVENTFRAME_PAIRED_V60_AUDIT:path.resolve(root+'/diagnostic.jsonl')});
 unchanged();const artifacts={};for(const p of fs.readdirSync(root).filter(p=>/\.(json|jsonl|log)$/.test(p)))artifacts[p]=await fileHash(root+'/'+p);save('completed.json',{time:new Date().toISOString(),checks,artifacts,sourceUnchanged:true,wholeGoals:'OPEN',goal:'ACTIVE',productionChanged:false});
}catch(e){save('failure.json',{time:new Date().toISOString(),error:e.message,checks,wholeGoals:'OPEN',goal:'ACTIVE'});throw e}
