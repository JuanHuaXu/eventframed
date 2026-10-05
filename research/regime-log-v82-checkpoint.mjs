import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const root='research/checkpoint-2026-10-05-regime-log-v82';assert(!fs.existsSync(root));const json=p=>JSON.parse(fs.readFileSync(p));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const parentPath='research/checkpoint-2026-10-05-regime-protected-v81/manifest.json';assert.equal(await hash(parentPath),'5d2a36f6311a98ab1d4bb1b3199247205b73e0e9e500205165b7e321b30a5dbf');const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
assert.deepEqual(execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort(),Object.keys(parent.trackedHashes).sort());for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
for(const old of ['research/regime-incremental-v80-initial','research/regime-protected-v81-initial','research/regime-protected-v81-screen-initial'])for(const[p,h]of Object.entries(json(old+'/freeze.json').files))assert.equal(await hash(p),h,p);
const runs=['research/regime-log-v82-initial','research/regime-log-v82-bound-diagnostic','research/regime-log-v82-bounds-repair'];
for(const run of runs){
 const f=json(run+'/freeze.json'),c=json(run+'/completed.json');assert(c.allJobsTerminal&&c.sourceUnchanged);
 for(const[p,h]of Object.entries(f.files)){assert.equal(await hash(run+'/source/'+p.replace(/^\//,'')),h,p);if(run.endsWith('bounds-repair'))assert.equal(await hash(p),h,p)}
 if(c.checks)for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256);else{assert.equal(c.exitCode,0);assert.equal(await hash(run+'/diagnostic.log'),c.logSHA256)}
}
const failed=json(runs[0]+'/completed.json'),done=json(runs[2]+'/completed.json'),r=json(runs[2]+'/readback.json');assert(!failed.allChecksPass&&failed.checks[0].exitCode===1);assert(done.allChecksPass&&done.checks.length===4&&done.checks.every(c=>c.exitCode===0));
assert.equal(r.sourceSHA256,await hash('research/regime-log-v82-readback.mjs'));assert.equal(r.benchmarkSHA256,await hash(runs[2]+'/benchmark.log'));assert.equal(r.auditSHA256,await hash(runs[2]+'/audit.json'));
assert(r.audit.LongLogRescue&&r.audit.MemberOddsRescue&&r.audit.MaxOracle<2e-11&&r.audit.MaxLogOracle<5e-8);assert.equal(r.audit.EnvelopeOracleChecks,280);assert.equal(r.audit.MaxOldEnvelopeDifference,1);
fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(p){const o=root+'/saved/'+p;fs.mkdirSync(path.dirname(o),{recursive:true,mode:0o700});fs.copyFileSync(p,o,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(o,0o600);copies[p]=await hash(p);assert.equal(await hash(o),copies[p])}
async function tree(d){for(const n of fs.readdirSync(d)){const p=d+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const run of runs)await tree(run);await tree('internal/researchregimelog');await tree('cmd/research-regime-log-bound-diagnostic');
for(const n of fs.readdirSync('research').filter(n=>/^regime-log-v82.*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['go.mod','go.sum','research-direction.md','research/regime-v77-v78-next.md','docs/experiments/mmm-regime-log-v82-protocol.md','docs/experiments/mmm-regime-log-v82-results.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,trackedHashes:parent.trackedHashes,goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,
 previousGoalTurn:'PROGRESS(V81 protected-support controlled screen and cost)',thisTurn:'PROGRESS(log-state/rare-query and member-endpoint rescue; failed bounds parity preserved/diagnosed; independent dense-TV checks and same-command performance)',
 componentAudit:r.audit,costReadback:r,archiveIsChainedNotStandalone:true,scientificAdoption:false,allSevenEmpiricalGoalsValidated:false,productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,
 limitations:['old envelope numeric parity FAIL remains preserved','actual small-state approximationTV max.698246 despite coherence/positivity','prepared log costs worse than V81, not whole-core or loaded latency','no new varied fitting/native Full/Adaptive/recovery/non-harm or agent/externally useful split result','no certified FP/target-law AP coverage','durable mixed-write freshness and equal TOTAL-cost observation remain required'],
 next:'Test output-only summary/empty-initialization cost rescue without altering persistent log state; then original population/native controls, recovery/non-harm/resources and remaining whole-goal evidence.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',goals:'seven OPEN',weeklyUsage:usage},null,2));
