// Archive an equivalent-computation lead, not a narrower success criterion.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const root='research/checkpoint-2026-10-05-regime-incremental-v80';assert(!fs.existsSync(root));
const json=p=>JSON.parse(fs.readFileSync(p));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const parentPath='research/checkpoint-2026-10-05-regime-ledger-v79/manifest.json';
assert.equal(await hash(parentPath),'c5582c55cf000fbfa2d34d91c1da54f7a6e06fdd0493f2113f403b162b05c839');
const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
for(const old of ['research/regime-v77-initial','research/regime-frozen-v78-initial','research/regime-frozen-v78-long-initial','research/regime-ledger-v79-bounded-audit']) {
  for(const[p,h]of Object.entries(json(old+'/freeze.json').files))assert.equal(await hash(p),h,p);
}
const run='research/regime-incremental-v80-initial',f=json(run+'/freeze.json'),c=json(run+'/completed.json'),r=json(run+'/readback.json');
assert(c.allJobsTerminal&&c.sourceUnchanged&&c.allChecksPass&&c.checks.length===4);
for(const[p,h]of Object.entries(f.files)) {assert.equal(await hash(p),h,p);assert.equal(await hash(run+'/source/'+p.replace(/^\//,'')),h,p)}
for(const x of c.checks){assert.equal(x.exitCode,0);assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256)}
assert.equal(r.sourceSHA256,await hash('research/regime-incremental-v80-readback.mjs'));assert.equal(r.benchmarkSHA256,await hash(run+'/benchmark.log'));assert.equal(r.auditSHA256,await hash(run+'/audit.json'));
assert.equal(r.audit.Checks,143038);assert.equal(r.audit.IncrementalChecks,10364);assert.equal(r.audit.MaxReference,0);assert.equal(r.audit.CacheChecks,214);
assert(r.audit.Full64&&r.audit.FullWorkload&&r.audit.OldReplay&&!r.audit.WholeGoalValidation);
assert(r.audit.MaxOracle<2e-11&&r.audit.MaxTower<2e-11&&r.audit.MaxReceipt<2e-11);
assert.equal(r.benchmarkRows,132);assert.equal(Object.keys(r.groups.researchregimeledger).length,32);assert.equal(Object.keys(r.groups.researchregimeincremental).length,34);
assert(!r.scientificQualityRescueEstablished&&!r.equalTotalCostSuperiorityEstablished&&!r.loadedServingEstablished&&!r.wholeMemoryLimitCertified);
fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(p){const out=root+'/saved/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[p]=await hash(p);assert.equal(await hash(out),copies[p]);}
async function tree(d){for(const n of fs.readdirSync(d)){const p=d+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
await tree(run);await tree('internal/researchregimeincremental');
for(const n of fs.readdirSync('research').filter(n=>/^regime-incremental-v80.*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['go.mod','go.sum','research-direction.md','research/regime-v77-v78-next.md','docs/experiments/mmm-regime-incremental-v80-protocol.md','docs/experiments/mmm-regime-incremental-v80-results.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,trackedHashes:parent.trackedHashes,
  goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,completedSHA256:copies[run+'/completed.json'],
  previousGoalTurn:'PROGRESS(V79 canonical support/law ledger and full-workload cost/unsupported-evidence audits)',
  thisTurn:'PROGRESS(isolated equivalent forward/suffix replay, immutable four-prefix cache and old-evidence fallback; dense/reference/as-of/cache audits; full150/200-member same-command comparative cost)',
  componentAudit:r.audit,costReadback:r,archiveIsChainedNotStandalone:true,scientificAdoption:false,allSevenEmpiricalGoalsValidated:false,
  productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,
  limitations:['working-law approximation/noisy-support failures remain','no new broad recovery/non-harm or useful downstream split result','timing serial prepared-state, not growing ingestion/full round/loaded serving','old candidate cost not paired with an independently timed old reference','payload and total allocations not peak/RSS/full allowed-maximum bound','full publication fingerprint still O(T*K)','untouched agent outcomes/durable mixed-write freshness/equal total cost still required'],
  next:'Fully bound incremental identity/full resources and positive-support/abstention rescue; then unchanged controlled varied-generator/noise/delay/recovery/non-harm/equal-TOTAL-cost gates.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
