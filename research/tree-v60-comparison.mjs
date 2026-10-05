// Post-collection descriptive comparison. No new acceptance gates or retuning.
import fs from 'node:fs';import readline from 'node:readline';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const root='research/tree-v60-diagnostic';
async function load(file){const rows=[];for await(const line of readline.createInterface({input:fs.createReadStream(file),crlfDelay:Infinity}))rows.push(JSON.parse(line));return rows}
const referenceRoot=process.argv[2]==='v58'?'research/tree-v58-diagnostic':'research/paired-v57-diagnostic';
const current=await load(root+'/diagnostic.jsonl'),previous=await load(referenceRoot+'/diagnostic.jsonl');
const readback=JSON.parse(fs.readFileSync(root+'/readback.json'));assert.equal(current.length,41);assert.equal(previous.length,41);
const mean=v=>v.reduce((a,b)=>a+b,0)/v.length,oldChanges={},expiry={},schedules={},stationary={},controls=[];
const modes=['full','adaptive','no_pair','random','uncertainty','information','falsification','predictive'];
for(const mode of modes){oldChanges[mode]=[];expiry[mode]={requested:0,expired:0,expiredAvailable:0,missing:0};stationary[mode]=[];schedules[mode]={}}
const stripped=a=>{const b=structuredClone(a);delete b.Costs;delete b.Breakdown;return b};
for(let i=1;i<current.length;i++) {
 const w=current[i],old=previous[i];assert.deepEqual(w.Population,old.Population);
 for(let j=0;j<24;j++) {const a=w.Arms[j],b=old.Arms[j];assert.equal(a.Mode,b.Mode);assert.equal(a.Schedule,b.Schedule);
  if(j%8<2){assert.deepEqual(stripped(a),stripped(b));controls.push([i,j])}
  const t=a.Snapshots.at(-1),u=b.Snapshots.at(-1);oldChanges[a.Mode].push({riskGain:b.IssuedBrier-a.IssuedBrier,terminalGain:u.Brier-t.Brier,utilityGain:t.PacketUsefulness-u.PacketUsefulness,recoveryGain:b.Recovery-a.Recovery,oldMS:b.Costs.ElapsedNS/1e6,newMS:a.Costs.ElapsedNS/1e6});
  for(const o of a.AuditOutcomes??[]){const expired=o.Trial<Math.min(o.ArrivedAt+1,2400)-600;assert.equal(o.Expired,expired);const e=expiry[a.Mode];e.requested++;e.expired+=+expired;e.expiredAvailable+=+(expired&&o.Available);e.missing+=+!o.Available}
  (schedules[a.Mode][a.Schedule]??=[]).push(a.IssuedBrier);
  if(!(w.Population.World.Changes??[]).length)stationary[a.Mode].push({risk:a.IssuedBrier,adaptiveHarm:a.IssuedBrier-w.Arms[Math.floor(j/8)*8+1].IssuedBrier});
 }
}
const changes=Object.fromEntries(modes.map(m=>[m,{riskGain:mean(oldChanges[m].map(x=>x.riskGain)),terminalGain:mean(oldChanges[m].map(x=>x.terminalGain)),utilityGain:mean(oldChanges[m].map(x=>x.utilityGain)),recoveryGain:mean(oldChanges[m].map(x=>x.recoveryGain)),wins:oldChanges[m].filter(x=>x.riskGain>1e-12).length,losses:oldChanges[m].filter(x=>x.riskGain< -1e-12).length,oldTotalMS:oldChanges[m].reduce((s,x)=>s+x.oldMS,0),newTotalMS:oldChanges[m].reduce((s,x)=>s+x.newMS,0),stationaryCells:stationary[m].length,stationaryMeanRisk:mean(stationary[m].map(x=>x.risk)),stationaryMaxAdaptiveHarm:Math.max(...stationary[m].map(x=>x.adaptiveHarm)),scheduleRisk:Object.fromEntries(Object.entries(schedules[m]).map(([s,x])=>[s,mean(x)]))}]));
const result={stage:'Descriptive audit, no added gates',referenceRoot,controlsBitwiseEqual:controls.length,allPopulationsIdentical:true,changes,expiry,gates:Object.fromEntries(Object.entries(readback.summaries).map(([m,s])=>[m,{meanGain01:s.fullRiskGain>=.01,noAdaptiveHarm01:s.adaptiveHarmOver01===0,completeLoop400:s.over400MS===0,recoveryBetterThanAdaptive:s.recoveryGainAdaptive>0,equalsTotalCostNotEstablished:true}])),wholeGoals:Array(7).fill('OPEN'),goal:'ACTIVE',productionChanged:false};
fs.writeFileSync(root+(process.argv[2]==='v58'?'/comparison-v58.json':'/comparison.json'),JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify(result,null,2));
