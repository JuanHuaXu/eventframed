// Full exact experiment equivalence, linked to the frozen independent V71 replay.
import fs from 'node:fs';import path from 'node:path';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/dynvarcache-v72-diagnostic',prior='research/dynvariance-v71-diagnostic';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));
async function* load(p){for await(const s of readline.createInterface({input:fs.createReadStream(p),crlfDelay:Infinity}))yield JSON.parse(s)}
const freeze=json(root+'/freeze.json'),oldFreeze=json(prior+'/freeze.json'),oldDone=json(prior+'/completed.json'),oldScore=json(prior+'/readback.json');
for(const[p,h]of Object.entries({...freeze.files,...freeze.protectedFiles,...freeze.data}))assert.equal(await hash(p),h,p);
assert(oldDone.allJobsTerminal&&oldDone.sourceUnchanged&&oldDone.checks.length===8&&oldDone.checks.every(x=>x.exitCode===0));
assert(!fs.existsSync(prior+'/failure.json'));
for(const[p,h]of Object.entries(oldDone.artifacts))assert.equal(await hash(prior+'/'+p),h,p);
for(const x of oldDone.checks)assert.equal(await hash(prior+'/'+x.name+'.log'),x.logSHA256,x.name);
for(const[p,h]of Object.entries(oldFreeze.files)){assert.equal(await hash(p),h,p);assert.equal(await hash(prior+'/source/'+p),h,p)}
for(const[p,h]of Object.entries(oldFreeze.generatedCompilerCopies))assert.equal(await hash(prior+'/'+p),h,p);
assert.equal(oldScore.modelArms,1920);assert.equal(oldScore.independentIssuedPackets,4608000);assert.equal(oldScore.scalarIssuedComparisons,9216000);
const checkpoint=json(freeze.parent.path);assert.equal(await hash(freeze.parent.path),freeze.parent.sha256);
for(const[p,h]of Object.entries(checkpoint.copies))assert.equal(await hash(path.dirname(freeze.parent.path)+'/saved/'+p),h,p);
assert.equal(checkpoint.copies[prior+'/completed.json'],await hash(prior+'/completed.json'));
assert.equal(checkpoint.copies[prior+'/diagnostic.jsonl'],await hash(prior+'/diagnostic.jsonl'));
const streams=[load(root+'/diagnostic.jsonl'),load(prior+'/diagnostic.jsonl')];
const m=(await streams[0].next()).value,oldMeta=(await streams[1].next()).value;
assert.equal(m.Worlds,40);assert.equal(m.SeedBase,2026105407);assert.deepEqual(m.Sources,freeze.files);assert.equal(oldMeta.SeedBase,m.SeedBase);
const strip=a=>{const b=structuredClone(a);delete b.Costs;delete b.Breakdown;return b};
let arms=0,models=0,issued=0;
for(let n=0;n<40;n++){
 const w=(await streams[0].next()).value,old=(await streams[1].next()).value;
 assert(w&&old);assert.deepEqual(w.Population,old.Population);assert.equal(w.Arms.length,54);assert.equal(old.Arms.length,54);
 for(let j=0;j<54;j++){
  const a=w.Arms[j],b=old.Arms[j];assert.deepEqual(strip(a),strip(b),`world ${n} arm ${j}`);arms++;
  assert.equal(a.Issued.length,2400);assert.equal(a.Receipts.length,2400);
  if(j%18>=2){assert.equal(a.ObservedIssued.length,2400);models++;issued+=a.Issued.length}
 }
}
for(const s of streams)assert((await s.next()).done);
assert.equal(arms,2160);assert.equal(models,1920);assert.equal(issued,4608000);
const result={allArmsBitwiseEqual:arms,modelArmsBitwiseEqual:models,populationsIdentical:40,excludedFields:['Costs','Breakdown'],compared:'every other serialized arm field including clean/noisy issue laws, values, choices, requests, receipts, snapshots and metrics',priorIndependentReplayVerified:true,priorCompletedSHA256:await hash(prior+'/completed.json'),priorDataSHA256:await hash(prior+'/diagnostic.jsonl'),dataSHA256:await hash(root+'/diagnostic.jsonl'),transitiveIndependentIssuedPackets:issued,transitiveScalarIssuedComparisons:2*issued,directIndependentFullReplayThisStudy:false,smallAndFullJournalIndependentAuditsThisStudy:true,scientificQualityRescueEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/transitive-audit.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));
