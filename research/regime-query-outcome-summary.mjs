import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createInterface} from 'node:readline';
import {auditRegimeOutcome} from './regime-query-outcome-audit.mjs';

// Attach both iterators before awaiting: hash listeners put streams into flow.
function input(file) { const stream = fs.createReadStream(file), hash = crypto.createHash('sha256'); stream.on('data', b => hash.update(b)); return {iterator: createInterface({input: stream, crlfDelay: Infinity})[Symbol.asyncIterator](), hash}; }
const raw = input(process.argv[2]), source = input(process.argv[3]);
const rawHeader = JSON.parse((await raw.iterator.next()).value), sourceHeader = JSON.parse((await source.iterator.next()).value);
assert.equal(rawHeader.Version, 'regime-query-outcome-v1'); assert.equal(sourceHeader.Version, 'soft-learners-v120');
for (const [key, value] of Object.entries({Records:2688,Workers:4,DecisionClock:160,RevealClock:161,Forecasts:31,BaseCap:63,PublicationCap:64,Hazard:.01,GenericPrior:.95})) assert.equal(rawHeader[key], value);
const sha = p => crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
for (const [file, expected] of Object.entries(rawHeader.Hashes)) { const absolute = path.resolve('internal/observationlearners', file); assert(absolute.startsWith(process.cwd() + path.sep)); assert.equal(sha(absolute), expected); }
const records = [], ids = new Set(); let checks = 0;
for (;;) {
  const a = await raw.iterator.next(), b = await source.iterator.next();
  if (a.done || b.done) { assert(a.done && b.done); break; }
  const r = JSON.parse(a.value), s = JSON.parse(b.value), id = `${r.Phase}:${r.Case}:${r.Index}:${r.Schedule}`;
  assert(!ids.has(id)); ids.add(id); records.push(auditRegimeOutcome(r, s)); checks += 4 * 31;
}
assert.equal(records.length, 2688); assert.equal(checks, 333312);
const inputSHA256 = source.hash.digest('hex'), artifactSHA256 = raw.hash.digest('hex');
assert.equal(inputSHA256, rawHeader.InputSHA256); assert.equal(inputSHA256, '5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f');
const mean = xs => xs.reduce((a, b) => a + b, 0) / xs.length;
const ci = xs => { assert.equal(xs.length, 32); const m = mean(xs), se = Math.sqrt(xs.reduce((s, x) => s + (x - m) ** 2, 0) / 31 / 32); return {mean:m, lower:m-3.5*se, upper:m+3.5*se}; };
const groups = [];
for (let phase=0;phase<2;phase++) for (let c=0;c<21;c++) for (let schedule=0;schedule<2;schedule++) {
  const rs = records.filter(r => r.phase===phase && r.case===c && r.schedule===schedule); assert.equal(rs.length,32); assert.equal(new Set(rs.map(r=>r.index)).size,32);
  groups.push({phase,case:c,schedule,brier:[0,1,2,3].map(a=>mean(rs.map(r=>r.brier[a]))),
    candidateGains:[0,1,2].map(a=>ci(rs.map(r=>r.brier[a]-r.brier[3]))),
    costs:[0,1,2,3].map(a=>rs.reduce((s,r)=>s+r.costs[a],0)),
    redundant:[0,1,2,3].map(a=>rs.reduce((s,r)=>s+r.redundant[a],0)),
    supportChanged:[0,1,2,3].map(a=>rs.reduce((s,r)=>s+r.supportChanged[a],0)),
    predictedGain:[0,1,2,3].map(a=>mean(rs.map(r=>r.predictedGain[a])))});
}
const totals = {queryBatches:records.filter(r=>r.poolSize>0).length, publicationFits:records.reduce((s,r)=>s+r.publicationFits,0),
  costs:[0,1,2,3].map(a=>records.reduce((s,r)=>s+r.costs[a],0)), redundant:[0,1,2,3].map(a=>records.reduce((s,r)=>s+r.redundant[a],0)),
  supportChanged:[0,1,2,3].map(a=>records.reduce((s,r)=>s+r.supportChanged[a],0))};
console.log(JSON.stringify({scope:'Consumed all-case one-decision diagnostic, not full-stream validation or a whole-goal pass/fail', inputSHA256,artifactSHA256,
  scriptSHA256:sha(import.meta.filename),auditSHA256:sha(new URL('./regime-query-outcome-audit.mjs',import.meta.url)),checks,totals,records,groups}));
