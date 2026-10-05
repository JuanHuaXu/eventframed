// Frozen public lead screen. No existing package is edited by this runner.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawn} from 'node:child_process';
const root='research/regime-protected-v81-screen-initial';assert(!fs.existsSync(root));
const json=p=>JSON.parse(fs.readFileSync(p));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const parentPath='research/checkpoint-2026-10-05-regime-incremental-v80/manifest.json';
assert.equal(await hash(parentPath),'d2ae1995bfaa0caff9a2ddee5864847606fd532ec1ae7cd27e16d5a50fcd965e');
const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
assert.deepEqual(execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort(),Object.keys(parent.trackedHashes).sort());
const initial='research/regime-protected-v81-initial',done=json(initial+'/completed.json');
assert(done.allJobsTerminal&&done.allChecksPass&&done.sourceUnchanged);
for(const[p,h]of Object.entries(json(initial+'/freeze.json').files))assert.equal(await hash(p),h,p);
for(const c of done.checks)assert.equal(await hash(initial+'/'+c.name+'.log'),c.logSHA256);
const closure=JSON.parse(execFileSync('go',['run','./cmd/research-go-list-closure','./cmd/research-regime-protected-screen'],{maxBuffer:64*1024*1024}));
const files={};
for(const row of closure)for(const k of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[k]??[]){const p=path.resolve(row.Dir,n);files[p]=await hash(p)}
for(const p of ['go.mod','go.sum','research/regime-protected-v81-screen-run.mjs','docs/experiments/mmm-regime-protected-v81-screen-protocol.md'])files[path.resolve(p)]=await hash(p);
fs.mkdirSync(root,{mode:0o700});
const save=(p,v)=>fs.writeFileSync(root+'/'+p,JSON.stringify(v,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(out),h,p)}
save('freeze.json',{time:new Date().toISOString(),files,closure,protectedFiles:parent.trackedHashes,parentPath,parentSHA256:await hash(parentPath),preflightCompletedSHA256:await hash(initial+'/completed.json'),toolchain:execFileSync('go',['version'],{encoding:'utf8'}).trim(),scope:'48 public DEVELOPMENT trajectories, not original full native controls, untouched confirmation, equal total cost or all-seven validation'});
async function unchanged(){for(const[p,h]of Object.entries({...files,...parent.trackedHashes}))assert.equal(await hash(p),h,p)}
const checks=[];
for(const[name,args]of [
 ['unit',['test','-v','-count=1','./cmd/research-regime-protected-screen','-timeout=10m']],
 ['race',['test','-race','-v','-count=1','./cmd/research-regime-protected-screen','-timeout=10m']],
 ['vet',['vet','./cmd/research-regime-protected-screen']],
 ['experiment',['run','./cmd/research-regime-protected-screen',root+'/data']],
]){
 await unchanged();const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];
 const fd=fs.openSync(root+'/'+name+'.log','wx',0o600),begin=performance.now();console.log('START',name,new Date().toISOString());
 let code;try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});c.on('error',reject);for(const s of [c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',x=>resolve(x??-1))})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,exitCode:code,seconds:(performance.now()-begin)/1000,logSHA256:await hash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();console.log(JSON.stringify(row));if(code!==0)break;
}
const dataHashes={};if(fs.existsSync(root+'/data'))for(const n of fs.readdirSync(root+'/data'))dataHashes[n]=await hash(root+'/data/'+n);
const pass=checks.length===4&&checks.every(c=>c.exitCode===0);
save('completed.json',{time:new Date().toISOString(),checks,dataHashes,allJobsTerminal:true,allChecksPass:pass,sourceUnchanged:true,scientificConfirmation:false,originalFullAdaptiveGatesTested:false,equalTotalCostComparison:false,loadedServing:false,goal:'ACTIVE',goals:Array(7).fill('OPEN')});
if(!pass)process.exitCode=1;
