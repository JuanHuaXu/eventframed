// Postcollection accounting; never reinterpret request equality as cost equality.
import fs from 'node:fs';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/mean-anchor-v75-diagnostic',source='research/mean-anchor-v75-cohort-cost-audit.mjs';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const done=JSON.parse(fs.readFileSync(root+'/completed.json')),read=JSON.parse(fs.readFileSync(root+'/readback.json'));
assert(done.allJobsTerminal&&done.checks.length===4&&done.checks.every(x=>x.exitCode===0));assert.equal(await hash(root+'/diagnostic.jsonl'),done.artifacts['diagnostic.jsonl']);assert.equal(read.newModelArms,1440);assert.equal(read.controlsBitwiseEqual,2160);
const phases=['ScheduleNS','SetupNS','IssueNS','FirstResolveNS','ProposalNS','RequestNS','SecondResolveNS','SnapshotNS'],totals={};let arms=0;
for await(const line of readline.createInterface({input:fs.createReadStream(root+'/diagnostic.jsonl'),crlfDelay:Infinity})){
 const w=JSON.parse(line);if(!w.Arms)continue;
 for(const a of w.Arms){
  if(!a.Mode.startsWith('mean'))continue;
  arms++;const c=a.Breakdown,t=totals[a.Mode]??={arms:0,elapsedNS:0,accountedNS:0,phases:Object.fromEntries(phases.map(p=>[p,0])),worst:null};
  assert.equal(c.AccountedNS,phases.reduce((s,p)=>s+c[p],0));assert(c.AccountedNS<=c.ElapsedNS);t.arms++;t.elapsedNS+=c.ElapsedNS;t.accountedNS+=c.AccountedNS;for(const p of phases)t.phases[p]+=c[p];
  if(!t.worst||c.ElapsedNS>t.worst.elapsedNS)t.worst={geometry:w.Population.World.Geometry,regime:w.Population.World.Regime,schedule:a.Schedule,elapsedNS:c.ElapsedNS,phases:Object.fromEntries(phases.map(p=>[p,c[p]]))};
 }
}
assert.equal(arms,1440);
for(const [mode,t]of Object.entries(totals)){assert.equal(t.arms,120);t.phaseFractions=Object.fromEntries(phases.map(p=>[p,t.phases[p]/t.elapsedNS]));t.unattributedFraction=1-t.accountedNS/t.elapsedNS;t.meanElapsedMS=t.elapsedNS/t.arms/1e6;t.gates=read.summaries[mode].gates;t.over400MS=read.summaries[mode].over400MS}
const paired={};for(const config of ['mean','meanlocal','meanindividual'])for(const policy of ['random','uncertainty']){const candidate=config+'_falsification',against=config+'_'+policy;paired[candidate+'/'+against]={riskGain:read.totals[against].risk-read.totals[candidate].risk,coreCostRatio:read.totals[candidate].totalMS/read.totals[against].totalMS,equalRequests:true,equalTotalCost:false}}
const result={source,sourceSHA256:await hash(source),dataSHA256:done.artifacts['diagnostic.jsonl'],newModelArms:arms,totals,paired,scope:'phase elapsed accounting on complete consumed screen, not causal profiling, fresh confirmation, loaded serving or full new-arm independent replay',equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/cost-audit.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({totals:Object.fromEntries(Object.entries(totals).map(([m,t])=>[m,{meanMS:t.meanElapsedMS,worstMS:t.worst.elapsedNS/1e6,over400MS:t.over400MS,phaseFractions:t.phaseFractions,gates:t.gates}])),paired},null,2));
