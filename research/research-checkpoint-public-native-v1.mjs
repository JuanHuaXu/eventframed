// Preserve full evidence, including failed native starts/imports, without
// duplicating immutable large public models or rewriting original artifacts.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const read=p=>JSON.parse(fs.readFileSync(p));
const root='research/checkpoint-2026-10-04-public-native-v1';
const pool='research/public-task-pilot/scifact-pool-v1',fit='research/public-task-pilot/scifact-native-fit-v1';
const p=read(pool+'/artifact-audit.json'),f=read(fit+'/failure-audit.json');
assert.equal(p.sources,171);assert.equal(p.commands,6);assert.equal(p.nativePilot.futurePositiveAndNegativeControls,true);
assert.equal(p.manifestSHA256,await sha(pool+'/manifest.json'));
assert.equal(f.manifestSHA256,await sha(fit+'/manifest.json'));assert.equal(f.fitPredictions,0);assert(f.partialCorpusNotScored&&f.rootCauseNotProven);
const previous='research/checkpoint-2026-10-04-public-frame-v1/manifest.json';assert.equal(read(previous).goal,'ACTIVE');
const status=execFileSync('git',['status','--short','--untracked-files=no'],{encoding:'utf8'});
assert.equal(status.split('\n').filter(Boolean).length,14);
const expected=['.learnings/ERRORS.md','internal/frame/turn.go','internal/model/api.go','internal/model/capture.go',
 'internal/productioneval/codex.go','internal/productioneval/codex_test.go','internal/retrieval/libravdb.go','internal/service/service.go',
 'internal/store/libravdbstore/store.go','plugin/openclaw.plugin.json','plugin/src/event.ts','plugin/src/index.test.ts','plugin/src/index.ts','plugin/src/types.ts'];
assert.deepEqual(status.split('\n').filter(Boolean).map(l=>l.slice(3)).sort(),expected.sort());
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
fs.mkdirSync(root,{mode:0o700});
const copies=[],links=[],seen=new Set();
async function save(p){
 const abs=path.resolve(p);if(seen.has(abs))return;seen.add(abs);
 const bytes=fs.statSync(p).size;if(bytes>8*1024*1024||!abs.startsWith(process.cwd()+'/')){
  links.push({path:p,sha256:await sha(p),bytes});return;
 }
 const rel=path.relative(process.cwd(),abs),dest='saved/'+rel,b=fs.readFileSync(p);
 fs.mkdirSync(path.dirname(root+'/'+dest),{recursive:true,mode:0o700});fs.writeFileSync(root+'/'+dest,b,{flag:'wx',mode:0o600});
 copies.push({path:p,copy:dest,sha256:hash(b),bytes});
}
async function walk(d){for(const e of fs.readdirSync(d,{withFileTypes:true})){const p=d+'/'+e.name;if(e.isDirectory())await walk(p);else if(e.isFile())await save(p);}}
const roots=[pool,fit,'research/public-task-pilot/native-loader-probe-v1','research/public-task-pilot/nativefit-v1',
 ...Array.from({length:6},(_,i)=>'research/public-task-pilot/nativepool-v'+(i+1))];
for(const r of roots)await walk(r);
for(const r of roots.filter(r=>/native(pool-v|fit-v)/.test(r))){
 const m=read(r+'/manifest.json');assert(m.allOwnedJobsTerminal);
 for(const[input,s]of Object.entries(m.inputs)){await save(input);const saved=[...copies,...links].find(e=>path.resolve(e.path)===path.resolve(input));assert(saved);assert.equal(saved.sha256,s.sha256);}
}
const scripts=fs.readdirSync('research/public-task-pilot').filter(n=>/^scifact-(pool|native)/.test(n)&&/\.(mjs|json|yaml)$/.test(n));
for(const n of scripts)await save('research/public-task-pilot/'+n);
for(const p of ['research-direction.md','docs/experiments/scifact-pool-native-v1-results.md','docs/experiments/scifact-native-fit-v1-results.md',
 'docs/experiments/scifact-native-fit-v1-protocol.md','docs/experiments/scifact-pool-v1-protocol.md','docs/experiments/scifact-native-quant-v4-protocol.md',
 'cmd/research-public-native-fit/main.go','cmd/research-public-native-fit/main_test.go','research/research-checkpoint-public-native-v1.mjs',
 'research/public-task-pilot/scifact-v1/corpus.json','research/public-task-pilot/scifact-v1/split-plan.json',
 'research/public-task-pilot/scifact-frame-v1-closure/frames.json'])await save(p);
for(const e of copies)assert.equal(await sha(root+'/'+e.copy),e.sha256);
assert.equal(execFileSync('git',['status','--short','--untracked-files=no'],{encoding:'utf8'}),status);
fs.writeFileSync(root+'/tracked-status-before-after.txt',status,{flag:'wx',mode:0o600});
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await sha(previous)},copies,links,weeklyUsage:usage,
 completeTestDependencyClosures:{pool:171,fit:30},coreTerminalCommands:10,nativePilotsTerminal:6,fullNativeDiagnosticTerminal:true,
 additionalIndependentAudits:{pool:path.resolve(pool+'/artifact-audit.json'),fitFailure:path.resolve(fit+'/failure-audit.json')},
 allRequiredJobsTerminal:true,previousResearchGoalTurn:'PROGRESS',thisGoalTurn:'PROGRESS',
 goals:Object.fromEntries([1,2,3,4,5,6,7].map(k=>[k,'OPEN'])),goal:'ACTIVE',
 qualityExperimentCompleted:false,fitPredictions:0,calibrationPredictions:0,confirmationPredictions:0,
 preservedFullCorpusFailure:true,productionPrivateCorporaWhitepaperUntouched:true,noCommitPushDeployment:true,
 next:'Separate bounded offline native import/recovery in fresh owned copies with actual stored-prefix verification, same full corpus/RSS/contracts and counted restarts; all-head alignment and loaded admission/persistence leads also remain.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,copies:copies.length,links:links.length,sourceClosures:manifest.completeTestDependencyClosures,
 nativePilotsTerminal:6,fitImportFailurePreserved:true,fitPredictions:0,weeklyUsage:usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE'}));
