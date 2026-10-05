// New isolated harness and frozen study; never edit the V71 archive.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),outputs={};
function emit(out,source,s){assert(!fs.existsSync(out));fs.writeFileSync(out,s,{flag:'wx',mode:0o600});outputs[out]={source,sourceSHA256:hash(fs.readFileSync(source)),initialSHA256:hash(s)}}
let source='internal/researchdispersion/dynvariance_v71_test.go';
let s=fs.readFileSync(source,'utf8').replaceAll('DynVarianceV71','DynVarCacheV72').replaceAll('dynVarianceV71','dynVarCacheV72').replaceAll('researchdynvariance"','researchdynvarcache"').replaceAll('EVENTFRAME_DYNVARIANCE_V71','EVENTFRAME_DYNVARCACHE_V72');
emit('internal/researchdispersion/dynvarcache_v72_test.go',source,s);
function replace(from,to){assert.equal(s.split(from).length-1,1,from);s=s.replace(from,to)}
source='research/dynvariance-v71-readback.mjs';s=fs.readFileSync(source,'utf8').replaceAll('dynvariance-v71','dynvarcache-v72');
replace("load('research/class-v69-diagnostic/diagnostic.jsonl')","load('research/dynvariance-v71-diagnostic/diagnostic.jsonl')");
replace('oldLocal=old.Arms[schedule*10+6]','oldLocal=old.Arms[schedule*18+14]');
replace('old.Arms[schedule*10].Schedule','old.Arms[schedule*18].Schedule');
replace('old.Arms[schedule*10+i]','old.Arms[schedule*18+i]');
replace('const summaries={};',`const audit=JSON.parse(fs.readFileSync(root+'/transitive-audit.json'));assert.equal(audit.allArmsBitwiseEqual,2160);assert.equal(audit.modelArmsBitwiseEqual,1920);assert(audit.priorIndependentReplayVerified);\nconst summaries={};`);
s=s.replaceAll('v69LatentGain','v71LatentGain');
replace('independentIssuedPackets:4608000,scalarIssuedComparisons:9216000','directIndependentFullReplayThisStudy:false,transitiveIndependentIssuedPackets:4608000,transitiveScalarIssuedComparisons:9216000,allArmsBitwiseEqual:audit.allArmsBitwiseEqual,modelArmsBitwiseEqual:audit.modelArmsBitwiseEqual');
replace('2160-arm consumed-cohort dispersion diagnostic','2160-arm consumed-cohort equivalent-cache diagnostic');
emit('research/dynvarcache-v72-readback.mjs',source,s);
source='research/dynvariance-v71-run.mjs';s=fs.readFileSync(source,'utf8').replaceAll('dynvariance-v71','dynvarcache-v72').replaceAll('researchdynvariance\'','researchdynvarcache\'').replaceAll('EVENTFRAME_DYNVARIANCE_V71','EVENTFRAME_DYNVARCACHE_V72').replaceAll('TestDynVarianceV71','TestDynVarCacheV72');
replace('research/checkpoint-2026-10-04-class-v69-rate-v70/manifest.json','research/checkpoint-2026-10-04-dynvariance-v71/manifest.json');
replace('da8c45a0551b895cc578b919bae78d1776b662f6dccc5d946f4ff9e17c1480a3','9d6cb6dcb7fc8281b407ebc9cb5035c27a2e8b29bcf9230013517e2da9a2b297');
replace("'./internal/researchdynvarianceref'];","'./internal/researchdynvarianceref','./internal/researchdynvarcachecheck'];");
const a=s.indexOf('const extras='),b=s.indexOf('const files=',a);assert(a>0&&b>a);
s=s.slice(0,a)+`const extras=['go.mod','go.sum','research/dynvarcache-v72-run.mjs','research/dynvarcache-v72-study-gen.mjs','research/dynvarcache-v72-study-generation.json','research/dynvarcache-v72-readback.mjs','research/dynvarcache-v72-transitive-audit.mjs','research/dynvarcache-v72-preflight.mjs','docs/experiments/mmm-dynvarcache-v72-protocol.md'];\n`+s.slice(b);
replace("['research/class-v69-diagnostic/diagnostic.jsonl','research/class-v69-diagnostic/freeze.json','research/class-v69-diagnostic/readback.json']","['research/dynvariance-v71-diagnostic/diagnostic.jsonl','research/dynvariance-v71-diagnostic/freeze.json','research/dynvariance-v71-diagnostic/readback.json','research/dynvariance-v71-diagnostic/completed.json']");
replace('full eighteen-arm consumed-cohort dynamic-dispersion diagnostic','full eighteen-arm consumed-cohort equivalent-cache diagnostic; independent V71 full replay plus exact all-arm equivalence, not a fresh full replay');
replace("'./internal/researchdynvarcache','./internal/researchdynvarianceref','-count=1'","'./internal/researchdynvarcache','./internal/researchdynvarcachecheck','./internal/researchdynvarianceref','-count=1'");
replace("await run('audit','go',['test','./internal/researchdispersion','-run','^TestDynVarCacheV72Audit$','-count=1','-v','-timeout=80m'],{EVENTFRAME_DYNVARCACHE_V72_AUDIT:path.resolve(root+'/diagnostic.jsonl')});","await run('audit','node',['research/dynvarcache-v72-transitive-audit.mjs']);");
emit('research/dynvarcache-v72-run.mjs',source,s);
fs.writeFileSync('research/dynvarcache-v72-study-generation.json',JSON.stringify({outputs,changes:'isolated cached-model import and test namespace; full identical eighteen arms, forty consumed worlds, three schedules, scoring and scientific gates; explicit transitive audit of all arms against independently replayed V71'},null,2)+'\n',{flag:'wx',mode:0o600});
