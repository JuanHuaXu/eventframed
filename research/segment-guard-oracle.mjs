import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {pointwiseGuard, testPointwiseGuard} from './pointwise-brier-guard.mjs';

// Evaluation only: Q is generator truth, never a deployable input. This gives
// an upper bound on achievable gain for these fixed heads and this exact guard.
const [source, output] = process.argv.slice(2);
assert(source && output);
testPointwiseGuard();
const groups = new Map();
let header = true, count = 0, optimalityChecks = 0;
for await (const line of readline.createInterface({input: fs.createReadStream(source), crlfDelay: Infinity})) {
  const r = JSON.parse(line);
  if (header) { assert.equal(r.Cohort, 'spike-independent-v1'); header = false; continue; }
  assert.equal(r.Steps.length, 256);
  const key = [r.Phase, r.Case, r.Schedule].join(':');
  if (!groups.has(key)) groups.set(key, {key, case: r.Case, indices: [], gain: [0, 0, 0, 0], terminalGain: [0, 0, 0, 0]});
  const g = groups.get(key);
  assert(!g.indices.includes(r.Index)); g.indices.push(r.Index);
  for (let t = 0; t < 256; t++) {
    const s = r.Steps[t], b = s.P[12], q = s.Q;
    assert(q >= 0 && q <= 1);
    const gains = [10, 13].map(head => {
      const c = s.P[head], d = c - b, cap = pointwiseGuard(b, c, 1).w;
      const w = d === 0 ? 0 : Math.max(0, Math.min(cap, (q - b) / d));
      const p = b + w * d;
      // A dense grid independently checks the analytic constrained minimizer.
      for (let j = 0; j <= 20; j++) {
        assert((p - q) ** 2 <= (b + d * cap * j / 20 - q) ** 2 + 1e-12);
        optimalityChecks++;
      }
      const gain = (b - q) ** 2 - (p - q) ** 2;
      assert(gain >= -1e-12);
      return gain;
    });
    // Also allow the oracle to choose either head per frame. This is a larger
    // feasible class than either implemented two-expert combiner, still bounded.
    gains.push(Math.max(...gains));
    // Solve both binary-outcome constraints over all p in [0,1], not just
    // a segment between available heads. The nearest feasible p to Q is exact.
    const lo = Math.max(0, 1 - Math.sqrt((1-b)**2 + .01));
    const hi = Math.min(1, Math.sqrt(b*b + .01));
    const universal = Math.max(lo, Math.min(hi, q));
    assert(Math.max(universal**2 - b*b, (1-universal)**2 - (1-b)**2) <= .01 + 1e-12);
    for (let j = 0; j <= 20; j++) {
      assert((universal-q)**2 <= (lo+(hi-lo)*j/20-q)**2 + 1e-12);
      optimalityChecks++;
    }
    const universalGain = (b-q)**2 - (universal-q)**2;
    assert(universalGain >= gains[2] - 1e-12);
    gains.push(universalGain);
    gains.forEach((gain, a) => { g.gain[a] += gain / 256; if (t >= 192) g.terminalGain[a] += gain / 64; });
  }
  count++;
}
assert.equal(count, 672); assert.equal(groups.size, 84);
const cells = [...groups.values()].map(g => {
  assert.deepEqual(g.indices.sort((a,b) => a-b), [0,1,2,3,4,5,6,7]);
  return {key: g.key, changing: g.case < 9 ? g.case % 3 !== 0 : g.case >= 19,
    gain: g.gain.map(v => v/8), terminalGain: g.terminalGain.map(v => v/8)};
});
const changing = cells.filter(g => g.changing);
assert.equal(changing.length, 32);
const summary = {
  arms: ['segmentOracle', 'staticOracle', 'eitherHeadOracle', 'anyForecastOracle'], count, optimalityChecks,
  recoveryCells: changing.length,
  recoveryCellsBelowRequiredMeanGain: [0,1,2,3].map(a => changing.filter(g => g.terminalGain[a] < .005 - 1e-12).length),
  wholeMeanGain: [0,1,2,3].map(a => cells.reduce((s,g) => s + g.gain[a], 0)/cells.length),
  cells,
  limitations: 'Consumed-data evaluator-only upper bound on expected gain versus the fixed Markov baseline, using inaccessible Q. The anyForecast arm bounds every binary prediction satisfying the .01 worst-outcome pointwise constraint on these records. Not a learner, confidence interval, fresh confirmation or impossibility result for other guards/baselines. Terminal means use the same eight indices; this is not the original full 32-index protocol.'
};
fs.writeFileSync(output, JSON.stringify(summary, null, 2)+'\n', {flag:'wx'});
console.log(JSON.stringify({...summary, cells: undefined}, null, 2));
