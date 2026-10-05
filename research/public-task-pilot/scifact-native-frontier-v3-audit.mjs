// Independent full-corpus/whole-frontier audit. Outcome access is below the
// complete structural, immutable-input and actual process-terminal checks.
import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';import readline from 'node:readline';
import {metrics} from './scifact-native-fit-v1-audit.mjs';
const read=p=>JSON.parse(fs.readFileSync(p)),sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const root='research/public-task-pilot/scifact-native-frontier-v3-executed',native='research/public-task-pilot/nativefrontier-v3';
const m=read(root+'/manifest.json'),n=read(native+'/manifest.json'),term=read(native+'/daemon-terminal.json');
assert(m.allCommandsTerminal&&n.allOwnedJobsTerminal);assert.equal(m.nativeCode,0);assert.equal(n.result.code,0);assert.equal(term.code,0);
assert.equal(n.error,null);assert.equal(term.watcherError,null);assert.equal(term.actualPriority,10);assert(term.sampledPeakRSSKiB<=2*1024*1024);
for(const[p,s]of Object.entries(m.sources)){assert.equal(await sha(p),s.sha256);assert.equal(await sha(root+'/'+s.copy),s.sha256);}
for(const[p,h]of Object.entries(m.artifacts))if(typeof h==='string')assert.equal(await sha(root+'/'+p),h);else assert.equal(await sha(h.path),h.sha256);
for(const[p,h]of Object.entries(n.artifacts))assert.equal(await sha(native+'/'+p),h);for(const[p,s]of Object.entries(n.inputs))assert.equal(await sha(p),s.sha256);
assert.equal(Object.keys(m.sources).length,39);assert(m.completeTestDependencyClosure);
const commands=read(root+'/commands.json');assert.deepEqual(commands.map(c=>[c.name,c.code,c.signal]),[['race',0,null],['vet',0,null],['build',0,null],['native',0,null]]);for(const c of commands)assert.equal(await sha(c.log),c.logSHA256);
const copies=read(native+'/copies-before.json'),previous=read('research/public-task-pilot/scifact-native-resume-v2/failure-audit.json');assert.equal(copies.length,4);
for(const c of copies){assert.equal(await sha(c.original),c.sha256);assert.equal(n.inputs[c.original].sha256,c.sha256);assert.equal(c.sha256,previous.preservedStores.find(s=>s.path===c.original.replace(process.cwd()+'/','')).sha256);}
const pool=read('research/public-task-pilot/scifact-pool-v1/pool.json'),byID=new Map(pool.map(e=>[e.candidate.ID,e]));
const prefix=read(root+'/prefix.json'),fit=read(root+'/fit-only.json'),plan=read('research/public-task-pilot/scifact-v1/split-plan.json'),mixed=read('research/public-task-pilot/scifact-v1/queries.json');
assert.equal(prefix.ids.length,5183);assert.deepEqual(prefix.ids,pool.map(e=>e.candidate.ID));assert.equal(fit.partition,'fit');assert.deepEqual(fit.queries.map(q=>q.id),plan.splits.fit);
const qmap=new Map(mixed.map(q=>[q.id,q]));for(const q of fit.queries)assert.deepEqual(q,qmap.get(q.id));
const heldOut=new Set([...plan.splits.calibration,...plan.splits.confirmation,...plan.splits.excludedOfficialTrain]);assert(fit.queries.every(q=>!heldOut.has(q.id)));
function bind(rows,cap=200){assert(Array.isArray(rows)&&rows.length<=cap);const seen=new Set();
 for(const c of rows){const e=byID.get(c.ID);assert(e);assert(!seen.has(c.ID));seen.add(c.ID);assert.equal(c.Text,e.candidate.Text);assert(Number.isFinite(c.Score));
  const actual=JSON.parse(Buffer.from(c.Metadata,'base64')),expected=JSON.parse(Buffer.from(e.candidate.Metadata,'base64'));
  for(const k of ['collection','ts','source_document_id','source_sha256','available_at','frame_contract','pool_contract','span_ids'])assert.deepEqual(actual[k],expected[k]);}
 return rows.map(c=>byID.get(c.ID).source_id);
}
function queryCheck(r){
 assert.equal(r.k,200);assert.equal(r.k1,r.search.length);assert.equal(r.k2,r.search.length);assert.equal(r.nativeCalled,r.search.length>0);
 const a=bind(r.search??[]),b=bind(r.ranked??[]),advice=bind(r.nativeRanked??[]);assert.deepEqual([...a].sort(),[...b].sort());assert.deepEqual(r.sources,b);
 const nom=new Set(r.search.map(c=>c.ID));assert((r.nativeRanked??[]).every(c=>nom.has(c.ID)));
 assert.equal(r.nativeComplete,advice.length===a.length);
 if(r.nativeComplete){assert.deepEqual(r.ranked,r.nativeRanked);}
 else{assert.deepEqual(r.ranked,r.search);}
 for(const name of ['searchNS','bindNS','rankNS','rankBindNS'])assert(Number.isFinite(r[name])&&r[name]>=0);
}
let start=null,complete=null;const verified=[],queries=[];
const stream=readline.createInterface({input:fs.createReadStream(native+'/trace.ndjson'),crlfDelay:Infinity});
for await(const line of stream){const r=JSON.parse(line);assert(!complete);
 if(r.kind==='start'){assert(!start);start=r;assert.equal(r.documents,5183);assert.equal(r.queries,351);assert.equal(r.partition,'fit');}
 else if(r.kind==='verify'){assert(start&&queries.length===0);assert.equal(r.index,verified.length);assert(r.ok);assert.equal(r.id,prefix.ids[r.index]);assert.equal(r.source,pool[r.index].source_id);
  assert.equal(r.rows.length,1);assert.equal(r.rows[0].ID,r.id);assert(r.ns>=0);assert.deepEqual(bind(r.rows,1),[r.source]);verified.push(r);}
 else if(r.kind==='query'){assert.equal(verified.length,5183);assert.equal(r.index,queries.length);assert.deepEqual({id:r.id,text:r.text},fit.queries[r.index]);queryCheck(r);queries.push(r);}
 else if(r.kind==='complete'){assert.equal(verified.length,5183);assert.equal(queries.length,351);assert.equal(r.documents,5183);assert.equal(r.queries,351);assert.equal(r.verifiedPrefix,5183);
  assert.equal(r.labelsRead,false);assert.equal(r.confirmationPredictions,0);assert(r.totalNS>=r.importNS);complete=r;}
 else throw new Error('unexpected trace phase: '+r.kind);
}
assert(start&&complete);assert.equal(queries.length,351);
const first=queries.find(q=>q.search.length>=2);assert(first);
const mutations=[q=>q.search[0].Text+='bad',q=>q.search[0].ID='unknown',q=>q.search[0].Metadata=Buffer.from('{}').toString('base64'),
 q=>q.search[0].Score=Infinity,q=>q.ranked.pop(),q=>q.ranked.push(q.ranked[0]),q=>q.sources=['wrong'],q=>q.k2=10,
 q=>q.nativeComplete=!q.nativeComplete,q=>q.nativeCalled=false,q=>q.nativeRanked=[q.search[0],q.search[0]],
 q=>{const ids=new Set(q.search.map(c=>c.ID));q.nativeRanked=[pool.find(e=>!ids.has(e.candidate.ID)).candidate];}];
