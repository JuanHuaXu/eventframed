// Postcollection phase accounting and matched V71 cost comparison, not profiling.
import fs from 'node:fs';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/dynvarcache-v72-diagnostic',source='research/dynvarcache-v72-cost-audit.mjs',prior='research/dynvariance-v71-diagnostic';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const done=JSON.parse(fs.readFileSync(root+'/completed.json')),read=JSON.parse(fs.readFileSync(root+'/readback.json')),old=JSON.parse(fs.readFileSync(prior+'/readback.json'));
assert(done.allJobsTerminal&&done.checks.length===8&&done.checks.every(x=>x.exitCode===0));
assert.equal(await hash(root+'/diagnostic.jsonl'),done.artifacts['diagnostic.jsonl']);assert.equal(read.allArmsBitwiseEqual,2160);
const phases=['ScheduleNS','SetupNS','IssueNS','FirstResolveNS','ProposalNS','RequestNS','SecondResolveNS','SnapshotNS'];
const totals={};let arms=0;
for await(const line of readline.createInterface({input:fs.createReadStream(root+'/diagnostic.jsonl'),crlfDelay:Infinity})){
 const w=JSON.parse(line);if(!w.Arms)continue;
 for(const a of w.Arms){
  if(a.Mode==='full'||a.Mode==='adaptive')continue;
  arms++;const c=a.Breakdown,t=totals[a.Mode]??={arms:0,elapsedNS:0,accountedNS:0,phases:Object.fromEntries(phases.map(p=>[p,0])),worst:null};
  assert.equal(c.AccountedNS,phases.reduce((s,p)=>s+c[p],0));assert(c.AccountedNS<=c.ElapsedNS);t.arms++;t.elapsedNS+=c.ElapsedNS;t.accountedNS+=c.AccountedNS;
  for(const p of phases)t.phases[p]+=c[p];
  if(!t.worst||c.ElapsedNS>t.worst.elapsedNS)t.worst={geometry:w.Population.World.Geometry,regime:w.Population.World.Regime,schedule:a.Schedule,elapsedNS:c.ElapsedNS,phases:Object.fromEntries(phases.map(p=>[p,c[p]]))};
 }
}
assert.equal(arms,1920);
for(const [mode,t]of Object.entries(totals)){
 assert.equal(t.arms,120);assert.equal(read.totals[mode].risk,old.totals[mode].risk);assert.equal(read.totals[mode].terminalRisk,old.totals[mode].terminalRisk);
 t.phaseFractions=Object.fromEntries(phases.map(p=>[p,t.phases[p]/t.elapsedNS]));t.unattributedFraction=1-t.accountedNS/t.elapsedNS;t.meanElapsedMS=t.elapsedNS/t.arms/1e6;
 t.v71MeanElapsedMS=old.totals[mode].totalMS/120;t.v71WorstMS=old.totals[mode].maxMS;t.coreCostRatioToV71=t.meanElapsedMS/t.v71MeanElapsedMS;
 t.runtimeGate=read.summaries[mode].gates.completeLoop400;t.scientificGatesUnchanged=['gain01','noAdaptiveHarm01','recovery'].every(g=>read.summaries[mode].gates[g]===old.summaries[mode].gates[g]);assert(t.scientificGatesUnchanged);
}
const comparison={};for(const against of ['learn_random','learn_uncertainty'])comparison['learn_falsification/'+against]={riskGain:read.totals[against].risk-read.totals.learn_falsification.risk,coreCostRatio:read.totals.learn_falsification.totalMS/read.totals[against].totalMS,equalRequests:true,equalTotalCost:false};
const result={source,sourceSHA256:await hash(source),dataSHA256:done.artifacts['diagnostic.jsonl'],priorDataSHA256:await hash(prior+'/diagnostic.jsonl'),scope:'matched consumed-cohort phase elapsed accounting; all scientific outputs bitwise; different serial runs are not a randomized causal profiling comparison',modelArms:arms,totals,comparison,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/cost-audit.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({totals:Object.fromEntries(Object.entries(totals).map(([m,t])=>[m,{meanElapsedMS:t.meanElapsedMS,v71MeanElapsedMS:t.v71MeanElapsedMS,worstMS:t.worst.elapsedNS/1e6,v71WorstMS:t.v71WorstMS,ratio:t.coreCostRatioToV71,runtimeGate:t.runtimeGate,phaseFractions:t.phaseFractions}])),comparison},null,2));
