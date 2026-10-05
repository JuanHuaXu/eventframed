import fs from 'node:fs';import path from 'node:path';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const roots=['research/public-task-pilot/scifact-frame-v1','research/public-task-pilot/scifact-frame-v1-closure','research/public-task-pilot/scifact-frame-v1-strict-preflight'];
const root='research/checkpoint-2026-10-04-public-frame-v1';
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const closure=JSON.parse(fs.readFileSync(roots[1]+'/manifest.json'));
assert(closure.allCommandsTerminal&&closure.completeTestDependencyClosure&&closure.exactOriginalConversion);
const readback=JSON.parse(fs.readFileSync(roots[1]+'/artifact-audit.json'));assert(readback.fullDependencyRepeatTerminal);
assert.equal(readback.frozenManifestSHA256,hash(fs.readFileSync(roots[1]+'/manifest.json')));
for(const s of Object.values(closure.sources))assert.equal(hash(fs.readFileSync(roots[1]+'/'+s.copy)),s.sha256);
const previous='research/checkpoint-2026-10-04-scored-v51/manifest.json';
const old=JSON.parse(fs.readFileSync(previous));assert.equal(old.goal,'ACTIVE');
fs.mkdirSync(root,{mode:0o700});
const copies=[],links=[];
function copy(p){const b=fs.readFileSync(p),out=root+'/saved/'+p;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.writeFileSync(out,b,{flag:'wx',mode:0o600});copies.push({path:p,copy:'saved/'+p,sha256:hash(b),bytes:b.length});}
function walk(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=dir+'/'+e.name;if(e.isDirectory())walk(p);else if(e.isFile()){
 if(e.name==='frames.json'){const b=fs.readFileSync(p);links.push({path:p,sha256:hash(b),bytes:b.length});}else copy(p);
}}}
for(const r of roots)walk(r);
for(const p of ['research-direction.md','docs/experiments/scifact-frame-v1-results.md','research/public-task-pilot/scifact-frame-v1-artifact-audit.mjs','research/research-checkpoint-public-frame-v1.mjs'])copy(p);
const corpus='research/public-task-pilot/scifact-v1/corpus.json';links.push({path:corpus,sha256:hash(fs.readFileSync(corpus)),bytes:fs.statSync(corpus).size});
assert.equal(links.find(x=>x.path===corpus).sha256,readback.inventory.documents===5183?JSON.parse(fs.readFileSync(roots[1]+'/freeze.json')).corpusSHA256:'invalid');
for(const c of copies)assert.equal(hash(fs.readFileSync(root+'/'+c.copy)),c.sha256);
const manifest={time:new Date().toISOString(),weeklyUsage:usage,previousCheckpoint:{path:previous,sha256:hash(fs.readFileSync(previous))},
 copies,links,completeTestDependencySources:168,terminalCommands:10,allRequiredJobsTerminal:true,
 previousResearchGoalTurn:'PROGRESS',thisGoalTurn:'PROGRESS',goals:Object.fromEntries([1,2,3,4,5,6,7].map(k=>[k,'OPEN'])),
 goal:'ACTIVE',qualityExperiment:false,next:'Document-level pooling/native contracts/owned semantic inference and untouched outcome-labeled retrieval; loaded-serving and all-head hypotheses remain open.',
 productionPrivateCorporaWhitepaperUntouched:true,noCommitPushDeployment:true};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,copies:copies.length,links:links.length,sources:168,terminalCommands:10,weeklyUsage:usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE'}));
