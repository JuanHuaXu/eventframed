import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root = path.resolve(import.meta.dirname, '..');
const cases = {
  early_bit2: {change:128, noise:.05, missing:.20, delay:0, known:true},
  late_bit2: {change:384, noise:.05, missing:.20, delay:0, known:true},
  noisy_bit2: {change:256, noise:.20, missing:.20, delay:0, known:true},
  delayed_bit2: {change:256, noise:.05, missing:.25, delay:32, known:true},
  skew_bit2: {change:256, noise:.05, missing:.20, delay:0, known:true, skew:true},
  stable20: {change:512, noise:.20, missing:.20, delay:0, stable:true},
  majority_ood: {change:256, noise:.05, missing:.20, delay:0, ood:true},
};
const knownCases = Object.keys(cases).filter(c => cases[c].known);
const near = (a,b,label) => assert.ok(Number.isFinite(a) && Math.abs(a-b)<1e-9, `${label}: ${a} != ${b}`);
const avg = a => a.reduce((s,x)=>s+x,0)/a.length;
function interval(rows, candidate, control, field) {
  const fit=Array.from({length:16},(_,f)=>avg(rows.filter(r=>r.Fit===f)
    .map(r=>r.Arms[control][field]-r.Arms[candidate][field])));
  const mean=avg(fit), variance=fit.reduce((s,x)=>s+(x-mean)**2,0)/15;
  return {mean,low:mean-3.5*Math.sqrt(variance/16),high:mean+3.5*Math.sqrt(variance/16)};
}
function truth(name,x,clock) {
  const c=cases[name];
  let bit=Boolean(x&64)!==Boolean(x&128);
  bit=bit!==Boolean(x&256);
  if (clock>=c.change && !c.stable) {
    bit=c.ood ? [1,2,4].filter(b=>x&b).length>=2 : Boolean(x&4);
  }
  return bit ? 1-c.noise : c.noise;
}
function recovery(expected,c) {
  if (!c.known || c.noise>=.20) return {delay:-1,missed:false};
  let consecutive=0;
  for (let clock=c.change+31;clock<512;clock++) {
    const mean=avg(expected.slice(clock-31,clock+1));
    consecutive=mean<=.12 ? consecutive+1 : 0;
    if (consecutive===16) return {delay:clock-c.change,missed:false};
  }
  return {delay:512-c.change,missed:true};
}

