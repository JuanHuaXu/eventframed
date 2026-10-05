// Reuse frozen truth/cost scoring; cohort equivalence is measured, not assumed.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex'),outputs={};
function emit(out,source,s){fs.writeFileSync(out,s,{flag:'wx',mode:0o600});outputs[out]={source,sourceSHA256:sha(fs.readFileSync(source)),initialSHA256:sha(s)}}
let source='research/mean-joint-v74-readback.mjs',s=fs.readFileSync(source,'utf8').replaceAll('mean-joint-v74','mean-anchor-v75');
function rep(a,b){assert.equal(s.split(a).length-1,1,a);s=s.replace(a,b)}
rep("import assert from 'node:assert/strict';","import assert from 'node:assert/strict';\nimport {compare,selftest} from './mean-anchor-v75-comparison.mjs';");
rep("load('research/dynvarcache-v72-diagnostic/diagnostic.jsonl')","load('research/mean-joint-v74-diagnostic/diagnostic.jsonl')");
rep("['race-fixture','vet','allocation','benchmark','experiment']","['vet','allocation','experiment']");
rep('const totals={},cells=[];let controls=0;','const totals={},cells=[],equivalence={selftestChecks:selftest(),arms:0,passed:0,scalars:0,maxDefect:0,failed:[],selectionChangedArms:0,selectionChangedRounds:0,sharedPrefixIssuedComparisons:0,sharedPrefixMaxDefect:0};let controls=0;');
rep('old.Arms[schedule*18+9]','old.Arms[schedule*30+9]');rep('old.Arms[schedule*18].Schedule','old.Arms[schedule*30].Schedule');
rep('if(i<18){assert.deepEqual(strip(a),strip(old.Arms[schedule*18+i]));controls++}',`if(i<18){assert.deepEqual(strip(a),strip(old.Arms[schedule*30+i]));controls++}else{
   const prior=old.Arms[j],cmp=compare(strip(a),strip(prior));equivalence.arms++;equivalence.scalars+=cmp.scalars;equivalence.maxDefect=Math.max(equivalence.maxDefect,cmp.maxDefect);
   if(cmp.differences===0)equivalence.passed++;else equivalence.failed.push({geometry:world.Geometry,regime:world.Regime,schedule:a.Schedule,mode,comparison:cmp});
   const changes=[];assert.equal((a.Decisions??[]).length,(prior.Decisions??[]).length);for(let z=0;z<(a.Decisions??[]).length;z++)if(JSON.stringify(a.Decisions[z].Members)!==JSON.stringify(prior.Decisions[z].Members))changes.push({round:z,at:Math.min(a.Decisions[z].At,prior.Decisions[z].At)});
   if(changes.length){equivalence.selectionChangedArms++;equivalence.selectionChangedRounds+=changes.length}
   const before=changes.length?changes[0].at:2400;for(let k=0;k<Math.min(2400,before+1);k++)for(const field of ['Issued','ObservedIssued']){const d=Math.abs(a[field][k]-prior[field][k]);assert(d<=2e-10,'law discrepancy before first changed nomination');equivalence.sharedPrefixIssuedComparisons++;equivalence.sharedPrefixMaxDefect=Math.max(equivalence.sharedPrefixMaxDefect,d)}
  }`);
rep('fullNewArmIndependentReplay:false,','fullNewArmIndependentReplay:false,equivalence,allNewArmsEquivalent:equivalence.passed===1440,');
rep('freshDetailedFixtureArms:3,freshDetailedFixtureIssuedPackets:7200,freshDetailedFixtureCleanNoisyScalarComparisons:14400,','freshDetailedFixtureArms:0,freshDetailedFixtureIssuedPackets:0,freshDetailedFixtureCleanNoisyScalarComparisons:0,');
rep('console.log(JSON.stringify({allocation,totals,summaries:','console.log(JSON.stringify({equivalence:{...equivalence,failed:equivalence.failed.slice(0,3)},allocation,totals,summaries:');
emit('research/mean-anchor-v75-cohort-readback.mjs',source,s);
source='research/mean-joint-v74-cost-audit.mjs';s=fs.readFileSync(source,'utf8').replaceAll('mean-joint-v74','mean-anchor-v75').replace("source='research/mean-anchor-v75-cost-audit.mjs'","source='research/mean-anchor-v75-cohort-cost-audit.mjs'").replace('done.checks.length===6','done.checks.length===4');
emit('research/mean-anchor-v75-cohort-cost-audit.mjs',source,s);
fs.writeFileSync('research/mean-anchor-v75-cohort-tools-generation.json',JSON.stringify({outputs,scope:'full native-policy screen; declared recursive non-cost comparison, exact discrete fields and 2e-10 float tolerance; independent unchanged truth/cost recomputation, no fresh full dense replay'},null,2)+'\n',{flag:'wx',mode:0o600});
