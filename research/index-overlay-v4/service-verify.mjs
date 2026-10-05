import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {evaluate} from './service-evaluate.mjs';
import {independent} from './service-independent.mjs';
const adjacent=['CaptureTurnEnrichesInsideServiceAndObservePreservesAuthoredFields','RecallExcludesUnavailableEventsBeforeCandidateLimit','RerankingRunsBeforeIndependentPackingCap','CalibrationChangesForecastLawWithoutChangingRetrievalScore','RerankingReceivesOnlyEventFrameCorpus','DefaultPolicyActivatesCompleteBoundedFrontier','VectorHydrationIsExplicit','StoreRoundTripAndAvailabilityGate','DurableStateSurvivesRestartAndPinsEmbeddingContract','BayesianJournalSurvivesRestartAndRejectsConflict','ConcurrentBayesianJournalsRemainDurable','BayesianPosteriorUpdateSurvivesRestart','AntiPigeonRevisionSurvivesRestart','SharedEvidenceDiscountSurvivesRestart','PredictiveSnapInvalidationRollbackAndRestart','DeleteRetentionCompactAndBackup'];
export async function verify(root,{sources=true}={}) {
  const read=f=>fs.readFileSync(path.join(root,f)),json=f=>JSON.parse(read(f)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
  const m=json('service-manifest.json'),run=json('service-run-results.json');
  assert.equal(hash(read('SERVICE_PROTOCOL.md')),m.serviceProtocolSHA256);
  assert.equal(hash(read('SOURCE_PINS.json')),m.parentPinsSHA256);
  assert.equal(m.bothUnsharded,true);
  if(sources)for(const [arm,files]of Object.entries(m.sources))for(const [f,h]of Object.entries(files))assert.equal(hash(fs.readFileSync(path.join(m.paths[arm],f))),h,'source drift '+arm+'/'+f);
  assert.equal(run.completed,true);assert.equal(run.preflightPassed,true);assert.equal(run.functionalPass,true);assert.equal(run.commands.length,8);
  const preflight=run.commands.filter(c=>!c.full);assert.equal(preflight.length,4);
  for(const arm of ['control','candidate'])for(const race of [false,true]) {
    const c=preflight.find(c=>c.arm===arm&&c.race===race);assert.ok(c);
    assert.equal(c.status,0);assert.equal(c.signal,null);assert.equal(c.error,null);assert.equal(c.dataRaceReported,false);
    const b=read(c.transcript),t=b.toString();assert.equal(hash(b),c.sha256);assert.equal(b.length,c.bytes);
    assert.doesNotMatch(t,/WARNING: DATA RACE|--- SKIP:|--- FAIL:|panic:/);
    for(const name of adjacent)assert.match(t,new RegExp('^--- PASS: Test'+name+' \\(','m'));
  }
  assert.deepEqual(run.commands.filter(c=>c.full).map(c=>[c.pair,c.arm]),[[0,'control'],[0,'candidate'],[1,'candidate'],[1,'control']]);
  const derivations=json('SERVICE_EVALUATOR_DERIVATION.json');
  for(const d of derivations.derivations)assert.equal(hash(read(d.destination)),d.destinationSHA256);
  if(fs.existsSync(path.join(root,'ARCHIVES.json'))) {
   const archives=json('ARCHIVES.json');assert.equal(archives.losslessVerified,true);assert.equal(archives.archives.length,4);
   for(const x of archives.archives) {
    assert.equal(hash(read(x.archive)),x.archiveSHA256);assert.equal(read(x.archive).length,x.archiveBytes);
    const c=run.commands.find(c=>c.transcript===x.original);assert.ok(c);assert.equal(x.originalSHA256,c.sha256);assert.equal(x.originalBytes,c.bytes);
   }
  }
  const actual=await evaluate(root);assert.deepEqual(actual,json('service-evaluation.json'));
  const math=await independent(root);assert.deepEqual(math,json('service-independent-results.json'));
  return {verified:true,commands:8,adjacentCases:64,ordinaryServiceRecalls:3072,exactReferenceRecalls:3072,currentSourceReadback:sources,annAbsolutePass:actual.annAbsolutePass,packetQualityPass:actual.packetQualityPass,repeatEqualityPass:actual.repeatEqualityPass,counterfactualEqualityPass:actual.counterfactualEqualityPass,wholeGoalValidation:false};
}
if(process.argv[1]&&fs.realpathSync(process.argv[1])===fs.realpathSync(fileURLToPath(import.meta.url)))console.log(JSON.stringify(await verify(path.dirname(fileURLToPath(import.meta.url)))));
