import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import { performance } from 'node:perf_hooks';

const sigmoid = x => x >= 0 ? 1 / (1 + Math.exp(-x)) : Math.exp(x) / (1 + Math.exp(x));
const loss = (p, y) => (p - Number(y)) ** 2;
const key = r => [r.Phase, r.Case, r.Index, r.Schedule].join(':');

// Feedback refers to stored issue-time probabilities, never a newly fitted law.
function aggregate(rows) {
  const consumed = new Set(), forecasts = [];
  let odds = 0;
  for (let i = 0; i < rows.length; i++) {
    const arrivals = [];
    for (let j = 0; j < i; j++) {
      const r = rows[j];
      if (!r.missing && j + r.delay <= i && !consumed.has(j)) {
        odds += loss(r.incumbent, r.y) - loss(r.challenger, r.y);
        consumed.add(j);
        arrivals.push(j);
      }
    }
    const w = sigmoid(odds), r = rows[i];
    forecasts.push({ p: (1 - w) * r.incumbent + w * r.challenger, w, odds, arrivals, feedback: consumed.size });
  }
  return forecasts;
}

function selfTest() {
  const rows = Array.from({length: 8}, () => ({incumbent: .2, challenger: .8, y: true, delay: 0, missing: false}));
  const a = aggregate(rows);
  assert.equal(a[0].p, .5);
  assert(Math.abs(a[1].odds - .6) < 1e-14);
  assert(Math.abs(a[7].odds - 4.2) < 1e-14);
  assert.equal(a[7].feedback, 7);
  const poisoned = rows.map((r, i) => ({...r, y: i >= 3 ? false : r.y}));
  const b = aggregate(poisoned);
  assert.deepEqual(a.slice(0, 4), b.slice(0, 4));
  assert.notEqual(a[4].p, b[4].p);
  const quiet = rows.map(r => ({...r, missing: true, y: NaN}));
  assert(aggregate(quiet).every(r => r.p === .5 && r.feedback === 0));
  const delayed = rows.map(r => ({...r, delay: 3}));
  const d = aggregate(delayed);
  assert(d.slice(0, 3).every(r => r.p === .5));
  assert.deepEqual(d[3].arrivals, [0]);
  assert.equal(d[7].feedback, 5);
  delayed[7].y = false;
  assert.deepEqual(d, aggregate(delayed));
  console.log('feedback self-tests PASS');
}

async function* readJSONL(path) {
  const input = fs.createReadStream(path);
  const lines = readline.createInterface({ input, crlfDelay: Infinity });
  try { for await (const line of lines) if (line.trim()) yield JSON.parse(line); }
  finally { lines.close(); input.destroy(); }
}

selfTest();
const [source, cadence, output, mode, clockArg] = process.argv.slice(2);
const startClock=clockArg===undefined?128:Number(clockArg);assert([0,128,224].includes(startClock));
assert(mode===undefined||mode==='eight');
const indices=mode==='eight'?8:1, expectedCount=84*indices;
if (source) {
  assert(cadence && output);
  const start = performance.now(), candidates = new Map();
  for await (const r of readJSONL(cadence)) {
    assert.equal(r.Clock??128,startClock);
    assert(r.Index>=0&&r.Index<indices);
    assert(!candidates.has(key(r)));
    candidates.set(key(r), r);
  }
  assert.equal(candidates.size, expectedCount);
  const results = [];
  let header = true, maxWeightError = 0, updates = 0;
  for await (const src of readJSONL(source)) {
    if (header) { assert.equal(src.Version, 'soft-learners-v120'); header = false; continue; }
    const k = key(src), candidate = candidates.get(k);
    if (!candidate) continue;
    candidates.delete(k);
    const rows = candidate.P.map((p, i) => {
      const s = src.Steps[startClock+i];
      assert.equal(candidate.Y[i], s.Y);
      assert.equal(candidate.Q[i], s.Q);
      assert.deepEqual(candidate.Control[i], s.P);
      assert(Number.isFinite(p) && p >= 0 && p <= 1);
      assert(Number.isFinite(s.P[12]) && s.P[12] >= 0 && s.P[12] <= 1);
      assert(Number.isInteger(s.Delay) && s.Delay >= 0);
      return { incumbent: s.P[12], challenger: p, y: s.Y, delay: s.Delay, missing: s.Missing };
    });
    assert.equal(rows.length, 32);
    const forecasts = aggregate(rows), expected = [0,0,0,0], realized = [0,0,0,0];
    for (let i = 0; i < rows.length; i++) {
      // Independent batch recomputation has no incremental state or consumed set.
      let total = 0, count = 0;
      for (let j = 0; j < i; j++) {
        const r = rows[j];
        if (!r.missing && j + r.delay <= i) {
          total += (r.incumbent-r.challenger) * (r.incumbent+r.challenger-2*Number(r.y));
          count++;
        }
      }
      maxWeightError = Math.max(maxWeightError, Math.abs(sigmoid(total)-forecasts[i].w));
      assert(Math.abs(sigmoid(total)-forecasts[i].w) < 1e-12);
      assert.equal(forecasts[i].feedback, count);
      updates += forecasts[i].arrivals.length;
      const r = rows[i], q = candidate.Q[i];
      assert(q >= 0 && q <= 1);
      const probabilities = [forecasts[i].p, r.incumbent, r.challenger, (r.incumbent+r.challenger)/2];
      for (let arm = 0; arm < 4; arm++) {
        expected[arm] += ((probabilities[arm]-q)**2+q*(1-q))/32;
        realized[arm] += loss(probabilities[arm], r.y)/32;
      }
    }
    results.push({key:k, expected, realized, forecasts});
  }
  assert.equal(candidates.size, 0);
  assert.equal(results.length, expectedCount);
  const mean = [0,0,0,0], cells = [];
  for (const r of results) for (let j=0;j<4;j++) mean[j] += r.expected[j]/expectedCount;
  for (let phase=0;phase<2;phase++) for (let schedule=0;schedule<2;schedule++) {
    const selected = results.filter(r=>r.key.startsWith(`${phase}:`) && r.key.endsWith(`:${schedule}`));
    assert.equal(selected.length,21*indices);
    const expected=[0,0,0,0], realized=[0,0,0,0];
    for (const r of selected) for(let j=0;j<4;j++){expected[j]+=r.expected[j]/selected.length;realized[j]+=r.realized[j]/selected.length;}
    cells.push({phase,schedule,expected,realized});
  }
  const harm = results.filter(r=>r.expected[0]-r.expected[1]>.01);
  const summary = { startClock, arms:['feedbackMixture','Markov','arrivalRefit','fixedHalf'], records:expectedCount, forecasts:expectedCount*32,
    updates, maxWeightError, expected:mean, improved:results.filter(r=>r.expected[0]<r.expected[1]).length,
    harmGreaterThanPoint01:harm.length, harmfulKeys:harm.map(r=>r.key), cells,
    conclusion:harm.length?'REJECT unconditional incumbent protection on this pilot':'Broader test needed; not validated',
    limitations:'Consumed-data short-block replay. No confidence guarantee; no equal-compute claim. All challenger fit cost retained.' };
  const processingMilliseconds = performance.now()-start;
  fs.writeFileSync(output, JSON.stringify({summary,processingMilliseconds,results},null,2)+'\n', {flag:'wx'});
  console.log(JSON.stringify({...summary,processingMilliseconds},null,2));
}
