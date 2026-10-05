import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/checkpoint-2026-10-04-paired-v55',run='research/paired-v55-rescue',failed='research/paired-v55-diagnostic';
assert(!fs.existsSync(root));const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');}
const previous='research/checkpoint-2026-10-04-paired-v54/manifest.json';assert.equal(await fileHash(previous),'af17d510571473b5e70506eaae191675a6043b3b74d7dc73dd6ad278e13a2765');
const old=JSON.parse(fs.readFileSync(previous)),freeze=JSON.parse(fs.readFileSync(run+'/freeze.json')),completed=JSON.parse(fs.readFileSync(run+'/completed.json')),readback=JSON.parse(fs.readFileSync(run+'/readback.json')),equivalence=JSON.parse(fs.readFileSync(run+'/equivalence.json'));
assert.deepEqual(freeze.protectedFiles,old.trackedHashes);assert.equal(completed.checks.length,7);assert(completed.checks.every(c=>c.exitCode===0));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles}))assert.equal(await fileHash(p),h,p);
for(const[p,h]of Object.entries(completed.artifacts))assert.equal(await fileHash(run+'/'+p),h,p);
for(const[p,h]of Object.entries(old.copies))assert.equal(await fileHash(path.dirname(previous)+'/saved/'+p),h,p);
const badFreeze=JSON.parse(fs.readFileSync(failed+'/freeze.json')),bad=JSON.parse(fs.readFileSync(failed+'/failure.json'));assert.equal(bad.checks.at(-1).name,'vet');assert.equal(bad.checks.at(-1).exitCode,1);assert(!fs.existsSync(failed+'/diagnostic.jsonl'));
for(const[p,h]of Object.entries(badFreeze.files))assert.equal(await fileHash(failed+'/source/'+p),h,p);
assert(readback.sourceClosureVerified);assert.equal(readback.arms,840);assert.equal(readback.distinctUnderlyingOutcomes,96000);assert.equal(equivalence.totals.arms,840);
const unitRoot='research/paired-v56-unit-preflight',unit=JSON.parse(fs.readFileSync(unitRoot+'/completed.json')),unitFreeze=JSON.parse(fs.readFileSync(unitRoot+'/freeze.json'));assert(unit.commands.length===2&&unit.commands.every(c=>c.exitCode===0)&&!unit.fullExperimentRun&&!unit.performanceMeasured);
for(const[p,h]of Object.entries(unitFreeze.files))assert.equal(await fileHash(p),h,p);
fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(src){const dest=root+'/saved/'+src;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(src,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);copies[src]=await fileHash(src);assert.equal(await fileHash(dest),copies[src]);}
for(const p of Object.keys(freeze.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
for(const p of fs.readdirSync(failed).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(failed+'/'+p);
for(const p of Object.keys(badFreeze.files))await copy(failed+'/source/'+p);
for(const p of['docs/experiments/mmm-paired-v55-results.md','research-direction.md','research/paired-v55-checkpoint.mjs','research/paired-v55-post-audit.md','research/paired-v56-module-gen.mjs','research/paired-v56-module-generation.json','internal/researchpairedmemo/model.go','internal/researchpairedmemo/model_test.go','internal/researchpairedmemo/reference_test.go','internal/researchpairedmemo/paths_test.go','internal/researchpairedmemo/cache_test.go'])await copy(p);
await copy('research/paired-v56-unit-preflight.mjs');for(const p of fs.readdirSync(unitRoot).filter(p=>/\.(json|log)$/.test(p)))await copy(unitRoot+'/'+p);
for(const[p,h]of Object.entries(unitFreeze.files))assert.equal(copies[p],h,p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await fileHash(previous),copiedArtifactsVerified:Object.keys(old.copies).length},copies,trackedHashes:freeze.protectedFiles,frozenCompilerInputs:freeze.compilerFiles.length,frozenSourceFiles:Object.keys(freeze.files).length,terminalCommands:completed.checks.length,failedPreExperimentFreezePreserved:true,allRequiredJobsTerminal:true,worlds:40,arms:840,distinctUnderlyingOutcomes:96000,consumedV54Populations:true,executionEquivalent:equivalence.executionEquivalent,executionDifferences:equivalence.totals,diagnosticOnly:true,nPerCell:1,qualityAdoption:false,equalTotalCostSuperiorityEstablished:false,previousGoalTurn:'PROGRESS (full paired experiment, independent negative result and checkpoint)',thisGoalTurn:'PROGRESS',weeklyUsage:{usedPercent:usage,windowDurationMins:10080,recordedAt:new Date().toISOString()},goals:['OPEN','OPEN','OPEN','OPEN','OPEN','OPEN','OPEN'],goal:'ACTIVE',productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,prospectiveRepositoryCompilerSourceClosureComplete:true,archiveIsChainedNotStandalone:true,next:'Continue from the full cost/equivalence outcome without weakening gates; preserve V54 broad failures and untouched task cohorts'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
for(const[p,h]of Object.entries(copies))assert.equal(await fileHash(root+'/saved/'+p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await fileHash(root+'/manifest.json'),copies:Object.keys(copies).length,parentVerified:true,weeklyUsage:usage,executionEquivalent:equivalence.executionEquivalent,wholeGoals:'OPEN',goal:'ACTIVE'},null,2));
