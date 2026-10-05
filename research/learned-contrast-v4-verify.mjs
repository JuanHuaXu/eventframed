import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root = path.resolve(import.meta.dirname, '..');
const cases = ['stable', 'bit2_shift', 'bit2_delayed', 'bit0_shift', 'interaction_shift', 'null', 'majority_ood'];
const correct = {bit2_shift: 3, bit2_delayed: 3, bit0_shift: 1, interaction_shift: 10};
const changed = c => !['stable', 'null'].includes(c);
const near = (a, b, label) => assert.ok(Number.isFinite(a) && Math.abs(a-b) < 1e-9, `${label}: ${a} != ${b}`);
const avg = a => a.reduce((s, x) => s+x, 0)/a.length;
function interval(rows, candidate, control, field) {
  const fit = Array.from({length: 16}, (_, f) => avg(rows.filter(r => r.Fit === f)
    .map(r => r.Arms[control][field] - r.Arms[candidate][field])));
  const mean = avg(fit), variance = fit.reduce((s, x) => s+(x-mean)**2, 0)/15;
  return {mean, low: mean-3.5*Math.sqrt(variance/16), high: mean+3.5*Math.sqrt(variance/16)};
}
function truth(c, x, clock) {
  if (c === 'null') return .5;
  let bit = Boolean(x & 64) !== Boolean(x & 128);
  bit = bit !== Boolean(x & 256);
  if (clock >= 256) {
    if (c === 'bit2_shift' || c === 'bit2_delayed') bit = Boolean(x & 4);
    if (c === 'bit0_shift') bit = Boolean(x & 1);
    if (c === 'interaction_shift') bit = Boolean(x & 1) !== Boolean(x & 2) !== Boolean(x & 4);
    if (c === 'majority_ood') bit = [1,2,4].filter(b => x & b).length >= 2;
  }
  return bit ? .95 : .05;
}
function recover(expected) {
  let consecutive = 0;
  for (let clock=287; clock<512; clock++) {
    const mean = avg(expected.slice(clock-31, clock+1));
    consecutive = mean <= .12 ? consecutive+1 : 0;
    if (consecutive === 16) return {delay: clock-256, missed: false};
  }
  return {delay: 256, missed: true};
}

