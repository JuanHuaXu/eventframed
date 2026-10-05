import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),m=JSON.parse(fs.readFileSync(path.join(out,'service-manifest.json')));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'service-run-results.json')))throw Error('preserve original run');
for(const[arm,files]of Object.entries(m.sources))for(const[f,h]of Object.entries(files))if(hash(fs.readFileSync(path.join(m.paths[arm],f)))!==h)throw Error('source drift '+arm+'/'+f);
const adjacent=['CaptureTurnEnrichesInsideServiceAndObservePreservesAuthoredFields','RecallExcludesUnavailableEventsBeforeCandidateLimit','RerankingRunsBeforeIndependentPackingCap','CalibrationChangesForecastLawWithoutChangingRetrievalScore','RerankingReceivesOnlyEventFrameCorpus','DefaultPolicyActivatesCompleteBoundedFrontier','VectorHydrationIsExplicit','StoreRoundTripAndAvailabilityGate','DurableStateSurvivesRestartAndPinsEmbeddingContract','BayesianJournalSurvivesRestartAndRejectsConflict','ConcurrentBayesianJournalsRemainDurable','BayesianPosteriorUpdateSurvivesRestart','AntiPigeonRevisionSurvivesRestart','SharedEvidenceDiscountSurvivesRestart','PredictiveSnapInvalidationRollbackAndRestart','DeleteRetentionCompactAndBackup'];
const commands=[];
function run(arm,label,args,extra={}){
 const transcript=label+'.txt',file=path.join(out,transcript),fd=fs.openSync(file,'wx'),start=performance.now();
 const child=spawnSync('go',args,{cwd:m.paths[arm],env:{...process.env,GOMAXPROCS:'10'},stdio:['ignore',fd,fd]});fs.closeSync(fd);
 const data=fs.readFileSync(file);const c={arm,label,transcript,args,...extra,status:child.status,signal:child.signal,error:child.error?.message??null,wallMS:performance.now()-start,sha256:hash(data),bytes:data.length,dataRaceReported:data.includes(Buffer.from('WARNING: DATA RACE'))};commands.push(c);
 fs.writeFileSync(path.join(out,'service-run-results.json'),JSON.stringify({completed:false,commands},null,2)+'\n');console.log(JSON.stringify(c));if(child.status!==0)process.exit(child.status??1);
}
for(const arm of['control','candidate'])for(const race of[false,true])run(arm,`adjacent-${arm}-${race?'race':'ordinary'}`,['test',...(race?['-race']:[]),'./internal/service','./internal/store/libravdbstore','-run','^Test('+adjacent.join('|')+')$','-count=1','-v','-timeout=5m'],{full:false,race});
for(let pair=0;pair<2;pair++)for(const arm of(pair?['candidate','control']:['control','candidate']))run(arm,`service-quality-${pair}-${arm}`,['test','./internal/service','-run','^TestResearchShardedQualityV1$','-count=1','-v','-timeout=10m'],{full:true,pair});
fs.writeFileSync(path.join(out,'service-run-results.json'),JSON.stringify({completed:true,preflightPassed:true,functionalPass:true,commands},null,2)+'\n');
console.log(JSON.stringify({completed:true,commands:commands.length,wholeGoalValidation:false}));
