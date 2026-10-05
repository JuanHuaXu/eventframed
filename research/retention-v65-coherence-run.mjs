// Small frozen counterfactual diagnostic, separate from performance collection.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync,spawnSync} from 'node:child_process';
const root='research/retention-v65-coherence';assert(!fs.existsSync(root));const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const raw=execFileSync('go',['run','./cmd/research-go-list-closure','./internal/researchretentioncoherence'],{maxBuffer:32*1024*1024});const compiled=new Set(['go.mod','go.sum','research/retention-v65-coherence-run.mjs']);
for(const x of JSON.parse(raw)){if(!x.Dir?.startsWith(process.cwd()+'/'))continue;for(const key of['GoFiles','CgoFiles','EmbedFiles'])for(const n of x[key]??[]){const p=path.resolve(x.Dir,n);if(p.startsWith(process.cwd()+'/'))compiled.add(path.relative(process.cwd(),p))}}
const files=Object.fromEntries([...compiled].sort().map(p=>[p,hash(fs.readFileSync(p))])),parent=JSON.parse(fs.readFileSync('research/checkpoint-2026-10-04-retention-v62/manifest.json'));
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(hash(fs.readFileSync(p)),h,p);
fs.mkdirSync(root,{mode:0o700});const save=(n,x)=>fs.writeFileSync(root+'/'+n,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('freeze.json',{time:new Date().toISOString(),files,compilerClosureSHA256:hash(raw),protectedFiles:parent.trackedHashes,scope:'20 public-API counterfactual branches, not quality/adoption/totalcost proof'});
for(const[p,h]of Object.entries(files)){const out=root+'/source/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL);assert.equal(hash(fs.readFileSync(out)),h)}
const env={...process.env};for(const k of Object.keys(env))if(k.startsWith('EVENTFRAME_'))delete env[k];env.EVENTFRAME_RETENTION_TOWER_OUT=path.resolve(root+'/tower.json');
const args=['test','./internal/researchretentioncoherence','-run','^TestOperationalMixtureTowerDiagnostic$','-count=1','-v'];const start=new Date().toISOString(),r=spawnSync('go',args,{env,encoding:'utf8',maxBuffer:8*1024*1024});
fs.writeFileSync(root+'/audit.log',r.stdout+r.stderr,{flag:'wx',mode:0o600});save('command.json',{args,start,end:new Date().toISOString(),exitCode:r.status,logSHA256:hash(r.stdout+r.stderr)});assert.equal(r.status,0);
for(const[p,h]of Object.entries({...files,...parent.trackedHashes}))assert.equal(hash(fs.readFileSync(p)),h,p);
const x=JSON.parse(fs.readFileSync(root+'/tower.json'));save('completed.json',{allJobsTerminal:true,exitCode:0,rows:x.Rows.length,maxAbsoluteDefect:x.MaxAbsoluteDefect,towerIdentityConfirmed:x.MaxAbsoluteDefect<=2e-10,goals:Array(7).fill('OPEN'),goal:'ACTIVE',productionChanged:false});console.log(r.stdout);console.log(JSON.stringify(x,null,2));
