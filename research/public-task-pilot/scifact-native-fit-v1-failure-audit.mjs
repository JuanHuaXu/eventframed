// Failure-path readback never opens citation labels or scores a partial corpus.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import readline from 'node:readline';
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const read=p=>JSON.parse(fs.readFileSync(p));
const root='research/public-task-pilot/scifact-native-fit-v1',native='research/public-task-pilot/nativefit-v1';
const m=read(root+'/manifest.json'),n=read(native+'/manifest.json'),term=read(native+'/daemon-terminal.json');
assert(m.allCommandsTerminal&&n.allOwnedJobsTerminal);assert.equal(m.nativeCode,1);assert.equal(n.result.code,1);assert(n.error);
assert.equal(term.code,2);assert.equal(term.actualPriority,10);assert.equal(term.ownedProcessOnly,true);
assert.equal(term.watcherError,'owned daemon exceeded sampled 2GiB ceiling');assert(term.sampledPeakRSSKiB>2*1024*1024);
assert.equal(Object.keys(m.sources).length,30);assert(m.completeTestDependencyClosure);
for(const[p,s]of Object.entries(m.sources)){assert.equal(await sha(p),s.sha256);assert.equal(await sha(root+'/'+s.copy),s.sha256);}
for(const[p,h]of Object.entries(m.artifacts))if(typeof h==='string')assert.equal(await sha(root+'/'+p),h);else assert.equal(await sha(h.path),h.sha256);
for(const[p,s]of Object.entries(n.inputs))assert.equal(await sha(p),s.sha256);
for(const[p,h]of Object.entries(n.artifacts))assert.equal(await sha(native+'/'+p),h);
const commands=read(root+'/commands.json');assert.deepEqual(commands.map(c=>[c.name,c.code,c.signal]),[
 ['race',0,null],['vet',0,null],['build',0,null],['native',1,null]]);
for(const c of commands)assert.equal(await sha(c.log),c.logSHA256);
const race=fs.readFileSync(root+'/race.log','utf8');assert(!race.includes('SKIP'));assert(!race.includes('DATA RACE'));
for(const name of ['TestNativeFitOnlyBoundary','TestNativeFitOwnedPath']){
 assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${name}`).length,3);
 assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${name} (`)).length,3);
}
const plan=read('research/public-task-pilot/scifact-v1/split-plan.json'),fit=read(root+'/fit-only.json'),mixed=read('research/public-task-pilot/scifact-v1/queries.json');
assert.equal(fit.partition,'fit');assert.deepEqual(fit.queries.map(q=>q.id),plan.splits.fit);
const qmap=new Map(mixed.map(q=>[q.id,q]));for(const q of fit.queries)assert.deepEqual(q,qmap.get(q.id));
const heldOut=new Set([...plan.splits.calibration,...plan.splits.confirmation,...plan.splits.excludedOfficialTrain]);assert(fit.queries.every(q=>!heldOut.has(q.id)));
const pool=read('research/public-task-pilot/scifact-pool-v1/pool.json');
let start=null,attempts=0,successful=0,failure=null,queries=0,rpcNS=0;
const stream=readline.createInterface({input:fs.createReadStream(native+'/trace.ndjson'),crlfDelay:Infinity});
for await(const line of stream){const r=JSON.parse(line);assert(!failure,'continued after failed import');
 if(r.kind==='start'){assert(!start);start=r;assert.equal(r.documents,5183);assert.equal(r.queries,351);assert.equal(r.partition,'fit');}
 else if(r.kind==='import'){
  assert(start);assert.equal(r.index,attempts);assert.equal(r.id,pool[attempts].candidate.ID);assert.equal(r.source,pool[attempts].source_id);assert(r.ns>0);rpcNS+=r.ns;attempts++;
  if(r.ok){assert(!r.error);successful++;}else{assert(r.error.includes('Unavailable'));failure=r;}
 }else if(r.kind==='query')queries++;else throw new Error('partial run must not contain completion');
}
assert.equal(successful,4250);assert.equal(attempts,4251);assert.equal(queries,0);assert(failure);assert(!fs.existsSync(root+'/audit-results.json'));
const log=fs.readFileSync(native+'/daemon.log','utf8');assert(log.includes('panic: runtime error: index out of range [1083] with length 0'));
assert(log.includes('cognitive-engine/causal.(*DirtyJournal).Drain'));assert(log.includes('dirty_journal.go:377'));
assert(log.includes('context deadline exceeded'));assert(log.includes('GGUF embedding backend ready'));
const store=native+'/store.libravdb';assert(fs.statSync(store).size>0);
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const report={time:new Date().toISOString(),manifestSHA256:await sha(root+'/manifest.json'),nativeManifestSHA256:await sha(native+'/manifest.json'),
 traceSHA256:await sha(native+'/trace.ndjson'),daemonLogSHA256:await sha(native+'/daemon.log'),
 preservedOwnedStore:{path:store,bytes:fs.statSync(store).size,sha256:await sha(store),recoveryOrCompletenessNotEstablished:true},
 sources:30,completeTestDependencyClosure:true,commands:4,allCommandsTerminal:true,raceRoots:2,executionsPerRoot:3,
 successfulImports:successful,attemptedImports:attempts,requiredImports:5183,fitQueriesRequired:351,fitPredictions:queries,
 calibrationPredictions:0,confirmationPredictions:0,evaluatorLabelsRead:false,partialCorpusNotScored:true,
 cost:{sumAttemptedImportRPCNS:rpcNS,wholeOwnedProcessMS:Date.parse(term.ended)-Date.parse(term.began),sampledDaemonPeakRSSKiB:term.sampledPeakRSSKiB},
 confirmed:['full-corpus fit diagnostic failed','sampled daemon RSS exceeded fixed2GiB ceiling','native DirtyJournal.Drain panic present','client socket became unavailable'],
 needsInvestigation:['whether dirty-journal panic preceded or followed monitor-triggered shutdown','native memory growth attribution and complete offline import recovery'],
 rootCauseNotProven:true,weeklyUsage:usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE',
 scriptSHA256:await sha('research/public-task-pilot/scifact-native-fit-v1-failure-audit.mjs')};
fs.writeFileSync(root+'/failure-audit.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(report,null,2));
