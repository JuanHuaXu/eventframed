// Reproducible new research state, with a hashed link to prior negative evidence.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/checkpoint-2026-10-04-specialist-v47';assert(!fs.existsSync(root));
const run='research/specialist-v47-study-diagnostic',pair='research/specialist-v47-pair';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function digest(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const done=JSON.parse(fs.readFileSync(run+'/completed.json'));
const freeze=JSON.parse(fs.readFileSync(run+'/freeze.json'));
const pd=JSON.parse(fs.readFileSync(pair+'/completed.json'));
assert(done.sourceUnchanged&&done.checks.every(c=>c.code===0));
assert(pd.sourceUnchanged&&pd.independentGlobalReference&&pd.exactSharedExpertAdvice);
assert.equal(JSON.parse(fs.readFileSync(pair+'/command.json')).code,0);
assert(JSON.parse(fs.readFileSync(run+'/artifact-audit.json')).sourceAndArtifactHashes);
assert(!JSON.parse(fs.readFileSync(run+'/readback.json')).qualityAdoption);
const source={...freeze.files,...pd.files};
for(const p of ['research-direction.md','docs/experiments/mmm-specialist-v47-results.md','research/specialist-v47-artifact-audit.mjs','research/specialist-v47-diagnostic-summary.mjs','research/research-checkpoint-specialist-v47.mjs'])source[p]=hash(fs.readFileSync(p));
for(const[p,h]of Object.entries(source))assert.equal(hash(fs.readFileSync(p)),h,'current source '+p);
const records=[];
for(const dir of [run,pair])for(const p of fs.readdirSync(dir))if(p.endsWith('.json')||p.endsWith('.log'))records.push(dir+'/'+p);
const previous='research/checkpoint-2026-10-04-cohort-v46/manifest.json';records.push(previous);
const artifacts={};
for(const p of [run+'/diagnostic-fixture.jsonl',run+'/diagnostic.jsonl',pair+'/global-static.jsonl'])artifacts[p]={sha256:await digest(p),bytes:fs.statSync(p).size};
assert.equal(artifacts[run+'/diagnostic-fixture.jsonl'].sha256,done.artifacts['diagnostic-fixture.jsonl']);
assert.equal(artifacts[run+'/diagnostic.jsonl'].sha256,done.artifacts['diagnostic.jsonl']);
assert.equal(artifacts[pair+'/global-static.jsonl'].sha256,pd.globalRawSHA256);
for(const[p,h]of Object.entries(done.artifacts))assert.equal(await digest(run+'/'+p),h,'complete artifact '+p);
fs.mkdirSync(root,{mode:0o700});
function copy(p,folder,h){const b=fs.readFileSync(p);assert.equal(hash(b),h);const d=root+'/'+folder+'/'+p;fs.mkdirSync(path.dirname(d),{recursive:true,mode:0o700});const fd=fs.openSync(d,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd)}finally{fs.closeSync(fd)}assert.equal(hash(fs.readFileSync(d)),h)}
for(const[p,h]of Object.entries(source))copy(p,'source',h);
const hashes={};for(const p of records){hashes[p]=hash(fs.readFileSync(p));copy(p,'records',hashes[p])}
const manifest={time:new Date().toISOString(),source,records:hashes,artifacts,previousCheckpoint:{path:previous,sha256:hashes[previous]},previousResearchTurn:'PROGRESS',interveningSocialExchangeResearchProgress:false,thisTurn:'PROGRESS',goal:'ACTIVE',allSevenWholeGoals:'OPEN',weeklyUsageAtLastReadPercent:18,allRequiredManagedJobsTerminal:true,productionPrivateCorporaWhitepaperUntouched:true,pushOrDeploymentPerformed:false,scope:'isolated V47 full diagnostic and same-world attribution; negative quality retained; no normal or whole-goal completion'};
const fd=fs.openSync(root+'/manifest.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(manifest,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify({root,sources:Object.keys(source).length,records:records.length,artifacts:Object.keys(artifacts).length,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2));
