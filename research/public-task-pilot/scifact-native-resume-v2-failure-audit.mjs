// Verify the recovered full import and preserve the distinct native rank failure.
// No citation labels or scoring of missing predictions.
import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';import readline from 'node:readline';
const read=p=>JSON.parse(fs.readFileSync(p)),sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const root='research/public-task-pilot/scifact-native-resume-v2',native='research/public-task-pilot/nativeresume-v2';
const m=read(root+'/manifest.json'),n=read(native+'/manifest.json'),term=read(native+'/daemon-terminal.json');
assert(m.allCommandsTerminal&&n.allOwnedJobsTerminal);assert.equal(m.nativeCode,1);assert.equal(n.result.code,1);assert.equal(term.code,0);
assert.equal(term.watcherError,null);assert.equal(term.actualPriority,10);assert(term.sampledPeakRSSKiB<2*1024*1024);
assert.equal(Object.keys(m.sources).length,36);assert(m.completeTestDependencyClosure);
for(const[p,s]of Object.entries(m.sources)){assert.equal(await sha(p),s.sha256);assert.equal(await sha(root+'/'+s.copy),s.sha256);}
for(const[p,h]of Object.entries(m.artifacts))if(typeof h==='string')assert.equal(await sha(root+'/'+p),h);else assert.equal(await sha(h.path),h.sha256);
for(const[p,h]of Object.entries(n.artifacts))assert.equal(await sha(native+'/'+p),h);for(const[p,s]of Object.entries(n.inputs))assert.equal(await sha(p),s.sha256);
const commands=read(root+'/commands.json');assert.deepEqual(commands.map(c=>[c.name,c.code,c.signal]),[['race',0,null],['vet',0,null],['build',0,null],['native',1,null]]);
for(const c of commands)assert.equal(await sha(c.log),c.logSHA256);
const copied=read(native+'/copies-before.json');assert.equal(copied.length,4);for(const c of copied)assert.equal(await sha(c.original),c.sha256);
const pool=read('research/public-task-pilot/scifact-pool-v1/pool.json'),prefix=read(root+'/prefix.json');assert.deepEqual(prefix.ids,pool.slice(0,4250).map(e=>e.candidate.ID));
let started=false;const verified=[],imports=[];let completeQueries=0;
const stream=readline.createInterface({input:fs.createReadStream(native+'/trace.ndjson'),crlfDelay:Infinity});
for await(const line of stream){const r=JSON.parse(line);
 if(r.kind==='start'){assert(!started);started=true;assert.equal(r.documents,5183);assert.equal(r.queries,351);}
 else if(r.kind==='verify'){
  assert(started&&imports.length===0);assert.equal(r.index,verified.length);assert.equal(r.id,pool[r.index].candidate.ID);assert.equal(r.source,pool[r.index].source_id);assert(r.ok&&r.ns>=0);
  assert.equal(r.rows.length,1);const row=r.rows[0],e=pool[r.index];assert.equal(row.ID,e.candidate.ID);assert.equal(row.Text,e.candidate.Text);assert(Number.isFinite(row.Score));
  const actual=JSON.parse(Buffer.from(row.Metadata,'base64')),expected=JSON.parse(Buffer.from(e.candidate.Metadata,'base64'));
  for(const k of ['collection','ts','source_document_id','source_sha256','available_at','frame_contract','pool_contract','span_ids'])assert.deepEqual(actual[k],expected[k]);verified.push(r);
 }else if(r.kind==='import'){assert.equal(verified.length,4250);assert.equal(r.index,4250+imports.length);assert.equal(r.id,pool[r.index].candidate.ID);assert.equal(r.source,pool[r.index].source_id);assert(r.ok&&r.ns>0);imports.push(r);}
 else if(r.kind==='query')completeQueries++;else throw new Error('unexpected completed phase in failed run');
}
assert.equal(verified.length,4250);assert.equal(imports.length,933);assert.equal(completeQueries,0);
const log=fs.readFileSync(native+'/daemon.log','utf8'),fit=read(root+'/fit-only.json');
assert(log.includes('post-L7 hits=200 for query='+JSON.stringify(fit.queries[0].text).replaceAll('"','\\"'))||log.includes('post-L7 hits=200'));
assert(log.includes('raw_union=200 cert_union=0 total_union=200 top_k_output=0'));
assert(fs.readFileSync(native+'/client.log','utf8').includes('native ranker shortened public frontier'));assert(!log.includes('panic:'));
const race=fs.readFileSync(root+'/race.log','utf8');for(const name of ['TestResumePrefixRequiresActualWholeRead','TestResumeRejectsMissingCorruptAndFuture','TestResumeCancellationAndSinkFailure','TestResumeActualTypedRPCAndOwnedSocket','TestNativeFitOnlyBoundary','TestNativeFitOwnedPath']){
 assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${name}`).length,3);assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${name} (`)).length,3);}
assert(!race.includes('SKIP')&&!race.includes('DATA RACE'));
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const report={time:new Date().toISOString(),manifestSHA256:await sha(root+'/manifest.json'),nativeManifestSHA256:await sha(native+'/manifest.json'),traceSHA256:await sha(native+'/trace.ndjson'),
 sources:36,commands:4,allCommandsTerminal:true,actualVerifiedPrefix:4250,newAcknowledgedImports:933,fullCorpusImported:5183,
 fullPostResumeDurabilityUnverified:true,fitQueriesAttempted:1,fitQueriesCompleted:0,calibrationPredictions:0,confirmationPredictions:0,evaluatorLabelsRead:false,
 nativeRankFailure:{searchReturned200:true,nativeRankLogReturned0:true,rawResponseBytesNotJournaledBeforeGate:true,gateNotRelaxed:true},
 cost:{verifyRPCNS:verified.reduce((s,r)=>s+r.ns,0),newImportRPCNS:imports.reduce((s,r)=>s+r.ns,0),processMS:Date.parse(term.ended)-Date.parse(term.began),sampledPeakRSSKiB:term.sampledPeakRSSKiB},
 preservedStores:await Promise.all(['store.libravdb','daemon.libravdb','store.libravdb.embedding.json','daemon.libravdb.embedding.json'].map(async p=>({path:native+'/'+p,sha256:await sha(native+'/'+p),bytes:fs.statSync(native+'/'+p).size}))),
 nativeMemoryLeakOrPanicRootCauseNotFixed:true,allSevenWholeGoals:'OPEN',goal:'ACTIVE',weeklyUsage:usage,
 scriptSHA256:await sha('research/public-task-pilot/scifact-native-resume-v2-failure-audit.mjs')};
fs.writeFileSync(root+'/failure-audit.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(report,null,2));
