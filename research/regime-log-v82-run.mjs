// Freeze an isolated attempt before its first test; preserve every failure.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawn} from 'node:child_process';
const stage=process.argv[2];assert(/^[a-z0-9-]+$/.test(stage));const root=`research/regime-log-v82-${stage}`;assert(!fs.existsSync(root));
const json=p=>JSON.parse(fs.readFileSync(p));async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const parentPath='research/checkpoint-2026-10-05-regime-protected-v81/manifest.json';assert.equal(await hash(parentPath),'5d2a36f6311a98ab1d4bb1b3199247205b73e0e9e500205165b7e321b30a5dbf');const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
assert.deepEqual(execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort(),Object.keys(parent.trackedHashes).sort());
for(const run of ['research/regime-incremental-v80-initial','research/regime-protected-v81-initial','research/regime-protected-v81-screen-initial'])for(const[p,h]of Object.entries(json(run+'/freeze.json').files))assert.equal(await hash(p),h,p);
const closure=JSON.parse(execFileSync('go',['run','./cmd/research-go-list-closure','./internal/researchregimeprotected','./internal/researchregimelog'],{maxBuffer:64*1024*1024})),files={};
for(const row of closure)for(const k of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[k]??[]){const p=path.resolve(row.Dir,n);files[p]=await hash(p)}
for(const p of ['go.mod','go.sum','research/regime-log-v82-run.mjs','research/regime-log-v82-gen.mjs','research/regime-log-v82-generation.json','research/regime-log-v82-preparation-repair.md','docs/experiments/mmm-regime-log-v82-protocol.md'])files[path.resolve(p)]=await hash(p);
fs.mkdirSync(root,{mode:0o700});const save=(p,v)=>fs.writeFileSync(root+'/'+p,JSON.stringify(v,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(out),h,p)}
save('freeze.json',{time:new Date().toISOString(),stage,files,closure,protectedFiles:parent.trackedHashes,parentPath,parentSHA256:await hash(parentPath),toolchain:execFileSync('go',['version'],{encoding:'utf8'}).trim(),scope:'persistent log-state/log-branch numerical contract; not quality or full-seven-goal validation'});
async function unchanged(){for(const[p,h]of Object.entries({...files,...parent.trackedHashes}))assert.equal(await hash(p),h,p)}
const checks=[];
for(const[name,args]of [
 ['unit',['test','-v','-count=1','./internal/researchregimelog','-timeout=20m']],
 ['race',['test','-race','-v','-count=1','./internal/researchregimelog','-timeout=30m']],
 ['vet',['vet','./internal/researchregimelog']],
 ['benchmark',['test','-run=^$','-bench=Benchmark','-benchtime=100ms','-count=2','./internal/researchregimeprotected','./internal/researchregimelog','-timeout=20m']],
]){
 await unchanged();const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];if(name==='unit')env.EVENTFRAME_REGIME_V82_REPORT=path.resolve(root+'/audit.json');
 const fd=fs.openSync(root+'/'+name+'.log','wx',0o600),begin=performance.now();console.log('START',name,new Date().toISOString());let code;
 try{code=await new Promise((resolve,reject)=>{const c=spawn('go',args,{env,stdio:['ignore','pipe','pipe']});c.on('error',reject);for(const s of [c.stdout,c.stderr])s.on('data',b=>{fs.writeSync(fd,b);process.stdout.write(b)});c.on('close',x=>resolve(x??-1))})}finally{fs.fsyncSync(fd);fs.closeSync(fd)}
 const row={name,args,exitCode:code,seconds:(performance.now()-begin)/1000,logSHA256:await hash(root+'/'+name+'.log')};checks.push(row);save(name+'-command.json',row);await unchanged();console.log(JSON.stringify(row));if(code!==0)break;
}
const pass=checks.length===4&&checks.every(x=>x.exitCode===0);
save('completed.json',{time:new Date().toISOString(),stage,checks,allJobsTerminal:true,allChecksPass:pass,sourceUnchanged:true,scientificQualityRescueEstablished:false,scalableApproximationCertified:false,wholeCohortTested:false,loadedServingEstablished:false,goal:'ACTIVE',goals:Array(7).fill('OPEN')});if(!pass)process.exitCode=1;