const result = {splits: {}};
for (const split of ['design', 'confirmation']) {
  const file = path.join(root, `docs/experiments/mmm-learned-contrast-v4-${split}.jsonl.gz`);
  const reader = readline.createInterface({input: fs.createReadStream(file).pipe(zlib.createGunzip())});
  const byCase = new Map(cases.map(c => [c, []])), identities = new Set();
  let manifest, count = 0;
  for await (const line of reader) {
    const row = JSON.parse(line);
    if (!manifest) {
      manifest = row;
      assert.equal(row.kind, 'manifest');
      assert.equal(row.split, split);
      assert.equal(row.seedBase, split === 'design' ? 2026102401 : 2026102402);
      assert.equal(row.fitOffset, split === 'design' ? 800 : 900);
      assert.equal(row.trialsPerCase, 256);
      assert.equal(row.budget, 128);
      assert.ok(row.stateBytes > 0 && row.stateBytes < 4096);
      near(row.prior.reduce((a,b) => a+b, 0), 1, 'prior sum');
      for (const [source, expected] of Object.entries(row.hashes)) {
        const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(root, source))).digest('hex');
        assert.equal(actual, expected, `source changed: ${source}`);
      }
      continue;
    }
    assert.equal(row.Kind, 'trial');
    assert.equal(row.Split, split);
    assert.ok(byCase.has(row.Scenario));
    assert.ok(row.Fit >= 0 && row.Fit < 16 && row.Stream >= 0 && row.Stream < 16);
    const id = `${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!identities.has(id)); identities.add(id);
    assert.equal(row.Tape.length, 512);
    assert.equal(row.Arms.length, 8);
    const delay = row.Scenario === 'bit2_delayed' ? 16 : 0;
    const summary = {Fit: row.Fit, Arms: []};
    for (let clock=0; clock<512; clock++) {
      const event = row.Tape[clock];
      assert.ok(Number.isInteger(event.X) && event.X >= 0 && event.X < 512);
      near(event.PTrue, truth(row.Scenario, event.X, clock), `${id}/truth/${clock}`);
    }
    for (const [armIndex, wrapped] of row.Arms.entries()) {
      const arm = wrapped.Data.Data;
      assert.equal(arm.Policy, ['random','uncertainty','learned_disagreement','learned_disagreement','learned_disagreement','random','uncertainty','learned_disagreement'][armIndex]);
      assert.equal(arm.Ticks.length, 512);
      let full=0, first=0, post=0, early=0, expectedPost=0;
      let nominated=0, nominatedPost=0, correctPost=0, arrived=0, pending=0, gateCount=0;
      const selected = new Set(), delivered = new Set(), expected=[];
      for (let clock=0; clock<512; clock++) {
        const event=row.Tape[clock], tick=arm.Ticks[clock];
        assert.ok(Number.isFinite(tick.P) && tick.P>0 && tick.P<1);
        assert.ok(tick.Alternate>=1 && tick.Alternate<=10);
        const score=(tick.P-Number(event.Y))**2;
        const proper=event.PTrue*(1-event.PTrue)+(tick.P-event.PTrue)**2;
        expected.push(proper);
        full+=score/512;
        if (clock<64) first+=score/64;
        if (clock>=256) {
          post+=score/256; expectedPost+=proper/256;
          if (clock<320) early+=score/64;
        }
        if (tick.Selected) {
          selected.add(clock); nominated++;
          if (clock>=256) {
            nominatedPost++;
            if (tick.Alternate===correct[row.Scenario]) correctPost++;
          }
          if (!event.Missing && clock+delay>=512) pending++;
        }
        for (const origin of tick.Delivered ?? []) {
          assert.ok(origin<=clock && selected.has(origin), 'future or unrequested label');
          assert.ok(!row.Tape[origin].Missing, 'missing label delivered');
          assert.equal(origin+delay, clock, 'wrong delivery clock');
          assert.ok(!delivered.has(origin), 'duplicate label');
          delivered.add(origin); arrived++;
        }
        if (wrapped.GateAt?.[clock]) gateCount++;
      }
      for (const origin of selected) {
        if (!row.Tape[origin].Missing && origin+delay<512) assert.ok(delivered.has(origin));
      }
      assert.equal(nominated, 128);
      assert.equal(arm.Nominated, nominated);
      assert.equal(arm.NominatedPost, nominatedPost);
      assert.equal(arm.CorrectPost, correctPost);
      assert.equal(arm.Arrived, arrived);
      assert.equal(arm.Pending, pending);
      near(full, arm.FullBrier, `${id}/${armIndex}/full`);
      near(post, arm.PostBrier, `${id}/${armIndex}/post`);
      near(early, arm.EarlyBrier, `${id}/${armIndex}/early`);
      near(expectedPost, wrapped.Data.PostExpected, `${id}/${armIndex}/expected`);
      const recovery=changed(row.Scenario)?recover(expected):{delay:-1,missed:false};
      assert.equal(wrapped.Data.RecoveryDelay, recovery.delay);
      assert.equal(wrapped.Data.MissedRecovery, recovery.missed);
      if (armIndex>=5) {
        assert.equal(wrapped.GateAt.length, 512);
        assert.equal(wrapped.GateCount, gateCount);
        assert.equal(wrapped.FirstGate, wrapped.GateAt.indexOf(true));
      } else {
        assert.equal(gateCount, 0);
      }
      summary.Arms.push({FullBrier:full, FirstBrier:first, PostBrier:post,
        EarlyBrier:early, PostExpected:expectedPost, RecoveryDelay:recovery.delay,
        MissedRecovery:Number(recovery.missed), Arrived:arrived,
        GateRateAfter100:armIndex>=5?wrapped.GateAt.slice(100).filter(Boolean).length/412:0});
    }
    for (const [a,b] of [[0,5],[1,6],[2,4],[2,7]]) {
      for (let clock=0; clock<512; clock++) {
        assert.equal(row.Arms[a].Data.Data.Ticks[clock].Selected,
          row.Arms[b].Data.Data.Ticks[clock].Selected, `selection mismatch ${id}/${a}/${b}/${clock}`);
      }
    }
    byCase.get(row.Scenario).push(summary); count++;
  }
  assert.ok(manifest);
  assert.equal(count, 1792);
  const splitReport={trajectories:count,cases:{},pass:true};
  for (const scenario of cases) {
    const rows=byCase.get(scenario);
    assert.equal(rows.length,256);
    const mean=field=>Array.from({length:8},(_,a)=>avg(rows.map(r=>r.Arms[a][field])));
    const full=mean('FullBrier'),first=mean('FirstBrier'),post=mean('PostBrier');
    const early=mean('EarlyBrier'),expected=mean('PostExpected');
    const recovery=mean('RecoveryDelay'),missed=mean('MissedRecovery');
    const arrived=mean('Arrived'),gate=mean('GateRateAfter100');
    const gain=(a,b,field)=>interval(rows,a,b,field);
    const checks={};
    if (scenario==='null') {
      checks.null=[0,1,2].every(i=>full[i+5]<=.27 &&
        gain(i+5,i,'FullBrier').mean>=.05 && gain(i+5,i,'FullBrier').low>0);
    }
    if (scenario==='stable') {
      checks.stable=gain(7,2,'FullBrier').low>-.01 && gate[7]<.05;
    }
    if (scenario==='bit2_shift'||scenario==='bit2_delayed') {
      checks.known=[5,6].every(i=>{
        const g=gain(7,i,'PostBrier');
        return g.mean>=.01 && g.low>0 && first[7]-first[i]<=.005 &&
          recovery[i]-recovery[7]>=10 && missed[7]<=missed[i]+.05;
      }) && gain(7,2,'RecoveryDelay').low> -5;
    }
    if (['bit2_shift','bit2_delayed','bit0_shift','interaction_shift'].includes(scenario)) {
      checks.knownHarm=gain(7,2,'PostBrier').low>-.01;
    }
    if (scenario==='majority_ood') {
      const g=gain(7,2,'PostExpected');
      checks.ood=expected[7]<=.26 && g.mean>=.01 && g.low>0;
    }
    checks.arrival=Math.abs(arrived[7]-arrived[5])<=2 && Math.abs(arrived[7]-arrived[6])<=2;
    const pass=Object.values(checks).every(Boolean);
    splitReport.cases[scenario]={full,first,post,early,expected,recovery,missed,arrived,gate,
      nullFullGain:[0,1,2].map(i=>gain(i+5,i,'FullBrier')),
      recoveryHarm:gain(7,2,'RecoveryDelay'),
      gainVsOld:gain(7,2,'PostBrier'),gainVsRandom:gain(7,5,'PostBrier'),
      gainVsUncertainty:gain(7,6,'PostBrier'),oodGain:gain(7,2,'PostExpected'),checks,pass};
    splitReport.pass &&=pass;
  }
  result.splits[split]=splitReport;
}
result.efficacyPass=result.splits.design.pass&&result.splits.confirmation.pass;
console.log(JSON.stringify(result,null,2));
