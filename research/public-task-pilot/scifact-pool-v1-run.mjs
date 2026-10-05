import fs from'node:fs';import path from'node:path';import assert from'node:assert/strict';import crypto from'node:crypto';import{execFileSync,spawn}from'node:child_process';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),root='research/public-task-pilot/scifact-pool-v1';
const corpus=path.resolve('research/public-task-pilot/scifact-v1/corpus.json');
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
fs.mkdirSync(root,{mode:0o700});const sourcePaths=new Set(['go.mod','go.sum','docs/experiments/scifact-pool-v1-protocol.md',
 'research/public-task-pilot/scifact-pool-v1-run.mjs','research/public-task-pilot/scifact-pool-v1-audit.mjs',
 'research/public-task-pilot/scifact-native-pilot-v1-config.yaml']);
const template='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const text=execFileSync('go',['list','-test','-deps','-f',template,'./internal/researchpublicpool','./internal/researchpublicframe','./cmd/research-public-pool','./cmd/research-public-native-pilot'],{encoding:'utf8',maxBuffer:8<<20});
const generated=[];for(const line of text.split('\n').filter(Boolean)){const[dir,...names]=line.split('\t');for(const n of names){const a=path.resolve(dir,n),p=path.relative(process.cwd(),a);if(p.startsWith('..')){assert(path.isAbsolute(n));generated.push(a);}else sourcePaths.add(p);}}
const sources={};function copy(p,target){const b=fs.readFileSync(p),out=root+'/'+target;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.writeFileSync(out,b,{flag:'wx',mode:0o600});sources[p]={sha256:hash(b),copy:target};}
for(const p of [...sourcePaths].sort())copy(p,'source/'+p);for(let i=0;i<generated.length;i++)copy(generated[i],'generated/testmain-'+i+'.go');
const freeze={time:new Date().toISOString(),sources,usage,corpusSHA256:hash(fs.readFileSync(corpus)),allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
assert.equal(freeze.corpusSHA256,JSON.parse(fs.readFileSync('research/public-task-pilot/scifact-v1/source.json')).artifacts['corpus.json']);
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const commands=[];async function run(name,cmd,args){const log=root+'/'+name+'.log',fd=fs.openSync(log,'wx',0o600),start=new Date().toISOString();const p=spawn(cmd,args,{env:{...process.env,EVENTFRAME_PUBLIC_CORPUS:corpus},stdio:['ignore',fd,fd]});const r=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});fs.closeSync(fd);commands.push({name,command:[cmd,...args],start,end:new Date().toISOString(),...r,log,logSHA256:hash(fs.readFileSync(log))});fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});console.log(name,r.code);assert.equal(r.code,0);for(const[p,s]of Object.entries(sources))assert.equal(hash(fs.readFileSync(p)),s.sha256);}
await run('race','go',['test','-race','-count=3','-v','./internal/researchpublicpool','./internal/researchpublicframe']);
await run('vet','go',['vet','./internal/researchpublicpool','./cmd/research-public-pool','./cmd/research-public-native-pilot']);
await run('pool','go',['run','./cmd/research-public-pool',corpus,root+'/pool.json']);
await run('audit','node',['research/public-task-pilot/scifact-pool-v1-audit.mjs','research/public-task-pilot/scifact-frame-v1-closure/frames.json',root+'/pool.json',root+'/audit.json']);
await run('cost','go',['test','-run','^$','-bench','^BenchmarkPublicPoolV1','-benchtime','3x','-count','3','-benchmem','./internal/researchpublicpool']);
await run('native-client-build','go',['build','-o',root+'/native-client','./cmd/research-public-native-pilot']);
const artifacts={};for(const p of ['pool.json','audit.json','freeze.json','commands.json','native-client',...commands.map(c=>c.name+'.log')])artifacts[p]=hash(fs.readFileSync(root+'/'+p));
fs.writeFileSync(root+'/manifest.json',JSON.stringify({time:new Date().toISOString(),sources,artifacts,commands:commands.length,allCommandsTerminal:true,usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({root,sources:Object.keys(sources).length,commands:commands.length,allCommandsTerminal:true}));
