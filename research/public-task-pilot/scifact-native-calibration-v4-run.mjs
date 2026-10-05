// New owned trial only. No target annotations are read by this launcher.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync,spawn} from 'node:child_process';
const root='research/public-task-pilot/scifact-native-calibration-v4';
const native='research/public-task-pilot/nativecal-v4';
const prior='research/public-task-pilot/scifact-native-frontier-v3-executed';
const data='research/public-task-pilot/scifact-v1';
const read=p=>JSON.parse(fs.readFileSync(p));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
assert(!fs.existsSync(root)&&!fs.existsSync(native),'never overwrite a previous trial');
fs.mkdirSync(root,{mode:0o700});
const write=(p,v)=>fs.writeFileSync(root+'/'+p,JSON.stringify(v,null,2)+'\n',{flag:'wx',mode:0o600});
const plan=read(data+'/split-plan.json'),mixed=read(data+'/queries.json'),byID=new Map(mixed.map(q=>[q.id,q]));
assert.equal(plan.splits.calibration.length,180);assert.equal(plan.units.calibration.length,104);
write('calibration-only.json',{partition:'calibration',queries:plan.splits.calibration.map(id=>{const q=byID.get(id);assert(q);assert.deepEqual(Object.keys(q).sort(),['id','text']);return q;})});
const models={},modelOrigins=[];
for(const[arm,study]of [['primary','scifact-sourceonly-pairrank-v1'],['secondary','scifact-sourceonly-lambdarank-v1']]){
 const dir='research/public-task-pilot/'+study,a=read(dir+'/audit-results.json'),m=read(dir+'/manifest.json'),p=read(dir+'/predictions.json');
 assert.equal(await sha(dir+'/manifest.json'),a.manifestSHA256);assert.equal(await sha(dir+'/predictions.json'),m.artifacts['predictions.json']);
 assert.equal(p.models[0].fold,-1);assert.deepEqual(p.models[0].trainIDs,plan.splits.fit);assert.equal(a.calibrationPredictions,0);assert.equal(a.confirmationPredictions,0);
 models[arm]=p.models[0].model;assert.equal(models[arm].weights.length,8);assert.equal(models[arm].weights[6],0);assert.equal(models[arm].weights[7],0);
 modelOrigins.push({arm,study,predictionsSHA256:await sha(dir+'/predictions.json'),manifestSHA256:await sha(dir+'/manifest.json'),auditSHA256:await sha(dir+'/audit-results.json'),fold:-1,trainQueries:351});
}
write('models.json',models);write('model-origins.json',modelOrigins);
const template=prior+'/owned-native-run.mjs',templateBytes=fs.readFileSync(template,'utf8');
assert.equal(await sha(template),read('research/public-task-pilot/nativefrontier-v3/manifest.json').inputs[path.resolve(template)].sha256);
const config=fs.readFileSync(prior+'/config.yaml','utf8').replaceAll('nativefrontier-v3','nativecal-v4');
fs.writeFileSync(root+'/config.yaml',config,{flag:'wx',mode:0o600});
let runner=templateBytes.replaceAll(prior,root).replaceAll('nativefrontier-v3','nativecal-v4').replaceAll('fit-only.json','calibration-only.json').replaceAll('docs/experiments/scifact-native-frontier-v3-protocol.md','docs/experiments/scifact-native-calibration-v4-protocol.md');
const oldArgs=`const args=['${data}/corpus.json','${root}/calibration-only.json',root,root+'/trace.ndjson','${root}/prefix.json'];`;
const newArgs=`const args=['${data}/corpus.json','${root}/calibration-only.json',root,'${root}/models.json'];`;
assert.equal(runner.split(oldArgs).length,2);runner=runner.replace(oldArgs,newArgs);
assert.equal(runner.split(`,'${root}/prefix.json'`).length,2);
runner=runner.replace(`,'${root}/prefix.json'`,'');
const modelMarker=`paths.push('docs/experiments/scifact-native-quant-v4-protocol.md');`;
assert.equal(runner.split(modelMarker).length,2);
runner=runner.replace(modelMarker,modelMarker+`\npaths.push('${root}/models.json','${root}/model-origins.json','${data}/corpus.json');`);
fs.writeFileSync(root+'/owned-native-run.mjs',runner,{flag:'wx',mode:0o600});
write('launcher-generation.json',{template,templateSHA256:hash(templateBytes),generatedSHA256:hash(runner),configSHA256:hash(config),nativeRoot:native,originalClosedStore:'research/public-task-pilot/nativeresume-v2',transformations:['owned study/root/config/protocol paths','calibration-only instead of fit-only','four explicit predictor args','remove obsolete prefix input','freeze models/origins/corpus'],scientificModelsNotRefit:true});
const paths=new Set(['go.mod','go.sum',template,root+'/owned-native-run.mjs',root+'/config.yaml',root+'/models.json',root+'/model-origins.json',root+'/calibration-only.json',root+'/launcher-generation.json','docs/experiments/scifact-native-calibration-v4-protocol.md','research/public-task-pilot/scifact-native-calibration-v4-run.mjs','research/public-task-pilot/scifact-native-calibration-v4-audit.mjs','research/public-task-pilot/scifact-native-fit-v1-audit.mjs']);
const packages=['./internal/researchpublicresume','./internal/researchpublichybrid','./internal/researchpublicpairrank','./internal/researchpublicrankmask','./cmd/research-public-native-calibration'];
const format='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const listing=execFileSync('go',['list','-deps','-test','-f',format,...packages],{encoding:'utf8',maxBuffer:8<<20});
for(const l of listing.split('\n').filter(Boolean)){const[d,...names]=l.split('\t');for(const n of names)paths.add(path.resolve(d,n));}
const sources={};let generated=0;
for(const p of [...paths].sort()){const a=path.resolve(p),rel=path.relative(process.cwd(),a),dest=rel.startsWith('..')?'generated/testmain-'+generated+++'.go':'source/'+rel,b=fs.readFileSync(a);fs.mkdirSync(path.dirname(root+'/'+dest),{recursive:true,mode:0o700});fs.writeFileSync(root+'/'+dest,b,{flag:'wx',mode:0o600});sources[a]={copy:dest,sha256:hash(b)};}
const inputs={};for(const p of [data+'/corpus.json',data+'/queries.json',data+'/split-plan.json','research/public-task-pilot/scifact-pool-v1/pool.json',...modelOrigins.flatMap(o=>['predictions.json','manifest.json','audit-results.json'].map(n=>'research/public-task-pilot/'+o.study+'/'+n))])inputs[p]={sha256:await sha(p),bytes:fs.statSync(p).size};
write('freeze.json',{time:new Date().toISOString(),sources,inputs,usage,completeTestDependencyClosure:true,goVersion:execFileSync('go',['version'],{encoding:'utf8'}).trim(),calibrationQueries:180,confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});
const commands=[];
async function run(name,cmd,args){const log=root+'/'+name+'.log',fd=fs.openSync(log,'wx',0o600),start=new Date().toISOString(),p=spawn(cmd,args,{stdio:['ignore',fd,fd]});
 const result=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});fs.closeSync(fd);
 commands.push({name,command:[cmd,...args],start,end:new Date().toISOString(),...result,log,logSHA256:await sha(log)});fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});
 for(const[p,s]of Object.entries(sources))assert.equal(await sha(p),s.sha256);console.log(name,result.code);return result;
}
assert.equal((await run('race','go',['test','-race','-count=3','-v',...packages])).code,0);
assert.equal((await run('vet','go',['vet',...packages])).code,0);
assert.equal((await run('build','go',['build','-o',root+'/native-client','./cmd/research-public-native-calibration'])).code,0);
const result=await run('native','node',[root+'/owned-native-run.mjs']);
for(const[p,s]of Object.entries(inputs))assert.equal(await sha(p),s.sha256);
const artifacts={};for(const n of ['freeze.json','commands.json','native-client',...commands.map(c=>c.name+'.log')])artifacts[n]=await sha(root+'/'+n);
artifacts.nativeManifest={path:native+'/manifest.json',sha256:await sha(native+'/manifest.json')};
write('manifest.json',{time:new Date().toISOString(),sources,inputs,artifacts,commands:4,allCommandsTerminal:true,nativeCode:result.code,completeTestDependencyClosure:true,usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE'});
console.log(JSON.stringify({root,native,nativeCode:result.code,allCommandsTerminal:true}));process.exitCode=result.code;