for(const mutation of mutations){const q=structuredClone(first);mutation(q);assert.throws(()=>queryCheck(q));}
const race=fs.readFileSync(root+'/race.log','utf8');assert(!race.includes('SKIP'));assert(!race.includes('WARNING: DATA RACE'));
const roots=['TestResumePrefixRequiresActualWholeRead','TestResumeRejectsMissingCorruptAndFuture','TestResumeCancellationAndSinkFailure','TestResumeActualTypedRPCAndOwnedSocket',
 'TestCompleteAdviceOrWholeSearch','TestCorruptAdviceNeverFallsBack','TestCompleteRankCapsAndCancellation','TestNativeFitOnlyBoundary','TestNativeFitOwnedPath'];
for(const name of roots){assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${name}`).length,3);assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${name} (`)).length,3);}
// Outcomes are FIT citation relevance only, accessed after complete predictions.
const fitIDs=new Set(plan.splits.fit),labels=read('research/public-task-pilot/scifact-v1/labels-train.json'),sourceIDs=new Set(pool.map(e=>e.source_id));
const targets=new Map(plan.splits.fit.map(id=>[id,new Set()]));for(const l of labels)if(fitIDs.has(l.query)){assert.equal(l.score,1);assert(sourceIDs.has(l.document));targets.get(l.query).add(l.document);}for(const v of targets.values())assert(v.size>0);
const details=queries.map(q=>({id:q.id,candidates:q.search.length,nativeCandidates:q.nativeRanked.length,nativeComplete:q.nativeComplete,
 search:metrics(bind(q.search),targets.get(q.id)),completeOrSearch:metrics(q.sources,targets.get(q.id)),
 searchNS:q.searchNS,rankNS:q.rankNS,bindNS:q.bindNS,rankBindNS:q.rankBindNS,relevant:[...targets.get(q.id)].sort()}));
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length,summary={};for(const arm of ['search','completeOrSearch'])summary[arm]=Object.fromEntries(Object.keys(details[0][arm]).map(k=>[k,mean(details.map(d=>d[arm][k]))]));
const percentile=(a,p)=>[...a].sort((a,b)=>a-b)[Math.ceil(a.length*p)-1],distribution=a=>({min:Math.min(...a),mean:mean(a),p50:percentile(a,.5),p95:percentile(a,.95),p99:percentile(a,.99),max:Math.max(...a)});
const unitMeans=plan.units.fit.map(id=>{const c=plan.components.find(c=>c.id===id);assert(c);const rows=details.filter(d=>c.queries.includes(d.id));assert(rows.length);return{id,queries:rows.length,
 search:Object.fromEntries(Object.keys(rows[0].search).map(k=>[k,mean(rows.map(r=>r.search[k]))])),completeOrSearch:Object.fromEntries(Object.keys(rows[0].completeOrSearch).map(k=>[k,mean(rows.map(r=>r.completeOrSearch[k]))]))};});assert.equal(unitMeans.length,221);
