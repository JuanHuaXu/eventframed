import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/checkpoint-2026-10-04-noise-v53-final';
assert(!fs.existsSync(root),'exclusive supplemental checkpoint');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const parent='research/checkpoint-2026-10-04-noise-v53/manifest.json';
const parentHash=hash(fs.readFileSync(parent));assert.equal(parentHash,'e8a353b4c3d6f26521d5c40047cbdd7cb4f2ff19546b7f1a968b487175506818');
const old=JSON.parse(fs.readFileSync(parent));for(const[p,h]of Object.entries(old.copies))assert.equal(hash(fs.readFileSync('research/checkpoint-2026-10-04-noise-v53/saved/'+p)),h,p);
for(const[p,h]of Object.entries(old.trackedHashes))assert.equal(hash(fs.readFileSync(p)),h,p);
const closure=JSON.parse(fs.readFileSync('research/noise-v53-diagnostic/closure-readback.json'));assert(closure.savedCopiesVerified&&closure.repositoryClosureProspectivelyComplete&&closure.repositoryCompilerInputs===56);
const identities=JSON.parse(fs.readFileSync('research/noise-v54-identities.json'));assert(identities.parameterCases===18&&identities.stagedCases===68&&identities.proposalOnly&&!identities.learnedImplementation);
for(const[p,h]of Object.entries(identities.sources))assert.equal(hash(fs.readFileSync(p)),h,p);
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
fs.mkdirSync(root,{mode:0o700});const copies={};
for(const p of['research-direction.md','docs/experiments/mmm-noise-v53-results.md','research/noise-v53-closure-readback.mjs','research/noise-v53-diagnostic/closure-readback.json','research/noise-v53-post-audit.md','research/noise-v54-direction.md','research/noise-v54-identities.mjs','research/noise-v54-identities.json','research/noise-v53-final-checkpoint.mjs']){const b=fs.readFileSync(p),out=root+'/saved/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.writeFileSync(out,b,{flag:'wx',mode:0o600});copies[p]=hash(b);assert.equal(hash(fs.readFileSync(out)),copies[p])}
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parent,sha256:parentHash,copiedArtifactsVerified:Object.keys(old.copies).length},copies,trackedHashes:old.trackedHashes,prospectiveRepositoryCompilerSourceClosureComplete:true,postRunSupplementalBoundaryRepairRecorded:true,allRequiredJobsTerminal:true,noiseV53DiagnosticFailurePreserved:true,noiseV54IdentityChecksOnly:true,goals:['OPEN','OPEN','OPEN','OPEN','OPEN','OPEN','OPEN'],goal:'ACTIVE',thisGoalTurn:'PROGRESS',weeklyUsage:{usedPercent:usage,windowDurationMins:10080},productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,next:'Implement/evaluate prospective independent-measurement identification at equal total acquisition cost; retain correlated-source failures and continue other outstanding goals'};
const out=root+'/manifest.json',fd=fs.openSync(out,'wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify({path:out,sha256:hash(fs.readFileSync(out)),copies:Object.keys(copies).length,parentVerified:true,weeklyUsage:usage,wholeGoals:'OPEN',goal:'ACTIVE'},null,2));
