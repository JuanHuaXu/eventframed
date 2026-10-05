import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const hash=bytes=>createHash('sha256').update(bytes).digest('hex');
const clamp=x=>Math.min(1,Math.max(0,x));
const average=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const risk=(q,p)=>p*(1-q)**2+(1-p)*q*q;
const near=(a,b,why)=>assert.ok(Number.isFinite(a)&&Math.abs(a-b)<1e-11,`${why}: ${a} != ${b}`);
const geometries=[['tight',.005],['wide',.020]];
const regimes=['independent','aligned','reversed','calibrated'];
const names=['baseline','served','flat','context'];
const fields={baseline:'Base',served:'Served',flat:'Flat',context:'Context'};
function interval(xs) {
  const mean=average(xs);
  const se=Math.sqrt(xs.reduce((sum,x)=>sum+(x-mean)**2,0)/(xs.length-1)/xs.length);
  return {mean,lower:mean-3.5*se,upper:mean+3.5*se};
}
function verify(split) {
  const raw=readFileSync(resolve(root,`docs/experiments/mmm-packet-prior-v26-${split}.jsonl`));
  const [manifest,...rows]=raw.toString('utf8').trim().split('\n').map(JSON.parse);
  assert.equal(rows.length,256);
  assert.equal(manifest.kind,'manifest');
  assert.equal(manifest.split,split);
  assert.equal(manifest.worldsPerGeometryRegime,32);
  assert.deepEqual(manifest.geometries,['tight','wide']);
  assert.equal(manifest.seedBase,split==='design'?2026102603:2026102604);
  assert.equal(Object.keys(manifest.hashes).length,19);
  for(const [name,expected] of Object.entries(manifest.hashes)) assert.equal(hash(readFileSync(resolve(root,name))),expected,`source drift: ${name}`);
  const result={split,sha256:hash(raw),worlds:256,labels:8192,geometries:{},maxRecallMS:0,maxOutcomeMS:0};
  for(const [gi,[geometry,step]] of geometries.entries()) {
    const grouped={};
    for(const [ri,regime] of regimes.entries()) {
      const worlds=rows.slice(gi*128+ri*32,gi*128+ri*32+32);
      const utilities=Object.fromEntries(names.map(name=>[name,[]]));
      const packetBriers=Object.fromEntries(names.map(name=>[name,[]]));
      const falseItems=Object.fromEntries(names.map(name=>[name,[]]));
      const lawRisks=Object.fromEntries(names.map(name=>[name,[]]));
      let positiveLabels=0, differentProposalPackets=0, servedMonitoredPacked=0;
      for(const [wi,row] of worlds.entries()) {
        assert.equal(row.Kind,'trial');
        assert.equal(row.Split,split); assert.equal(row.Geometry,geometry); assert.equal(row.Regime,regime);
        assert.equal(row.AngleStep,step); assert.equal(row.World,wi);
        assert.equal(row.Seed,manifest.seedBase+gi*10000000+ri*1000000+wi*1000);
        assert.equal(row.InitialEpoch,217); assert.equal(row.LearnedEpoch,row.InitialEpoch);
        assert.equal(row.InitialCertified,true); assert.equal(row.LearnedCertified,true);
        assert.equal(row.Candidates.length,150);
        assert.equal(new Set(row.Candidates.map(c=>c.ID)).size,150);
        assert.equal(row.Candidates.filter(c=>c.Monitored).length,32);
        assert.equal(row.Candidates.filter(c=>c.Belief).length,32);
        assert.equal(row.RecallNS.length,2); assert.equal(row.OutcomeNS.length,32);
        result.maxRecallMS=Math.max(result.maxRecallMS,...row.RecallNS.map(x=>x/1e6));
        result.maxOutcomeMS=Math.max(result.maxOutcomeMS,...row.OutcomeNS.map(x=>x/1e6));
        const actualLaw=c=>c.Monitored ? (2*c.PriorBase+c.Alpha-1)/(c.Alpha+c.Beta) : c.Base;
        let high=0;
        for(const [i,c] of row.Candidates.entries()) {
          assert.ok(!c.ID.startsWith('future'));
          assert.equal(c.Monitored,i<32); assert.equal(c.Belief,c.Monitored);
          for(const field of ['Probability','PriorBase','Base','Served','Flat','Context','Law','FlatLaw','ContextLaw']) assert.ok(c[field]>=0 && c[field]<=1,field);
          let angle=0;
          if(c.ID.startsWith('eligible-')) angle=step*(Number(c.ID.slice(9))+1);
          else assert.ok(c.ID==='past100'||c.ID==='at120');
          const afterAge=c.ID==='at120'?4:5;
          const priorAge=c.ID==='at120'?1:2;
          const oracleBase=age=>.65*(Math.cos(angle)+1)/2+.15+.1*Math.exp(-age/(3600*24*30))+.025;
          assert.ok(Math.abs(c.Base-oracleBase(afterAge))<1e-6,'unit-cosine baseline');
          assert.ok(Math.abs(c.PriorBase-oracleBase(priorAge))<1e-6,'committed prior baseline');
          if(regime==='independent') {
            assert.ok(c.Probability===.2||c.Probability===.8);
            if(c.Probability===.8) high++;
          } else if(regime==='aligned') near(c.Probability,.9-.8*i/149,'aligned law');
          else if(regime==='reversed') near(c.Probability,.1+.8*i/149,'reversed law');
          else near(c.Probability,c.PriorBase,'calibrated law');
          let flatLaw=c.Base,contextLaw=c.Base,served=c.Base,flatRank=c.Base,contextRank=c.Base;
          if(c.Monitored) {
            if(c.Useful) positiveLabels++;
            const successes=c.Useful?1:0;
            assert.equal(c.Alpha,1+successes); assert.equal(c.Beta,2-successes);
            flatLaw=(1+successes)/3;
            contextLaw=(2*c.PriorBase+successes)/3;
            const law=clamp(c.Base+.1*(flatLaw-c.Base));
            served=contextLaw;
            flatRank=clamp(c.Base+.1*(flatLaw-.5));
            contextRank=clamp(c.Base+.1*(contextLaw-c.PriorBase));
          } else {
            assert.equal(c.Alpha,0); assert.equal(c.Beta,0); assert.equal(c.Useful,false);
          }
          near(c.Law,actualLaw(c),'actual scored law');
          near(c.Law,c.ContextLaw,'served/context joint predictive identity');
          near(c.Served,served,'actual rank');
          near(c.FlatLaw,flatLaw,'flat joint predictive');
          near(c.ContextLaw,contextLaw,'context joint predictive');
          near(c.Flat,flatRank,'flat innovation');
          near(c.Context,contextRank,'context innovation');
        }
        if(regime==='independent') assert.equal(high,75);
        const baseRisk=average(row.Candidates.map(c=>risk(c.Base,c.Probability)));
        const servedRisk=average(row.Candidates.map(c=>risk(c.Law,c.Probability)));
        const flatRisk=average(row.Candidates.map(c=>risk(c.FlatLaw,c.Probability)));
        const contextRisk=average(row.Candidates.map(c=>risk(c.ContextLaw,c.Probability)));
        near(row.PopulationBaseBrier,baseRisk,'baseline population risk');
        near(row.PopulationBrier,servedRisk,'actual population risk');
        near(row.PopulationFlatBrier,flatRisk,'flat population risk');
        near(row.PopulationContextBrier,contextRisk,'context population risk');
        lawRisks.baseline.push(baseRisk); lawRisks.served.push(servedRisk); lawRisks.flat.push(flatRisk); lawRisks.context.push(contextRisk);
        const byID=new Map(row.Candidates.map(c=>[c.ID,c]));
        assert.equal(row.Controls.length,4);
        for(const [ni,name] of names.entries()) {
          const control=row.Controls[ni]; assert.equal(control.Name,name);
          const ids=[...row.Candidates].sort((a,b)=>b[fields[name]]-a[fields[name]]).slice(0,50).slice(0,10).map(c=>c.ID);
          assert.deepEqual(control.IDs,ids,'packet reconstruction'); assert.equal(new Set(ids).size,10);
          const utility=average(ids.map(id=>byID.get(id).Probability));
          const falseCount=ids.reduce((sum,id)=>sum+1-byID.get(id).Probability,0);
          const packetRisk=average(ids.map(id=>risk(byID.get(id).Law,byID.get(id).Probability)));
          near(control.Utility,utility,'packet utility'); near(control.FalseItems,falseCount,'false items'); near(control.PacketBrier,packetRisk,'packet Brier');
          if(name==='served') { assert.deepEqual(row.ActualPacket,ids); servedMonitoredPacked+=ids.filter(id=>byID.get(id).Monitored).length; }
          utilities[name].push(utility); falseItems[name].push(falseCount); packetBriers[name].push(packetRisk);
        }
        if(JSON.stringify(row.Controls[2].IDs)!==JSON.stringify(row.Controls[3].IDs)) differentProposalPackets++;
        assert.deepEqual(row.InitialPacket,[...row.Candidates].sort((a,b)=>b.PriorBase-a.PriorBase).slice(0,10).map(c=>c.ID),'initial packet');
      }
      const ranks={},laws={};
      for(const name of names) {
        ranks[name]={utility:average(utilities[name]),falseItems:average(falseItems[name]),packetBrier:average(packetBriers[name]),gain:interval(utilities[name].map((x,i)=>x-utilities.baseline[i])),harm:interval(utilities[name].map((x,i)=>utilities.baseline[i]-x))};
        laws[name]={brier:average(lawRisks[name]),gain:interval(lawRisks[name].map((x,i)=>lawRisks.baseline[i]-x)),harm:interval(lawRisks[name].map((x,i)=>x-lawRisks.baseline[i]))};
      }
      grouped[regime]={worlds:32,positiveLabels,differentProposalPackets,servedMonitoredPacked,ranks,laws};
    }
    const rankScreens={},forecastScreens={};
    for(const name of ['served','flat','context']) {
      const rankGain=['independent','reversed'].every(r=>grouped[r].ranks[name].gain.mean>=.02 && grouped[r].ranks[name].gain.lower>0);
      const rankSafety=['aligned','calibrated'].every(r=>grouped[r].ranks[name].harm.upper<=.01);
      const lawGain=['independent','reversed'].every(r=>grouped[r].laws[name].gain.mean>=.005 && grouped[r].laws[name].gain.lower>0);
      const lawSafety=['aligned','calibrated'].every(r=>grouped[r].laws[name].harm.upper<=.01);
      rankScreens[name]={pass:rankGain&&rankSafety,gainPass:rankGain,safetyPass:rankSafety};
      forecastScreens[name]={pass:lawGain&&lawSafety,gainPass:lawGain,safetyPass:lawSafety};
    }
    result.geometries[geometry]={regimes:grouped,rankScreens,forecastScreens};
  }
  return result;
}
const result={design:verify('design'),confirmation:verify('confirmation')};
const outIndex=process.argv.indexOf('--out');
if(outIndex>=0) writeFileSync(resolve(root,process.argv[outIndex+1]),JSON.stringify(result,null,2)+'\n',{flag:'wx'});
for(const [split,data] of Object.entries(result)) {
  for(const [geometry,metrics] of Object.entries(data.geometries)) {
    process.stdout.write(JSON.stringify({split,geometry,rankScreens:metrics.rankScreens,forecastScreens:metrics.forecastScreens,proposalPacketsDiffer:regimes.reduce((n,r)=>n+metrics.regimes[r].differentProposalPackets,0),alignedRankHarmUpper:Object.fromEntries(['flat','context'].map(n=>[n,metrics.regimes.aligned.ranks[n].harm.upper])),calibratedForecastHarmUpper:Object.fromEntries(['flat','context'].map(n=>[n,metrics.regimes.calibrated.laws[n].harm.upper]))})+'\n');
  }
}
