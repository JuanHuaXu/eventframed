import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {spawnSync} from 'node:child_process';
const root='research/retention-v64c-audit',hash=b=>crypto.createHash('sha256').update(b).digest('hex'),source='research/retention-v64d-readback.mjs';
assert(JSON.parse(fs.readFileSync(root+'/completed.json')).allJobsTerminal);
const sourceSHA256=hash(fs.readFileSync(source));fs.copyFileSync(source,root+'/stream-readback-source.mjs',fs.constants.COPYFILE_EXCL);
const start=new Date().toISOString(),r=spawnSync(process.execPath,[source],{encoding:'utf8',maxBuffer:8*1024*1024});
fs.writeFileSync(root+'/readback.log',r.stdout+r.stderr,{flag:'wx',mode:0o600});fs.writeFileSync(root+'/readback-command.json',JSON.stringify({args:[source],start,end:new Date().toISOString(),exitCode:r.status,sourceSHA256,logSHA256:hash(r.stdout+r.stderr),scope:'postcollection streaming report; not serving/learning cost'},null,2)+'\n',{flag:'wx',mode:0o600});
assert.equal(r.status,0);assert.equal(hash(fs.readFileSync(source)),sourceSHA256);console.log(r.stdout);
