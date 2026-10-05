// Reproducible research checkpoint; no commit, deployment, or production access.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/checkpoint-2026-10-04-memo-v49';assert(!fs.existsSync(root),'exclusive checkpoint');
const v48='research/hybrid-v48-env-repair-diagnostic',v49='research/memo-v49-ablation',bad='research/hybrid-v48-study-diagnostic',p3='research/public-task-pilot/scifact-provenance-v3';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const sources={},records={},artifacts={};
for(const dir of [v48,v49]){
 const f=JSON.parse(fs.readFileSync(dir+'/freeze.json')),d=JSON.parse(fs.readFileSync(dir+'/completed.json'));assert(d.sourceUnchanged&&d.checks.every(c=>c.code===0));
 for(const[p,h]of Object.entries(f.files)){assert.equal(hash(fs.readFileSync(p)),h,'current '+p);assert.equal(hash(fs.readFileSync(dir+'/source/'+p)),h,'frozen '+p);if(sources[p])assert.equal(sources[p],h);sources[p]=h}
 for(const[p,h]of Object.entries(d.artifacts)){assert.equal(await digest(dir+'/'+p),h,'artifact '+p);if(p.endsWith('.jsonl'))artifacts[dir+'/'+p]={sha256:h,bytes:fs.statSync(dir+'/'+p).size}}
}
assert(JSON.parse(fs.readFileSync(v48+'/artifact-audit.json')).sourceAndArtifactHashes);
const readback=JSON.parse(fs.readFileSync(v49+'/readback.json'));assert(readback.exactAllFieldsExceptCost&&!readback.qualityConfirmation);assert.equal(readback.pairedArms,252);assert(readback.total.memo.loopPass&&readback.screen.memo.allocationPass);
const failedFreeze=JSON.parse(fs.readFileSync(bad+'/freeze.json'));for(const[p,h]of Object.entries(failedFreeze.files))assert.equal(hash(fs.readFileSync(bad+'/source/'+p)),h,'failed frozen copy '+p);assert(fs.existsSync(bad+'/failure.json'));
const p=JSON.parse(fs.readFileSync(p3+'/report.json')),a=JSON.parse(fs.readFileSync(p3+'/audit.json'));assert(a.independentRelationFirstReconstruction&&a.preparedDataUnchanged);assert.equal(hash(fs.readFileSync(p3+'/report.json')),a.reportSHA256);
for(const[file,x]of Object.entries(p.sources)){assert.equal(hash(fs.readFileSync(file)),x.sha256,file);if(file.endsWith('.gz'))artifacts[file]={sha256:x.sha256,bytes:x.bytes};else sources[file]=x.sha256}
for(const file of ['research-direction.md','docs/experiments/mmm-hybrid-v48-results.md','docs/experiments/mmm-memo-v49-results.md','docs/experiments/scifact-provenance-v3-results.md','research/public-task-pilot/scifact-provenance-v3-audit.mjs','research/research-checkpoint-memo-v49.mjs'])sources[file]=hash(fs.readFileSync(file));
for(const dir of [v48,v49,bad,p3])for(const file of fs.readdirSync(dir))if(file.endsWith('.json')||file.endsWith('.log'))records[dir+'/'+file]=hash(fs.readFileSync(dir+'/'+file));
const previous='research/checkpoint-2026-10-04-specialist-v47/manifest.json';records[previous]=hash(fs.readFileSync(previous));
for(const file of ['research/public-task-pilot/scifact-v1/source.json','research/public-task-pilot/scifact-v1/split-plan.json','research/public-task-pilot/scifact-v1/split-audit.json'])records[file]=hash(fs.readFileSync(file));
fs.mkdirSync(root,{mode:0o700});
function copy(p,folder,h){const b=fs.readFileSync(p);assert.equal(hash(b),h);const dest=root+'/'+folder+'/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const fd=fs.openSync(dest,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}assert.equal(hash(fs.readFileSync(dest)),h)}
for(const[p,h]of Object.entries(sources))copy(p,'source',h);for(const[p,h]of Object.entries(records))copy(p,'records',h);
for(const[p,x]of Object.entries(artifacts))assert.equal(await digest(p),x.sha256,'raw checkpoint '+p);
const manifest={time:new Date().toISOString(),source:sources,records,artifacts,previousCheckpoint:{path:previous,sha256:records[previous]},previousResearchTurn:'PROGRESS',interveningSocialExchange:'NO_RESEARCH_PROGRESS',noProgressRevalidated:true,thisTurn:'PROGRESS',goal:'ACTIVE',allSevenWholeGoals:'OPEN',weeklyUsageAtLastReadPercent:19,allRequiredManagedJobsTerminal:true,productionPrivateCorporaWhitepaperUntouched:true,pushOrDeploymentPerformed:false,scope:'V48 complete negative quality diagnostic; V49 exact paired cost rescue; SciFact citation/annotation provenance, no prediction or new confirmation',next:'unchanged normal hybrid cohorts and public source-preserving retrieval contract; loaded serving remains open'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify({root,sources:Object.keys(sources).length,records:Object.keys(records).length,artifacts:Object.keys(artifacts).length,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2));
