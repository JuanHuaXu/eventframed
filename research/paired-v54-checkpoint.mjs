import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const root='research/checkpoint-2026-10-04-paired-v54',run='research/paired-v54-diagnostic';
assert(!fs.existsSync(root),'exclusive checkpoint');
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');}
const previous='research/checkpoint-2026-10-04-noise-v53-final/manifest.json';
assert.equal(await fileHash(previous),'951cecf65133571926db6ac27b3711b90018afbf0957869facf9a7328da58829');
const old=JSON.parse(fs.readFileSync(previous)),freeze=JSON.parse(fs.readFileSync(run+'/freeze.json'));
const completed=JSON.parse(fs.readFileSync(run+'/completed.json')),readback=JSON.parse(fs.readFileSync(run+'/readback.json'));
assert.deepEqual(freeze.protectedFiles,old.trackedHashes);assert.equal(completed.checks.length,7);assert(completed.checks.every(c=>c.exitCode===0));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles}))assert.equal(await fileHash(p),h,p);
for(const[p,h]of Object.entries(completed.artifacts))assert.equal(await fileHash(run+'/'+p),h,p);
// Verify the immutable parent and its immediate full-run archive, not current
// historical working documents that this continuation legitimately supersedes.
const parentRoot=path.dirname(previous);
for(const[p,h]of Object.entries(old.copies))assert.equal(await fileHash(parentRoot+'/saved/'+p),h,p);
const grand=old.previousCheckpoint;assert.equal(await fileHash(grand.path),grand.sha256);
const grandManifest=JSON.parse(fs.readFileSync(grand.path));
for(const[p,h]of Object.entries(grandManifest.copies))assert.equal(await fileHash(path.dirname(grand.path)+'/saved/'+p),h,p);
assert.equal(readback.worlds,40);assert.equal(readback.arms,840);assert.equal(readback.distinctUnderlyingOutcomes,96000);
assert(readback.sourceClosureVerified);assert.equal(readback.wholeGoals,'OPEN');
const identity=JSON.parse(fs.readFileSync('research/paired-v55-smoothing-identities.json'));
assert.equal(identity.cases,162);assert.equal(identity.zeroSupportChecks,24);assert.equal(identity.unknownEvidenceForks,81);assert(identity.proposalOnly&&!identity.implementedLearner&&!identity.performanceMeasured);
for(const[p,h]of Object.entries(identity.sources))assert.equal(await fileHash(p),h,p);
fs.mkdirSync(root,{mode:0o700});const copies={};
async function copy(src){const dest=root+'/saved/'+src;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(src,dest,fs.constants.COPYFILE_EXCL);fs.chmodSync(dest,0o600);copies[src]=await fileHash(src);assert.equal(await fileHash(dest),copies[src]);}
for(const p of Object.keys(freeze.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
for(const p of ['docs/experiments/mmm-paired-v54-results.md','research-direction.md','research/paired-v54-checkpoint.mjs','research/paired-v55-direction.md','research/paired-v55-smoothing-identities.mjs','research/paired-v55-smoothing-identities.json'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await fileHash(previous),copiedArtifactsVerified:Object.keys(old.copies).length,precedingFullRunCopiesVerified:Object.keys(grandManifest.copies).length},copies,trackedHashes:freeze.protectedFiles,frozenCompilerInputs:freeze.compilerFiles.length,frozenSourceFiles:Object.keys(freeze.files).length,terminalCommands:completed.checks.length,allRequiredJobsTerminal:true,worlds:40,arms:840,distinctUnderlyingOutcomes:96000,secondaryRequests:192000,secondaryRequestsAreNotDistinctLatentOutcomes:true,futurePrefixCases:15,semanticCorruptionsRejected:38,domainSeparatedWorldSeeds:3840,domainSeparatedChannels:23040,diagnosticOnly:true,nPerCell:1,qualityAdoption:false,equalTotalCostSuperiorityEstablished:false,previousGoalTurn:'PROGRESS (paired implementation, preflight repairs and live serialized run); intervening social clarification contained no research',thisGoalTurn:'PROGRESS',weeklyUsage:{usedPercent:usage,windowDurationMins:10080,recordedAt:new Date().toISOString()},goals:['OPEN','OPEN','OPEN','OPEN','OPEN','OPEN','OPEN'],goal:'ACTIVE',productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,prospectiveRepositoryCompilerSourceClosureComplete:true,next:'Use preserved paired-observation tradeoffs to choose a fresh prospective lead; do not dispatch confirmation for a failed broad or cost screen'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
for(const[p,h]of Object.entries(copies))assert.equal(await fileHash(root+'/saved/'+p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await fileHash(root+'/manifest.json'),copies:Object.keys(copies).length,parentVerified:true,weeklyUsage:usage,wholeGoals:'OPEN',goal:'ACTIVE'},null,2));
