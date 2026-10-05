import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';

const data = JSON.parse(fs.readFileSync('research/public-task-pilot/zero-window-results.json'));
assert.equal(data.Arms.length, 32);
for (const [path, hash] of Object.entries(data.Hashes)) {
  assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path)).digest('hex'), hash, path);
}
const keys = new Set();
const q = (rows, key, p) => [...rows].map(r => r[key]).sort((a,b) => a-b)[Math.ceil(rows.length*p)-1]/1e6;
let reads = 0, writes = 0, errors = 0, queuedCanceled = 0;
for (const a of data.Arms) {
  const key = [a.Repeat,a.Mode,a.ReadIntervalMS,a.Lease,a.Enabled].join('/');
  assert(!keys.has(key)); keys.add(key);
  assert.equal(a.N,200); assert.equal(a.Workers,4);
  assert([0,1].includes(a.Repeat)); assert([5,20].includes(a.ReadIntervalMS));
  assert(['future','in-window'].includes(a.Mode));
  assert.equal(typeof a.Lease,'boolean'); assert.equal(typeof a.Enabled,'boolean');
  assert.equal(a.Samples.length,32); assert.equal(a.Writes.length,16);
  reads += a.Samples.length; writes += a.Writes.length;
  for (const r of [...a.Samples,...a.Writes]) {
    assert(r.NS > 0 && r.DispatchNS >= 0 && r.WaitNS >= r.DispatchNS && r.NS >= r.WaitNS);
    assert.equal(typeof r.Entered,'boolean');
    if (!r.Entered) { assert(r.Error); queuedCanceled++; }
    if (r.Error) errors++;
  }
  for (const r of a.Samples.filter(r=>!r.Error)) {
    assert(r.JournalMatched); assert.equal(r.FutureRecords,0); assert(r.Packed>0);
    assert(r.Nominated>=200 && r.Nominated<= (a.Mode==='future'?200:216));
    if(a.Enabled) assert(r.ExplanationBytes>0);
  }
  const re = a.Samples.filter(r=>r.Error).length;
  const we = a.Writes.filter(r=>r.Error).length;
  const late = [...a.Samples,...a.Writes].filter(r=>r.NS>100e6).length;
  assert.equal(a.Warmups.length,2);
  assert(a.DatabasePath && fs.existsSync(a.DatabasePath));
  const warmupErrors=a.Warmups.filter(r=>r.Error).length;
  const reopenErrors=a.ReopenErrors??[];
  assert(a.ReopenedJournals<=32-re && a.ReopenedWrites<=16-we);
  if(!reopenErrors.length) {
    assert.equal(a.ReopenedJournals,32-re);
    assert.equal(a.ReopenedWrites,16-we);
  }
  const pass = !re && !we && !late && !a.StaleRejections && !warmupErrors && !reopenErrors.length;
  console.log(JSON.stringify({key,pass,warmupErrors,reopenErrors,restoredJournals:a.ReopenedJournals,restoredWrites:a.ReopenedWrites,readErrors:re,writeErrors:we,late,stale:a.StaleRejections,
    readP95MS:q(a.Samples,'NS',.95),readMaxMS:q(a.Samples,'NS',1),
    writeMaxMS:q(a.Writes,'NS',1),writeWaitMaxMS:q(a.Writes,'WaitNS',1),
    dispatchMaxMS:q([...a.Samples,...a.Writes],'DispatchNS',1)}));
}
console.log(JSON.stringify({integrity:'PASS',reads,writes,errors,queuedCanceled}));



const base='research/public-task-pilot/libravdb-zero-window-source-v1/internal/storage/singlefile/engine.go';
const overlay='research/public-task-pilot/zero-window-local-overlay-v1/source-4.go.txt';
assert.equal(fs.readFileSync(overlay,'utf8'),fs.readFileSync(base,'utf8').replace('groupCommitWindow.Store(int64(1 * time.Millisecond))','groupCommitWindow.Store(0)'));
console.log('PASS: dependency overlay is exactly one coalescing initialization change');

