// Archive completed component evidence, including failed approximation claims.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
const usage = Number(process.argv[2]);
assert(Number.isFinite(usage) && usage >= 0 && usage <= 100);
const root = 'research/checkpoint-2026-10-05-regime-v77-frozen-v78';
assert(!fs.existsSync(root));
const parentPath = 'research/checkpoint-2026-10-05-mean-anchor-v75-ratio-v76/manifest.json';
const json = p => JSON.parse(fs.readFileSync(p));
async function hash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
assert.equal(await hash(parentPath), 'acd27d5b83186333790ae27fa3c07a27e540b039e6c1e41313ea3f8147b33554');
const parent = json(parentPath);
for (const [p,h] of Object.entries(parent.copies)) assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p), h, p);
for (const [p,h] of Object.entries(parent.generatedCompilerCopies ?? {})) assert.equal(await hash(path.dirname(parentPath)+'/'+p), h, p);
const tracked = execFileSync('git', ['diff','--name-only'], {encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();
assert.deepEqual(tracked, Object.keys(parent.trackedHashes).sort());
for (const [p,h] of Object.entries(parent.trackedHashes)) assert.equal(await hash(p), h, p);
const runs = ['research/regime-v77-initial','research/regime-frozen-v78-initial','research/regime-frozen-v78-long-initial'];
for (const run of runs) {
  const f=json(run+'/freeze.json'), c=json(run+'/completed.json');
  assert(c.allJobsTerminal && c.allChecksPass && c.sourceUnchanged && c.checks.length===4);
  assert(!c.scientificQualityRescueEstablished && !c.wholeCohortTested && !c.loadedServingEstablished);
  for (const [p,h] of Object.entries(f.files)) {
    assert.equal(await hash(p),h,p);
    assert.equal(await hash(run+'/source/'+p.replace(/^\//,'')),h,p);
  }
  for (const x of c.checks) {
    assert.equal(x.exitCode,0);
    assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256);
  }
}
const exact=json(runs[0]+'/audit.json'), frozen=json(runs[1]+'/audit.json'), long=json(runs[2]+'/audit.json');
assert.equal(exact.Checks,42787);
assert.equal(exact.Configurations,48);
assert.equal(exact.IndependentPathChecks,432);
assert.equal(exact.BeamCases,144);
assert.equal(exact.VacuousBeamEnvelopes,79);
assert(exact.MaxProjectionTV>.029 && exact.MaxBeamTV>.31 && exact.MaxBeamTowerDefect>.008);
assert.equal(frozen.FrozenChecks,2244);
assert(frozen.MaxFrozenOracleDefect<2e-11 && frozen.MaxFrozenTowerDefect<2e-11 && frozen.MaxBeamTowerDefect>.05);
assert.equal(long.cases,36);
assert.equal(long.checks,2589);
assert.equal(long.zeroBranches,3);
assert(long.maxOracle<2e-11 && long.maxTower<2e-11 && !long.wholeGoalValidation);
fs.mkdirSync(root,{mode:0o700});
const copies={};
async function copy(p) {
  const out=root+'/saved/'+p;
  fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});
  fs.copyFileSync(p,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);
  fs.chmodSync(out,0o600);
  copies[p]=await hash(p);
  assert.equal(await hash(out),copies[p]);
}
async function tree(d) {
  for(const n of fs.readdirSync(d)) {
    const p=d+'/'+n;
    if(fs.statSync(p).isDirectory()) await tree(p); else await copy(p);
  }
}
for(const run of runs) await tree(run);
for(const d of ['internal/researchregime','internal/researchregimefrozen','internal/researchregimefrozencheck']) await tree(d);
for(const n of fs.readdirSync('research').filter(n=>/^regime-(v77|frozen-v78|v77-v78).*\.(mjs|json|md)$/.test(n))) await copy('research/'+n);
for(const p of ['go.mod','go.sum','research-direction.md','docs/experiments/mmm-regime-v77-protocol.md','docs/experiments/mmm-regime-v77-results.md','docs/experiments/mmm-regime-frozen-v78-protocol.md','docs/experiments/mmm-regime-frozen-v78-results.md']) await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},
  copies,trackedHashes:parent.trackedHashes,goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},
  previousGoalTurn:'PROGRESS(V77 exact changing-explanation model; measured projection/beam failures; V78 isolated frozen-support conditioning; full64 oracle and timing)',
  allRequiredJobsTerminal:true,commands:runs.map(p=>({path:p,completedSHA256:copies[p+'/completed.json']})),componentAudits:{exact,frozen,long},
  scientificAdoption:false,allSevenEmpiricalGoalsValidated:false,productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,
  privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,archiveIsChainedNotStandalone:true,
  limitations:['Journal still re-prunes; conditional helper not canonical publication law','TV up to .318; 79/144 envelopes vacuous','64-row two-member conditional replay allocates about 10MB, not RSS','no full150/200-member resource or recovery/harm/equal-total-cost study','goals3/5/6 remain open'],
  next:'Wire isolated canonical frozen-support journal, then incremental/bounded replay, omitted-mass handling and unchanged full quality gates.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies)) assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.trackedHashes)) assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:usage},null,2));
