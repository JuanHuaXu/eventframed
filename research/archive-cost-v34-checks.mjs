import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
const dir=path.resolve('research/archive-cost-v34/technical-prefreeze');
fs.mkdirSync(dir,{recursive:true});
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const below=d=>fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[]);
const files=[...below('internal'),'go.mod','go.sum','cmd/research-archive-cost-gen/main.go','research/archive-cost-v34-checks.mjs','docs/experiments/mmm-archive-cost-v34-protocol.md'].sort();
const hashes=Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]));
fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),files:hashes},null,2)+'\n',{flag:'wx'});
const commands=[
 ['test','-race','./internal/store/libravdbstore','-run','^TestResearch(BatchArchive(Retry|Lifecycle|Interruption)V34|ArchiveCost(Authority|Generation)V34)$','-count=1','-v','-timeout=3m'],
 ['vet','./internal/store/libravdbstore','./cmd/research-archive-cost-gen'],
 ['test','-race','./internal/researchadmission','./internal/researchvalidity','-count=3','-timeout=2m']
];
const checks=[];
for(let i=0;i<commands.length;i++){
 const args=commands[i],begin=performance.now();let code=0,log;
 try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,EVENTFRAME_RUN_ARCHIVE_COST_V34:'1'},timeout:190000,maxBuffer:8*1024*1024});}
 catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message;}
 fs.writeFileSync(path.join(dir,`check-${i}.log`),log,{flag:'wx'});
 checks.push({args,code,wall_ms:performance.now()-begin,log_sha256:hash(log)});console.log(log);
}
const changed=Object.entries(hashes).filter(([p,h])=>hash(fs.readFileSync(p))!==h).map(([p])=>p);
fs.writeFileSync(path.join(dir,'checks.json'),JSON.stringify({time:new Date().toISOString(),checks,source_files:files.length,changed},null,2)+'\n',{flag:'wx'});
if(changed.length||checks.some(c=>c.code))process.exitCode=1;
