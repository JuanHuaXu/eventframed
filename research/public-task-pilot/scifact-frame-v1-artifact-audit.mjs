import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const original='research/public-task-pilot/scifact-frame-v1',root=original+'-closure';
const read=p=>JSON.parse(fs.readFileSync(p));
const first=read(original+'/manifest.json'),m=read(root+'/manifest.json'),f=read(root+'/freeze.json');
assert.equal(m.allCommandsTerminal,true);assert.equal(m.completeTestDependencyClosure,true);assert.equal(m.exactOriginalConversion,true);
assert.equal(f.originalManifestSHA256,hash(fs.readFileSync(original+'/manifest.json')));
for(const[p,h]of Object.entries(first.sources))assert.equal(hash(fs.readFileSync(original+'/source/'+p)),h);
for(const[p,h]of Object.entries(first.artifacts))assert.equal(hash(fs.readFileSync(original+'/'+p)),h);
for(const[p,s]of Object.entries(m.sources)){
 assert.equal(hash(fs.readFileSync(root+'/'+s.copy)),s.sha256,'frozen-copy drift');
 assert.equal(hash(fs.readFileSync(s.path)),s.sha256,'live source drift');
 if(p!=='generated-testmain')assert.equal(s.path,p);
}
for(const[p,h]of Object.entries(m.artifacts))assert.equal(hash(fs.readFileSync(root+'/'+p)),h);
assert.equal(Object.keys(first.sources).length,21);assert.equal(Object.keys(m.sources).length,168);
assert(m.sources['internal/service/service.go']);assert(m.sources['internal/store/memorystore/store.go']);assert(m.sources['internal/embed/hash.go']);
const commands=read(root+'/commands.json');assert.equal(commands.length,5);assert.deepEqual(commands.map(c=>c.name),['race','vet','convert','audit','cost']);
for(const c of commands){assert.equal(c.code,0);assert.equal(c.signal,null);assert.equal(hash(fs.readFileSync(c.log)),c.logSHA256);assert(new Date(c.ended)>=new Date(c.began));}
const roots=['DecodeAndBounds','StrictJSONIdentity','CoverageIdentityAndMetadata','AvailabilityOwnershipAndCancellation','RealObserveRecallContract','FullCorpus'].map(n=>'TestPublicFrameV1'+n);
const race=fs.readFileSync(root+'/race.log','utf8');assert(!race.includes('SKIP'));assert(!race.includes('WARNING: DATA RACE'));
for(const name of roots){assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${name}`).length,3);assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${name} (`)).length,3);}
const a=read(root+'/audit.json'),b=read(original+'/audit.json');
assert.deepEqual(a.inventory,b.inventory);assert.equal(a.framesSHA256,b.framesSHA256);assert.equal(a.corruptionRejections,20);
assert.equal(a.framesSHA256,hash(fs.readFileSync(root+'/frames.json')));assert.equal(a.framesSHA256,hash(fs.readFileSync(original+'/frames.json')));
const source='research/public-task-pilot/scifact-v1/corpus.json';assert.equal(hash(fs.readFileSync(source)),a.corpusSHA256);assert.equal(a.corpusSHA256,f.corpusSHA256);
assert.equal(a.inventory.documents,5183);assert.equal(a.inventory.frames,10869);assert.equal(a.inventory.maxField,2048);assert.equal(a.inventory.totalCacheEntries,11700);
const cost=fs.readFileSync(root+'/cost.log','utf8');const costs={};
for(const name of ['FullCorpus','SerializeCorpus']){
 const lines=cost.split('\n').filter(l=>l.startsWith(`BenchmarkPublicFrameV1${name}-`));assert.equal(lines.length,3);
 costs[name]=lines.map(l=>{const match=l.match(/\s+3\s+([0-9.]+) ns\/op\s+([0-9.]+) B\/op\s+([0-9.]+) allocs\/op$/);assert(match,l);
  return{ns:Number(match[1]),bytes:Number(match[2]),allocs:Number(match[3])};});
 for(const r of costs[name])assert(Number.isFinite(r.ns)&&r.ns>0&&r.bytes>0&&r.allocs>0);
}
const pref='research/public-task-pilot/scifact-frame-v1-strict-preflight',bad=read(pref+'/result.json');assert.equal(bad.code,1);assert.equal(bad.signal,null);
assert.equal(hash(fs.readFileSync(pref+'/test.log')),bad.logSHA256);
for(const[p,h]of Object.entries(bad.sources))assert.equal(hash(fs.readFileSync(pref+'/source/'+p)),h);
assert.equal(fs.readFileSync(pref+'/test.log','utf8').split('accepted ambiguous identity or non-scalar source:').length-1,3);
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const report={time:new Date().toISOString(),sources:168,originalSources:21,originalIncompleteDependencyFreezePreserved:true,
 fullDependencyRepeatTerminal:true,commands:5,actualRaceRoots:roots,executionsPerRoot:3,strictRegressionRepaired:true,
 corruptionRejections:20,exactRepeatFrameSHA256:a.framesSHA256,inventory:a.inventory,costs,
 frozenManifestSHA256:hash(fs.readFileSync(root+'/manifest.json')),scriptSHA256:hash(fs.readFileSync('research/public-task-pilot/scifact-frame-v1-artifact-audit.mjs')),
 weeklyUsage:usage,noQualityExperiment:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync(root+'/artifact-audit.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({sources:168,commands:5,raceRoots:6,executionsPerRoot:3,corruptions:20,
 frames:10869,exactRepeat:true,costs,weeklyUsage:usage,allSevenWholeGoals:'OPEN'},null,2));
