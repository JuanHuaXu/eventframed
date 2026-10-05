// Preserve positive and negative evidence without claiming whole-goal success.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const root='research/checkpoint-2026-10-05-regime-protected-v81';assert(!fs.existsSync(root));
const json=p=>JSON.parse(fs.readFileSync(p));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const parentPath='research/checkpoint-2026-10-05-regime-incremental-v80/manifest.json';
assert.equal(await hash(parentPath),'d2ae1995bfaa0caff9a2ddee5864847606fd532ec1ae7cd27e16d5a50fcd965e');
const parent=json(parentPath);for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
assert.deepEqual(execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort(),Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
for(const old of ['research/regime-v77-initial','research/regime-frozen-v78-initial','research/regime-frozen-v78-long-initial','research/regime-incremental-v80-initial'])for(const[p,h]of Object.entries(json(old+'/freeze.json').files))assert.equal(await hash(p),h,p);
const runs=['research/regime-protected-v81-initial','research/regime-protected-v81-screen-initial'];
for(const run of runs){
 const f=json(run+'/freeze.json'),c=json(run+'/completed.json');assert(c.allJobsTerminal&&c.sourceUnchanged&&c.allChecksPass&&c.checks.length===4);
 for(const[p,h]of Object.entries(f.files)){assert.equal(await hash(p),h,p);assert.equal(await hash(run+'/source/'+p.replace(/^\//,'')),h,p)}
 for(const x of c.checks){assert.equal(x.exitCode,0);assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256)}
}
const audit=json(runs[0]+'/readback.json'),screen=json(runs[1]+'/readback.json');
assert.equal(audit.sourceSHA256,await hash('research/regime-protected-v81-readback.mjs'));assert.equal(audit.benchmarkSHA256,await hash(runs[0]+'/benchmark.log'));assert.equal(audit.auditSHA256,await hash(runs[0]+'/audit.json'));
assert.equal(audit.audit.Checks,144812);assert.equal(audit.audit.IncrementalChecks,10364);assert.equal(audit.audit.CacheChecks,214);assert.equal(audit.audit.ProtectedChecks,2592);
assert.equal(audit.audit.MaxReference,0);assert(audit.audit.NoResetUnderflow&&audit.audit.ExactLongPairLogProbability<-1000);
assert.equal(screen.sourceSHA256,await hash('research/regime-protected-v81-screen-readback.mjs'));assert.equal(screen.completedSHA256,await hash(runs[1]+'/completed.json'));
for(const[p,h]of Object.entries(json(runs[1]+'/completed.json').dataHashes))assert.equal(await hash(runs[1]+'/data/'+p),h,p);
assert.equal(screen.forecasts,115200);assert.equal(screen.attempts,172800);assert.equal(screen.paired.length,48);
assert(screen.maxLossError<2e-14&&screen.maxMetadataError<2e-15);
assert(!screen.screen.rejectQualityRescueInterpretation);assert.equal(screen.screen.shiftedPositiveCases,36);assert.equal(screen.aggregate.v81_protected36.rejectedSecond,0);assert.equal(screen.aggregate.v80_top36.rejectedSecond,1679);
assert(screen.goals.every(x=>x==='OPEN')&&!screen.scientificAdoption);
fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(p){const out=root+'/saved/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[p]=await hash(p);assert.equal(await hash(out),copies[p])}
async function tree(d){for(const n of fs.readdirSync(d)){const p=d+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const run of runs)await tree(run);await tree('internal/researchregimeprotected');await tree('cmd/research-regime-protected-screen');
for(const n of fs.readdirSync('research').filter(n=>/^regime-protected-v81.*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['go.mod','go.sum','research-direction.md','research/regime-v77-v78-next.md','docs/experiments/mmm-regime-protected-v81-protocol.md','docs/experiments/mmm-regime-protected-v81-results.md','docs/experiments/mmm-regime-protected-v81-screen-protocol.md','docs/experiments/mmm-regime-protected-v81-screen-results.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,trackedHashes:parent.trackedHashes,
 goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,
 previousResearchGoalTurn:'PROGRESS(frozen V81 mathematical preflight started); intervening social clarification NO_RESEARCH_PROGRESS',
 thisTurn:'PROGRESS(verified V81 terminal unit/race/vet/benchmark; frozen full48 public development trajectories; independent all-loss/packet readback; numerical negative preserved; cost measured)',
 componentAudit:audit.audit,screenAudit:screen,archiveIsChainedNotStandalone:true,scientificAdoption:false,allSevenEmpiricalGoalsValidated:false,
 productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,
 preservedReadbackFailures:['null empty packet slice','last-bit cross-language metadata arithmetic'],
 limitations:['cap9 forced memory loss','reset0 long positive event rejected through underflow','support positivity not truncation-error or AP authority certificate','known synthetic baseline, no independent fit/missingness/interactions/native Full/Adaptive or untouched confirmation','same offered packets not same accepted-label or equal TOTAL-cost budget','whole core +10.8%,3.18s mean at800slots; microbenchmarks do not pass400ms/loaded latency','payload and cumulative allocations not peak/RSS/full bound','external useful splits/agent utility/durable mixed-epoch freshness/equal cost remain required'],
 next:'Log-domain persistent component/query contract and fully bound incremental identity/storage; separately frozen unchanged original population/native controls, recovery/non-harm/resources, then remaining whole-goal evidence.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
