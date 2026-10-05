// Post-collection accounting, not a new timed workload or a profile.
import fs from 'node:fs';import readline from 'node:readline';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/shared-v68-diagnostic',source='research/shared-v68-cost-audit.mjs';
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const done=JSON.parse(fs.readFileSync(root+'/completed.json'));assert(done.allJobsTerminal&&done.checks.every(x=>x.exitCode===0));
assert.equal(await hash(root+'/diagnostic.jsonl'),done.artifacts['diagnostic.jsonl']);
const phases=['ScheduleNS','SetupNS','IssueNS','FirstResolveNS','ProposalNS','RequestNS','SecondResolveNS','SnapshotNS'],totals={};let arms=0;
for await(const line of readline.createInterface({input:fs.createReadStream(root+'/diagnostic.jsonl'),crlfDelay:Infinity})){
 const w=JSON.parse(line);if(!w.Arms)continue;
 for(const a of w.Arms){if(a.Mode==='full'||a.Mode==='adaptive')continue;arms++;const c=a.Breakdown,t=totals[a.Mode]??={arms:0,elapsedNS:0,accountedNS:0,phases:Object.fromEntries(phases.map(p=>[p,0])),worst:null};
  assert.equal(c.AccountedNS,phases.reduce((s,p)=>s+c[p],0));assert(c.AccountedNS<=c.ElapsedNS);t.arms++;t.elapsedNS+=c.ElapsedNS;t.accountedNS+=c.AccountedNS;for(const p of phases)t.phases[p]+=c[p];
  if(!t.worst||c.ElapsedNS>t.worst.elapsedNS)t.worst={geometry:w.Population.World.Geometry,regime:w.Population.World.Regime,schedule:a.Schedule,elapsedNS:c.ElapsedNS,phases:Object.fromEntries(phases.map(p=>[p,c[p]]))};
 }
}
assert.equal(arms,1440);
for(const t of Object.values(totals)){assert.equal(t.arms,120);t.phaseFractions=Object.fromEntries(phases.map(p=>[p,t.phases[p]/t.elapsedNS]));t.unattributedFraction=1-t.accountedNS/t.elapsedNS;t.meanElapsedMS=t.elapsedNS/t.arms/1e6}
const result={source,sourceSHA256:await hash(source),dataSHA256:done.artifacts['diagnostic.jsonl'],scope:'measured phase wall-time accounting only; no unique causal diagnosis,profiler or production benchmark',modelArms:arms,totals,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
fs.writeFileSync(root+'/cost-audit.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(Object.fromEntries(Object.entries(totals).map(([m,t])=>[m,{meanElapsedMS:t.meanElapsedMS,phaseFractions:t.phaseFractions,worstMS:t.worst.elapsedNS/1e6}])),null,2));
