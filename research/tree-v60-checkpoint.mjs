// Chained reproducible checkpoint: negative full study plus distinct next-model
// components. No scientific adoption, git commit/upload or production changes.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-tree-v60',run='research/tree-v60-diagnostic';
const previous='research/checkpoint-2026-10-04-tree-v59/manifest.json';
const usage=Number(process.argv[2]);assert(!fs.existsSync(root));assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const json=p=>JSON.parse(fs.readFileSync(p));
async function hash(p) {const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');}
assert.equal(await hash(previous),'6bff139a1658a268b7286dde9b27eef0ae5e8f416de7982857d1323bfb57780c');
const parent=json(previous),freeze=json(run+'/freeze.json'),complete=json(run+'/completed.json'),readback=json(run+'/readback.json');
assert.deepEqual(freeze.protectedFiles,parent.trackedHashes);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked,Object.keys(freeze.protectedFiles).sort());assert.equal(tracked.length,14);
assert.equal(freeze.compilerFiles.length,83);assert.equal(Object.keys(freeze.files).length,99);
assert.deepEqual(complete.checks.map(c=>c.name),['race-model','race-fixture','closure-helper','vet','microbench','allocation','experiment','audit']);
assert(complete.checks.every(c=>c.exitCode===0));assert(!fs.existsSync(run+'/failure.json'));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(freeze.files))assert.equal(await hash(run+'/source/'+p),h,p);
for(const[p,h]of Object.entries(complete.artifacts))assert.equal(await hash(run+'/'+p),h,p);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(previous)+'/saved/'+p),h,p);
assert.equal(readback.worlds,40);assert.equal(readback.arms,960);assert.equal(readback.distinctUnderlyingOutcomes,96000);assert(readback.sourceClosureVerified);
const comparisons=['comparison.json','comparison-v58.json','comparison-v59.json'].map(p=>json(run+'/'+p));
for(const c of comparisons){assert.equal(c.controlsBitwiseEqual,240);assert(c.allPopulationsIdentical)}
for(const s of Object.values(readback.summaries)){assert.equal(s.adaptiveHarmOver01,105);assert.equal(s.over400MS,0);assert(s.fullRiskGain<0&&s.recoveryGainAdaptive<0)}
assert.equal(readback.allocationBytes.modelAllocatedBytes,2716544);
const branch=json(run+'/branch-audit.json');assert.equal(branch.forecastChecks,108000);assert.equal(branch.sourceSHA256,await hash('research/tree-v60-branch-audit.mjs'));
const probes=[1200,2400].map(w=>json(run+'/window-probe-'+w+'.json'));
for(const p of probes){assert.equal(p.forecastChecks,108000);assert(p.forecastChecksAreProbabilityShapesNotPublished600Equality);assert.equal(p.sourceSHA256,await hash('research/tree-v60-window-probe.mjs'))}
const component='research/retention-v61-component',cf=json(component+'/freeze.json'),cc=json(component+'/completed.json');
assert.equal(cf.compilerFiles.length,3);assert.equal(Object.keys(cf.files).length,7);
assert.deepEqual(cc.checks.map(c=>c.name),['race','vet','benchmark']);assert(cc.checks.every(c=>c.exitCode===0));assert(!fs.existsSync(component+'/failure.json'));
for(const[p,h]of Object.entries({...cf.files,...cf.protectedFiles}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(cf.files))assert.equal(await hash(component+'/source/'+p),h,p);
for(const[p,h]of Object.entries(cc.artifacts))assert.equal(await hash(component+'/'+p),h,p);
for(const c of cc.checks)assert.equal(await hash(component+'/'+c.name+'.log'),c.logSHA256);
const performance=json(component+'/performance-audit.json');assert.equal(performance.Timings.length,12);
for(const[p,h]of Object.entries(performance.sources))assert.equal(await hash(p),h,p);
const law='research/retention-v61-law',lf=json(law+'/freeze.json'),lc=json(law+'/completed.json');
assert(lc.checks.every(c=>c.exitCode===0));assert(!fs.existsSync(law+'/failure.json'));
for(const[p,h]of Object.entries(lf.files)){assert.equal(await hash(p),h,p);assert.equal(await hash(law+'/source/'+p),h,p)}
for(const c of lc.checks)assert.equal(await hash(law+'/'+c.name+'.log'),c.logSHA256);
const preflight=json('research/retention-v61-preflight/failure.json');assert.equal(preflight.failedSourceSHA256,await hash('research/retention-v61-preflight/failed-run.mjs'));

fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(src) {
 if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}
 const dest=root+'/saved/'+src;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(src,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);
 copies[src]=await hash(src);assert.equal(await hash(dest),copies[src]);
}
for(const p of [...Object.keys(freeze.files),...Object.keys(cf.files),...Object.keys(lf.files),...Object.keys(performance.sources)])await copy(p);
for(const dir of [run,component,law,'research/retention-v61-preflight'])for(const p of fs.readdirSync(dir).filter(p=>/\.(json|jsonl|log|mjs|md)$/.test(p)))await copy(dir+'/'+p);
for(const p of ['docs/experiments/mmm-tree-v60-results.md','docs/experiments/mmm-retention-v61-component-results.md','research-direction.md','research/tree-v60-checkpoint.mjs',
 'research/tree-v60-parent-comparison.mjs','research/tree-v60-branch-audit.mjs','research/tree-v60-window-probe.mjs','research/tree-v60-window-probe-gen.mjs','research/tree-v60-window-probe-generation.json',
 'research/retention-v61-preflight-save.mjs'])await copy(p);
