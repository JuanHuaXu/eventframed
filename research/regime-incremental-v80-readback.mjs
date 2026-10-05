// Parse same-command reference/candidate costs and retain measurement limits.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const run='research/regime-incremental-v80-initial',json=p=>JSON.parse(fs.readFileSync(p)),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const done=json(run+'/completed.json'),audit=json(run+'/audit.json');
assert(done.allJobsTerminal&&done.allChecksPass&&done.sourceUnchanged&&done.checks.length===4);
for(const c of done.checks){assert.equal(c.exitCode,0);assert.equal(hash(fs.readFileSync(run+'/'+c.name+'.log')),c.logSHA256)}
assert.equal(audit.Checks,143038);assert.equal(audit.IncrementalChecks,10364);assert.equal(audit.CacheChecks,214);assert.equal(audit.MaxReference,0);
assert(audit.Full64&&audit.FullWorkload&&audit.OldReplay&&audit.RepruningChanged&&!audit.WholeGoalValidation);
assert(audit.MaxOracle<2e-11&&audit.MaxTower<2e-11&&audit.MaxReceipt<2e-11);
const groups={},log=fs.readFileSync(run+'/benchmark.log','utf8');let pkg;
for(const line of log.split('\n')) {
  if(line.startsWith('pkg: ')){pkg=line.slice(5).split('/').at(-1);groups[pkg]={};continue}
  const m=line.match(/^(Benchmark(?:LedgerOperations|Constructor|FullFrontier(?:OldFirst)?)\/.+)-\d+\s+(\d+)\s+(\d+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/);
  if(!m)continue;
  assert(pkg==='researchregimeledger'||pkg==='researchregimeincremental');
  (groups[pkg][m[1]]??=[]).push({iterations:Number(m[2]),nsPerOp:Number(m[3]),totalAllocatedBytesPerOp:Number(m[4]),allocationsPerOp:Number(m[5])});
}
assert.equal(Object.keys(groups.researchregimeledger).length,32);assert.equal(Object.keys(groups.researchregimeincremental).length,34);
for(const g of Object.values(groups))for(const rows of Object.values(g)){assert.equal(rows.length,2);for(const r of rows)assert(r.iterations>=1&&r.nsPerOp>0&&r.totalAllocatedBytesPerOp>0)}
assert.deepEqual(Object.keys(groups.researchregimeincremental).filter(n=>!n.startsWith('BenchmarkFullFrontierOldFirst')).sort(),Object.keys(groups.researchregimeledger).sort());
const mean=rows=>rows.reduce((a,b)=>a+b.nsPerOp,0)/rows.length;
const matched={};
for(const n of Object.keys(groups.researchregimeledger)) {
  const control=groups.researchregimeledger[n],candidate=groups.researchregimeincremental[n];
  matched[n]={control,candidate,meanTimeRatio:mean(candidate)/mean(control),timingConfidenceInterval:null};
}
const out={time:new Date().toISOString(),sourceSHA256:hash(fs.readFileSync('research/regime-incremental-v80-readback.mjs')),benchmarkSHA256:hash(Buffer.from(log)),auditSHA256:hash(fs.readFileSync(run+'/audit.json')),
  audit,groups,matched,benchmarkRows:132,measurementScope:'same serial command, reference then candidate; prepared equivalent laws; not randomized paired trials or loaded serving; fixtures excluded from timing; old-query candidate has no separately measured matching old reference',
  supportPayloadBytes:{M150T2400:1381968,M200T3200:1842768},cachedComponentPayloadBytes:233856,
  resourceScope:'type-size payload only; not live heap, peak/RSS, serialized fingerprint or whole allowed-maximum memory bound',
  scientificQualityRescueEstablished:false,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,wholeMemoryLimitCertified:false,goals:Array(7).fill('OPEN')};
fs.writeFileSync(run+'/readback.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({audit,rows:132,fullFrontier:Object.fromEntries(Object.entries(matched).filter(([n])=>n.startsWith('BenchmarkFullFrontier'))),oldQueries:Object.fromEntries(Object.entries(groups.researchregimeincremental).filter(([n])=>n.startsWith('BenchmarkFullFrontierOldFirst')))},null,2));
