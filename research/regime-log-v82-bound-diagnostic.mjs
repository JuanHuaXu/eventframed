import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/regime-log-v82-bound-diagnostic';assert(!fs.existsSync(root));const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const closure=JSON.parse(execFileSync('go',['run','./cmd/research-go-list-closure','./cmd/research-regime-log-bound-diagnostic'],{maxBuffer:64*1024*1024})),files={};
for(const row of closure)for(const k of ['GoFiles','CgoFiles','EmbedFiles'])for(const n of row[k]??[]){const p=path.resolve(row.Dir,n);files[p]=hash(fs.readFileSync(p))}
files[path.resolve('research/regime-log-v82-bound-diagnostic.mjs')]=hash(fs.readFileSync('research/regime-log-v82-bound-diagnostic.mjs'));
fs.mkdirSync(root,{mode:0o700});for(const[p,h]of Object.entries(files)){const o=root+'/source/'+p.replace(/^\//,'');fs.mkdirSync(path.dirname(o),{recursive:true,mode:0o700});fs.copyFileSync(p,o,fs.constants.COPYFILE_EXCL);assert.equal(hash(fs.readFileSync(o)),h)}
fs.writeFileSync(root+'/freeze.json',JSON.stringify({time:new Date().toISOString(),files,closure,scope:'diagnostic only; preserve failed prior unit run'},null,2)+'\n',{flag:'wx',mode:0o600});
let code=0,out;try{out=execFileSync('go',['run','./cmd/research-regime-log-bound-diagnostic'],{encoding:'utf8'})}catch(e){code=e.status??-1;out=String(e.stdout??'')+String(e.stderr??'')}
fs.writeFileSync(root+'/diagnostic.log',out,{flag:'wx',mode:0o600});for(const[p,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(p)),h,p);
fs.writeFileSync(root+'/completed.json',JSON.stringify({exitCode:code,allJobsTerminal:true,sourceUnchanged:true,logSHA256:hash(Buffer.from(out))},null,2)+'\n',{flag:'wx',mode:0o600});process.stdout.write(out);if(code!==0)process.exitCode=code;
