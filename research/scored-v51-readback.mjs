// Independent metric arithmetic and exact same-world factorial/legacy pairing.
import fs from 'node:fs';
import crypto from 'node:crypto';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const root = 'research/scored-v51-diagnostic', done = JSON.parse(fs.readFileSync(root + '/completed.json'));
const freeze = JSON.parse(fs.readFileSync(root + '/freeze.json'));
assert(done.sourceUnchanged && done.checks.length === 16 && done.checks.every(x => x.code === 0));
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
for (const [p, h] of Object.entries(freeze.files)) { assert.equal(hash(fs.readFileSync(p)), h); assert.equal(hash(fs.readFileSync(root + '/source/' + p)), h); }
const usage = Number(process.env.EVENTFRAME_WEEKLY_USAGE); assert(Number.isFinite(usage) && usage >= 0 && usage <= 80);
const allocation = JSON.parse(fs.readFileSync(root + '/allocation.json')).constructorBytes;
const mean = x => x.reduce((s, v) => s + v / x.length, 0);
async function* lines(name) {
  const digest = crypto.createHash('sha256'), input = fs.createReadStream(root + '/' + name);
  input.on('data', b => digest.update(b));
  for await (const line of readline.createInterface({ input, crlfDelay: Infinity })) yield JSON.parse(line);
  assert.equal(digest.digest('hex'), done.artifacts[name], name);
}
const policies = ['static', 'slow', 'round'], schedules = ['immediate', 'fixed150', 'uniform299'];
const stationary = new Set(['aligned', 'independent', 'curved', 'baseline_matched', 'stationary_noise10']);
function near(a, b) { assert(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) < 3e-12); }
function derive(w, a) {
  assert.equal(a.Issued.length, 2400); assert(a.Issued.every(q => Number.isFinite(q) && q > 0 && q < 1));
  const loss = a.Issued.map((q, j) => { const p = w.Rates[Math.floor(j / 150)][j % 150]; return q * q - 2 * q * p + p; });
  const snapshots = a.Snapshots.map(s => {
    const rates = w.Rates[Math.min(15, Math.floor(s.Tick / 150))]; assert.equal(s.Forecast.length, 150);
    const l = s.Forecast.map((q, j) => { assert(Number.isFinite(q) && q > 0 && q < 1); return q * q - 2 * q * rates[j] + rates[j]; });
    const top = s.Forecast.map((q, j) => ({ q, j })).sort((a, b) => b.q - a.q || a.j - b.j).slice(0, 10);
    return { tick: s.Tick, brier: mean(l), priority: l.reduce((x, v, j) => x + v * (j < 10 ? 3 : 1) / 170, 0), usefulness: mean(top.map(v => rates[v.j])) };
  });
  let recovery = 0;
  for (let k = 0; k < (w.Changes ?? []).length; k++) {
    const start = w.Changes[k], end = w.Changes[k + 1] ?? 16; let consecutive = 0, delay = end - start + 1;
    for (const s of snapshots) {
      if (s.tick >= 2400) continue;
      const round = Math.floor(s.tick / 150); if (round < start || round >= end) continue;
      consecutive = s.brier <= .20 && s.usefulness >= .75 ? consecutive + 1 : 0;
      if (consecutive === 2) { delay = round - start + 1; break; }
    }
    recovery += delay / w.Changes.length;
  }
  const last = snapshots.at(-1);
  return { IssuedBrier: mean(loss), IssuedPriority: loss.reduce((x, v, j) => x + v * (j % 150 < 10 ? 3 : 1) / (16 * 170), 0),
    FinalBrier: last.brier, FinalPriority: last.priority, FinalUsefulness: last.usefulness, Recovery: recovery };
}
const fi = lines('diagnostic-fixture.jsonl')[Symbol.asyncIterator]();
const streams = Object.fromEntries(freeze.styles.map(style => [style, lines(style + '.jsonl')[Symbol.asyncIterator]() ]));
const fm = (await fi.next()).value;
assert.equal(fm.Kind, 'fixture_manifest'); assert.equal(fm.Worlds, 28); assert.equal(fm.SeedBase, 2026105107);
for (const [style, it] of Object.entries(streams)) {
  const rm = (await it.next()).value; assert.equal(rm.Style, style); delete rm.Style; rm.Kind = fm.Kind; assert.deepEqual(rm, fm);
}
const groups = Object.fromEntries(freeze.styles.map(s => [s, {}]));
let worlds = 0, legacyPairs = 0, sharedAdvicePairs = 0;
const seeds = new Set(), maxMS = Object.fromEntries(freeze.styles.map(s => [s, 0]));
for (;;) {
  const next = await fi.next(), raw = {};
  for (const [style, it] of Object.entries(streams)) { const r = await it.next(); assert.equal(r.done, next.done); if (!r.done) raw[style] = r.value; }
  if (next.done) break;
  const f = next.value, w = f.World.Population;
  assert(!seeds.has(w.Seed)); seeds.add(w.Seed); worlds++;
  for (const r of Object.values(raw)) { assert.equal(r.Seed, w.Seed); assert.equal(r.Arms.length, 9); }
  for (let j = 0; j < 9; j++) {
    const { Costs: oldCost, ...old } = raw.legacy.Arms[j], { Costs: newCost, ...replacement } = raw.log_mean.Arms[j];
    assert.deepEqual(old, replacement, 'every non-cost legacy field'); legacyPairs++;
    const controls = f.World.Arms.slice(Math.floor(j / 3) * 3, Math.floor(j / 3) * 3 + 2).map(a => { const m = derive(w, a); near(m.IssuedBrier, a.IssuedBrier); near(m.IssuedPriority, a.IssuedPriority); near(m.Recovery, a.Recovery); return m; });
    const legacy = derive(w, raw.legacy.Arms[j]);
    for (const style of freeze.styles) {
      const a = raw[style].Arms[j]; assert.equal(a.Mode, policies[j % 3]); assert.equal(a.Schedule, schedules[Math.floor(j / 3)]);
      assert.deepEqual(a.Advice, raw.legacy.Arms[j].Advice); assert.deepEqual(a.Heads, raw.legacy.Arms[j].Heads);
      for (let k = 0; k < a.Snapshots.length; k++) {
        const v = a.Snapshots[k], b = raw.legacy.Arms[j].Snapshots[k];
        for (const name of ['Advice', 'Heads', 'Weights', 'GlobalWeights']) assert.deepEqual(v[name], b[name]);
      }
      if (style !== 'legacy') sharedAdvicePairs++;
      const c = a.Costs; assert(c.ElapsedNS >= c.AccountedNS && c.AccountedNS === c.SetupNS + c.IssueNS + c.ResolveNS + c.SnapshotNS);
      assert(c.SetupNS > 0 && c.IssueNS > 0 && c.ResolveNS > 0 && c.SnapshotNS > 0); maxMS[style] = Math.max(maxMS[style], c.ElapsedNS / 1e6);
      const key = w.Geometry + '/' + w.Regime + '/' + a.Schedule; assert(!groups[style][key]?.[a.Mode]);
      groups[style][key] ??= {}; const candidate = derive(w, a), gains = {};
      for (const [name, control] of [['full', controls[0]], ['adaptive', controls[1]], ['legacy', legacy]]) {
        gains[name] = Object.fromEntries(Object.keys(candidate).map(k => [k, (control[k] - candidate[k]) * (k === 'FinalUsefulness' ? -1 : 1)]));
      }
      groups[style][key][a.Mode] = { candidate, controls, gains, elapsedMS: c.ElapsedNS / 1e6 };
    }
  }
}
assert.equal(worlds, 28); assert.equal(legacyPairs, 252); assert.equal(sharedAdvicePairs, 1008);
const summaries = {};
for (const style of freeze.styles) {
  assert.equal(Object.keys(groups[style]).length, 84);
  summaries[style] = { maximumLoopMS: maxMS[style], loopScreenPass: maxMS[style] <= 400,
    constructorBytes: allocation[style], allocationScreenPass: allocation[style] <= 8 << 20, policies: {} };
  for (const mode of policies) {
    const rows = Object.entries(groups[style]).map(([key, v]) => ({ key, ...v[mode] })); assert.equal(rows.length, 84);
    const shifted = rows.filter(v => !stationary.has(v.key.split('/')[1])); assert.equal(shifted.length, 54);
    summaries[style].policies[mode] = {
      issuedBrierVsFull: mean(rows.map(v => v.gains.full.IssuedBrier)),
      issuedBrierVsAdaptive: mean(rows.map(v => v.gains.adaptive.IssuedBrier)),
      shiftedIssuedBrierVsFull: mean(shifted.map(v => v.gains.full.IssuedBrier)),
      issuedBrierVsLegacy: mean(rows.map(v => v.gains.legacy.IssuedBrier)),
      finalUsefulnessVsAdaptive: mean(rows.map(v => v.gains.adaptive.FinalUsefulness)),
      finalUsefulnessVsLegacy: mean(rows.map(v => v.gains.legacy.FinalUsefulness)),
      legacyIssuedBrierWins: rows.filter(v => v.gains.legacy.IssuedBrier > 0).length,
      legacyIssuedBrierLosses: rows.filter(v => v.gains.legacy.IssuedBrier < 0).length,
      cellsWithMeanAdaptiveHarmBeyond01: rows.filter(v => ['IssuedBrier', 'IssuedPriority', 'FinalBrier', 'FinalPriority', 'FinalUsefulness'].some(k => v.gains.adaptive[k] < -.01)).length,
    };
  }
}
const out = { time: new Date().toISOString(), sourceAndRawHashes: true, independentMetricArithmetic: true,
  worlds, distinctLabels: worlds * 2400, candidateArms: worlds * 9 * 4, legacyControlArms: worlds * 9,
  exactLegacyPairs: legacyPairs, exactSharedAdvicePairs: sharedAdvicePairs, groups, summaries,
  nPerCell: 1, intervalsClaimed: false, qualityAdoption: false, weeklyUsage: usage, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' };
fs.writeFileSync(root + '/readback.json', JSON.stringify(out, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify({ worlds, distinctLabels: out.distinctLabels, exactLegacyPairs: legacyPairs, exactSharedAdvicePairs: sharedAdvicePairs, summaries, qualityAdoption: false }, null, 2));
