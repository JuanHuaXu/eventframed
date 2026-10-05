// Preparation reads mixed query identities once; the native predictor receives
// only the exact fit projection, with no labels or held-out query file.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync, spawn} from 'node:child_process';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const root='research/public-task-pilot/scifact-native-fit-v1';
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);
assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
fs.mkdirSync(root,{mode:0o700});
const read=p=>JSON.parse(fs.readFileSync(p));
const planPath='research/public-task-pilot/scifact-v1/split-plan.json';
const queryPath='research/public-task-pilot/scifact-v1/queries.json';
const plan=read(planPath),mixed=read(queryPath),ids=plan.splits.fit;
assert.equal(ids.length,351);assert.equal(new Set(ids).size,351);
const excluded=new Set([...plan.splits.calibration,...plan.splits.confirmation,...plan.splits.excludedOfficialTrain]);
assert(ids.every(id=>!excluded.has(id)));
const byID=new Map(mixed.map(q=>[q.id,q]));assert.equal(byID.size,mixed.length);
const fit={partition:'fit',queries:ids.map(id=>{assert(byID.has(id));return byID.get(id);})};
const fitPath=root+'/fit-only.json';
fs.writeFileSync(fitPath,JSON.stringify(fit)+'\n',{flag:'wx',mode:0o600});
const base='research/public-task-pilot/scifact-native-pilot-v6-run.mjs';
const old=fs.readFileSync(base,'utf8'),baseManifest=read('research/public-task-pilot/nativepool-v6/manifest.json');
assert.equal(hash(old),baseManifest.inputs[path.resolve(base)].sha256);
assert.equal(baseManifest.error,null);assert.equal(baseManifest.result.code,0);
const configPath=root+'/config.yaml';
const config=fs.readFileSync('research/public-task-pilot/scifact-native-pilot-v6-config.yaml','utf8').replaceAll('nativepool-v6','nativefit-v1');
fs.writeFileSync(configPath,config,{flag:'wx',mode:0o600});
const nativePath=root+'/owned-native-run.mjs';
let runner=old.replaceAll('nativepool-v6','nativefit-v1')
 .replaceAll('research/public-task-pilot/scifact-native-pilot-v6-config.yaml',configPath)
 .replaceAll('research/public-task-pilot/scifact-native-pilot-v6-run.mjs',nativePath)
 .replaceAll('research/public-task-pilot/scifact-pool-v1/native-client',root+'/native-client');
const oldArgs="const args=['research/public-task-pilot/scifact-v1/corpus.json',root,root+'/result.json'];";
const newArgs=`const args=['research/public-task-pilot/scifact-v1/corpus.json','${fitPath}',root,root+'/trace.ndjson'];`;
assert.equal(runner.split(oldArgs).length,2);runner=runner.replace(oldArgs,newArgs);
const marker="paths.push('docs/experiments/scifact-native-quant-v4-protocol.md');";
const extra=`\npaths.push('${fitPath}','${root}/freeze.json','docs/experiments/scifact-native-fit-v1-protocol.md');`;
assert.equal(runner.split(marker).length,2);runner=runner.replace(marker,marker+extra);
const oldLog="if(n.endsWith('.json')||n.endsWith('.log'))";
const newLog="if(n.endsWith('.json')||n.endsWith('.log')||n.endsWith('.ndjson'))";
assert.equal(runner.split(oldLog).length,2);runner=runner.replace(oldLog,newLog);
const inverse=runner.replace(newLog,oldLog).replace(marker+extra,marker).replace(newArgs,oldArgs)
 .replaceAll(root+'/native-client','research/public-task-pilot/scifact-pool-v1/native-client')
 .replaceAll(nativePath,'research/public-task-pilot/scifact-native-pilot-v6-run.mjs')
 .replaceAll(configPath,'research/public-task-pilot/scifact-native-pilot-v6-config.yaml')
 .replaceAll('nativefit-v1','nativepool-v6');
assert.equal(inverse,old);
fs.writeFileSync(nativePath,runner,{flag:'wx',mode:0o600});
const paths=new Set(['go.mod','go.sum',base,nativePath,configPath,fitPath,planPath,queryPath,
 'docs/experiments/scifact-native-fit-v1-protocol.md','research/public-task-pilot/scifact-native-fit-v1-run.mjs']);
const template='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const listed=execFileSync('go',['list','-test','-deps','-f',template,'./cmd/research-public-native-fit'],{encoding:'utf8',maxBuffer:8<<20});
const sources={};let n=0;
for(const line of listed.split('\n').filter(Boolean)){
 const[d,...files]=line.split('\t');for(const f of files)paths.add(path.resolve(d,f));
}
for(const p of [...paths].sort()){
 const a=path.resolve(p),rel=path.relative(process.cwd(),a),dest=rel.startsWith('..')?'generated/testmain-'+(n++)+'.go':'source/'+rel;
 const b=fs.readFileSync(a);fs.mkdirSync(path.dirname(root+'/'+dest),{recursive:true,mode:0o700});
 fs.writeFileSync(root+'/'+dest,b,{flag:'wx',mode:0o600});sources[a]={sha256:hash(b),copy:dest};
}
const corpus='research/public-task-pilot/scifact-v1/corpus.json';
const freeze={time:new Date().toISOString(),sources,usage,completeTestDependencyClosure:true,
 baseManifestSHA256:hash(fs.readFileSync('research/public-task-pilot/nativepool-v6/manifest.json')),
 runnerCompleteInverseEquality:true,fitQueries:351,corpusSHA256:hash(fs.readFileSync(corpus)),
 confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
assert.equal(freeze.corpusSHA256,read('research/public-task-pilot/scifact-v1/source.json').artifacts['corpus.json']);
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const commands=[];
async function run(name,cmd,args){
 const log=root+'/'+name+'.log',fd=fs.openSync(log,'wx',0o600),start=new Date().toISOString();
 const p=spawn(cmd,args,{stdio:['ignore',fd,fd]});
 const end=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});
 fs.closeSync(fd);commands.push({name,command:[cmd,...args],start,end:new Date().toISOString(),...end,log,logSHA256:hash(fs.readFileSync(log))});
 fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});
 for(const[p,s]of Object.entries(sources))assert.equal(hash(fs.readFileSync(p)),s.sha256);
 console.log(name,end.code);return end;
}
assert.equal((await run('race','go',['test','-race','-count=3','-v','./cmd/research-public-native-fit'])).code,0);
assert.equal((await run('vet','go',['vet','./cmd/research-public-native-fit'])).code,0);
assert.equal((await run('build','go',['build','-o',root+'/native-client','./cmd/research-public-native-fit'])).code,0);
const native=await run('native','node',[nativePath]);
const artifacts={};for(const p of ['freeze.json','fit-only.json','commands.json','native-client',...commands.map(c=>c.name+'.log')])artifacts[p]=hash(fs.readFileSync(root+'/'+p));
const nativeManifest='research/public-task-pilot/nativefit-v1/manifest.json';
if(fs.existsSync(nativeManifest))artifacts.nativeManifest={path:nativeManifest,sha256:hash(fs.readFileSync(nativeManifest))};
fs.writeFileSync(root+'/manifest.json',JSON.stringify({time:new Date().toISOString(),sources,artifacts,commands:commands.length,
 allCommandsTerminal:true,nativeCode:native.code,completeTestDependencyClosure:true,usage,
 confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,nativeCode:native.code,sources:Object.keys(sources).length,allCommandsTerminal:true}));
process.exitCode=native.code;
