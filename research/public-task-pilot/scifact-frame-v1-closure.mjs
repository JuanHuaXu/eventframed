// Supplemental complete test-dependency freeze. The original run is immutable.
import fs from 'node:fs';import path from 'node:path';import assert from 'node:assert/strict';
import crypto from 'node:crypto';import {execFileSync,spawn} from 'node:child_process';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const root='research/public-task-pilot/scifact-frame-v1-closure';
const original='research/public-task-pilot/scifact-frame-v1';
const prior=JSON.parse(fs.readFileSync(original+'/manifest.json'));
for(const[p,h]of Object.entries(prior.sources))assert.equal(hash(fs.readFileSync(p)),h);
for(const[p,h]of Object.entries(prior.artifacts))assert.equal(hash(fs.readFileSync(original+'/'+p)),h);
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
fs.mkdirSync(root,{mode:0o700});
const sources={...Object.fromEntries(Object.keys(prior.sources).map(p=>[p,{path:p,copy:'source/'+p}]))};
sources['research/public-task-pilot/scifact-frame-v1-closure.mjs']={path:'research/public-task-pilot/scifact-frame-v1-closure.mjs',copy:'source/research/public-task-pilot/scifact-frame-v1-closure.mjs'};
const template='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const lines=execFileSync('go',['list','-test','-deps','-f',template,'./internal/researchpublicframe','./cmd/research-public-frames'],{encoding:'utf8',maxBuffer:8<<20});
for(const line of lines.split('\n').filter(Boolean)){
 const[dir,...files]=line.split('\t');for(const f of files){
  const absolute=path.resolve(dir,f),p=path.relative(process.cwd(),absolute);
  if(p.startsWith('..')){assert(path.isAbsolute(f),'only Go-reported synthetic testmain outside module');sources['generated-testmain']={path:absolute,copy:'generated/testmain.go'};}
  else sources[p]={path:p,copy:'source/'+p};
 }
}
for(const s of Object.values(sources)){
 const b=fs.readFileSync(s.path),target=root+'/'+s.copy;fs.mkdirSync(path.dirname(target),{recursive:true,mode:0o700});
 fs.writeFileSync(target,b,{flag:'wx',mode:0o600});s.sha256=hash(b);
}
const corpus=path.resolve('research/public-task-pilot/scifact-v1/corpus.json');
const corpusHash=hash(fs.readFileSync(corpus));assert.equal(corpusHash,JSON.parse(fs.readFileSync(original+'/freeze.json')).corpusSHA256);
const freeze={time:new Date().toISOString(),usage,sources,corpusSHA256:corpusHash,
 originalManifestSHA256:hash(fs.readFileSync(original+'/manifest.json')),reason:'go list -deps omitted imports used only by external integration tests',
 originalScientificValuesUnchanged:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const commands=[];const env={...process.env,EVENTFRAME_PUBLIC_CORPUS:corpus};
async function run(name,command,args){
 const log=root+'/'+name+'.log',fd=fs.openSync(log,'wx',0o600),began=new Date().toISOString();
 const p=spawn(command,args,{env,stdio:['ignore',fd,fd]});
 const result=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});fs.closeSync(fd);
 commands.push({name,command:[command,...args],began,ended:new Date().toISOString(),...result,log,logSHA256:hash(fs.readFileSync(log))});
 fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});console.log(name,result.code);assert.equal(result.code,0);
 for(const s of Object.values(sources))assert.equal(hash(fs.readFileSync(s.path)),s.sha256,'live source drift');
 assert.equal(hash(fs.readFileSync(corpus)),corpusHash,'corpus drift');
}
await run('race','go',['test','-race','-count=3','-v','./internal/researchpublicframe']);
await run('vet','go',['vet','./internal/researchpublicframe','./cmd/research-public-frames']);
await run('convert','go',['run','./cmd/research-public-frames',corpus,root+'/frames.json']);
assert.equal(hash(fs.readFileSync(root+'/frames.json')),prior.artifacts['frames.json'],'conversion changed');
await run('audit','node',['research/public-task-pilot/scifact-frame-v1-audit.mjs',corpus,root+'/frames.json',root+'/audit.json']);
await run('cost','go',['test','-run','^$','-bench','^BenchmarkPublicFrameV1','-benchtime','3x','-count','3','-benchmem','./internal/researchpublicframe']);
const artifacts={};for(const p of ['frames.json','audit.json','freeze.json','commands.json',...commands.map(c=>c.name+'.log')])artifacts[p]=hash(fs.readFileSync(root+'/'+p));
fs.writeFileSync(root+'/manifest.json',JSON.stringify({time:new Date().toISOString(),artifacts,sources,commands:commands.length,
 allCommandsTerminal:true,usage,completeTestDependencyClosure:true,exactOriginalConversion:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,sources:Object.keys(sources).length,commands:commands.length,allCommandsTerminal:true}));
