import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/checkpoint-2026-10-04-noise-v53',run='research/noise-v53-diagnostic';
assert(!fs.existsSync(root),'exclusive checkpoint');
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const previous='research/checkpoint-2026-10-04-scored-v52/manifest.json';
const previousHash=hash(fs.readFileSync(previous));assert.equal(previousHash,'436d7a90ad32aa75cc1977076eeaa296d9a3cf7e62a4d12c2f4c04c23d51c3be');
const old=JSON.parse(fs.readFileSync(previous)),freeze=JSON.parse(fs.readFileSync(run+'/freeze.json')),completed=JSON.parse(fs.readFileSync(run+'/completed.json')),readback=JSON.parse(fs.readFileSync(run+'/readback.json'));
assert.deepEqual(freeze.protectedFiles,old.trackedHashes);assert(completed.checks.every(c=>c.exitCode===0));
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles}))assert.equal(hash(fs.readFileSync(p)),h,p);
for(const[p,h]of Object.entries(completed.artifacts))assert.equal(hash(fs.readFileSync(run+'/'+p)),h,p);
assert.equal(readback.worlds,36);assert.equal(readback.arms,756);assert.equal(readback.distinctLabels,86400);
fs.mkdirSync(root,{mode:0o700});
const copies={};
function copy(src){const b=fs.readFileSync(src),dest=root+'/saved/'+src;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,b,{flag:'wx',mode:0o600});copies[src]=hash(b);assert.equal(hash(fs.readFileSync(dest)),copies[src]);}
for(const p of Object.keys(freeze.files))copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))copy(run+'/'+p);
for(const p of ['docs/experiments/mmm-noise-v53-results.md','research-direction.md','research/noise-v53-checkpoint.mjs'])copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:previousHash},copies,trackedHashes:freeze.protectedFiles,frozenCompilerInputs:freeze.compilerFiles.length,frozenSourceFiles:Object.keys(freeze.files).length,terminalCommands:completed.checks.length,allRequiredJobsTerminal:true,worlds:36,arms:756,distinctOutcomes:86400,exactLegacyPairs:108,futurePrefixCases:12,semanticCorruptionsRejected:84,diagnosticOnly:true,nPerCell:1,qualityAdoption:false,previousGoalTurn:'NO_RESEARCH_PROGRESS (social clarification); preceding researchV52 PROGRESS',thisGoalTurn:'PROGRESS',weeklyUsage:{usedPercent:usage,windowDurationMins:10080,recordedAt:new Date().toISOString()},goals:['OPEN','OPEN','OPEN','OPEN','OPEN','OPEN','OPEN'],goal:'ACTIVE',productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,prospectiveRepositoryCompilerSourceClosureComplete:true,next:'Do not confirm failed noise model; investigate observation identification and unaffected outstanding goals with fresh prospective protocols'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
for(const[p,h]of Object.entries(copies))assert.equal(hash(fs.readFileSync(root+'/saved/'+p)),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:hash(fs.readFileSync(root+'/manifest.json')),copies:Object.keys(copies).length,weeklyUsage:usage,wholeGoals:'OPEN',goal:'ACTIVE'},null,2));
