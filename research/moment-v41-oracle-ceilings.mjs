// Post-hoc feasibility diagnostic only. Truth is NEVER candidate input.
import fs from 'node:fs';
import crypto from 'node:crypto';
import readline from 'node:readline';
import assert from 'node:assert/strict';

const root='research/moment-v41-normal';
const completed=JSON.parse(fs.readFileSync(root+'/completed.json'));
assert(completed.source_unchanged&&completed.stage==='normal');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const stationary=new Set(['aligned','independent','curved','baseline_matched','stationary_noise10']);
const schedules=['immediate','fixed150','uniform299'];
function mean(x){return x.reduce((sum,value)=>sum+value/x.length,0)}
function oracle(w){
 assert.equal(w.Rates.length,16);
 const rounds=w.Rates.map(p=>{assert.equal(p.length,150);assert(p.every(x=>Number.isFinite(x)&&x>=0&&x<=1));return{brier:mean(p.map(x=>x*(1-x))),priority:p.reduce((s,x,i)=>s+x*(1-x)*(i<10?3:1)/170,0),usefulness:mean([...p].sort((a,b)=>b-a).slice(0,10))}});
 let recovery=0;
 const phases=[];
 for(let j=0;j<(w.Changes??[]).length;j++){
  const start=w.Changes[j],end=w.Changes[j+1]??16;let consecutive=0,delay=end-start+1;
  for(let r=start;r<end;r++){consecutive=rounds[r].brier<=.20&&rounds[r].usefulness>=.75?consecutive+1:0;if(consecutive>=2){delay=r-start+1;break}}
  phases.push({start,end,oracleDelay:delay,oracleMiss:delay===end-start+1});recovery+=delay/w.Changes.length;
 }
 return{brier:mean(rounds.map(r=>r.brier)),priority:mean(rounds.map(r=>r.priority)),recovery,phases};
}
const reports={};
for(const split of ['design','confirmation']){
 const groups=new Map();let count=0,manifest;
 const digest=crypto.createHash('sha256');
 const raw=root+'/'+split+'.jsonl';
 const input=fs.createReadStream(raw);input.on('data',b=>digest.update(b));
 for await(const line of readline.createInterface({input,crlfDelay:Infinity})){
  const row=JSON.parse(line);
  if(!manifest){manifest=row;assert.equal(manifest.Split,split);assert.equal(manifest.Worlds,448);continue}
  count++;const w=row.Population,o=oracle(w);
  for(let s=0;s<3;s++){
   const full=row.Arms[s*11];assert.equal(full.Mode,'full');assert.equal(full.Schedule,schedules[s]);
   const brierGain=full.IssuedBrier-o.brier,priorityGain=full.IssuedPriority-o.priority,recoveryGain=full.Recovery-o.recovery;
   assert(brierGain>=-1e-12&&priorityGain>=-1e-12&&recoveryGain>=-1e-12,'control below attainable oracle lower bound');
   const key=w.Geometry+'/'+w.Regime+'/'+schedules[s];if(!groups.has(key))groups.set(key,[]);
   groups.get(key).push({seed:w.Seed,brierGain,priorityGain,recoveryGain,fullRecovery:full.Recovery,phases:o.phases});
  }
 }
 assert.equal(count,448);assert.equal(groups.size,84);
 const sha=digest.digest('hex');assert.equal(sha,completed.artifacts[split+'.jsonl'],'immutable audited raw');
 const cells={};
 for(const[key,rows]of groups){assert.equal(rows.length,16);const regime=key.split('/')[1];const shifted=!stationary.has(regime),changes=rows[0].phases.length>0;
  const maxBrierGain=mean(rows.map(r=>r.brierGain)),maxPriorityGain=mean(rows.map(r=>r.priorityGain)),maxRecoveryGain=mean(rows.map(r=>r.recoveryGain)),fullRecovery=mean(rows.map(r=>r.fullRecovery));
  cells[key]={maximumIssuedBrierMeanGain:maxBrierGain,maximumIssuedPriorityMeanGain:maxPriorityGain,maximumRecoveryMeanGain:maxRecoveryGain,fullRecoveryMean:fullRecovery,
   issuedBrierMeanTargetImpossible:shifted&&maxBrierGain<.01-1e-12,issuedPriorityMeanTargetImpossible:shifted&&maxPriorityGain<.01-1e-12,recoveryMeanTargetImpossible:changes&&maxRecoveryGain<.1*fullRecovery-1e-12,
   oracleMissPhases:rows.reduce((s,r)=>s+r.phases.filter(p=>p.oracleMiss).length,0),phaseCount:rows.reduce((s,r)=>s+r.phases.length,0)};
 }
 const impossible=Object.entries(cells).filter(([,c])=>c.issuedBrierMeanTargetImpossible||c.issuedPriorityMeanTargetImpossible||c.recoveryMeanTargetImpossible).map(([name,c])=>({name,...c}));
 reports[split]={worlds:count,rawSHA256:sha,cells,impossibleCells:impossible};
}
const result={time:new Date().toISOString(),sourceSHA256:hash(fs.readFileSync('research/moment-v41-oracle-ceilings.mjs')),postHocDiagnostic:true,truthNeverCandidateInput:true,existingGatesChanged:false,scope:'finite audited cohorts; perfect truth predictor ignores observation costs and delay, therefore supplies optimistic upper bounds, NOT an implementable or validated rescue',reports,goal:'ACTIVE',allSevenWholeGoals:'OPEN'};
const f=fs.openSync(root+'/oracle-ceilings.json','wx',0o600);try{fs.writeFileSync(f,JSON.stringify(result,null,2)+'\n');fs.fsyncSync(f)}finally{fs.closeSync(f)}
console.log(JSON.stringify({impossibleCells:Object.fromEntries(Object.entries(reports).map(([s,r])=>[s,r.impossibleCells])),existingGatesChanged:false},null,2));
