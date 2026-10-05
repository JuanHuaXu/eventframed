import fs from 'node:fs';
import crypto from 'node:crypto';

const raw = fs.readFileSync('docs/experiments/mmm-age-challenger-v87.jsonl');
const [header, ...rows] = raw.toString().trim().split('\n').map(JSON.parse);
if (header.Version !== 'v87' || rows.length !== 2560) throw Error('incomplete artifact');
for (const [p, h] of Object.entries(header.Hashes)) {
  if (crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex') !== h) throw Error(`source changed ${p}`);
}
const phases = ['design', 'confirmation'];
const cases = ['stable', 'member_shift', 'common_shift', 'recurring', 'null'];
const schedules = [
  { Name: 'immediate', Delay: 0, Jitter: false, Missing: 0 },
  { Name: 'delay16', Delay: 16, Jitter: false, Missing: 0 },
  { Name: 'missing20', Delay: 0, Jitter: false, Missing: .2 },
  { Name: 'jitter31_missing20', Delay: 0, Jitter: true, Missing: .2 },
];
const integer = (x, lo, hi) => Number.isInteger(x) && x >= lo && x <= hi;
const mean = xs => xs.reduce((a, b) => a + b, 0) / xs.length;
const interval = xs => {
  if (xs.length !== 64) throw Error('trajectory count');
  const m = mean(xs), radius = 3.5 * Math.sqrt(xs.reduce((s, x) => s + (x - m) ** 2, 0) / 63 / 64);
  return { mean: m, lower: m - radius, upper: m + radius };
};
const latent = new Map();
for (let k = 0; k < rows.length; k++) {
  const r = rows[k], phase = Math.floor(k / 1280), c = Math.floor(k % 1280 / 256), s = Math.floor(k % 256 / 64);
  if (r.Split !== phases[phase] || r.Scenario !== cases[c] || r.Index !== k % 64 ||
      r.StreamBase !== 2026118701 + phase || JSON.stringify(r.Schedule) !== JSON.stringify(schedules[s])) throw Error('configuration/order');
  if (!integer(r.AgeFits, 0, r.SubsetFits) || !integer(r.AgeSamples, 16*r.AgeFits, 64*r.AgeFits)) throw Error('age accounting');
  const key = `${r.Split}/${r.Scenario}/${r.Index}`;
  if (latent.has(key) && latent.get(key) !== r.LatentTape) throw Error('schedule changed latent trajectory');
  latent.set(key, r.LatentTape);
  if (![r.Tape, r.LatentTape].every(h => /^[a-f0-9]{64}$/.test(h))) throw Error('tape');
  for (const name of ['Received', 'ReceivedAudits', 'Missing', 'PendingPackets', 'MaxPending', 'CountFits', 'SubsetFits', 'MonitorCost', 'AuditCost']) {
    if (!integer(r[name], 0, name.endsWith('Cost') ? 18 * 512 : 512)) throw Error(`counter ${name}`);
  }
  if (r.Received + r.Missing + r.PendingPackets !== 512 || r.ReceivedAudits > r.Received || r.MaxPending > 64 ||
      r.CountFits !== 3 * r.SubsetFits || r.SubsetFits !== (r.ReceivedAudits < 32 ? 0 : 1 + Math.floor((r.ReceivedAudits - 32) / 16)) ||
      r.AuditCost % 18 !== 0 || r.ReceivedAudits > r.AuditCost / 18 || !integer(r.GateFirst, -1, 511)) throw Error('accounting');
  if (r.Arms.length !== 2 || r.Stats.length !== 2 || JSON.stringify(r.Stats[0]) !== JSON.stringify(r.Stats[1]) || r.Arms[0].SplitAt !== r.Arms[1].SplitAt) throw Error('unequal policy');
  const st = r.Stats[0];
  if (!Object.values(st).every(v => integer(v, 0, 512)) || st.Pending !== 0 || st.Applied + st.Stale !== r.Received || st.Censored !== r.Missing + r.PendingPackets) throw Error('journal accounting');
  if (s === 0 && (st.Applied !== 512 || st.Stale !== 0 || st.Censored !== 0)) throw Error('immediate mismatch');
  for (let a = 0; a < 2; a++) {
    const arm = r.Arms[a];
    if (arm.Arm !== ['retained_control', 'age_challenger'][a] || !integer(arm.SplitAt, -1, 511)) throw Error('arm');
    for (const [part, n] of [['Full', 512], ['Post', r.Scenario === 'recurring' ? 384 : 256]]) {
      const m = arm[part];
      if (m.N !== n || !Number.isFinite(m.Brier) || m.Brier < 0 || m.Brier > n || !Number.isFinite(m.LogLoss) || m.LogLoss < 0 ||
          !integer(m.Correct, 0, n) || !integer(m.Cost, 0, 6 * n)) throw Error('metric');
    }
  }
}
let pass = true;
const summaries = [], comparisons = [], versusImmediate = [];
for (const split of phases) for (const scenario of cases) for (const schedule of schedules) {
  const rs = rows.filter(r => r.Split === split && r.Scenario === scenario && r.Schedule.Name === schedule.Name);
  const immediate = rows.filter(r => r.Split === split && r.Scenario === scenario && r.Schedule.Name === 'immediate');
  if (rs.length !== 64 || immediate.length !== 64) throw Error('cell count');
  const postN = scenario === 'recurring' ? 384 : 256;
  for (let a = 0; a < 2; a++) {
    summaries.push({ split, scenario, schedule: schedule.Name, arm: rs[0].Arms[a].Arm,
      fullBrier: mean(rs.map(r => r.Arms[a].Full.Brier / 512)), postBrier: mean(rs.map(r => r.Arms[a].Post.Brier / postN)),
      postAccuracy: mean(rs.map(r => r.Arms[a].Post.Correct / postN)), postLogLoss: mean(rs.map(r => r.Arms[a].Post.LogLoss / postN)),
      foreground: mean(rs.map(r => r.Arms[a].Full.Cost / 512)),
      splitRate: mean(rs.map(r => +(r.Arms[a].SplitAt >= 0))),
    });
    versusImmediate.push({ split, scenario, schedule: schedule.Name, arm: rs[0].Arms[a].Arm,
      postBrierIncrease: interval(rs.map((r, i) => (r.Arms[a].Post.Brier - immediate[i].Arms[a].Post.Brier) / postN)),
    });
  }
  const fullGain = interval(rs.map(r => (r.Arms[0].Full.Brier - r.Arms[1].Full.Brier) / 512));
  const postGain = interval(rs.map(r => (r.Arms[0].Post.Brier - r.Arms[1].Post.Brier) / postN));
  const cellPass = schedule.Name === 'jitter31_missing20' && ['member_shift', 'common_shift'].includes(scenario) ? postGain.mean >= .005 && postGain.lower > 0 : fullGain.lower >= -.01 && postGain.lower >= -.01;
  pass &&= cellPass;
  comparisons.push({ split, scenario, schedule: schedule.Name, fullGain, postGain, cellPass,
    ageFits: mean(rs.map(r => r.AgeFits)), ageSamples: mean(rs.map(r => r.AgeSamples)),
    applied: mean(rs.map(r => r.Stats[0].Applied)), stale: mean(rs.map(r => r.Stats[0].Stale)), censored: mean(rs.map(r => r.Stats[0].Censored)),
    received: mean(rs.map(r => r.Received)), receivedAudits: mean(rs.map(r => r.ReceivedAudits)), subsetFits: mean(rs.map(r => r.SubsetFits)),
    maxPending: Math.max(...rs.map(r => r.MaxPending)), monitorCost: mean(rs.map(r => r.MonitorCost / 512)), auditCost: mean(rs.map(r => r.AuditCost / 512)),
  });
}
console.log(JSON.stringify({ sha256: crypto.createHash('sha256').update(raw).digest('hex'), sourceCount: Object.keys(header.Hashes).length,
  underlyingTrajectories: latent.size, scheduleRuns: rows.length, overall: pass ? 'PASS finite age-challenger screen conditional on fitted bases' : 'FAIL',
  summaries, comparisons, versusImmediate }, null, 2));
