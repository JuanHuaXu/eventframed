// Independent raw-artifact readback, including native request/response content.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {audit as poolAudit} from './scifact-pool-v1-audit.mjs';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const read=p=>JSON.parse(fs.readFileSync(p));
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const root='research/public-task-pilot/scifact-pool-v1',m=read(root+'/manifest.json');
assert(m.allCommandsTerminal);assert.equal(m.commands,6);assert.equal(Object.keys(m.sources).length,171);
for(const[p,s]of Object.entries(m.sources)){
 assert.equal(await sha(root+'/'+s.copy),s.sha256);
 assert.equal(await sha(p),s.sha256);
}
for(const[p,h]of Object.entries(m.artifacts))assert.equal(await sha(root+'/'+p),h);
const commands=read(root+'/commands.json');assert.equal(commands.length,6);
assert.deepEqual(commands.map(c=>c.name),['race','vet','pool','audit','cost','native-client-build']);
for(const c of commands){assert.equal(c.code,0);assert.equal(c.signal,null);assert(new Date(c.end)>=new Date(c.start));assert.equal(await sha(c.log),c.logSHA256);}
const race=fs.readFileSync(root+'/race.log','utf8');assert(!race.includes('SKIP'));assert(!race.includes('WARNING: DATA RACE'));
const roots=[...['CoverageOwnershipAndIdentity','ResponsesAndWholeFrontier','AvailabilityAndMetadataOnlyPoisoning','FullCorpus'].map(n=>'TestPublicPoolV1'+n),
 ...['DecodeAndBounds','StrictJSONIdentity','CoverageIdentityAndMetadata','AvailabilityOwnershipAndCancellation','RealObserveRecallContract','FullCorpus'].map(n=>'TestPublicFrameV1'+n)];
for(const n of roots){assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${n}`).length,3);assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${n} (`)).length,3);}
const entries=read(root+'/pool.json'),frames=read('research/public-task-pilot/scifact-frame-v1-closure/frames.json');
const inventory=poolAudit(frames,entries),a=read(root+'/audit.json');assert.deepEqual(inventory,a.inventory);assert.equal(a.corruptionRejections,14);
const cost=fs.readFileSync(root+'/cost.log','utf8'),costs={};
for(const n of ['Build','Bind50','Bind200']){
 const lines=cost.split('\n').filter(l=>l.startsWith('BenchmarkPublicPoolV1'+n+'-'));assert.equal(lines.length,3);
 costs[n]=lines.map(l=>{const v=l.match(/\s+3\s+([0-9.]+) ns\/op\s+([0-9.]+) B\/op\s+([0-9.]+) allocs\/op$/);assert(v,l);return{ns:Number(v[1]),bytes:Number(v[2]),allocs:Number(v[3])};});
}
const native=[],verified=new Map();
for(let i=1;i<=6;i++){
 const r='research/public-task-pilot/nativepool-v'+i,f=read(r+'/freeze.json'),nm=read(r+'/manifest.json'),term=read(r+'/daemon-terminal.json');
 assert(nm.allOwnedJobsTerminal);assert.equal(term.ownedProcessOnly,true);assert(term.code!==undefined||term.error);
 for(const[p,s]of Object.entries(nm.inputs)){
  assert.deepEqual(f.inputs[p],s);
  if(!verified.has(p))verified.set(p,{sha256:await sha(p),bytes:fs.statSync(p).size});assert.deepEqual(verified.get(p),s);
 }
 for(const[p,h]of Object.entries(nm.artifacts))assert.equal(await sha(r+'/'+p),h);
 assert.equal(await sha(r+'/daemon.log'),term.logSHA256);
 if(i<6)assert(nm.error,'failed native pilot must remain failed');else{assert.equal(nm.error,null);assert.equal(nm.result.code,0);assert.equal(term.code,0);assert.equal(term.actualPriority,10);assert(term.sampledPeakRSSKiB<2*1024*1024);}
 native.push({variant:i,manifestSHA256:await sha(r+'/manifest.json'),client:nm.result,daemonCode:term.code,error:nm.error,watcherError:term.watcherError,sampledPeakRSSKiB:term.sampledPeakRSSKiB});
}
const probe=read('research/public-task-pilot/native-loader-probe-v1/result.json');assert(probe.wrapperDropsDyldVerified);
assert.deepEqual(probe.results.map(x=>({code:x.code,signal:x.signal,dyld:x.parsed.dyld})),[
 {code:0,signal:null,dyld:'/owned-test-library-path'},{code:0,signal:null,dyld:null}]);