const generatedCompilerCopies={};
for(const[p,h]of Object.entries(cf.generatedCompilerFiles)){
 assert.equal(await hash(p),h);const name='generated/'+h+'-test-main.go',dest=root+'/'+name;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(p,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);assert.equal(await hash(dest),h);generatedCompilerCopies[name]=h;
}
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await hash(previous),copiedArtifactsVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,
 trackedHashes:freeze.protectedFiles,frozenCompilerInputs:83,frozenSourceFiles:99,terminalV60Commands:8,terminalV61ComponentCommands:3,terminalJointLawCommands:2,publicReadbackTerminal:true,
 allRequiredJobsTerminal:true,worlds:40,arms:960,distinctUnderlyingOutcomes:96000,consumedPopulations:true,controlsBitwiseEqualV57V58V59:240,scientificAdoption:false,
 predictionSummary:readback.summaries.predictive,predictionTotals:readback.totals.predictive,constructorAllocatedBytes:2716544,predictive400MSFailures:0,
 branch:{forecastChecks:branch.forecastChecks,summary:branch.summaries.predictive},retentionProbes:probes.map(p=>({window:p.retainedIssuedPositions,summary:p.summaries.predictive,scope:p.scope})),
 nextSelectorComponents:{exactPathComparisons:2000,selectorRaceRoots:7,jointLawRaceRoots:3,performance:performance.Timings,integratedBaseExperts:false,
  unsharedMemoryEstimate150Members:3*2716544+performance.Timings.find(x=>x.Name==='constructor_150').BytesPerOp,notLoadedServingOrEmpiricalRescue:true},
 goals:Array(7).fill('OPEN'),goal:'ACTIVE',populationConfirmationDispatched:false,privateOrSealedLabelsOpened:false,
 previousGoalTurn:'PROGRESS(V60 full implementation/collection/independent audit); intervening social clarification NO_RESEARCH_PROGRESS',
 thisGoalTurn:'PROGRESS(terminal handle revalidation,2,400-position probe,negative result documentation,bounded member-specific selector and coherent joint adapter,race/reference/performance tests)',
 weeklyUsage:{usedPercent:usage,windowDurationMins:10080,recordedAt:new Date().toISOString()},productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,
 archiveIsChainedNotStandalone:true,next:'Share immutable templates with law/lifecycle equivalence; expose real issued (Y,W1,W2) kernels; integrate scored mixture and its acquisition; full controls/quality/recovery/cost then all remaining seven-goal evidence'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(freeze.protectedFiles))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,goals:'all seven OPEN',goal:'ACTIVE',weeklyUsage:usage,allRequiredJobsTerminal:true},null,2));
