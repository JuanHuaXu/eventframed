import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
const dir=path.resolve('research/archive-boundary-v29');fs.mkdirSync(dir,{recursive:true});
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function below(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?below(path.join(d,e.name)):e.name.endsWith('.go')?[path.join(d,e.name)]:[]);}
const files=[...below('internal'),'go.mod','go.sum','cmd/research-archive-gen/main.go','research/archive-boundary-v29-run.mjs','research/archive-boundary-v29-audit.mjs','docs/experiments/mmm-archive-boundary-v29-protocol.md'].sort();
const hashes=Object.fromEntries(files.map(p=>[p,hash(fs.readFileSync(p))]));
fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),files:hashes},null,2)+'\n',{flag:'wx'});
const args=['test','-race','./internal/store/libravdbstore','-run','^TestResearch(Archive.*V29|HandoffCoverageV28)$','-count=1','-v','-timeout=3m'];
const start=new Date().toISOString(),begin=performance.now();let code=0,log;
try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,EVENTFRAME_ARCHIVE_BOUNDARY_V29_OUTPUT:path.join(dir,'raw.json')},timeout:190000,maxBuffer:8*1024*1024});}catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message;}
fs.writeFileSync(path.join(dir,'test.log'),log,{flag:'wx'});
let changed=[];for(const[p,h]of Object.entries(hashes))if(hash(fs.readFileSync(p))!==h)changed.push(p);
fs.writeFileSync(path.join(dir,'run.json'),JSON.stringify({start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,args,source_files:files.length,changed,raw_sha256:fs.existsSync(path.join(dir,'raw.json'))?hash(fs.readFileSync(path.join(dir,'raw.json'))):null,log_sha256:hash(log)},null,2)+'\n',{flag:'wx'});
console.log(log);if(code||changed.length)process.exitCode=1;
