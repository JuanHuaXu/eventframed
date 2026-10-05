import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';

const out=path.dirname(fileURLToPath(import.meta.url)),read=f=>fs.readFileSync(path.join(out,f)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const m=JSON.parse(read('manifest.json')),arms=['rebuild','overlay','shard','both'],orders=[arms,[...arms].reverse()];
if(fs.existsSync(path.join(out,'run-started.json')))throw Error('preserve existing run; inspect the original process before resuming');
const check=()=>{if(hash(read('PROTOCOL.md'))!==m.protocolSHA256)throw Error('protocol drift');for(const[a,files]of Object.entries(m.sources))for(const[f,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(m.paths[a],f)))!==h)throw Error('source drift '+a+'/'+f);};
const adjacent=['CaptureTurnEnrichesInsideServiceAndObservePreservesAuthoredFields','RecallExcludesUnavailableEventsBeforeCandidateLimit','RerankingRunsBeforeIndependentPackingCap','CalibrationChangesForecastLawWithoutChangingRetrievalScore','RerankingReceivesOnlyEventFrameCorpus','DefaultPolicyActivatesCompleteBoundedFrontier','VectorHydrationIsExplicit','StoreRoundTripAndAvailabilityGate','DurableStateSurvivesRestartAndPinsEmbeddingContract','BayesianJournalSurvivesRestartAndRejectsConflict','ConcurrentBayesianJournalsRemainDurable','BayesianPosteriorUpdateSurvivesRestart','AntiPigeonRevisionSurvivesRestart','SharedEvidenceDiscountSurvivesRestart','PredictiveSnapInvalidationRollbackAndRestart','DeleteRetentionCompactAndBackup'].map(x=>'Test'+x);
const plan=[];
for(const[a,old]of[['library-control','control-ordinary'],['library-candidate','preflight-full-ordinary']]){
 const receipt=JSON.parse(read('../index-overlay-v6/'+old+'.json'));
 const required=[...read('../index-overlay-v6/'+old+'.txt').toString().matchAll(/^--- PASS: (\S+) /gm)].map(x=>x[1]);
 for(const race of[false,true])plan.push({arm:a,kind:'preflight',label:a+'-'+(race?'race':'ordinary'),args:[receipt.args[0],...(race?['-race']:[]),...receipt.args.slice(1)],required,race});
}
for(const arm of arms)for(const race of[false,true])plan.push({arm,kind:'preflight',label:arm+'-adjacent-'+(race?'race':'ordinary'),args:['test',...(race?['-race']:[]),'-mod=readonly','./internal/service','./internal/store/libravdbstore','-run','^('+adjacent.join('|')+')$','-count=1','-v','-timeout=5m'],required:adjacent,race});
for(const arm of['shard','both'])for(const race of[false,true])plan.push({arm,kind:'preflight',label:arm+'-topology-'+(race?'race':'ordinary'),args:['test',...(race?['-race']:[]),'-mod=readonly','./internal/store/libravdbstore','-run','^TestResearchShardedTopologyReopenV1$','-count=1','-v','-timeout=90s'],required:['TestResearchShardedTopologyReopenV1'],race});
for(const arm of['overlay','both'])plan.push({arm,kind:'race-load',label:arm+'-load-race',args:['test','-race','-mod=readonly','./internal/service','-run','^TestResearchPublicCaptureLoadV1$','-count=1','-v','-timeout=4m'],required:[],race:true,frontier:200});
for(let pair=0;pair<2;pair++)for(const arm of orders[pair])plan.push({arm,kind:'quality',label:`quality-${pair}-${arm}`,args:['test','-mod=readonly','./internal/service','-run','^TestResearchShardedQualityV1$','-count=1','-v','-timeout=10m'],required:['TestResearchShardedQualityV1'],pair});
for(const frontier of[50,200])for(let pair=0;pair<2;pair++)for(const arm of orders[pair])plan.push({arm,kind:'load',label:`load-${frontier}-${pair}-${arm}`,args:['test','-mod=readonly','./internal/service','-run','^TestResearchPublicCaptureLoadV1$','-count=1','-v','-timeout=4m'],required:['TestResearchPublicCaptureLoadV1'],frontier,pair});
check();
fs.writeFileSync(path.join(out,'run-started.json'),JSON.stringify({startedAt:new Date().toISOString(),runnerSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),manifestSHA256:hash(read('manifest.json')),plan,workers:10,orders,wholeGoalValidation:false},null,2)+'\n');
const commands=[];
function saved(completed,preflightPassed){fs.writeFileSync(path.join(out,'run-results.json'),JSON.stringify({completed,preflightPassed,commands,wholeGoalValidation:false,productionTouched:false,privateDataUsed:false},null,2)+'\n');}
for(const p of plan){
 check();console.log(JSON.stringify({started:true,...p}));const transcript=p.label+'.txt',file=path.join(out,transcript),fd=fs.openSync(file,'wx'),start=performance.now();
 const child=spawnSync('go',p.args,{cwd:m.paths[p.arm],env:{...process.env,GOMAXPROCS:'10',...(p.frontier?{EVENTFRAME_PUBLIC_FRONTIER:String(p.frontier)}:{})},stdio:['ignore',fd,fd]});fs.closeSync(fd);check();
 const b=fs.readFileSync(file),t=b.toString(),passNames=[...t.matchAll(/^--- PASS: (\S+) /gm)].map(x=>x[1]),skipNames=[...t.matchAll(/^--- SKIP: (\S+) /gm)].map(x=>x[1]);
 const c={...p,transcript,command:['go',...p.args],status:child.status,signal:child.signal,error:child.error?.message??null,wallMS:performance.now()-start,sha256:hash(b),bytes:b.length,passNames,skipNames,dataRaceReported:/WARNING: DATA RACE/.test(t)};
 c.functional=c.status===0&&!c.dataRaceReported&&skipNames.length===0&&p.required.every(x=>passNames.includes(x));
 if(p.kind==='load'||p.kind==='race-load'){
  const matches=[...t.matchAll(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/g)];c.trace=matches.length===1?JSON.parse(matches[0][1]):null;
  c.testErrors=[...t.matchAll(/^\s+public_capture_load_test.go:\d+: (.+)$/gm)].map(x=>x[1]).filter(x=>!x.startsWith('PUBLIC_CAPTURE_TRACE='));
  c.instrumentedTimingOnly=p.kind==='race-load'&&c.status===1&&c.testErrors.length>0&&c.testErrors.every(x=>['frozen recall p99 <100ms screen failed','frozen publication p99 <250ms screen failed'].includes(x));
  const valid=c.trace?.functional===true&&!c.dataRaceReported&&!/panic:|fatal error:|--- SKIP:/.test(t);
  c.functional=valid&&(c.status===0||c.instrumentedTimingOnly||(p.kind==='load'&&c.status===1&&c.testErrors.length>0&&c.testErrors.every(x=>['frozen recall p99 <100ms screen failed','frozen publication p99 <250ms screen failed'].includes(x))));
 }
 commands.push(c);const preflightPassed=commands.filter(x=>x.kind==='preflight'||x.kind==='race-load').every(x=>x.functional);saved(false,preflightPassed);
 console.log(JSON.stringify({finished:true,arm:c.arm,kind:c.kind,label:c.label,status:c.status,functional:c.functional,wallMS:c.wallMS,testErrors:c.testErrors}));
 if((p.kind==='preflight'||p.kind==='race-load')&&!c.functional){console.log(JSON.stringify({stopped:true,reason:'functional preflight failure',completed:false}));process.exit(1);}
 if(p.kind==='quality'&&!c.functional){console.log(JSON.stringify({stopped:true,reason:'quality fixture execution failure',completed:false}));process.exit(1);}
}
saved(true,true);console.log(JSON.stringify({completed:true,commands:commands.length,wholeGoalValidation:false}));
