// Exact isolated-source/record checkpoint; large traces are byte-hashed in place.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root='research/checkpoint-2026-10-04-cohort-v46';
assert(!fs.existsSync(root),'exclusive checkpoint');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p){const h=crypto.createHash('sha256');for await(const bytes of fs.createReadStream(p))h.update(bytes);return h.digest('hex')}
const run='research/cohort-batch-v46-native-cap';
const done=JSON.parse(fs.readFileSync(run+'/completed.json'));
assert(done.sourceUnchanged&&done.checks.every(c=>c.code===0));
const freeze=JSON.parse(fs.readFileSync(run+'/freeze.json'));
for(const[p,h]of Object.entries(freeze.files))assert.equal(hash(fs.readFileSync(p)),h,'changed frozen runtime '+p);
const profile='research/eager-load-v45-profile';
const prof=JSON.parse(fs.readFileSync(profile+'/completed.json'));assert(prof.sourceUnchanged&&prof.checks.every(c=>c.code===0));
const old='research/cohort-batch-v46';
assert(fs.existsSync(old+'/failure.json'));assert.equal(JSON.parse(fs.readFileSync(old+'/failure-triage.json')).fullStudyAudit,'FAIL');
const source={...freeze.files};
for(const p of ['research-direction.md','research/research-checkpoint-cohort-v46.mjs','research/cohort-batch-v46-readback.mjs','research/cohort-batch-v46-failure-triage.mjs',
  'docs/experiments/mmm-eager-load-v45-profile-results.md','docs/experiments/mmm-cohort-batch-v46-results.md',
  'research/cohort-batch-v46-incomplete-getter-closure.go.txt','research/cohort-batch-v46-incomplete-getter-closure.json',
  'research/cohort-batch-v46-before-adjacent-controls.go.txt','research/cohort-batch-v46-before-adjacent-controls.json',
  'research/cohort-batch-v46-worker-only-generated.go.txt','research/cohort-batch-v46-worker-only-generation.json','research/cohort-batch-v46-audit-missing-vm-clone.mjs.txt']) source[p]=hash(fs.readFileSync(p));
const records=[];
for(const dir of [profile,old,'research/cohort-batch-v46-preflight','research/cohort-batch-v46-preflight-native-cap',run]){
  for(const name of fs.readdirSync(dir))if(name.endsWith('.json')||name.endsWith('.log'))records.push(dir+'/'+name);
}
records.push('research/checkpoint-2026-10-04-switch-load-public/manifest.json');
const artifacts={};
for(const p of [profile+'/raw.ndjson',old+'/raw.ndjson',run+'/raw.ndjson',profile+'/subject.test',...['cpu','block','mutex','alloc'].map(n=>profile+'/'+n+'.pprof')]) artifacts[p]={sha256:await fileHash(p),bytes:fs.statSync(p).size};
assert.equal(artifacts[run+'/raw.ndjson'].sha256,done.rawSHA256);
assert.equal(artifacts[profile+'/raw.ndjson'].sha256,prof.profiledRawSHA256);
assert.equal(artifacts[old+'/raw.ndjson'].sha256,JSON.parse(fs.readFileSync(old+'/failure-triage.json')).rawSHA256);
fs.mkdirSync(root,{mode:0o700});
for(const[p,h]of Object.entries(source)){
  const dest=root+'/source/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});const bytes=fs.readFileSync(p);assert.equal(hash(bytes),h);fs.writeFileSync(dest,bytes,{flag:'wx',mode:0o600});assert.equal(hash(fs.readFileSync(dest)),h);
}
const recordHashes={};
for(const p of records){const bytes=fs.readFileSync(p),dest=root+'/records/'+p;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.writeFileSync(dest,bytes,{flag:'wx',mode:0o600});recordHashes[p]=hash(bytes);assert.equal(hash(fs.readFileSync(dest)),recordHashes[p])}
const manifest={time:new Date().toISOString(),source,records:recordHashes,artifacts,
  previousTurn:'PROGRESS',thisTurn:'PROGRESS',goal:'ACTIVE',allSevenWholeGoals:'OPEN',weeklyUsageAtLastReadPercent:18,
  allRequiredManagedJobsTerminal:true,productionPrivateCorporaWhitepaperUntouched:true,pushOrDeploymentPerformed:false,
  scope:'isolated V45 diagnosis and V46 full-load test; failures retained, no whole-goal completion implied'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({root,sources:Object.keys(source).length,records:records.length,largeArtifacts:Object.keys(artifacts).length,allSevenWholeGoals:'OPEN',goal:'ACTIVE'},null,2));
