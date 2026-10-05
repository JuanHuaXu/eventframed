import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync,spawn} from 'node:child_process';
const root='research/public-task-pilot/scifact-hybrid-v1';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const prior='research/public-task-pilot/scifact-native-frontier-v3-executed',native='research/public-task-pilot/nativefrontier-v3';
const old=JSON.parse(fs.readFileSync(prior+'/audit-results.json'));
assert(old.all5183StoredRecordsVerified&&old.fitDiagnosticOnly);assert.equal(old.queries,351);
assert.equal(await sha(native+'/trace.ndjson'),old.traceSHA256);assert.equal(await sha(prior+'/manifest.json'),old.manifestSHA256);
const inputs={};for(const p of ['research/public-task-pilot/scifact-pool-v1/pool.json',prior+'/fit-only.json',native+'/trace.ndjson',prior+'/audit-results.json',prior+'/manifest.json','research/public-task-pilot/scifact-v1/split-plan.json'])inputs[p]={sha256:await sha(p),bytes:fs.statSync(p).size};
fs.mkdirSync(root,{mode:0o700});
const paths=new Set(['go.mod','go.sum','docs/experiments/scifact-hybrid-v1-protocol.md','research/public-task-pilot/scifact-hybrid-v1-run.mjs','research/public-task-pilot/scifact-hybrid-v1-audit.mjs','research/public-task-pilot/scifact-native-fit-v1-audit.mjs']);
const template='{{if .Module}}{{if eq .Module.Path "github.com/JuanHuaXu/eventframed"}}{{.Dir}}{{range .GoFiles}}{{printf "\t%s" .}}{{end}}{{end}}{{end}}';
const deps=execFileSync('go',['list','-deps','-test','-f',template,'./internal/researchpublichybrid','./cmd/research-public-hybrid'],{encoding:'utf8'});
const generated=[];for(const l of deps.split('\n').filter(Boolean)){const[dir,...names]=l.split('\t');for(const n of names){const p=path.relative(process.cwd(),path.resolve(dir,n));if(p.startsWith('..'))generated.push(path.resolve(dir,n));else paths.add(p);}}
const sources={};function copy(p,to){const b=fs.readFileSync(p);fs.mkdirSync(path.dirname(root+'/'+to),{recursive:true,mode:0o700});fs.writeFileSync(root+'/'+to,b,{flag:'wx',mode:0o600});sources[p]={sha256:hash(b),copy:to};}
for(const p of [...paths].sort())copy(p,'source/'+p);for(let i=0;i<generated.length;i++)copy(generated[i],'generated/testmain-'+i+'.go');
const freeze={time:new Date().toISOString(),sources,inputs,completeTestDependencyClosure:true,goVersion:execFileSync('go',['version'],{encoding:'utf8'}).trim(),nodeVersion:process.version,usage};
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
const commands=[];
async function run(name,cmd,args){const log=root+'/'+name+'.log',fd=fs.openSync(log,'wx',0o600),start=new Date().toISOString();const p=spawn(cmd,args,{stdio:['ignore',fd,fd]});const timer=setTimeout(()=>p.kill('SIGKILL'),240000);const r=await new Promise((resolve,reject)=>{p.once('error',reject);p.once('close',(code,signal)=>resolve({code,signal}));});clearTimeout(timer);fs.closeSync(fd);commands.push({name,command:[cmd,...args],start,end:new Date().toISOString(),...r,log,logSHA256:await sha(log)});fs.writeFileSync(root+'/commands.json',JSON.stringify(commands,null,2)+'\n',{mode:0o600});console.log(name,r.code);assert.equal(r.code,0);for(const[p,s]of Object.entries(sources))assert.equal(await sha(p),s.sha256);}
await run('race','go',['test','-race','-count=3','-v','./internal/researchpublichybrid']);
await run('vet','go',['vet','./internal/researchpublichybrid','./cmd/research-public-hybrid']);
await run('build','go',['build','-o',root+'/predictor','./cmd/research-public-hybrid']);
await run('predict','/usr/bin/time',['-l',path.resolve(root+'/predictor'),'research/public-task-pilot/scifact-pool-v1/pool.json',prior+'/fit-only.json',native+'/trace.ndjson',root+'/predictions.ndjson']);
await run('bench','go',['test','-run','^$','-bench','^BenchmarkHybridSearch$','-benchtime','30x','-count','3','-benchmem','./internal/researchpublichybrid']);
for(const[p,v]of Object.entries(inputs))assert.equal(await sha(p),v.sha256);
const artifacts={};for(const p of ['freeze.json','commands.json','predictor','predictions.ndjson',...commands.map(c=>c.name+'.log')])artifacts[p]=await sha(root+'/'+p);
fs.writeFileSync(root+'/manifest.json',JSON.stringify({...freeze,artifacts,commands:commands.length,allCommandsTerminal:true,calibrationPredictions:0,confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,sources:Object.keys(sources).length,commands:commands.length,goal:'ACTIVE'}));
