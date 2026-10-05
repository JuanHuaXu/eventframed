// Preserve the real component pass, remaining200-member failure and next lead.
// Archive is chained; no git commit, publishing, deployment or goal closure.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-retention-v62',run='research/retention-v62-integration';
const previous='research/checkpoint-2026-10-04-tree-v60/manifest.json',usage=Number(process.argv[2]);
assert(!fs.existsSync(root));assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));
assert.equal(await hash(previous),'0bc6cdb1eb7f5eb0c5c1e92e1f7f799237e075d116276fde7a29c76b6f963240');
const parent=json(previous),freeze=json(run+'/freeze.json'),complete=json(run+'/completed.json'),allocation=json(run+'/allocation.json');
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(previous)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(previous)+'/'+p),h,p);
assert.deepEqual(freeze.protectedFiles,parent.trackedHashes);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(freeze.protectedFiles).sort());assert.equal(tracked.length,14);
assert.equal(freeze.compilerFiles.length,27);assert.equal(Object.keys(freeze.files).length,39);
assert.deepEqual(complete.checks.map(c=>c.name),['race','vet','allocation','benchmark']);assert(complete.checks.every(c=>c.exitCode===0));assert(complete.allJobsTerminal);
assert(!fs.existsSync(run+'/failure.json'));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(freeze.files))assert.equal(await hash(run+'/source/'+p),h,p);
for(const[p,h]of Object.entries(complete.artifacts))assert.equal(await hash(run+'/'+p),h,p);
for(const c of complete.checks)assert.equal(await hash(run+'/'+c.name+'.log'),c.logSHA256);
for(const[p,h]of Object.entries(freeze.generatedCompilerCopies))assert.equal(await hash(run+'/'+p),h,p);
assert(allocation.AllocatedBytes.shared150<allocation.CapBytes);assert(allocation.AllocatedBytes.shared200>allocation.CapBytes);
assert(allocation.AllocatedBytes.unshared150>allocation.CapBytes);assert(allocation.AllocatedBytes.unshared200>allocation.CapBytes);
const race=fs.readFileSync(run+'/race.log','utf8');assert.equal((race.match(/^--- PASS: /gm)||[]).length,23);
assert(race.includes('independent bank forecast comparisons 25758'));assert(race.includes('exact path/reference checks 2000'));assert(race.includes('independent forecasts 4656 prediction values 120'));
const bench=fs.readFileSync(run+'/benchmark.log','utf8');
const loops=[...bench.matchAll(/^BenchmarkBankBounded2400Loop-\d+\s+\d+\s+([\d.]+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op/gm)].map(x=>({elapsedNS:Number(x[1]),allocatedBytes:Number(x[2]),allocations:Number(x[3])}));
assert.equal(loops.length,3);assert(loops.every(x=>x.elapsedNS>0&&x.elapsedNS<400e6));
const preflight=json('research/retention-v62-preflight/failure.json');assert.equal(preflight.failedSourceSHA256,await hash('research/retention-v62-preflight/failed-bank-test.go'));

fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(src){if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}
 const dest=root+'/saved/'+src;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(src,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);copies[src]=await hash(src);assert.equal(await hash(dest),copies[src])}
for(const p of Object.keys(freeze.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|log)$/.test(p)))await copy(run+'/'+p);
for(const p of ['docs/experiments/mmm-retention-v62-results.md','research/retention-v63-journal-direction.md','research-direction.md','research/retention-v62-checkpoint.mjs'])await copy(p);
const generatedCompilerCopies={};for(const[p,h]of Object.entries(freeze.generatedCompilerCopies)){const dest=root+'/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);assert.equal(await hash(dest),h);generatedCompilerCopies[p]=h}
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await hash(previous),copiedArtifactsVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,
 trackedHashes:freeze.protectedFiles,frozenCompilerInputs:27,frozenSourceFiles:39,terminalCommands:4,allRequiredJobsTerminal:true,
 scientificAdoption:false,allSevenEmpiricalGoalsValidated:false,selectorConnectedToActualScoredLaw:true,independentBankForecastComparisons:25758,exactPathComparisons:2000,
 legacyReferenceForecasts:4656,legacyPredictionValues:120,functionalRaceRoots:23,constructorAllocation:allocation,nominal150MemberMemoryPass:true,full200MemberMemoryPass:false,
 fixedNominationCoreLoop:loops,notFullExperimentOrLoadedServing:true,noQualityRecoveryOrEqualTotalCostResult:true,
 previousGoalTurn:'PROGRESS(V60 rejection and V61 selector/joint-adapter components)',
 thisGoalTurn:'PROGRESS(shared immutable tables,actual joint extraction,scored mixture,atomic bank integration,independent delayed/expiry/future audits,race/vet/allocation/core cost,preserved failure,next200-member memory lead)',
 goals:Array(7).fill('OPEN'),goal:'ACTIVE',weeklyUsage:{usedPercent:usage,windowDurationMins:10080,recordedAt:new Date().toISOString()},
 productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,
 archiveIsChainedNotStandalone:true,next:'Bank-owned canonical observed-event journal with child authority fences and complete V62 equivalence; then same-mixture acquisition and full controlled quality/recovery/cost, followed by all remaining seven-goal evidence'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(freeze.protectedFiles))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allJobsTerminal:true,goal:'ACTIVE',wholeGoals:'all seven OPEN',weeklyUsage:usage},null,2));
