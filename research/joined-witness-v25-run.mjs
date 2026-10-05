import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import os from 'node:os';
import {execFileSync} from 'node:child_process';
const dir=path.resolve('research/joined-witness-v25'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const freeze=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
for(const [p,h]of Object.entries(freeze.files))if(hash(fs.readFileSync(p))!==h)throw Error('changed source '+p);
const raw=path.join(dir,'raw.ndjson');if(fs.existsSync(raw)||fs.existsSync(path.join(dir,'run.json')))throw Error('exclusive outputs exist');
const args=['test','./internal/store/libravdbstore','-run','^TestResearchJoinedWitnessLoadV25$','-count=1','-v','-timeout=5m'];
const start=new Date().toISOString(),begin=performance.now();let code=0,log;
try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,EVENTFRAME_JOINED_WITNESS_V25_OUTPUT:raw},maxBuffer:16*1024*1024,timeout:310000});}
catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message;}
fs.writeFileSync(path.join(dir,'load.log'),log,{flag:'wx'});
fs.writeFileSync(path.join(dir,'run.json'),JSON.stringify({start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,args,freeze_sha256:hash(fs.readFileSync(path.join(dir,'freeze.json'))),raw_sha256:fs.existsSync(raw)?hash(fs.readFileSync(raw)):null,log_sha256:hash(log),node:process.version,os:{platform:os.platform(),release:os.release(),arch:os.arch(),cpu:os.cpus()[0]?.model,logical_cpus:os.cpus().length,total_memory:os.totalmem()}},null,2)+'\n',{flag:'wx'});
console.log(log);process.exitCode=code?1:0;
