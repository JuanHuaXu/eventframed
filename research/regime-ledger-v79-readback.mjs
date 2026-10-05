// Recompute cost tables from completed raw benchmark output, without reruns.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const run='research/regime-ledger-v79-bounded-audit';
const json=p=>JSON.parse(fs.readFileSync(p));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const done=json(run+'/completed.json'),audit=json(run+'/audit.json');
assert(done.allJobsTerminal&&done.allChecksPass&&done.sourceUnchanged&&done.checks.length===4);
for(const x of done.checks) {assert.equal(x.exitCode,0);assert.equal(hash(fs.readFileSync(run+'/'+x.name+'.log')),x.logSHA256)}
assert.equal(audit.Configurations,108);assert.equal(audit.Publications,2474);assert.equal(audit.Checks,132674);assert.equal(audit.Queries,2341);assert.equal(audit.Abstentions,1);
assert(audit.Full64&&audit.RepruningChanged&&!audit.WholeGoalValidation&&audit.MaxOracle<2e-11&&audit.MaxTower<2e-11&&audit.MaxReceipt<2e-11);
const log=fs.readFileSync(run+'/benchmark.log','utf8'),groups={};
for(const line of log.split('\n')) {
  const m=line.match(/^(Benchmark(?:LedgerOperations|Constructor|FullFrontier)\/.+)-\d+\s+(\d+)\s+(\d+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op$/);
  if(!m)continue;
  (groups[m[1]]??=[]).push({iterations:Number(m[2]),nsPerOp:Number(m[3]),totalAllocatedBytesPerOp:Number(m[4]),allocationsPerOp:Number(m[5])});
}
const expected=[];
for(const m of [2,150,200])for(const t of [32,64])for(const op of [t===64?'PendingSecondAbstention':'PendingSecond','PendingFirst','Reveal','Issue'])expected.push(`BenchmarkLedgerOperations/M${m}/T${t}/${op}`);
for(const m of [150,200]) {
  expected.push(`BenchmarkConstructor/M${m}`);
  for(const op of ['PendingFirst','RevealFirst','Issue'])expected.push(`BenchmarkFullFrontier/M${m}/T${m*16}/${op}`);
}
assert.deepEqual(Object.keys(groups).sort(),expected.sort());
for(const rows of Object.values(groups)) {assert.equal(rows.length,2);for(const r of rows)assert(r.iterations>=1&&r.nsPerOp>0&&r.totalAllocatedBytesPerOp>0&&r.allocationsPerOp>0)}
const out={time:new Date().toISOString(),sourceSHA256:hash(fs.readFileSync('research/regime-ledger-v79-readback.mjs')),benchmarkSHA256:hash(Buffer.from(log)),auditSHA256:hash(fs.readFileSync(run+'/audit.json')),
  groups,count:Object.values(groups).reduce((a,b)=>a+b.length,0),audit,
  measuredScope:'serial cold public operation, prepared as-of state; full frontier setup uses one excluded cold unknown-history replay; not whole ingestion, loaded serving, peak memory or fresh empirical quality',
  retainedSupportUpperBound:{keyBytes:16,maxRows:12800,maxCap:256,payloadBytes:16*12800*256,includesSliceHeaders:false,includesScratch:false,observedPeakRSS:false},
  equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,wholeMemoryLimitCertified:false,scientificQualityRescueEstablished:false,goals:Array(7).fill('OPEN')};
fs.writeFileSync(run+'/readback.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({groups:Object.keys(groups).length,rows:out.count,audit,fullFrontier:Object.fromEntries(Object.entries(groups).filter(([n])=>n.startsWith('BenchmarkFullFrontier')))},null,2));
