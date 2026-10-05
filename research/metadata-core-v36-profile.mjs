// Instrumented attribution repeat, never an untouched confirmation cohort.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import{execFileSync}from'node:child_process';
const dir=path.resolve('research/metadata-core-v36'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const f=JSON.parse(fs.readFileSync(path.join(dir,'freeze.json')));for(const[p,h]of Object.entries(f.files))if(hash(fs.readFileSync(p))!==h)throw Error('changed source '+p);
const names={raw:'profile-raw.ndjson',cpu:'profile-cpu.out',mutex:'profile-mutex.out',block:'profile-block.out',binary:'profile.test'};
for(const p of Object.values(names))if(fs.existsSync(path.join(dir,p)))throw Error('exclusive profile exists '+p);
const args=['test','./internal/store/libravdbstore','-run','^TestResearchMetadataCoreLoadV36$','-count=1','-timeout=5m','-cpuprofile',path.join(dir,names.cpu),'-mutexprofile',path.join(dir,names.mutex),'-mutexprofilefraction','1','-blockprofile',path.join(dir,names.block),'-o',path.join(dir,names.binary)];
const begin=performance.now();let code=0,log;
try{log=execFileSync('go',args,{encoding:'utf8',env:{...process.env,EVENTFRAME_METADATA_CORE_V36_OUTPUT:path.join(dir,names.raw)},timeout:310000,maxBuffer:8*1024*1024})}
catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message}
fs.writeFileSync(path.join(dir,'profile.log'),log,{flag:'wx'});
const summaries={};if(!code)for(const[kind,options]of[
 ['cpu',['-top','-cum','-nodecount=40']],['mutex',['-top','-cum','-nodecount=40']],['block',['-top','-cum','-nodecount=40']],
 ['serving-cpu',['-top','-cum','-nodecount=40','-focus=Service.Recall|applyV34|coreStoreV36.Search|GetBayesianPosterior']],
 ['owner-mutex',['-list=applyV34|coreStoreV36.Search|witnessStoreV23.Search|GetBayesianPosterior','-cum']]
]){
 const which=kind==='serving-cpu'?'cpu':kind==='owner-mutex'?'mutex':kind;
 const text=execFileSync('go',['tool','pprof',...options,path.join(dir,names.binary),path.join(dir,names[which])],{encoding:'utf8',maxBuffer:8*1024*1024});
 fs.writeFileSync(path.join(dir,`profile-${kind}.txt`),text,{flag:'wx'});summaries[kind]=text;
}
for(const[p,h]of Object.entries(f.files))if(hash(fs.readFileSync(p))!==h)throw Error('changed source '+p);
const report={type:'instrumented-attribution-repeat-not-confirmation',time:new Date().toISOString(),code,args,wall_ms:performance.now()-begin,scope:'entire 16-trial harness includes fixture construction; cumulative times overlap; mutex/block delay aggregates across goroutines, never per-request latency',script_sha256:hash(fs.readFileSync('research/metadata-core-v36-profile.mjs')),files:Object.fromEntries(Object.entries(names).filter(([,p])=>fs.existsSync(path.join(dir,p))).map(([k,p])=>[k,{path:p,sha256:hash(fs.readFileSync(path.join(dir,p)))}])),summary_files:Object.keys(summaries)};
fs.writeFileSync(path.join(dir,'profile.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(report,null,2));process.exitCode=code?1:0;
