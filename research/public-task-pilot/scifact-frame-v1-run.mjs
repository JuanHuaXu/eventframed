import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync,spawn} from 'node:child_process';

const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const root='research/public-task-pilot/scifact-frame-v1';
const corpus='research/public-task-pilot/scifact-v1/corpus.json';
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);
assert(Number.isFinite(usage)&&usage>=0&&usage<=80,'fresh bounded weekly usage required');
const original=JSON.parse(fs.readFileSync('research/public-task-pilot/scifact-v1/source.json'));
assert.equal(hash(fs.readFileSync(corpus)),original.artifacts['corpus.json']);
fs.mkdirSync(root,{mode:0o700});
const paths=new Set(['go.mod','go.sum','internal/researchpublicframe/document_test.go',
 'docs/experiments/scifact-frame-v1-protocol.md','research/public-task-pilot/scifact-frame-v1-audit.mjs',
 'research/public-task-pilot/scifact-frame-v1-run.mjs','research/public-task-pilot/scifact-frame-v1-strict-preflight.mjs']);
const template='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const packages=execFileSync('go',['list','-deps','-f',template,'./internal/researchpublicframe','./cmd/research-public-frames'],{encoding:'utf8',maxBuffer:8<<20});
for(const line of packages.split('\n').filter(Boolean)){
 const[dir,...names]=line.split('\t');for(const name of names){const p=path.relative(process.cwd(),path.join(dir,name));assert(!p.startsWith('..'));paths.add(p);}
}
const sources={};for(const p of [...paths].sort()){
 const b=fs.readFileSync(p),out=root+'/source/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});
 fs.writeFileSync(out,b,{flag:'wx',mode:0o600});sources[p]=hash(b);
}
const env={...process.env,EVENTFRAME_PUBLIC_CORPUS:path.resolve(corpus)};
const freeze={time:new Date().toISOString(),usage,sources,corpusSHA256:original.artifacts['corpus.json'],
 goEnv:execFileSync('go',['env','GOVERSION','GOOS','GOARCH'],{encoding:'utf8'}).trim().split('\n'),
 noExternalModelCalls:true,noProductionMutation:true,noLabelReads:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const commands=[];
async function run(name,command,args){
 const out=root+'/'+name+'.log',fd=fs.openSync(out,'wx',0o600);const began=new Date().toISOString();
 const p=spawn(command,args,{env,stdio:['ignore',fd,fd]});
 const result=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});
 fs.closeSync(fd);commands.push({name,command:[command,...args],began,ended:new Date().toISOString(),...result,log:out,logSHA256:hash(fs.readFileSync(out))});
 fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});
 console.log(name,result.code);assert.equal(result.code,0,`terminal command failed: ${name}`);
 for(const[p,h]of Object.entries(sources))assert.equal(hash(fs.readFileSync(p)),h,'source drift');
}
await run('race','go',['test','-race','-count=3','-v','./internal/researchpublicframe']);
await run('vet','go',['vet','./internal/researchpublicframe','./cmd/research-public-frames']);
await run('convert','go',['run','./cmd/research-public-frames',corpus,root+'/frames.json']);
await run('audit','node',['research/public-task-pilot/scifact-frame-v1-audit.mjs',corpus,root+'/frames.json',root+'/audit.json']);
await run('cost','go',['test','-run','^$','-bench','^BenchmarkPublicFrameV1','-benchtime','3x','-count','3','-benchmem','./internal/researchpublicframe']);
const artifacts={};for(const p of ['frames.json','audit.json','freeze.json','commands.json',...commands.map(c=>c.name+'.log')])artifacts[p]=hash(fs.readFileSync(root+'/'+p));
fs.writeFileSync(root+'/manifest.json',JSON.stringify({time:new Date().toISOString(),sources,artifacts,
 commands:commands.length,allCommandsTerminal:true,usage,qualityExperiment:false,
 labelPartitionsUntouchedByPrediction:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,sources:Object.keys(sources).length,commands:commands.length,allCommandsTerminal:true}));
