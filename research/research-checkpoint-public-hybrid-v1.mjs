// Preserve new positive and negative evidence without duplicating large models
// or modifying source runs. Hash-bound external links are not self-contained.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-public-hybrid-v1-executed';
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const read=p=>JSON.parse(fs.readFileSync(p));
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const previous='research/checkpoint-2026-10-04-public-native-v1/manifest.json';assert.equal(read(previous).goal,'ACTIVE');
const resume='research/public-task-pilot/scifact-native-resume-v2',frontier='research/public-task-pilot/scifact-native-frontier-v3-executed',hybrid='research/public-task-pilot/scifact-hybrid-v1';
const r=read(resume+'/failure-audit.json'),f=read(frontier+'/audit-results.json'),h=read(hybrid+'/audit-results.json');
for(const [d,a]of [[resume,r],[frontier,f],[hybrid,h]])assert.equal(await sha(d+'/manifest.json'),a.manifestSHA256);
assert.equal(r.actualVerifiedPrefix,4250);assert.equal(r.newAcknowledgedImports,933);assert.equal(r.fitQueriesCompleted,0);assert(r.nativeMemoryLeakOrPanicRootCauseNotFixed);
assert(f.all5183StoredRecordsVerified&&f.nativeRankFailureNotReclassifiedAsSuccess);assert.equal(f.queries,351);
assert.equal(h.independentQueryReconstructions,351);assert.equal(h.calibrationPredictions,0);assert.equal(h.confirmationPredictions,0);
const status=execFileSync('git',['status','--short','--untracked-files=no'],{encoding:'utf8'});
const expected=['.learnings/ERRORS.md','internal/frame/turn.go','internal/model/api.go','internal/model/capture.go','internal/productioneval/codex.go','internal/productioneval/codex_test.go','internal/retrieval/libravdb.go','internal/service/service.go','internal/store/libravdbstore/store.go','plugin/openclaw.plugin.json','plugin/src/event.ts','plugin/src/index.test.ts','plugin/src/index.ts','plugin/src/types.ts'];
assert.deepEqual(status.split('\n').filter(Boolean).map(l=>l.slice(3)).sort(),expected.sort());
const trackedHashes={};for(const p of expected)trackedHashes[p]=await sha(p);
fs.mkdirSync(root,{mode:0o700});const copies=[],links=[],seen=new Set(),omittedNonFiles=[];
async function save(p){const abs=path.resolve(p);if(seen.has(abs))return;seen.add(abs);const ls=fs.lstatSync(p),st=fs.statSync(p);assert(st.isFile(),'file target only');
 const bytes=st.size,s=await sha(p);
 if(ls.isSymbolicLink()||bytes>8*1024*1024||!abs.startsWith(process.cwd()+'/')){links.push({path:p,sha256:s,bytes,...(ls.isSymbolicLink()?{resolvedTarget:fs.realpathSync(p)}:{})});return;}
 const target='saved/'+path.relative(process.cwd(),abs);fs.mkdirSync(path.dirname(root+'/'+target),{recursive:true,mode:0o700});fs.copyFileSync(p,root+'/'+target,fs.constants.COPYFILE_EXCL);fs.chmodSync(root+'/'+target,0o600);
 assert.equal(await sha(root+'/'+target),s);copies.push({path:p,copy:target,sha256:s,bytes});}
async function walk(d){for(const e of fs.readdirSync(d,{withFileTypes:true})){const p=d+'/'+e.name;if(e.isDirectory())await walk(p);else if(e.isFile())await save(p);else omittedNonFiles.push(p);}}
for(const d of [resume,frontier,hybrid,'research/public-task-pilot/scifact-native-resume-v2-preflight','research/public-task-pilot/scifact-native-frontier-v3-gen-preflight','research/public-task-pilot/scifact-native-frontier-v3','research/public-task-pilot/nativeresume-v2','research/public-task-pilot/nativefrontier-v3','research/checkpoint-2026-10-04-public-hybrid-v1'])await walk(d);
for(const d of ['research/public-task-pilot/nativeresume-v2','research/public-task-pilot/nativefrontier-v3']){
 const m=read(d+'/manifest.json');assert(m.allOwnedJobsTerminal);const term=read(d+'/daemon-terminal.json');assert.equal(term.code,0);
 for(const[p,v]of Object.entries(m.inputs)){assert.equal(await sha(p),v.sha256);await save(p);}
}
for(const[p,v]of Object.entries(read(hybrid+'/manifest.json').inputs)){assert.equal(await sha(p),v.sha256);await save(p);}
for(const d of ['internal/researchpublicresume','internal/researchpublicrankguard','internal/researchpublichybrid','cmd/research-public-native-resume','cmd/research-public-native-frontier','cmd/research-public-hybrid'])await walk(d);
for(const n of fs.readdirSync('research/public-task-pilot').filter(n=>/^scifact-(native-resume-v2|native-frontier-v3|hybrid-v1)/.test(n)&&/\.(mjs|json)$/.test(n)))await save('research/public-task-pilot/'+n);
for(const p of ['research-direction.md','docs/experiments/scifact-native-resume-v2-protocol.md','docs/experiments/scifact-native-resume-v2-results.md','docs/experiments/scifact-native-frontier-v3-protocol.md','docs/experiments/scifact-native-frontier-v3-results.md','docs/experiments/scifact-hybrid-v1-protocol.md','docs/experiments/scifact-hybrid-v1-results.md','research/research-checkpoint-public-hybrid-v1.mjs','research/public-task-pilot/scifact-v1/split-plan.json'])await save(p);
for(const e of copies)assert.equal(await sha(root+'/'+e.copy),e.sha256);for(const e of links)assert.equal(await sha(e.path),e.sha256);
for(const[p,h]of Object.entries(trackedHashes))assert.equal(await sha(p),h);assert.equal(execFileSync('git',['status','--short','--untracked-files=no'],{encoding:'utf8'}),status);
fs.writeFileSync(root+'/tracked-status-before-after.txt',status,{flag:'wx',mode:0o600});
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await sha(previous)},copies,links,omittedNonFiles,trackedHashes,
 completeLocalTestDependencyClosures:{resume:36,frontier:39,hybrid:10},coreTerminalCommands:13,allRequiredJobsTerminal:true,
 actualStoreReadback:5183,fitQueries:351,fitLabelsConsumed:true,calibrationPredictions:0,confirmationPredictions:0,
 preservedNativeImportAndRankFailures:true,replayedHybridNotLiveServing:true,fusionOrderingNegativePreserved:true,
 previousResearchGoalTurn:'PROGRESS',thisGoalTurn:'PROGRESS',goals:Object.fromEntries([1,2,3,4,5,6,7].map(k=>[k,'OPEN'])),goal:'ACTIVE',weeklyUsage:usage,
 productionPrivateCorporaWhitepaperUntouched:true,noCommitPushDeployment:true,notSelfContainedLargeLinkedAssets:true,
 generatorReceiptCaveat:'Frontier-v3 run-generation receipt hashes initial generated runner; final prelaunch manual corrections are preserved in actual executed source freeze.',
 next:'Source-bound real pre-packing lexical hydration/common200frontier adapter; bounded learned correction versus stronger BM25 control, then frozen untouched tasks. Delayed-score alignment and loaded freshness/admission remain viable.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({root,copies:copies.length,links:links.length,sourceClosures:manifest.completeLocalTestDependencyClosures,calibrationPredictions:0,confirmationPredictions:0,weeklyUsage:usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE'}));