const original=read('research/public-task-pilot/scifact-native-fit-v1/failure-audit.json'),currentProcessMS=Date.parse(term.ended)-Date.parse(term.began);
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const report={time:new Date().toISOString(),manifestSHA256:await sha(root+'/manifest.json'),nativeManifestSHA256:await sha(native+'/manifest.json'),traceSHA256:await sha(native+'/trace.ndjson'),
 scriptSHA256:await sha('research/public-task-pilot/scifact-native-frontier-v3-audit.mjs'),labelsSHA256:await sha('research/public-task-pilot/scifact-v1/labels-train.json'),
 sources:39,commands:4,completeTestDependencyClosure:true,raceRoots:9,executionsPerRoot:3,corruptionRejections:mutations.length,
 originalStoresImmutable:true,all5183StoredRecordsVerified:true,documents:5183,queries:351,units:221,summary,
 candidates:distribution(details.map(d=>d.candidates)),undersizedFrontiers:details.filter(d=>d.candidates<50).length,
 nativeAdvice:{completeNonempty:details.filter(d=>d.candidates>0&&d.nativeComplete).length,rejectedIncomplete:details.filter(d=>!d.nativeComplete).length,
  emptyOnNonempty:details.filter(d=>d.candidates>0&&d.nativeCandidates===0).length,emptyNominations:details.filter(d=>d.candidates===0).length},
 costs:{verifyRPCNS:verified.reduce((s,r)=>s+r.ns,0),clientPreparationNS:complete.importNS,clientTotalNS:complete.totalNS,currentProcessMS,
  priorFailedProcessMS:original.cost.wholeOwnedProcessMS,priorResumeProcessMS:previous.cost.processMS,
  totalOwnedProcessMS:original.cost.wholeOwnedProcessMS+previous.cost.processMS+currentProcessMS,
  sampledDaemonPeakRSSKiB:term.sampledPeakRSSKiB,searchNS:distribution(details.map(d=>d.searchNS)),rankNS:distribution(details.map(d=>d.rankNS)),
  servingRPCAndBindNS:distribution(details.map(d=>d.searchNS+d.rankNS+d.bindNS+d.rankBindNS))},
 details,unitMeans,fitDiagnosticOnly:true,noLearnerFittingOrSelection:true,citationRelevanceNotTruth:true,
 nativeRankFailureNotReclassifiedAsSuccess:true,confirmationPredictions:0,calibrationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE',weeklyUsage:usage};
fs.writeFileSync(root+'/audit-results.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({sources:39,documents:5183,queries:351,units:221,summary,candidates:report.candidates,nativeAdvice:report.nativeAdvice,costs:report.costs,allSevenWholeGoals:'OPEN'},null,2));
