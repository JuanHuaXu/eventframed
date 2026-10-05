import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync,spawn} from 'node:child_process';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const read=p=>JSON.parse(fs.readFileSync(p));
const root='research/public-task-pilot/scifact-native-frontier-v3-executed';
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const oldRoot='research/public-task-pilot/scifact-native-resume-v2',oldNative='research/public-task-pilot/nativeresume-v2';
const failure=read(oldRoot+'/failure-audit.json');assert(failure.allCommandsTerminal);assert.equal(failure.fitQueriesCompleted,0);assert.equal(failure.fullCorpusImported,5183);
const receipts=fs.readFileSync(oldNative+'/trace.ndjson');assert.equal(hash(receipts),failure.traceSHA256);
const rows=receipts.toString().trim().split('\n').map(JSON.parse);
const successful=rows.filter(r=>r.kind==='verify'||r.kind==='import');assert.equal(successful.length,5183);assert(successful.every((r,i)=>r.ok&&r.index===i));
const pool=read('research/public-task-pilot/scifact-pool-v1/pool.json');for(let i=0;i<successful.length;i++)assert.equal(successful[i].id,pool[i].candidate.ID);
fs.mkdirSync(root,{mode:0o700});
const prefixPath=root+'/prefix.json';fs.writeFileSync(prefixPath,JSON.stringify({ids:successful.map(r=>r.id),originalTraceSHA256:hash(receipts),receiptsNotDurabilityProof:true})+'\n',{flag:'wx',mode:0o600});
const fitPath=root+'/fit-only.json',fit=fs.readFileSync(oldRoot+'/fit-only.json');
assert.equal(hash(fit),read(oldRoot+'/manifest.json').artifacts['fit-only.json']);
fs.writeFileSync(fitPath,fit,{flag:'wx',mode:0o600});
const source=oldRoot+'/owned-native-run.mjs',raw=fs.readFileSync(source,'utf8');
assert.equal(hash(raw),read(oldNative+'/manifest.json').inputs[path.resolve(source)].sha256);
const configPath=root+'/config.yaml',config=fs.readFileSync(oldRoot+'/config.yaml','utf8').replaceAll('nativeresume-v2','nativefrontier-v3');
fs.writeFileSync(configPath,config,{flag:'wx',mode:0o600});
const nativePath=root+'/owned-native-run.mjs';
let runner=raw.replaceAll('nativeresume-v2','nativefrontier-v3').replaceAll(oldRoot+'/config.yaml',configPath)
 .replaceAll(source,nativePath).replaceAll(oldRoot+'/native-client',root+'/native-client').replaceAll(oldRoot+'/fit-only.json',fitPath)
 .replaceAll(oldRoot+'/freeze.json',root+'/freeze.json').replaceAll('docs/experiments/scifact-native-resume-v2-protocol.md','docs/experiments/scifact-native-frontier-v3-protocol.md');
// The frozen v2 launcher already contains a copy block. Replace it, rather than
// nesting another clone of the older v1 store into the same destination.
const inheritedStart=runner.indexOf('\nconst copied=[];for');
const inheritedEnd=runner.indexOf('\nfs.copyFileSync(scanner',inheritedStart);
assert(inheritedStart>0&&inheritedEnd>inheritedStart);
const inheritedCopyBlock=runner.slice(inheritedStart,inheritedEnd);
runner=runner.slice(0,inheritedStart)+runner.slice(inheritedEnd);
const marker="paths.push(scanner);";
const copyBlock=`\nconst copied=[];for(const name of ['store.libravdb','store.libravdb.embedding.json','daemon.libravdb','daemon.libravdb.embedding.json']){\n const original=path.resolve('${oldNative}/'+name),dest=root+'/'+name;\n const before=await sha(original);fs.copyFileSync(original,dest,fs.constants.COPYFILE_EXCL);assert.equal(await sha(dest),before);\n copied.push({original,dest,sha256:before,bytes:fs.statSync(original).size});paths.push(original);\n}\njson('copies-before.json',copied);paths.push(root+'/copies-before.json','${prefixPath}');`;
assert.equal(runner.split(marker).length,2);runner=runner.replace(marker,marker+copyBlock);
const oldArgs=`const args=['research/public-task-pilot/scifact-v1/corpus.json','${fitPath}',root,root+'/trace.ndjson','${oldRoot}/prefix.json'];`;
const newArgs=`const args=['research/public-task-pilot/scifact-v1/corpus.json','${fitPath}',root,root+'/trace.ndjson','${prefixPath}'];`;
assert.equal(runner.split(oldArgs).length,2);runner=runner.replace(oldArgs,newArgs);
const inverse=runner.replace(newArgs,oldArgs).replace(marker+copyBlock,marker+inheritedCopyBlock)
 .replaceAll(root+'/native-client',oldRoot+'/native-client').replaceAll(configPath,oldRoot+'/config.yaml')
 .replaceAll(nativePath,source).replaceAll(fitPath,oldRoot+'/fit-only.json').replaceAll(root+'/freeze.json',oldRoot+'/freeze.json')
 .replaceAll('docs/experiments/scifact-native-frontier-v3-protocol.md','docs/experiments/scifact-native-resume-v2-protocol.md').replaceAll('nativefrontier-v3','nativeresume-v2');
