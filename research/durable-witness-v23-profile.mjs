import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';

// Instrumented replay is attribution, NOT a second confirmation or latency gate.
const dir=path.resolve('research/durable-witness-v23');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const freeze=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));
for(const [file,digest] of Object.entries(freeze.files)) if(hash(fs.readFileSync(file))!==digest)throw Error('frozen source changed: '+file);
const names={raw:'profile-raw.ndjson',cpu:'profile-cpu.out',mutex:'profile-mutex.out',block:'profile-block.out',binary:'profile.test'};
for(const name of Object.values(names))if(fs.existsSync(path.join(dir,name)))throw Error('exclusive profile exists: '+name);
const args=['test','./internal/store/libravdbstore','-run','^TestResearchDurableWitnessLoadV23$','-count=1','-timeout','5m','-cpuprofile',path.join(dir,names.cpu),'-mutexprofile',path.join(dir,names.mutex),'-mutexprofilefraction','1','-blockprofile',path.join(dir,names.block),'-o',path.join(dir,names.binary)];
const start=new Date().toISOString(), begin=performance.now();let code=0,log;
try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,EVENTFRAME_DURABLE_WITNESS_V23_OUTPUT:path.join(dir,names.raw)},maxBuffer:16*1024*1024,timeout:310000});}
catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message;}
fs.writeFileSync(path.join(dir,'profile.log'),log,{flag:'wx'});
const tops={};
if(code===0)for(const [kind,name]of Object.entries(names).filter(([k])=>['cpu','mutex','block'].includes(k))){
  const output=execFileSync('go',['tool','pprof','-top','-nodecount=25',path.join(dir,names.binary),path.join(dir,name)],{encoding:'utf8'});
  fs.writeFileSync(path.join(dir,`profile-${kind}-top.txt`),output,{flag:'wx'});tops[kind]=output;
}
const evidence={type:'instrumented-attribution-not-confirmation',start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,args,files:Object.fromEntries(Object.entries(names).filter(([,name])=>fs.existsSync(path.join(dir,name))).map(([k,name])=>[k,{path:name,sha256:hash(fs.readFileSync(path.join(dir,name)))}])),
  additional_diagnostic_source_sha256:hash(fs.readFileSync('internal/store/libravdbstore/research_durable_witness_v23_diagnostic_test.go')),script_sha256:hash(fs.readFileSync('research/durable-witness-v23-profile.mjs')),tops};
fs.writeFileSync(path.join(dir,'profile.json'),JSON.stringify(evidence,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({code,wall_ms:evidence.wall_ms,...tops},null,2));process.exitCode=code?1:0;