const result={splits:{}};
for (const split of ['design','confirmation']) {
  const file=path.join(root,`docs/experiments/mmm-learned-contrast-transfer-v1-${split}.jsonl.gz`);
  const reader=readline.createInterface({input:fs.createReadStream(file).pipe(zlib.createGunzip())});
  const byCase=new Map(Object.keys(cases).map(c=>[c,[]])), ids=new Set();
  let manifest,count=0;
  for await (const line of reader) {
    const row=JSON.parse(line);
    if (!manifest) {
      manifest=row;
      assert.equal(row.kind,'manifest');
      assert.equal(row.split,split);
      assert.equal(row.seedBase,split==='design'?2026102501:2026102502);
      assert.equal(row.fitOffset,split==='design'?1200:1300);
      assert.equal(row.trialsPerCase,256);
      assert.equal(row.budget,128);
      assert.ok(row.stateBytes>0&&row.stateBytes<4096);
      for (const [source,expected] of Object.entries(row.hashes)) {
        const actual=crypto.createHash('sha256').update(fs.readFileSync(path.join(root,source))).digest('hex');
        assert.equal(actual,expected,`source changed: ${source}`);
      }
      continue;
    }
    assert.equal(row.Kind,'trial');
    assert.equal(row.Split,split);
    const c=cases[row.Scenario];
    assert.ok(c);
    assert.ok(row.Fit>=0&&row.Fit<16&&row.Stream>=0&&row.Stream<16);
    const id=`${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!ids.has(id)); ids.add(id);
    assert.equal(row.Tape.length,512);
    assert.equal(row.Arms.length,3);
    const postStart=c.stable?256:c.change;
    const summary={Fit:row.Fit,Arms:[],SkewRate:0};
    for (let clock=0;clock<512;clock++) {
      const event=row.Tape[clock];
      assert.ok(Number.isInteger(event.X)&&event.X>=0&&event.X<512);
      if (c.skew) {
        summary.SkewRate+=Number(Boolean(event.X&4))/512;
      }
      near(event.PTrue,truth(row.Scenario,event.X,clock),`${id}/truth/${clock}`);
    }
    for (const [index,wrapped] of row.Arms.entries()) {
      const arm=wrapped.Data;
      assert.equal(arm.Policy,['random','uncertainty','learned_disagreement'][index]);
      assert.equal(arm.Ticks.length,512);
      let full=0,post256=0,early256=0,fullExpected=0,postExpected=0,firstExpected=0,postRealized=0;
      let nominated=0,nominatedPost=0,correctPost=0,arrived=0,pending=0;
      const selected=new Set(), delivered=new Set(), expected=[];
      for (let clock=0;clock<512;clock++) {
        const event=row.Tape[clock], tick=arm.Ticks[clock];
        assert.ok(Number.isFinite(tick.P)&&tick.P>0&&tick.P<1);
        assert.ok(tick.Alternate>=1&&tick.Alternate<=10);
        const score=(tick.P-Number(event.Y))**2;
        const proper=event.PTrue*(1-event.PTrue)+(tick.P-event.PTrue)**2;
        expected.push(proper);
        full+=score/512; fullExpected+=proper/512;
        if (clock>=256) {
          post256+=score/256;
          if (clock<320) early256+=score/64;
        }
        if (clock>=postStart) {
          postExpected+=proper/(512-postStart);
          postRealized+=score/(512-postStart);
          if (clock<postStart+64) firstExpected+=proper/64;
        }
        if (tick.Selected) {
          nominated++; selected.add(clock);
          if (clock>=256) {
            nominatedPost++;
            if (c.known&&tick.Alternate===3) correctPost++;
          }
          if (!event.Missing&&clock+c.delay>=512) pending++;
        }
        for (const origin of tick.Delivered??[]) {
          assert.ok(origin<=clock&&selected.has(origin),'future or unrequested label');
          assert.ok(!row.Tape[origin].Missing,'missing label delivered');
          assert.equal(origin+c.delay,clock,'wrong delivery clock');
          assert.ok(!delivered.has(origin),'duplicate label');
          delivered.add(origin); arrived++;
        }
      }
      for (const origin of selected) {
        if (!row.Tape[origin].Missing&&origin+c.delay<512) assert.ok(delivered.has(origin));
      }
      assert.equal(nominated,128);
      assert.equal(arm.Nominated,nominated);
      assert.equal(arm.NominatedPost,nominatedPost);
      assert.equal(arm.CorrectPost,correctPost);
      assert.equal(arm.Arrived,arrived);
      assert.equal(arm.Pending,pending);
      near(full,arm.FullBrier,`${id}/${index}/full`);
      near(post256,arm.PostBrier,`${id}/${index}/post256`);
      near(early256,arm.EarlyBrier,`${id}/${index}/early256`);
      near(fullExpected,wrapped.FullExpected,`${id}/${index}/fullExpected`);
      near(postExpected,wrapped.PostExpected,`${id}/${index}/postExpected`);
      near(firstExpected,wrapped.FirstExpected,`${id}/${index}/firstExpected`);
      near(postRealized,wrapped.PostRealized,`${id}/${index}/postRealized`);
      const recovered=recovery(expected,c);
      assert.equal(wrapped.RecoveryDelay,recovered.delay);
      assert.equal(wrapped.MissedRecovery,recovered.missed);
      summary.Arms.push({FullExpected:fullExpected,PostExpected:postExpected,
        FirstExpected:firstExpected,PostRealized:postRealized,
        RecoveryDelay:recovered.delay,MissedRecovery:Number(recovered.missed),Arrived:arrived});
    }
    byCase.get(row.Scenario).push(summary); count++;
  }
  assert.ok(manifest);
  assert.equal(count,1792);
  const report={trajectories:count,cases:{},pass:true};
  let strongKnown=0;
  for (const [name,c] of Object.entries(cases)) {
    const rows=byCase.get(name);
    assert.equal(rows.length,256);
    const mean=field=>[0,1,2].map(a=>avg(rows.map(r=>r.Arms[a][field])));
    const full=mean('FullExpected'),post=mean('PostExpected'),first=mean('FirstExpected');
    const realized=mean('PostRealized'),rec=mean('RecoveryDelay');
    const missed=mean('MissedRecovery'),arrived=mean('Arrived');
    const skewRate=avg(rows.map(r=>r.SkewRate));
    const gains=[0,1].map(i=>interval(rows,2,i,'PostExpected'));
    const fullGain=[0,1].map(i=>interval(rows,2,i,'FullExpected'));
    const checks={arrival:Math.abs(arrived[2]-arrived[0])<=2&&Math.abs(arrived[2]-arrived[1])<=2};
    if (c.known) {
      checks.nonharm=gains.every(g=>g.low>-.01);
      checks.strong=gains.every(g=>g.mean>=.01&&g.low>0);
      if (checks.strong) strongKnown++;
      if (c.noise===.05) {
        checks.early=[0,1].every(i=>first[2]-first[i]<=.005);
        if (c.delay===0) checks.recovery=[0,1].every(i=>rec[i]-rec[2]>=10);
      }
    }
    if (c.skew) checks.skew=skewRate>=.08&&skewRate<=.12;
    if (c.stable) checks.stable=fullGain.every(g=>g.low>-.01);
    if (c.ood) checks.ood=gains.every(g=>g.low>-.01);
    const pass=Object.entries(checks).filter(([k])=>k!=='strong').every(([,v])=>v);
    report.cases[name]={full,post,first,realized,recovery:rec,missed,arrived,skewRate,
      gains,fullGain,checks,pass};
    report.pass&&=pass;
  }
  report.strongKnown=strongKnown;
  report.pass&&=strongKnown>=4;
  result.splits[split]=report;
}
result.transferPass=result.splits.design.pass&&result.splits.confirmation.pass;
console.log(JSON.stringify(result,null,2));