assert.equal(inverse,raw);
fs.writeFileSync(nativePath,runner,{flag:'wx',mode:0o600});
const paths=new Set(['go.mod','go.sum',source,nativePath,configPath,fitPath,prefixPath,
 'docs/experiments/scifact-native-frontier-v3-protocol.md','research/public-task-pilot/scifact-native-frontier-v3-run.mjs',
 'research/public-task-pilot/scifact-native-frontier-v3-gen.mjs','research/public-task-pilot/scifact-native-frontier-v3-generation.json',
 oldRoot+'/failure-audit.json',oldNative+'/manifest.json']);
const template='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const listed=execFileSync('go',['list','-test','-deps','-f',template,'./internal/researchpublicresume','./internal/researchpublicrankguard','./cmd/research-public-native-frontier'],{encoding:'utf8',maxBuffer:8<<20});
for(const line of listed.split('\n').filter(Boolean)){const[d,...names]=line.split('\t');for(const n of names)paths.add(path.resolve(d,n));}
const sources={};let generated=0;
for(const p of [...paths].sort()){const a=path.resolve(p),rel=path.relative(process.cwd(),a),dest=rel.startsWith('..')?'generated/testmain-'+(generated++)+'.go':'source/'+rel;
 const b=fs.readFileSync(a);fs.mkdirSync(path.dirname(root+'/'+dest),{recursive:true,mode:0o700});fs.writeFileSync(root+'/'+dest,b,{flag:'wx',mode:0o600});sources[a]={sha256:hash(b),copy:dest};}
const freeze={time:new Date().toISOString(),sources,usage,completeTestDependencyClosure:true,runnerCompleteInverseEquality:true,
 claimedPrefix:5183,originalFailureSHA256:hash(fs.readFileSync(oldRoot+'/failure-audit.json')),originalManifestSHA256:hash(fs.readFileSync(oldNative+'/manifest.json')),
 corpusSHA256:hash(fs.readFileSync('research/public-task-pilot/scifact-v1/corpus.json')),fitQueries:351,confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const commands=[];async function run(name,cmd,args){const log=root+'/'+name+'.log',fd=fs.openSync(log,'wx',0o600),start=new Date().toISOString();
 const p=spawn(cmd,args,{stdio:['ignore',fd,fd]});const end=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});fs.closeSync(fd);
 commands.push({name,command:[cmd,...args],start,end:new Date().toISOString(),...end,log,logSHA256:hash(fs.readFileSync(log))});fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});
 for(const[p,s]of Object.entries(sources))assert.equal(hash(fs.readFileSync(p)),s.sha256);console.log(name,end.code);return end;}
assert.equal((await run('race','go',['test','-race','-count=3','-v','./internal/researchpublicresume','./internal/researchpublicrankguard','./cmd/research-public-native-frontier'])).code,0);
assert.equal((await run('vet','go',['vet','./internal/researchpublicresume','./internal/researchpublicrankguard','./cmd/research-public-native-frontier'])).code,0);
assert.equal((await run('build','go',['build','-o',root+'/native-client','./cmd/research-public-native-frontier'])).code,0);
const result=await run('native','node',[nativePath]);
const artifacts={};for(const p of ['freeze.json','prefix.json','fit-only.json','commands.json','native-client',...commands.map(c=>c.name+'.log')])artifacts[p]=hash(fs.readFileSync(root+'/'+p));
const native='research/public-task-pilot/nativeresume-v2/manifest.json';assert(fs.existsSync(native));artifacts.nativeManifest={path:native,sha256:hash(fs.readFileSync(native))};
fs.writeFileSync(root+'/manifest.json',JSON.stringify({time:new Date().toISOString(),sources,artifacts,commands:4,allCommandsTerminal:true,
 nativeCode:result.code,completeTestDependencyClosure:true,usage,confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,sources:Object.keys(sources).length,nativeCode:result.code,allCommandsTerminal:true}));process.exitCode=result.code;
