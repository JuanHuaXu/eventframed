// An evaluator, not part of prediction. Open only fit citation labels AFTER the
// client/daemon have terminated and the entire 351-query prediction is present.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import readline from 'node:readline';
import {fileURLToPath} from 'node:url';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const read=p=>JSON.parse(fs.readFileSync(p));
export function metrics(order, relevant) {
 assert(relevant.size>0);assert.equal(new Set(order).size,order.length);
 const gain=k=>order.slice(0,k).reduce((s,id)=>s+Number(relevant.has(id)),0);
 const top=order.slice(0,10),first=top.findIndex(id=>relevant.has(id));
 const dcg=top.reduce((s,id,i)=>s+(relevant.has(id)?1/Math.log2(i+2):0),0);
 let ideal=0;for(let i=0;i<Math.min(10,relevant.size);i++)ideal+=1/Math.log2(i+2);
 return{recallFrontier:gain(order.length)/relevant.size,recall10:gain(10)/relevant.size,mrr10:first<0?0:1/(first+1),ndcg10:dcg/ideal};
}
function preflight() {
 assert.deepEqual(metrics(['a','b'],new Set(['a'])),{recallFrontier:1,recall10:1,mrr10:1,ndcg10:1});
 assert.equal(metrics(['x','a'],new Set(['a'])).mrr10,.5);
 assert.equal(metrics(['a'],new Set(['a','b'])).recall10,.5);
 assert.deepEqual(metrics([],new Set(['a'])),{recallFrontier:0,recall10:0,mrr10:0,ndcg10:0});
 assert.equal(metrics(Array.from({length:10},(_,i)=>'x'+i).concat('a'),new Set(['a'])).recallFrontier,1);
 assert.equal(metrics(Array.from({length:10},(_,i)=>'x'+i).concat('a'),new Set(['a'])).recall10,0);
 assert.throws(()=>metrics(['a','a'],new Set(['a'])));assert.throws(()=>metrics([],new Set()));
}
preflight();
async function main(){
 const root='research/public-task-pilot/scifact-native-fit-v1',native='research/public-task-pilot/nativefit-v1';
 const m=read(root+'/manifest.json'),nm=read(native+'/manifest.json'),term=read(native+'/daemon-terminal.json');
 assert(m.allCommandsTerminal&&nm.allOwnedJobsTerminal);assert.equal(m.nativeCode,0);assert.equal(nm.error,null);assert.equal(nm.result.code,0);assert.equal(term.code,0);
 assert.equal(term.actualPriority,10);assert(term.sampledPeakRSSKiB<=2*1024*1024);assert.equal(term.watcherError,null);
 const commands=read(root+'/commands.json');assert.equal(commands.length,4);
 for(const c of commands){assert.equal(c.code,0);assert.equal(c.signal,null);assert.equal(await sha(c.log),c.logSHA256);}
 for(const[p,s]of Object.entries(m.sources)){assert.equal(await sha(p),s.sha256);assert.equal(await sha(root+'/'+s.copy),s.sha256);}
 for(const[p,h]of Object.entries(m.artifacts))if(typeof h==='string')assert.equal(await sha(root+'/'+p),h);else assert.equal(await sha(h.path),h.sha256);
 for(const[p,h]of Object.entries(nm.artifacts))assert.equal(await sha(native+'/'+p),h);
 for(const[p,s]of Object.entries(nm.inputs))assert.equal(await sha(p),s.sha256);
 const plan=read('research/public-task-pilot/scifact-v1/split-plan.json'),projection=read(root+'/fit-only.json'),mixed=read('research/public-task-pilot/scifact-v1/queries.json');
 assert.equal(projection.partition,'fit');assert.deepEqual(projection.queries.map(q=>q.id),plan.splits.fit);
 const qmap=new Map(mixed.map(q=>[q.id,q]));for(const q of projection.queries)assert.deepEqual(q,qmap.get(q.id));
 const heldOut=new Set([...plan.splits.calibration,...plan.splits.confirmation,...plan.splits.excludedOfficialTrain]);
 assert(projection.queries.every(q=>!heldOut.has(q.id)));
 const pool=read('research/public-task-pilot/scifact-pool-v1/pool.json'),byID=new Map(pool.map(e=>[e.candidate.ID,e]));
 const bind=rows=>{
  assert(Array.isArray(rows));assert(rows.length<=200);const seen=new Set();
  for(const c of rows){const e=byID.get(c.ID);assert(e);assert(!seen.has(c.ID));seen.add(c.ID);assert.equal(c.Text,e.candidate.Text);assert(Number.isFinite(c.Score));
   const meta=JSON.parse(Buffer.from(c.Metadata,'base64')),want=JSON.parse(Buffer.from(e.candidate.Metadata,'base64'));
   // Native access/authority decoration cannot alter canonical source binding.
   for(const k of ['collection','ts','source_document_id','source_sha256','available_at','frame_contract','pool_contract','span_ids'])assert.deepEqual(meta[k],want[k]);
  }return rows.map(c=>byID.get(c.ID).source_id);
 };
 let start=null,complete=null;const imports=[],queries=[];
 const stream=readline.createInterface({input:fs.createReadStream(native+'/trace.ndjson'),crlfDelay:Infinity});
 for await(const line of stream){assert(line.length>0);const row=JSON.parse(line);
  assert(!complete,'output after completion');
  if(row.kind==='start'){assert(!start&&imports.length===0&&queries.length===0);start=row;assert.equal(row.documents,5183);assert.equal(row.queries,351);assert.equal(row.partition,'fit');}
  else if(row.kind==='import'){assert(start&&queries.length===0);assert.equal(row.index,imports.length);assert(row.ok);assert.equal(row.id,pool[row.index].candidate.ID);assert.equal(row.source,pool[row.index].source_id);assert(row.ns>0);imports.push(row);}
  else if(row.kind==='query'){assert.equal(imports.length,5183);assert.equal(row.index,queries.length);assert.deepEqual({id:row.id,text:row.text},projection.queries[row.index]);assert.equal(row.k,200);
   assert.equal(row.k1,row.search.length);assert.equal(row.k2,row.search.length);
   const a=bind(row.search??[]),b=bind(row.ranked??[]);assert.deepEqual([...a].sort(),[...b].sort());assert.deepEqual(row.sources,b);
   for(const n of ['searchNS','bindNS','rankNS','rankBindNS'])assert(Number.isFinite(row[n])&&row[n]>=0);
   queries.push(row);
  }else if(row.kind==='complete'){assert.equal(imports.length,5183);assert.equal(queries.length,351);assert.equal(row.documents,5183);assert.equal(row.queries,351);assert.equal(row.labelsRead,false);assert.equal(row.confirmationPredictions,0);assert(row.totalNS>=row.importNS);complete=row;}
  else throw new Error('unexpected trace phase');
 }
 assert(complete);assert.equal(queries.length,351);
 const race=fs.readFileSync(root+'/race.log','utf8');assert(!race.includes('SKIP'));assert(!race.includes('DATA RACE'));
 for(const n of ['TestNativeFitOnlyBoundary','TestNativeFitOwnedPath']){assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${n}`).length,3);assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${n} (`)).length,3);}
 // Outcome access starts here, after all structural/temporal checks above.
 const fitIDs=new Set(plan.splits.fit),labels=read('research/public-task-pilot/scifact-v1/labels-train.json');
 const targets=new Map(plan.splits.fit.map(id=>[id,new Set()]));
 for(const l of labels)if(fitIDs.has(l.query)){assert.equal(l.score,1);assert(pool.some(e=>e.source_id===l.document));targets.get(l.query).add(l.document);}
 for(const v of targets.values())assert(v.size>0);
 const details=queries.map(q=>({id:q.id,candidates:q.search.length,search:metrics(bind(q.search),targets.get(q.id)),ranked:metrics(q.sources,targets.get(q.id)),
  searchNS:q.searchNS,rankNS:q.rankNS,bindNS:q.bindNS,rankBindNS:q.rankBindNS,relevant:[...targets.get(q.id)].sort()}));
 const mean=a=>a.reduce((s,v)=>s+v,0)/a.length,summary={};
 for(const arm of ['search','ranked'])summary[arm]=Object.fromEntries(Object.keys(details[0][arm]).map(k=>[k,mean(details.map(d=>d[arm][k]))]));
 const percentile=(a,p)=>[...a].sort((a,b)=>a-b)[Math.ceil(a.length*p)-1];
 const distribution=a=>({min:Math.min(...a),mean:mean(a),p50:percentile(a,.5),p95:percentile(a,.95),p99:percentile(a,.99),max:Math.max(...a)});
 const unitMeans=plan.units.fit.map(unit=>{const c=plan.components.find(c=>c.id===unit);assert(c);const rows=details.filter(d=>c.queries.includes(d.id));assert(rows.length);return{id:unit,queries:rows.length,
  search:Object.fromEntries(Object.keys(rows[0].search).map(k=>[k,mean(rows.map(r=>r.search[k]))])),
  ranked:Object.fromEntries(Object.keys(rows[0].ranked).map(k=>[k,mean(rows.map(r=>r.ranked[k]))]))};});
 assert.equal(unitMeans.length,221);
 const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
 const report={time:new Date().toISOString(),manifestSHA256:await sha(root+'/manifest.json'),nativeManifestSHA256:await sha(native+'/manifest.json'),
  traceSHA256:await sha(native+'/trace.ndjson'),labelsSHA256:await sha('research/public-task-pilot/scifact-v1/labels-train.json'),
  scriptSHA256:await sha('research/public-task-pilot/scifact-native-fit-v1-audit.mjs'),sources:Object.keys(m.sources).length,
  completeTestDependencyClosure:true,commands:4,raceRoots:2,executionsPerRoot:3,documents:imports.length,queries:queries.length,units:221,
  summary,candidates:distribution(details.map(d=>d.candidates)),undersizedFrontiers:details.filter(d=>d.candidates<50).length,
  costs:{importNS:complete.importNS,totalNS:complete.totalNS,sumImportRPCNS:imports.reduce((s,d)=>s+d.ns,0),
   searchNS:distribution(details.map(d=>d.searchNS)),rankNS:distribution(details.map(d=>d.rankNS)),
   servingRPCAndBindNS:distribution(details.map(d=>d.searchNS+d.rankNS+d.bindNS+d.rankBindNS)),sampledDaemonRSSKiB:term.sampledPeakRSSKiB},
  details,unitMeans,labelsMean:'citation relevance only',fitDiagnosticOnly:true,noFitting:true,confirmationPredictions:0,calibrationPredictions:0,
  weeklyUsage:usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
 fs.writeFileSync(root+'/audit-results.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
 console.log(JSON.stringify({documents:report.documents,queries:report.queries,units:221,sources:report.sources,summary,candidates:report.candidates,
  undersizedFrontiers:report.undersizedFrontiers,costs:report.costs,confirmationPredictions:0,allSevenWholeGoals:'OPEN'},null,2));
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 if(process.argv[2]==='preflight')console.log('independent metric controls PASS');else await main();
}
