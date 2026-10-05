// Bulk mechanical harness extension; no V51 artifact or source is modified.
import fs from 'node:fs';import assert from 'node:assert/strict';
function replace(s,a,b){assert(s.includes(a),a);return s.replaceAll(a,()=>b);}
function save(p,s){assert(!fs.existsSync('research/scored-v52-diagnostic'),'cannot regenerate after freeze');if(fs.existsSync(p)){assert(['research/scored-v52-diagnostic.mjs','research/scored-v52-readback.mjs'].includes(p));fs.writeFileSync(p,s,{mode:0o600});}else fs.writeFileSync(p,s,{flag:'wx',mode:0o600});}
let s=fs.readFileSync('research/scored-v51-diagnostic.mjs','utf8');
s=replace(s,"const root = 'research/scored-v51-diagnostic';","const root = 'research/scored-v52-diagnostic';");
s=replace(s,"'docs/experiments/mmm-scored-v51-protocol.md', 'go.mod'","'docs/experiments/mmm-scored-v52-protocol.md', 'research/scored-v52-gen.mjs', 'research/scored-v52-run-gen.mjs', 'research/scored-v52-generation.json', 'research/scored-v52-preflight-failure.md', 'research/scored-v52-diagnostic.mjs', 'research/scored-v52-readback.mjs', 'go.mod'");
s=replace(s,'2026105107','2026105207');s=replace(s,'2026105109','2026105209');s=replace(s,'2026105111','2026105211');
s=replace(s,'SCORED_V51','SCORED_V52');
s=replace(s,"'^(TestScoredV51(DenseDelayedReference|LegacyExactAndCancellation|LifecycleAndFutureFork|HybridLegacyExactAndFencing|StreamIdentity)|TestHybridV48(CompactEquivalent|CompactExtremeRevival|DelayedHeadsAndOriginalServedLaw|OwnershipCapsAndGlobalFence))$'","'^(TestScoredV51(DenseDelayedReference|LegacyExactAndCancellation|LifecycleAndFutureFork)|TestScoredV52(HybridLegacyExactAndFencing|AllHeadDenseDelayedReference|OwnershipCapsAndLocalClock|StreamIdentity)|TestHybridV48(CompactEquivalent|CompactExtremeRevival|DelayedHeadsAndOriginalServedLaw|OwnershipCapsAndGlobalFence))$'");
for(const name of ['SeedSeparation','FixtureChecks','Fixture','Allocation','StudyAudit','Study'])s=replace(s,'TestScoredV51'+name,'TestScoredV52'+name);
s=replace(s,"  unchanged(); const artifacts = {};",`  // Same fresh worlds compare against the earlier OUTER-only intervention.
  for (const style of styles.filter(s => s !== 'legacy')) {
    const raw = path.resolve(root + '/outer-' + style + '.jsonl');
    const env = { EVENTFRAME_SCORED_V51_FIXTURE: fixture, EVENTFRAME_SCORED_V51_STYLE: style };
    await run('outer-' + style, ['test', './internal/researchswitch', '-run', '^TestScoredV51Study$', '-v', '-count=1', '-timeout=30m'], { ...env, EVENTFRAME_SCORED_V51_OUT: raw });
    await run('outer-' + style + '-audit', ['test', './internal/researchswitch', '-run', '^TestScoredV51StudyAudit$', '-v', '-count=1', '-timeout=30m'], { ...env, EVENTFRAME_SCORED_V51_AUDIT: raw });
  }
  unchanged(); const artifacts = {};`);
s=replace(s,'const checks = [];',`const generation = JSON.parse(fs.readFileSync('research/scored-v52-generation.json'));
for (const [p,h] of Object.entries(generation.sources)) assert.equal(hash(fs.readFileSync(p)),h);
for (const [p,h] of Object.entries(generation.outputs)) assert.equal(hash(fs.readFileSync(p)),h);
const checks = [];`);
save('research/scored-v52-diagnostic.mjs',s);
s=fs.readFileSync('research/scored-v51-readback.mjs','utf8');
s=replace(s,'scored-v51-diagnostic','scored-v52-diagnostic');s=replace(s,'done.checks.length === 16','done.checks.length === 24');s=replace(s,'2026105107','2026105207');
s=replace(s,"const streams = Object.fromEntries(freeze.styles.map(style => [style, lines(style + '.jsonl')[Symbol.asyncIterator]() ]));",`const allStyles = [...freeze.styles, ...freeze.styles.filter(s=>s!=='legacy').map(s=>'outer-'+s)];
const streams = Object.fromEntries(allStyles.map(style => [style, lines(style + '.jsonl')[Symbol.asyncIterator]() ]));`);
s=replace(s,'assert.equal(rm.Style, style);','assert.equal(rm.Style, style.replace(/^outer-/,\'\'));');
s=replace(s,'freeze.styles.map(s => [s, {}])','allStyles.map(s => [s, {}])');s=replace(s,'freeze.styles.map(s => [s, 0])','allStyles.map(s => [s, 0])');
s=replace(s,'for (const style of freeze.styles)','for (const style of allStyles)');
s=replace(s,"assert.deepEqual(a.Advice, raw.legacy.Arms[j].Advice); assert.deepEqual(a.Heads, raw.legacy.Arms[j].Heads);","assert.deepEqual(a.Advice, raw.legacy.Arms[j].Advice); if (style.startsWith('outer-')) assert.deepEqual(a.Heads, raw.legacy.Arms[j].Heads);");
s=replace(s,"for (const name of ['Advice', 'Heads', 'Weights', 'GlobalWeights']) assert.deepEqual(v[name], b[name]);","assert.deepEqual(v.Advice, b.Advice); if (style.startsWith('outer-')) for (const name of ['Heads', 'Weights', 'GlobalWeights']) assert.deepEqual(v[name],b[name]);");
s=replace(s,'sharedAdvicePairs, 1008','sharedAdvicePairs, 2016');
s=replace(s,'candidateArms: worlds * 9 * 4','candidateArms: worlds * 9 * 4, outerOnlyComparisonArms: worlds * 9 * 4');
s=replace(s,'constructorBytes: allocation[style], allocationScreenPass: allocation[style] <= 8 << 20','constructorBytes: allocation[style] ?? null, allocationScreenPass: allocation[style] == null ? null : allocation[style] <= 8 << 20');
s=replace(s,'legacyIssuedBrierWins: rows.filter',`recoveryChangedVsLegacy: rows.filter(v=>v.gains.legacy.Recovery!==0).length,
      finalUsefulnessChangedVsLegacy: rows.filter(v=>v.gains.legacy.FinalUsefulness!==0).length,
      legacyIssuedBrierWins: rows.filter`);
s=replace(s,'const out = {',`const allHeadVsOuter={};
for (const style of freeze.styles.filter(s=>s!=='legacy')) {
 allHeadVsOuter[style]={};
 for (const mode of policies) {
  const rows=Object.keys(groups[style]).map(key=>({key,all:groups[style][key][mode].candidate,outer:groups['outer-'+style][key][mode].candidate}));
  allHeadVsOuter[style][mode]={};
  for (const k of Object.keys(rows[0].all)) {
   const gains=rows.map(v=>(v.outer[k]-v.all[k])*(k==='FinalUsefulness'?-1:1));
   allHeadVsOuter[style][mode][k]={meanGain:mean(gains),wins:gains.filter(v=>v>0).length,losses:gains.filter(v=>v<0).length,ties:gains.filter(v=>v===0).length};
  }
 }
}
const out = { allHeadVsOuter,`);
save('research/scored-v52-readback.mjs',s);
