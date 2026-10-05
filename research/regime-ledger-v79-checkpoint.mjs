// Preserve successful coherence evidence and ALL failed attempts; no adoption.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const root='research/checkpoint-2026-10-05-regime-ledger-v79';assert(!fs.existsSync(root));
const json=p=>JSON.parse(fs.readFileSync(p));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const parentPath='research/checkpoint-2026-10-05-regime-v77-frozen-v78/manifest.json';
assert.equal(await hash(parentPath),'5aa836d99096fe6c843edc2a8d855e261d3e0a0a526b2d1d7c4d4bc75cd05249');
const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
for(const run of ['research/regime-v77-initial','research/regime-frozen-v78-initial','research/regime-frozen-v78-long-initial']) {
  for(const[p,h]of Object.entries(json(run+'/freeze.json').files))assert.equal(await hash(p),h,p);
}
const stages=['initial','repaired','latest-fork','bounded-audit'],attempts=[];
for(const stage of stages) {
  const run='research/regime-ledger-v79-'+stage,f=json(run+'/freeze.json'),c=json(run+'/completed.json');
  assert(c.allJobsTerminal&&c.sourceUnchanged&&c.checks.length>=1);
  assert(!c.scientificQualityRescueEstablished&&!c.scalableApproximationCertified&&!c.wholeCohortTested&&!c.loadedServingEstablished);
  for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p.replace(/^\//,'')),h,p);
  for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256);
  assert.equal(c.allChecksPass,c.checks.length===4&&c.checks.every(x=>x.exitCode===0));
  if(stage==='bounded-audit')for(const[p,h]of Object.entries(f.files))assert.equal(await hash(p),h,p);
  if(stage==='initial'||stage==='repaired'){assert.equal(c.checks.length,1);assert.equal(c.checks[0].exitCode,1)}
  if(stage==='latest-fork'){assert.equal(c.checks.length,4);assert(c.checks.slice(0,3).every(x=>x.exitCode===0));assert.equal(c.checks[3].exitCode,1)}
  attempts.push({stage,path:run,completedSHA256:await hash(run+'/completed.json'),checks:c.checks,allJobsTerminal:true,allChecksPass:c.allChecksPass});
}
assert(attempts.at(-1).allChecksPass);
const run='research/regime-ledger-v79-bounded-audit',readback=json(run+'/readback.json');
assert.equal(readback.sourceSHA256,await hash('research/regime-ledger-v79-readback.mjs'));
assert.equal(readback.benchmarkSHA256,await hash(run+'/benchmark.log'));assert.equal(readback.auditSHA256,await hash(run+'/audit.json'));
assert.equal(Object.keys(readback.groups).length,32);assert.equal(readback.count,64);
assert.equal(readback.audit.Checks,132674);assert.equal(readback.audit.Configurations,108);assert.equal(readback.audit.Publications,2474);assert.equal(readback.audit.Queries,2341);assert.equal(readback.audit.Abstentions,1);
assert(readback.audit.Full64&&readback.audit.RepruningChanged&&!readback.audit.WholeGoalValidation);
assert(readback.audit.MaxOracle<2e-11&&readback.audit.MaxTower<2e-11&&readback.audit.MaxReceipt<2e-11);
assert(!readback.loadedServingEstablished&&!readback.wholeMemoryLimitCertified&&!readback.scientificQualityRescueEstablished&&!readback.equalTotalCostSuperiorityEstablished);
fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(p){const out=root+'/saved/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[p]=await hash(p);assert.equal(await hash(out),copies[p]);}
async function tree(d){for(const n of fs.readdirSync(d)){const p=d+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const a of attempts)await tree(a.path);
await tree('internal/researchregimeledger');
for(const n of fs.readdirSync('research').filter(n=>/^regime-ledger-v79.*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['go.mod','go.sum','research-direction.md','research/regime-v77-v78-next.md','docs/experiments/mmm-regime-ledger-v79-protocol.md','docs/experiments/mmm-regime-ledger-v79-results.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,trackedHashes:parent.trackedHashes,
  goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},attempts,allRequiredJobsTerminal:true,
  previousGoalTurn:'NO_RESEARCH_PROGRESS(intervening social clarification); preceding substantive V77/V78 research PROGRESS',
  thisTurn:'PROGRESS(verified/chained V77/V78 archive; canonical frozen-support ledger; dense as-of/branch/publication/receipt audits and lifecycle/domain repairs; measured full150/200-member cold replay costs and unsupported noisy-evidence limitation)',
  componentAudit:readback.audit,costReadback:readback,archiveIsChainedNotStandalone:true,
  scientificAdoption:false,allSevenEmpiricalGoalsValidated:false,productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,
  limitations:['restricted working-law coherence is not unrestricted-joint approximation accuracy','some frozen supports exclude all noisy explanations; contradictory second evidence unavailable','full cold queries148/187-188ms and cumulative allocation766MB/1.02GB; no loaded serving or peak/RSS bound','no complete growing-history ingestion/core benchmark or full allowed-maximum memory bound','no new broad quality/recovery/stationary-harm/equal-total-cost experiment','useful valid splits, untouched agent outcomes, durable mixed-write freshness still required'],
  next:'Equivalent incremental forward/bounded delayed rewind and full memory resources; declared positive-support/abstention contracts; then unchanged broad controlled quality and equal-TOTAL-cost goals.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
