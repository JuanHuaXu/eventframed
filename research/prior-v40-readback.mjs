// Independent artifact/statistical readback; does not rerun or fit learners.
import fs from 'node:fs';
import crypto from 'node:crypto';
import readline from 'node:readline';
import assert from 'node:assert/strict';

const modes = ['full', 'adaptive', 'local16', 'raw2_private', 'raw2_shared', 'raw4_private', 'raw4_shared', 'inverse2_private', 'inverse2_shared', 'inverse4_private', 'inverse4_shared'];
const schedules = ['immediate', 'fixed150', 'uniform299'];
const fields = ['IssuedBrier', 'IssuedPriority', 'Recovery', 'FinalBrier', 'FinalPriority', 'FinalUsefulness'];
const diagnosticOnly = process.argv[2] === 'diagnostic';
assert(process.argv[2] === undefined || diagnosticOnly, 'choose diagnostic or omit for complete study');
const stationary = new Set(['aligned', 'independent', 'curved', 'baseline_matched', 'stationary_noise10']);
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
async function fileHash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
function near(a, b, name) {
  assert(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) <= 3e-12, name + ': ' + a + ' vs ' + b);
}
function interval(x) {
  const mean = x.reduce((s, v) => s + v / x.length, 0);
  const se = Math.sqrt(x.reduce((s, v) => s + (v - mean) ** 2, 0) / (x.length - 1) / x.length);
  return {Mean: mean, SE: se, Lower: mean - 3.5 * se, Upper: mean + 3.5 * se};
}
function metric(a, f) {
  if (f.startsWith('Final')) return a.Snapshots.at(-1)[{FinalBrier: 'Brier', FinalPriority: 'PriorityBrier', FinalUsefulness: 'PacketUsefulness'}[f]];
  return a[f];
}
function gain(a, b, field) {
  const d = metric(a, field) - metric(b, field);
  return field === 'FinalUsefulness' ? -d : d;
}
function verifyCell(report, name, ws, n, allocation) {
  const cell = report.Groups[name];
  assert(cell, 'missing group ' + name);
  const max = Math.max(...ws.flatMap(w => w.arms.map(a => a.Costs.ElapsedNS)));
  for (let j = 0; j < 11; j++) {
    const row = cell[modes[j]];
    assert(row);
    for (const f of fields) {
      const x = ws.map(w => metric(w.arms[j], f));
      if (n === 1) {
        near(row[f].Mean, x[0], name + '/' + f);
        assert.equal(row[f].N, 1);
        assert.equal(row[f].Interval, null);
        assert.equal(row[f].DiagnosticOnly, true);
      } else {
        const v = interval(x);
        for (const k of Object.keys(v)) near(row[f][k], v[k], name + '/' + modes[j] + '/' + f + '/' + k);
      }
    }
    assert.equal(row.MaximumElapsedNS, Math.max(...ws.map(w => w.arms[j].Costs.ElapsedNS)));
  }
  for (const factor of ['SharedMinusPrivate', 'InverseMinusRaw', 'Strength2Minus4']) {
    for (const f of fields) {
      const x = ws.map(w => {
        const pairs = factor === 'SharedMinusPrivate' ? [[3,4],[5,6],[7,8],[9,10]] : factor === 'InverseMinusRaw' ? [[3,7],[4,8],[5,9],[6,10]] : [[5,3],[6,4],[9,7],[10,8]];
        return pairs.reduce((s, [a,b]) => s + gain(w.arms[a], w.arms[b], f) / 4, 0);
      });
      const row = report.FactorialContrasts[name][factor][f];
      if (n === 1) near(row.Mean, x[0], 'diagnostic contrast');
      else for (const [k, v] of Object.entries(interval(x))) near(row[k], v, name + '/' + factor + '/' + f);
    }
  }
  if (n === 1) return {};
  const decisions = {};
  for (let j = 3; j < 11; j++) {
    const checks = {WorkPass: max <= 400000000, AllocationPass: allocation <= 8 * 1024 * 1024};
    let pass = checks.WorkPass && checks.AllocationPass;
    for (let control = 0; control < 2; control++) for (const f of fields.filter(f => f !== 'Recovery')) {
      const v = interval(ws.map(w => gain(w.arms[control], w.arms[j], f)));
      const improve = control === 0 && !stationary.has(ws[0].regime) && (f === 'IssuedBrier' || f === 'IssuedPriority');
      const ok = improve ? v.Mean >= .01 && v.Lower > 0 : v.Lower >= -.01;
      const saved = report.Gates[name][modes[j]].Checks[modes[control] + '/' + f];
      for (const k of Object.keys(v)) near(saved.Gain[k], v[k], 'gate interval');
      assert.equal(saved.Pass, ok);
      assert.equal(saved.Require001Improvement, improve);
      checks[modes[control] + '/' + f] = ok;
      pass &&= ok;
    }
    if (ws[0].changes) {
      const v = interval(ws.map(w => w.arms[0].Recovery - w.arms[j].Recovery));
      const baseline = interval(ws.map(w => w.arms[0].Recovery));
      const ok = v.Lower > 0 && v.Mean >= .1 * baseline.Mean;
      const saved = report.Gates[name][modes[j]].Checks.Recovery;
      for (const k of Object.keys(v)) {
        near(saved.Gain[k], v[k], 'recovery gain');
        near(saved.Control[k], baseline[k], 'recovery control');
      }
      assert.equal(saved.Pass, ok);
      checks.Recovery = ok;
      pass &&= ok;
    }
    const saved = report.Gates[name][modes[j]];
    assert.equal(saved.Checks.WorkPass, checks.WorkPass);
    assert.equal(saved.Checks.AllocationPass, checks.AllocationPass);
    assert.equal(saved.Pass, pass);
    decisions[modes[j]] = {pass, failedChecks: Object.entries(checks).filter(([,v]) => !v).map(([k]) => k)};
  }
  return decisions;
}
const completed = [];
for (const stage of diagnosticOnly ? ['diagnostic'] : ['diagnostic', 'normal']) {
  const root = 'research/prior-v40-' + stage;
  const finish = JSON.parse(fs.readFileSync(root + '/completed.json'));
  const freeze = JSON.parse(fs.readFileSync(root + '/freeze.json'));
  assert.equal(finish.source_unchanged, true);
  assert.equal(finish.stage, stage);
  for (const [p, h] of Object.entries(freeze.files)) assert.equal(await fileHash(p), h, 'frozen source ' + p);
  for (const [p, h] of Object.entries(finish.artifacts)) assert.equal(await fileHash(root + '/' + p), h, 'artifact ' + p);
  for (const command of finish.checks) {
    assert.equal(command.code, 0);
    assert.equal(await fileHash(root + '/' + command.name + '.log'), command.log_sha256);
  }
  completed.push({stage, sha256: await fileHash(root + '/completed.json')});
}
const summary = {}, allSeeds = new Set();
let negativeControls = 0;
for (const split of diagnosticOnly ? ['diagnostic'] : ['diagnostic', 'design', 'confirmation']) {
  const root = split === 'diagnostic' ? 'research/prior-v40-diagnostic' : 'research/prior-v40-normal';
  const raw = root + '/' + split + '.jsonl';
  const report = JSON.parse(fs.readFileSync(root + '/' + split + '-audit.json'));
  const n = split === 'diagnostic' ? 1 : 16, groups = new Map();
  let manifest, worlds = 0, snapshots = 0;
  for await (const line of readline.createInterface({input: fs.createReadStream(raw), crlfDelay: Infinity})) {
    const w = JSON.parse(line);
    if (!manifest) { manifest = w; assert.equal(w.Split, split); assert.equal(w.Worlds, 28 * n); continue; }
    assert(!allSeeds.has(w.Population.Seed), 'cross-cohort seed reuse');
    allSeeds.add(w.Population.Seed);
    assert.equal(w.Arms.length, 33);
    worlds++;
    for (let s = 0; s < 3; s++) {
      const key = w.Population.Geometry + '/' + w.Population.Regime + '/' + schedules[s];
      const arms = w.Arms.slice(s * 11, (s + 1) * 11);
      for (let j = 0; j < 11; j++) {
        assert.equal(arms[j].Mode, modes[j]); assert.equal(arms[j].Schedule, schedules[s]);
        snapshots += arms[j].Snapshots.length;
      }
      const compact = {regime: w.Population.Regime, changes: w.Population.Changes?.length > 0, arms: arms.map(a => ({...a, Issued: undefined, ExpertIssued: undefined, Receipts: undefined, Snapshots: [a.Snapshots.at(-1)]}))};
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key).push(compact);
    }
  }
  assert.equal(worlds, 28 * n); assert.equal(snapshots, worlds * 550); assert.equal(groups.size, 84);
  assert.equal(report.SHA256, await fileHash(raw));
  const failures = Object.fromEntries(modes.slice(3).map(m => [m, 0])), reasons = {}, decisions = {};
  for (const [name, ws] of groups) {
    assert.equal(ws.length, n);
    const cell = verifyCell(report, name, ws, n, report.MaximumConstructorBytes);
    decisions[name] = cell;
    for (const [m, d] of Object.entries(cell)) {
      if (!d.pass) failures[m]++;
      reasons[m] ??= {};
      for (const reason of d.failedChecks) reasons[m][reason] = (reasons[m][reason] ?? 0) + 1;
    }
  }
  if (n === 16) assert.deepEqual(failures, report.FailedCells);
  else {assert.equal(report.QualityAdoptionEvaluated, false); assert.equal(report.Gates, null); assert.equal(report.FailedCells, null);}
  const [key, ws] = groups.entries().next().value;
  for (const mutate of [r => {r.Groups[key].full.IssuedBrier.Mean += .01;}, r => {r.FactorialContrasts[key].SharedMinusPrivate.IssuedBrier.Mean += .01;}]) {
    const r = structuredClone(report); mutate(r);
    assert.throws(() => verifyCell(r, key, ws, n, report.MaximumConstructorBytes)); negativeControls++;
  }
  if (n === 16) {
    const r = structuredClone(report); r.Gates[key].raw2_private.Pass = !r.Gates[key].raw2_private.Pass;
    assert.throws(() => verifyCell(r, key, ws, n, report.MaximumConstructorBytes)); negativeControls++;
  }
  summary[split] = {worlds, arms: worlds * 33, snapshots, distinctTrials: worlds * 2400, raw_sha256: report.SHA256, failures: n === 16 ? failures : null, reasons, decisions, maximumElapsedNS: Object.fromEntries(modes.map(m => [m, Math.max(...Object.values(report.Groups).map(c => c[m].MaximumElapsedNS))]))};
}
const usage = Number(process.env.EVENTFRAME_WEEKLY_USAGE);
assert(Number.isFinite(usage) && usage >= 0 && usage <= 100, 'explicit current weekly usage required');
const result = {time: new Date().toISOString(), completed, summary, negativeControls, sourceAndArtifactHashes: true, independentStatisticsAndGates: true, disjointWorldSeeds: true, weeklyUsage: usage, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE', productionChanged: false};
fs.writeFileSync((diagnosticOnly ? 'research/prior-v40-diagnostic' : 'research/prior-v40-normal') + '/readback.json', JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({negativeControls, failures: Object.fromEntries(Object.entries(summary).map(([k,v]) => [k,v.failures])), allSevenWholeGoals: 'OPEN'}));
