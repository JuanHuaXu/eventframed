// Mechanical reuse of frozen gates and collection; explicit evidence limits.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),outputs={};
function emit(out,source,s){fs.writeFileSync(out,s,{flag:'wx',mode:0o600});outputs[out]={source,sourceSHA256:hash(fs.readFileSync(source)),initialSHA256:hash(s)}}
let source='research/dynvariance-v71-readback.mjs',s=fs.readFileSync(source,'utf8').replaceAll('dynvariance-v71','mean-joint-v74');
function replace(a,b){assert.equal(s.split(a).length-1,1,a);s=s.replace(a,b)}
replace("load('research/class-v69-diagnostic/diagnostic.jsonl')","load('research/dynvarcache-v72-diagnostic/diagnostic.jsonl')");
const modes=JSON.parse(fs.readFileSync('research/mean-joint-v74-study-generation.json')).modes;
const a=s.indexOf('const modes='),b=s.indexOf('\n',a);assert(a>0);s=s.slice(0,a)+'const modes='+JSON.stringify(modes)+';'+s.slice(b);
replace('w.Arms.length,54','w.Arms.length,90');replace('j<54','j<90');
replace('i=j%18,schedule=Math.floor(j/18),mode=modes[i],full=w.Arms[schedule*18],adaptive=w.Arms[schedule*18+1],single=w.Arms[schedule*18+10],oldLocal=old.Arms[schedule*10+6]', 'i=j%30,schedule=Math.floor(j/30),mode=modes[i],full=w.Arms[schedule*30],adaptive=w.Arms[schedule*30+1],single=w.Arms[schedule*30+(i<18?10:18+4*Math.floor((i-18)/4))],oldLocal=old.Arms[schedule*18+9]');
replace('old.Arms[schedule*10].Schedule','old.Arms[schedule*18].Schedule');
replace('if(i<2){assert.deepEqual(strip(a),strip(old.Arms[schedule*10+i]));controls++}', 'if(i<18){assert.deepEqual(strip(a),strip(old.Arms[schedule*18+i]));controls++}');
replace('controls,240','controls,2160');s=s.replaceAll('v69LatentGain','v72BestComponentGain');
replace("['learn_model_class','learn_noise_class']","['mean_falsification','meanlocal_falsification','meanindividual_falsification']");
replace("['learn_random','learn_uncertainty','learn_falsification']","['mean_random','mean_uncertainty','meanlocal_random','meanlocal_uncertainty','meanindividual_random','meanindividual_uncertainty']");
replace('2160-arm consumed-cohort dispersion diagnostic','3600-arm consumed-cohort joint-mean diagnostic');
replace('arms:2160,modelArms:1920','arms:3600,modelArms:3360,newModelArms:1440');
replace('independentIssuedPackets:4608000,scalarIssuedComparisons:9216000','fullNewArmIndependentReplay:false,freshDetailedFixtureArms:3,freshDetailedFixtureIssuedPackets:7200,freshDetailedFixtureCleanNoisyScalarComparisons:14400');
replace("for(const n of ['race-model','race-fixture','vet','allocation','benchmark','experiment','audit'])", "for(const n of ['race-fixture','vet','allocation','benchmark','experiment'])");
replace('console.log(JSON.stringify({allocation,totals,summaries:', 'console.log(JSON.stringify({allocation,totals,summaries:');
emit('research/mean-joint-v74-readback.mjs',source,s);
source='research/dynvarcache-v72-run.mjs';s=fs.readFileSync(source,'utf8').replaceAll('dynvarcache-v72','mean-joint-v74').replaceAll('researchdynvarcachecheck','researchmeanjointcheck').replaceAll('researchdynvarcache\'','researchmeanjoint\'').replaceAll('researchdynvarianceref','researchmeanjointref').replaceAll('EVENTFRAME_DYNVARCACHE_V72','EVENTFRAME_MEANJOINT_V74').replaceAll('TestDynVarCacheV72','TestMeanJointV74');
replace('research/checkpoint-2026-10-04-dynvariance-v71/manifest.json','research/checkpoint-2026-10-04-dynvarcache-v72-mean-v73/manifest.json');
replace('9d6cb6dcb7fc8281b407ebc9cb5035c27a2e8b29bcf9230013517e2da9a2b297','d4045ac945e8044155f54fb0871d8b2fffb30fa8a4806d431beaf028ea22cc54');
const c=s.indexOf('const extras='),d=s.indexOf('const files=',c);assert(c>0&&d>c);
s=s.slice(0,c)+`const extras=['go.mod','go.sum','research/mean-joint-v74-run.mjs','research/mean-joint-v74-run-gen.mjs','research/mean-joint-v74-run-generation.json','research/mean-joint-v74-readback.mjs','research/mean-joint-v74-study-gen.mjs','research/mean-joint-v74-study-generation.json','research/mean-joint-v74-preflight.mjs','research/mean-joint-v74-long-preflight.mjs','docs/experiments/mmm-mean-joint-v74-protocol.md'];\n`+s.slice(d);
replace("['research/dynvariance-v71-diagnostic/diagnostic.jsonl','research/dynvariance-v71-diagnostic/freeze.json','research/dynvariance-v71-diagnostic/readback.json','research/dynvariance-v71-diagnostic/completed.json']", "['research/dynvarcache-v72-diagnostic/diagnostic.jsonl','research/dynvarcache-v72-diagnostic/freeze.json','research/dynvarcache-v72-diagnostic/readback.json','research/dynvarcache-v72-diagnostic/completed.json','research/mean-joint-v74-initial/completed.json','research/mean-joint-v74-long/completed.json']");
replace('full eighteen-arm consumed-cohort equivalent-cache diagnostic; independent V71 full replay plus exact all-arm equivalence, not a fresh full replay','full thirty-arm consumed-cohort joint-mean screen; old18 arms bitwise compared, three full independent new fixtures; NOTfull independent replay of all1440 new arms');
const start=s.indexOf(" await run('race-model'"),end=s.indexOf(" await run('race-fixture'",start);assert(start>0&&end>start);s=s.slice(0,start)+s.slice(end);
replace(" await run('audit','node',['research/mean-joint-v74-transitive-audit.mjs']);\n",'');
emit('research/mean-joint-v74-run.mjs',source,s);
fs.writeFileSync('research/mean-joint-v74-run-generation.json',JSON.stringify({outputs,changes:'new full3600-arm collection, all18 immutable controls and12 jointmean primary-policy arms; original40worlds/20regimes/twogeometries/threedelays/scientific gates; fresh detailed fixture replay, explicit no full1440newarm independent replay'},null,2)+'\n',{flag:'wx',mode:0o600});
