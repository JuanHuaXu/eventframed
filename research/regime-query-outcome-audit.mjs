import assert from 'node:assert/strict';
import crypto from 'node:crypto';

export const maximumOrigin = (pool, scores) => {
  assert.equal(pool.length, scores.length); assert(pool.length > 0);
  assert(scores.every(Number.isFinite));
  const best = Math.max(...scores);
  return Math.min(...pool.filter((_, i) => best - scores[i] <= 1e-10));
};
export const randomOrigin = (source, pool) => pool.map(origin => ({origin,
  hash: crypto.createHash('sha256').update(`regime-query-outcome-v1:${source.Phase}:${source.Case}:${source.Index}:160:${origin}`).digest('hex')
})).sort((a, b) => (a.hash < b.hash ? -1 : a.hash > b.hash ? 1 : 0) || a.origin - b.origin)[0].origin;

export function auditRegimeOutcome(r, source, probeStart = 153) {
  assert([137, 153].includes(probeStart), 'undeclared probe contract');
  assert(!r.Error);
  for (const key of ['Phase', 'Case', 'Index', 'Schedule']) assert.equal(r[key], source[key]);
  const d = r.Decision; assert.equal(d.Clock, 160); assert.equal(source.Steps.length, 256);
  const close = (x, y, tolerance = 1e-10) => { assert(Number.isFinite(x) && Number.isFinite(y)); assert(Math.abs(x - y) <= tolerance, `${x} != ${y}`); };
  const available = t => { const a = Array.from({length: 16}, (_, i) => i - 16); for (let j = 0; j < t; j++) if (!source.Steps[j].Missing && j + source.Steps[j].Delay <= t) a.push(j); return a; };
  assert.deepEqual(d.Origins, available(160).slice(-63));
  const pool = Array.from({length: 8}, (_, i) => 152 + i).filter(j => source.Steps[j].Missing || j + source.Steps[j].Delay > 160);
  assert.deepEqual(d.Pool ?? [], pool);
  assert.deepEqual(d.Probes, source.Steps.slice(probeStart, probeStart + 8).map(s => s.X));
  const values = d.Values ?? []; assert.equal(values.length, pool.length);
  const entropy = [], gains = [];
  values.forEach((v, i) => {
    assert.equal(v.Origin, pool[i]);
    assert.equal(v.Mass.length, 2); assert(v.Mass.every(p => Number.isFinite(p) && p > 0 && p < 1));
    close(v.Mass[0] + v.Mass[1], 1); assert.equal(v.Base.length, 8);
    assert.equal(v.Conditional.length, 2); assert.equal(v.LogEvidence.length, 2);
    for (let y = 0; y < 2; y++) { assert.equal(v.Conditional[y].length, 8); close(Math.exp(v.LogEvidence[y] - d.BaseLogEvidence), v.Mass[y]); }
    let gain = 0, riskReduction = 0;
    for (let p = 0; p < 8; p++) {
      const base = v.Base[p]; assert(base > 0 && base < 1); close(base, values[0].Base[p]);
      let marginal = 0, remaining = 0;
      for (let y = 0; y < 2; y++) {
        const c = v.Conditional[y][p]; assert(c > 0 && c < 1);
        marginal += v.Mass[y] * c; remaining += v.Mass[y] * c * (1 - c);
        gain += v.Mass[y] * (c - base) ** 2 / 8;
      }
      close(marginal, base); riskReduction += (base * (1 - base) - remaining) / 8;
    }
    close(v.Gain, gain); close(v.Gain, riskReduction);
    const p = v.Mass[1]; entropy.push(-p * Math.log(p) - (1 - p) * Math.log1p(-p)); gains.push(gain);
  });
  const selections = [-1, -1, -1, -1], costs = [0, 0, 0, 0];
  if (pool.length) {
    selections[1] = randomOrigin(source, pool); selections[2] = maximumOrigin(pool, entropy); selections[3] = maximumOrigin(pool, gains);
    costs[1] = costs[2] = costs[3] = 1;
  }
  assert.deepEqual(d.Selected, selections); assert.deepEqual(d.Costs, costs);
  assert.equal(r.Origins.length, 4); assert.equal(r.Predictions.length, 4); assert.equal(r.AtPublication.length, 4);
  assert.equal(r.Redundant.length, 4); assert.equal(r.LogEvidence.length, 4);
  const natural = available(161), supportSets = new Map(), losses = [], redundant = [], changed = [];
  for (let arm = 0; arm < 4; arm++) {
    const selected = selections[arm], support = [...new Set([...natural, ...(selected < 0 ? [] : [selected])])].sort((a, b) => a - b).slice(-64);
    assert.deepEqual(r.Origins[arm], support); assert(Number.isFinite(r.LogEvidence[arm]));
    const isRedundant = selected >= 0 && natural.includes(selected); assert.equal(r.Redundant[arm], isRedundant); redundant.push(Number(isRedundant));
    if (selected >= 0) assert(support.includes(selected));
    const key = support.join(',');
    if (supportSets.has(key)) { const prior = supportSets.get(key); assert.deepEqual(r.AtPublication[arm], r.AtPublication[prior]); assert.deepEqual(r.Predictions[arm], r.Predictions[prior]); assert.equal(r.LogEvidence[arm], r.LogEvidence[prior]); }
    supportSets.set(key, arm);
    assert.equal(r.Predictions[arm].length, 31); assert.equal(r.AtPublication[arm].length, 31);
    let loss = 0;
    for (let i = 0; i < 31; i++) {
      const base = r.AtPublication[arm][i], p = r.Predictions[arm][i], q = source.Steps[161 + i].Q;
      assert(base > 0 && base < 1 && p > 0 && p < 1); assert(Number.isFinite(q) && q >= 0 && q <= 1);
      close(p, .5 + .99 ** i * (base - .5), 1e-14);
      loss += ((p - q) ** 2 + q * (1 - q)) / 31;
    }
    losses.push(loss); changed.push(Number(key !== r.Origins[0].join(',')));
  }
  assert.equal(r.ActualFits, supportSets.size);
  if (source.Schedule === 0) { assert.equal(pool.length, 0); for (let a = 1; a < 4; a++) assert.deepEqual(r.Predictions[a], r.Predictions[0]); }
  const chosenGain = selections.map(j => j < 0 ? 0 : values[pool.indexOf(j)].Gain);
  return {phase: r.Phase, case: r.Case, index: r.Index, schedule: r.Schedule, brier: losses,
    selected: selections, costs, redundant, supportChanged: changed, predictedGain: chosenGain,
    poolSize: pool.length, publicationFits: r.ActualFits};
}
