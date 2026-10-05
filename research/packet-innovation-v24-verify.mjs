import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const hash = bytes => createHash('sha256').update(bytes).digest('hex');
const near = (a, b, message) => assert.ok(Number.isFinite(a) && Math.abs(a-b) < 1e-11, `${message}: ${a} != ${b}`);
const clamp = x => Math.min(1, Math.max(0, x));
const risk = (q, p) => p*(1-q)**2+(1-p)*q*q;
const average = xs => xs.reduce((a, b) => a+b, 0)/xs.length;
function interval(xs) {
  const mean = average(xs);
  const se = Math.sqrt(xs.reduce((sum, v) => sum+(v-mean)**2, 0)/(xs.length-1)/xs.length);
  return {mean, lower:mean-3.5*se, upper:mean+3.5*se};
}
const regimes = ['independent', 'aligned', 'reversed'];
const names = ['baseline', 'current', 'flat', 'context'];
const field = {baseline:'Base', current:'Current', flat:'Flat', context:'Context'};
function verify(split) {
  const path = resolve(root, `docs/experiments/mmm-packet-innovation-v24-${split}.jsonl`);
  const raw = readFileSync(path);
  const lines = raw.toString('utf8').trim().split('\n').map(line => JSON.parse(line));
  assert.equal(lines.length, 25);
  const [manifest, ...rows] = lines;
  assert.equal(manifest.kind, 'manifest');
  assert.equal(manifest.split, split);
  assert.equal(manifest.worldsPerRegime, 8);
  assert.equal(manifest.seedBase, split === 'design' ? 2026102403 : 2026102404);
  assert.equal(Object.keys(manifest.hashes).length, 13);
  for (const [name, expected] of Object.entries(manifest.hashes)) {
    assert.equal(hash(readFileSync(resolve(root, name))), expected, `source drift: ${name}`);
  }
  const result = {split, sha256:hash(raw), worlds:24, labels:768, regimes:{}, maxRecallMS:0, maxOutcomeMS:0};
  for (let ri=0; ri<regimes.length; ri++) {
    const regime = regimes[ri];
    const worlds = rows.slice(ri*8, ri*8+8);
    const utilities = Object.fromEntries(names.map(n => [n, []]));
    const packetBriers = Object.fromEntries(names.map(n => [n, []]));
    const falseItems = Object.fromEntries(names.map(n => [n, []]));
    const baseRisks=[], learnedRisks=[], positives=[];
    for (let wi=0; wi<8; wi++) {
      const row = worlds[wi];
      assert.equal(row.Kind, 'trial');
      assert.equal(row.Split, split);
      assert.equal(row.Regime, regime);
      assert.equal(row.World, wi);
      assert.equal(row.Seed, manifest.seedBase+ri*1000000+wi*1000);
      assert.equal(row.InitialEpoch, 217);
      assert.equal(row.LearnedEpoch, row.InitialEpoch);
      assert.equal(row.InitialCertified, true);
      assert.equal(row.LearnedCertified, true);
      assert.equal(row.Candidates.length, 150);
      assert.equal(new Set(row.Candidates.map(c=>c.ID)).size, 150);
      assert.equal(row.Candidates.filter(c=>c.Monitored).length, 32);
      assert.equal(row.Candidates.filter(c=>c.Belief).length, 32);
      assert.equal(row.RecallNS.length, 2);
      assert.equal(row.OutcomeNS.length, 32);
      assert.ok(row.ScoreNS >= 0);
      result.maxRecallMS = Math.max(result.maxRecallMS, ...row.RecallNS.map(x=>x/1e6));
      result.maxOutcomeMS = Math.max(result.maxOutcomeMS, ...row.OutcomeNS.map(x=>x/1e6));
      let high=0;
      const preRankLaw=c=>c.Monitored ? clamp(c.Base+.1*(c.Alpha/(c.Alpha+c.Beta)-c.Base)) : c.Base;
      const preRank=[...row.Candidates].sort((a,b)=>preRankLaw(b)-preRankLaw(a));
      const inside=preRank[9].Base, outside=preRank[10].Base;
      const certainty=clamp((inside-outside)/Math.max(Math.abs(inside),Math.abs(outside),1e-12));
      const elasticScale=.5+2*(1-certainty);
      for (const [i,c] of row.Candidates.entries()) {
        assert.ok(!c.ID.startsWith('future'));
        let angle=0;
        if(c.ID.startsWith('eligible-')) { angle=.005*(Number(c.ID.slice(9))+1); }
        else assert.ok(c.ID==='past100' || c.ID==='at120');
        const age=c.ID==='at120' ? 4 : 5;
        const expectedBase=.65*(Math.cos(angle)+1)/2+.15+.1*Math.exp(-age/(3600*24*30))+.025;
        assert.ok(Math.abs(c.Base-expectedBase)<1e-6, 'unit-cosine service baseline');
        assert.equal(c.Monitored, i<32);
        assert.equal(c.Belief, c.Monitored);
        for(const k of ['Probability','Base','Current','Flat','Context','Law']) assert.ok(c[k]>=0 && c[k]<=1, k);
        if (regime==='independent') {
          assert.ok(c.Probability===.2 || c.Probability===.8);
          if(c.Probability===.8) high++;
        } else near(c.Probability, regime==='aligned' ? .9-.8*i/149 : .1+.8*i/149, 'regime truth');
        let current=c.Base, flat=c.Base, contextual=c.Base, law=c.Base;
        if(c.Monitored) {
          assert.equal(c.Alpha, c.Useful ? 2 : 1);
          assert.equal(c.Beta, c.Useful ? 1 : 2);
          const posterior=c.Alpha/(c.Alpha+c.Beta);
          law=clamp(c.Base+.1*(posterior-c.Base));
          current=clamp(c.Base+(law-c.Base)*elasticScale);
          flat=clamp(c.Base+.1*(posterior-.5));
          const contextMean=(2*c.Base+c.Alpha-1)/(c.Alpha+c.Beta);
          contextual=clamp(c.Base+.1*(contextMean-c.Base));
        } else {
          assert.equal(c.Alpha, 0); assert.equal(c.Beta, 0); assert.equal(c.Useful, false);
        }
        near(c.Current,current,'actual current rank');
        near(c.Flat,flat,'flat score');
        near(c.Context,contextual,'context score');
        near(c.Law,law,'actual unmodified law');
      }
      if(regime==='independent') assert.equal(high,75);
      const byID=new Map(row.Candidates.map(c=>[c.ID,c]));
      const baseRisk=average(row.Candidates.map(c=>risk(c.Base,c.Probability)));
      const learnedRisk=average(row.Candidates.map(c=>risk(c.Law,c.Probability)));
      near(row.PopulationBrier,learnedRisk,'whole nominated population risk');
      baseRisks.push(baseRisk); learnedRisks.push(learnedRisk);
      positives.push(row.Candidates.filter(c=>c.Monitored&&c.Useful).length);
      assert.equal(row.Controls.length,4);
      for (const [ni,name] of names.entries()) {
        const control=row.Controls[ni];
        assert.equal(control.Name,name);
        // Stable sort follows the frozen journal order; fixture events have
        // distinct provenance, no correlated occupancy and no token overflow.
        const ordered=[...row.Candidates].sort((a,b)=>b[field[name]]-a[field[name]]);
        const ids=ordered.slice(0,50).slice(0,10).map(c=>c.ID);
        assert.deepEqual(control.IDs,ids,'reconstructed packet');
        assert.equal(new Set(ids).size,10);
        const utility=average(ids.map(id=>byID.get(id).Probability));
        const falseCount=ids.reduce((sum,id)=>sum+1-byID.get(id).Probability,0);
        const brier=average(ids.map(id=>risk(byID.get(id).Law,byID.get(id).Probability)));
        near(control.Utility,utility,'packet utility');
        near(control.FalseItems,falseCount,'false items');
        near(control.PacketBrier,brier,'packet risk');
        if(name==='current') assert.deepEqual(ids,row.ActualPacket,'actual service packet');
        utilities[name].push(utility); falseItems[name].push(falseCount); packetBriers[name].push(brier);
      }
      assert.equal(row.InitialPacket.length,10);
      assert.deepEqual(row.InitialPacket, [...row.Candidates].sort((a,b)=>b.Base-a.Base).slice(0,10).map(c=>c.ID), 'initial baseline packet');
    }
    const controls={};
    for(const name of names) {
      const gain=interval(utilities[name].map((v,i)=>v-utilities.baseline[i]));
      const harm=interval(utilities[name].map((v,i)=>utilities.baseline[i]-v));
      controls[name]={utility:average(utilities[name]), expectedFalseItems:average(falseItems[name]), packetBrier:average(packetBriers[name]), gain, harm};
    }
    result.regimes[regime]={positiveLabels:positives.reduce((a,b)=>a+b,0), baseBrier:average(baseRisks), learnedBrier:average(learnedRisks), controls};
  }
  result.proposals={};
  for(const name of ['flat','context']) {
    const independent=result.regimes.independent.controls[name].gain;
    const reversed=result.regimes.reversed.controls[name].gain;
    const aligned=result.regimes.aligned.controls[name].harm;
    result.proposals[name]={pass:independent.mean>=.02 && independent.lower>0 && reversed.mean>=.02 && reversed.lower>0 && aligned.upper<=.01};
  }
  return result;
}
const result={design:verify('design'), confirmation:verify('confirmation')};
const outIndex=process.argv.indexOf('--out');
if(outIndex>=0) {
  assert.ok(process.argv[outIndex+1]);
  writeFileSync(resolve(root,process.argv[outIndex+1]),JSON.stringify(result,null,2)+'\n',{flag:'wx'});
}
process.stdout.write(JSON.stringify(result,null,2)+'\n');
