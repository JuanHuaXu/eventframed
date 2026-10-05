import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const stage=process.argv[2];assert(/^[a-z0-9-]+$/.test(stage));const run=`research/regime-log-v82-${stage}`,json=p=>JSON.parse(fs.readFileSync(p)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const done=json(run+'/completed.json'),audit=json(run+'/audit.json');assert(done.allJobsTerminal&&done.allChecksPass&&done.sourceUnchanged&&done.checks.length===4);
for(const c of done.checks){assert.equal(c.exitCode,0);assert.equal(hash(fs.readFileSync(run+'/'+c.name+'.log')),c.logSHA256)}
assert(audit.Full64&&audit.FullWorkload&&audit.OldReplay&&audit.RepruningChanged&&!audit.WholeGoalValidation);
assert(audit.MaxOracle<2e-11&&audit.MaxReference<2e-11&&audit.MaxTower<2e-11&&audit.MaxReceipt<2e-11);
assert.equal(audit.Configurations,108);assert.equal(audit.IncrementalChecks,10364);assert.equal(audit.CacheChecks,214);assert.equal(audit.ProtectedChecks,2592);
assert(audit.LongLogRescue&&audit.MemberOddsRescue&&!audit.NoResetUnderflow);assert(audit.MaxLogOracle<5e-8);
assert.equal(audit.OldEnvelopeComparisons,286);assert.equal(audit.MaxOldEnvelopeDifference,1);assert.equal(audit.EnvelopeOracleChecks,280);assert.equal(audit.NonvacuousEnvelopes,225);
const groups={},log=fs.readFileSync(run+'/benchmark.log','utf8');let pkg,rows=0;
for(const line of log.split('\n')){
 if(line.startsWith('pkg: ')){pkg=line.slice(5).split('/').at(-1);groups[pkg]={};continue}
 const m=line.match(/^(Benchmark(?:LedgerOperations|Constructor|FullFrontier(?:OldFirst)?)\/.+)-\d+\s+(\d+)\s+(\d+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/);if(!m)continue;
 assert(pkg==='researchregimeprotected'||pkg==='researchregimelog');(groups[pkg][m[1]]??=[]).push({iterations:Number(m[2]),nsPerOp:Number(m[3]),totalAllocatedBytesPerOp:Number(m[4]),allocationsPerOp:Number(m[5])});rows++;
}
assert.equal(rows,136);for(const g of Object.values(groups)){assert.equal(Object.keys(g).length,34);for(const rs of Object.values(g))assert.equal(rs.length,2)}
const full={},mean=rs=>rs.reduce((a,b)=>a+b.nsPerOp,0)/rs.length;
for(const n of Object.keys(groups.researchregimelog).filter(n=>n.startsWith('BenchmarkFullFrontier'))){const old=groups.researchregimeprotected[n],candidate=groups.researchregimelog[n];assert(old);full[n]={old,candidate,meanTimeRatio:mean(candidate)/mean(old)}}
const out={time:new Date().toISOString(),stage,sourceSHA256:hash(fs.readFileSync('research/regime-log-v82-readback.mjs')),benchmarkSHA256:hash(Buffer.from(log)),auditSHA256:hash(fs.readFileSync(run+'/audit.json')),audit,groups,fullFrontier:full,benchmarkRows:rows,
 measurementScope:'serial same command V81 then V82; matched dimensions, not bitwise law equivalence, randomized timings, growing whole-core or loaded serving',
 analyticEnvelopeScope:'independent unrestricted dense tests permit2e-11 numerical tolerance; not certified interval arithmetic or AP target-law authority; old bound numeric parity FAIL preserved',
 supportPayloadBytes:{M150T2400:1381968,M200T3200:1842768},cachedComponentPayloadBytes:233856,
 numericalRescueEstablishedInFixtures:true,scientificQualityRescueEstablished:false,scalableApproximationCertified:false,wholeMemoryLimitCertified:false,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goal:'ACTIVE',goals:Array(7).fill('OPEN')};
fs.writeFileSync(run+'/readback.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({stage,audit,fullFrontier:full},null,2));