const corpus=read('research/public-task-pilot/scifact-v1/corpus.json');
const pilot=read('research/public-task-pilot/nativepool-v6/result.json');
function verifyPilot(p){
 assert.equal(p.Entries.length,9);assert.equal(p.Results.length,2);assert(p.ImportNS>0);
 assert.equal(p.SearchContract,'libravdb.ipc.v1.LibravDB/SearchTextCollections');assert.equal(p.RankContract,'libravdb.ipc.v1.LibravDB/RankCandidates');
 assert.deepEqual(p.Entries.slice(0,8),entries.slice(0,8));
 const future=p.Entries[8],d=frames[0],record=corpus[0],clock='2026-10-04T01:00:00Z';
 const digest=hash(JSON.stringify(['future-only-native-control',record.title,record.text]));
 const spanIDs=d.spans.map(s=>'public-span-'+hash(JSON.stringify(['public-source-assertion-span-v1','research-scifact','public-import',clock,'future-only-native-control',digest,s.section,s.start,s.end])));
 assert.equal(future.source_id,'future-only-native-control');assert.equal(future.source_sha256,digest);assert.equal(future.available_at,clock);assert.deepEqual(future.span_ids,spanIDs);
 assert.equal(future.candidate.Text,entries[0].candidate.Text);assert.equal(future.candidate.ID,'public-document-'+hash('public-source-document-pool-v1\0'+spanIDs[0]+'\0'+digest));
 assert.equal(p.FutureID,future.candidate.ID);
 const expectedMeta={...JSON.parse(Buffer.from(entries[0].candidate.Metadata,'base64')),ts:Date.parse(clock),source_document_id:future.source_id,source_sha256:digest,available_at:clock,span_ids:spanIDs};
 assert.deepEqual(JSON.parse(Buffer.from(future.candidate.Metadata,'base64')),expectedMeta);
 const byID=new Map(p.Entries.map(e=>[e.candidate.ID,e]));
 const validate=(rows,exclude)=>{
  const seen=new Set();for(const row of rows){const e=byID.get(row.ID);assert(e);assert(!seen.has(row.ID));seen.add(row.ID);
   assert.equal(row.Text,e.candidate.Text);assert(Number.isFinite(row.Score));assert(!exclude.includes(row.ID));
   const actual=JSON.parse(Buffer.from(row.Metadata,'base64')),expected=JSON.parse(Buffer.from(e.candidate.Metadata,'base64'));
   for(const[k,v]of Object.entries(expected))assert.deepEqual(actual[k],v);
  }return seen;
 };
 for(let i=0;i<p.Results.length;i++){
  const q=p.Results[i];assert.equal(q.Query,corpus[i].title);assert.deepEqual(q.Excluded,[p.FutureID]);assert(q.SearchNS>0&&q.RankNS>0);
  const n=validate(q.Search,q.Excluded),r=validate(q.Ranked,q.Excluded);assert(n.size>0);assert.deepEqual([...n].sort(),[...r].sort());
  assert.deepEqual(q.SourceIDs,q.Ranked.map(v=>byID.get(v.ID).source_id));assert(q.SourceIDs.includes(corpus[i].id));
 }
 validate(p.Unexcluded,[]);assert(p.Unexcluded.some(v=>v.ID===p.FutureID));
}
verifyPilot(pilot);
const mutations=[p=>p.Entries.pop(),p=>p.Results.pop(),p=>p.Results[0].Search[0].Text+='changed',
 p=>p.Results[0].Ranked=[],p=>p.Results[0].Ranked.push(p.Results[0].Ranked[0]),p=>p.Results[0].SourceIDs=['other'],
 p=>p.Results[0].Excluded=[],p=>p.FutureID='unknown',p=>p.Unexcluded=p.Unexcluded.filter(v=>v.ID!==p.FutureID),
 p=>p.Entries[8].available_at='2026-10-04T00:00:00Z',p=>p.Results[1].Query='wrong',p=>p.Results[0].Search[0].Score=Infinity];
for(const mutation of mutations){const p=structuredClone(pilot);mutation(p);assert.throws(()=>verifyPilot(p));}
const nativeLog=fs.readFileSync('research/public-task-pilot/nativepool-v6/daemon.log','utf8');
assert(nativeLog.includes('GGUF embedding backend ready'));assert(nativeLog.includes('nomic-embed-text-v1.5.Q8_0.gguf'));
assert(nativeLog.includes('raw hits=8'));assert(nativeLog.includes('post-L7 hits=1'));
assert(nativeLog.includes('dirty-anchor generation is stale')); // preserved native warning, not hidden by exit zero.
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const report={time:new Date().toISOString(),manifestSHA256:await sha(root+'/manifest.json'),sources:171,commands:6,
 actualRaceRoots:roots,executionsPerRoot:3,inventory,costs,poolCorruptionRejections:14,nativePilotCorruptionRejections:mutations.length,
 native,loaderBoundaryVerified:true,nativePilot:{documents:9,queries:2,importNS:pilot.ImportNS,
  queriesMeasured:pilot.Results.map(q=>({searchCandidates:q.Search.length,rankCandidates:q.Ranked.length,searchNS:q.SearchNS,rankNS:q.RankNS})),
  futurePositiveAndNegativeControls:true,raw8Returned1:true,backend:'Q8 GGUF',nativeStaleAnchorWarningPreserved:true},
 weeklyUsage:usage,qualityValidation:false,allSevenWholeGoals:'OPEN',goal:'ACTIVE',scriptSHA256:await sha('research/public-task-pilot/scifact-pool-native-v1-artifact-audit.mjs')};
fs.writeFileSync(root+'/artifact-audit.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({sources:171,commands:6,raceRoots:10,inventory,nativePilot:report.nativePilot,corruptions:14+mutations.length,allSevenWholeGoals:'OPEN'},null,2));
