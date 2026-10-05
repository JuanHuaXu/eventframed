// Separate evaluator: full prediction reconstruction precedes FIT label access.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import readline from 'node:readline';
import {metrics} from './scifact-native-fit-v1-audit.mjs';
const root='research/public-task-pilot/scifact-hybrid-v1',prior='research/public-task-pilot/scifact-native-frontier-v3-executed';
const read=p=>JSON.parse(fs.readFileSync(p));
const sha=async p=>{const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex');};
const m=read(root+'/manifest.json'),commands=read(root+'/commands.json');
assert(m.allCommandsTerminal&&m.completeTestDependencyClosure);assert.equal(commands.length,5);
assert.deepEqual(commands.map(c=>c.name),['race','vet','build','predict','bench']);
for(const c of commands){assert.equal(c.code,0);assert.equal(c.signal,null);assert.equal(await sha(c.log),c.logSHA256);}
for(const[p,s]of Object.entries(m.sources)){assert.equal(await sha(p),s.sha256);assert.equal(await sha(root+'/'+s.copy),s.sha256);}
for(const[p,v]of Object.entries(m.inputs))assert.equal(await sha(p),v.sha256);
for(const[p,h]of Object.entries(m.artifacts))assert.equal(await sha(root+'/'+p),h);
const old=read(prior+'/audit-results.json');assert(old.all5183StoredRecordsVerified);assert.equal(old.queries,351);
assert.equal(await sha(prior+'/manifest.json'),old.manifestSHA256);assert.equal(await sha('research/public-task-pilot/nativefrontier-v3/trace.ndjson'),old.traceSHA256);
const pool=read('research/public-task-pilot/scifact-pool-v1/pool.json'),fit=read(prior+'/fit-only.json'),plan=read('research/public-task-pilot/scifact-v1/split-plan.json');
assert.equal(pool.length,5183);assert.deepEqual(fit.queries.map(q=>q.id),plan.splits.fit);
const held=new Set([...plan.splits.calibration,...plan.splits.confirmation,...plan.splits.excludedOfficialTrain]);assert(fit.queries.every(q=>!held.has(q.id)));
const native=[];for await(const l of readline.createInterface({input:fs.createReadStream('research/public-task-pilot/nativefrontier-v3/trace.ndjson'),crlfDelay:Infinity})){const q=JSON.parse(l);if(q.kind==='query')native.push(q);}assert.equal(native.length,351);
// Go lowers individual Unicode runes; per-codepoint JS lower avoids Greek
// final-sigma context rules. The public corpus is audited, not silently reduced.
const tokenize=s=>Array.from(s,c=>c==='\u0130'?'i':c.toLowerCase()).join('').split(/[^\p{L}\p{Nd}]+/u).filter(Boolean);
const docs=new Map(),postings=new Map();let tokenCount=0,postingCount=0;
for(const p of [...pool].sort((a,b)=>a.candidate.ID<b.candidate.ID?-1:1)){
 assert(!docs.has(p.candidate.ID));assert(Date.parse(p.available_at)<=Date.parse('2026-10-04T00:00:00Z'));
 const terms=tokenize(p.candidate.Text),tf=new Map();assert(terms.length);for(const t of terms)tf.set(t,(tf.get(t)??0)+1);
 docs.set(p.candidate.ID,{source:p.source_id,length:terms.length,text:p.candidate.Text});tokenCount+=terms.length;
 for(const[t,n]of tf){if(!postings.has(t))postings.set(t,[]);postings.get(t).push({id:p.candidate.ID,tf:n});postingCount++;}
}
const avg=tokenCount/docs.size,sort=a=>a.sort((a,b)=>b.score-a.score||(a.id<b.id?-1:a.id>b.id?1:0));
function bm25(query){const values=new Map();for(const t of [...new Set(tokenize(query))].sort()){
 const p=postings.get(t)??[],idf=Math.log1p((docs.size-p.length+.5)/(p.length+.5));
 for(const v of p){const w=idf*v.tf*2.2/(v.tf+1.2*(.25+.75*docs.get(v.id).length/avg));values.set(v.id,(values.get(v.id)??0)+w);}
 }return values;}
function fusion(a,b){const v=new Map();for(const list of [a,b])for(let i=0;i<list.length;i++)v.set(list[i].id,(v.get(list[i].id)??0)+1/(60+i+1));return sort([...v].map(([id,score])=>({id,score}))).slice(0,200);}
function compare(actual,expected){assert.equal(actual.length,expected.length);assert(actual.length<=200);const seen=new Set();
 for(let i=0;i<actual.length;i++){const a=actual[i],e=expected[i];assert(docs.has(a.id));assert(!seen.has(a.id));seen.add(a.id);assert.equal(a.id,e.id);assert(Number.isFinite(a.score));assert(Math.abs(a.score-e.score)<=1e-11*Math.max(1,Math.abs(e.score)));}}
function check(q){const n=native[q.index];assert(n);assert.equal(q.id,n.id);assert.equal(q.text,n.text);
 const nom=n.search.map(h=>({id:h.ID,score:h.Score}));compare(q.native,nom);
 const values=bm25(q.text),lex=sort([...values].map(([id,score])=>({id,score}))).slice(0,200);
 compare(q.lexical,lex);compare(q.withinNative,sort(nom.map(h=>({id:h.id,score:values.get(h.id)??0}))));compare(q.fused,fusion(nom,lex));
 for(const f of ['lexicalNS','rerankNS','fusionNS'])assert(Number.isFinite(q[f])&&q[f]>=0);
}
const rows=[];let start=null,complete=null;
for await(const l of readline.createInterface({input:fs.createReadStream(root+'/predictions.ndjson'),crlfDelay:Infinity})){const q=JSON.parse(l);assert(!complete);
 if(q.kind==='start'){assert(!start);start=q;assert.equal(q.contract,'public-hybrid-bm25-rrf-v1');assert.equal(q.partition,'fit');assert.equal(q.documents,5183);assert.equal(q.queries,351);
 assert.deepEqual(q.stats,{documents:docs.size,terms:postings.size,postings:postingCount,tokens:tokenCount});assert(q.buildNS>0&&q.buildAllocatedBytes>0);}
 else if(q.kind==='query'){assert(start);assert.equal(q.index,rows.length);assert.deepEqual({id:q.id,text:q.text},fit.queries[q.index]);check(q);rows.push(q);}
 else if(q.kind==='complete'){assert.equal(rows.length,351);assert.equal(q.queries,351);assert.equal(q.labelsRead,false);assert.equal(q.confirmationPredictions,0);assert.equal(q.calibrationPredictions,0);complete=q;}
 else throw new Error('unknown prediction phase');
}
assert(start&&complete);assert.equal(rows.length,351);
const first=rows.find(r=>r.native.length>1&&r.lexical.length>1);assert(first);
const mutations=[q=>q.lexical.pop(),q=>q.fused[0].id='unknown',q=>q.lexical[0].score=Infinity,q=>q.lexical[1]=q.lexical[0],q=>q.withinNative.pop(),q=>q.fused.reverse(),q=>q.text+=' bad',q=>q.id='heldout',q=>q.native.reverse(),q=>q.lexicalNS=-1];
for(const mutate of mutations){const q=structuredClone(first);mutate(q);assert.throws(()=>check(q));}
const race=fs.readFileSync(root+'/race.log','utf8');assert(!race.includes('SKIP')&&!race.includes('WARNING: DATA RACE'));
const roots=['TestHybridFormula','TestHybridFutureAndOwnership','TestHybridFullFrontierAndFusion','TestHybridRejectAndCancel','TestHybridConcurrentDeterminism'];
for(const n of roots){assert.equal(race.split('\n').filter(l=>l===`=== RUN   ${n}`).length,3);assert.equal(race.split('\n').filter(l=>l.startsWith(`--- PASS: ${n} (`)).length,3);}
// Only now open FIT citation outcomes, never the official test labels.
const labelsPath='research/public-task-pilot/scifact-v1/labels-train.json',labels=read(labelsPath),targets=new Map(plan.splits.fit.map(id=>[id,new Set()]));
for(const l of labels)if(targets.has(l.query)){assert.equal(l.score,1);assert(pool.some(p=>p.source_id===l.document));targets.get(l.query).add(l.document);}for(const v of targets.values())assert(v.size);
const arms=['native','withinNative','lexical','fused'],mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const details=rows.map(r=>{const d={id:r.id};for(const arm of arms)d[arm]=metrics(r[arm].map(h=>docs.get(h.id).source),targets.get(r.id));return d;});
const summary={};for(const arm of arms)summary[arm]=Object.fromEntries(Object.keys(details[0][arm]).map(k=>[k,mean(details.map(d=>d[arm][k]))]));
assert.deepEqual(summary.native,old.summary.search);
const units=plan.units.fit.map(id=>{const c=plan.components.find(c=>c.id===id),d=details.filter(d=>c.queries.includes(d.id));assert(d.length);const u={id,queries:d.length};for(const arm of arms)u[arm]=Object.fromEntries(Object.keys(summary[arm]).map(k=>[k,mean(d.map(d=>d[arm][k]))]));return u;});assert.equal(units.length,221);
const unitSummary={};for(const arm of arms)unitSummary[arm]=Object.fromEntries(Object.keys(summary[arm]).map(k=>[k,mean(units.map(u=>u[arm][k]))]));
const paired={};for(const arm of arms.filter(a=>a!=='native'))paired[arm]=Object.fromEntries(Object.keys(summary[arm]).map(k=>{const d=units.map(u=>u[arm][k]-u.native[k]);return[k,{mean:mean(d),better:d.filter(v=>v>1e-14).length,worse:d.filter(v=>v< -1e-14).length,tied:d.filter(v=>Math.abs(v)<=1e-14).length}];}));
const pct=(a,p)=>[...a].sort((a,b)=>a-b)[Math.ceil(a.length*p)-1],dist=a=>({min:Math.min(...a),mean:mean(a),p50:pct(a,.5),p95:pct(a,.95),p99:pct(a,.99),max:Math.max(...a)});
const cost={indexBuildNS:start.buildNS,indexAllocatedBytes:start.buildAllocatedBytes,lexicalNS:dist(rows.map(r=>r.lexicalNS)),rerankNS:dist(rows.map(r=>r.rerankNS)),fusionNS:dist(rows.map(r=>r.fusionNS)),
 counterfactualSearchPlusLexicalFusionNS:dist(rows.map((r,i)=>native[i].searchNS+native[i].bindNS+r.lexicalNS+r.fusionNS)),queryPhaseNS:complete.queryPhaseNS,
 observedNativePriorProcessMS:old.costs.totalOwnedProcessMS,offlinePredictCommandMS:Date.parse(commands.find(c=>c.name==='predict').end)-Date.parse(commands.find(c=>c.name==='predict').start),
 liveLoadedServingMeasured:false};
const evictions=rows.map(r=>{const f=new Set(r.fused.map(h=>h.id)),n=new Set(r.native.map(h=>h.id));return{nativeEvicted:r.native.filter(h=>!f.has(h.id)).length,lexicalAdded:r.fused.filter(h=>!n.has(h.id)).length};});
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const report={time:new Date().toISOString(),manifestSHA256:await sha(root+'/manifest.json'),scriptSHA256:await sha('research/public-task-pilot/scifact-hybrid-v1-audit.mjs'),labelsSHA256:await sha(labelsPath),
 sources:Object.keys(m.sources).length,coreCommands:5,raceRoots:5,executionsPerRoot:3,independentQueryReconstructions:351,corruptionRejections:mutations.length,
 documents:5183,queries:351,units:221,stats:start.stats,summary,unitSummary,paired,cost,nomination:{meanNativeEvicted:mean(evictions.map(v=>v.nativeEvicted)),meanLexicalAdded:mean(evictions.map(v=>v.lexicalAdded)),fusedCandidates:dist(rows.map(r=>r.fused.length))},
 fitExploratoryOnly:true,noFittedLearner:true,citationRelevanceNotTruth:true,calibrationPredictions:0,confirmationPredictions:0,allSevenWholeGoals:'OPEN',goal:'ACTIVE',weeklyUsage:usage,details,unitsDetail:units};
fs.writeFileSync(root+'/audit-results.json',JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});delete report.details;delete report.unitsDetail;console.log(JSON.stringify(report,null,2));
