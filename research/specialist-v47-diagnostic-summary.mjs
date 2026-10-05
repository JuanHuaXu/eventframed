// Descriptive only: n=1 per cell cannot satisfy the unchanged n=16 gates.
import fs from 'node:fs';
import assert from 'node:assert/strict';
const root='research/specialist-v47-study-diagnostic';
const r=JSON.parse(fs.readFileSync(root+'/readback.json'));
assert(r.stage==='diagnostic'&&!r.qualityAdoption&&r.independentMetricArithmetic);
const s=r.summaries.diagnostic;assert.equal(s.worlds,28);assert.equal(Object.keys(s.pairedContrasts).length,84);
const stationary=new Set(['aligned','independent','curved','baseline_matched','stationary_noise10']);
const mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
const modes=['static','local','reset'],summary={};
for(const mode of modes){
 const all=Object.entries(s.pairedContrasts).map(([key,policies])=>({key,d:policies[mode]}));
 const shifted=all.filter(x=>!stationary.has(x.key.split('/')[1]));
 const stable=all.filter(x=>stationary.has(x.key.split('/')[1]));
 const gain=(rows,field)=>rows.map(x=>x.d[field].mean);
 summary[mode]={
  cells:all.length,
  issuedVsFull:{positive:gain(all,'full/IssuedBrier').filter(x=>x>0).length,macroMean:mean(gain(all,'full/IssuedBrier'))},
  issuedVsAdaptive:{positive:gain(all,'adaptive/IssuedBrier').filter(x=>x>0).length,macroMean:mean(gain(all,'adaptive/IssuedBrier'))},
  shiftedIssuedVsFull:{cells:shifted.length,macroMean:mean(gain(shifted,'full/IssuedBrier')),atLeastOnePoint:gain(shifted,'full/IssuedBrier').filter(x=>x>=.01).length},
  stationaryIssuedVsFull:{cells:stable.length,macroMean:mean(gain(stable,'full/IssuedBrier')),worst:Math.min(...gain(stable,'full/IssuedBrier'))},
  finalUsefulnessVsAdaptive:{macroMean:mean(gain(all,'adaptive/FinalUsefulness')),worseByMoreThanOnePoint:gain(all,'adaptive/FinalUsefulness').filter(x=>x<-.01).length},
  nPerCell:1,intervals:null,adoption:false
 };
}
const result={time:new Date().toISOString(),summary,maximumMeasuredLoopNS:s.maximumElapsedNS,wholeGoals:'OPEN',interpretation:'Equal-cell descriptive means only, no new gates or policy selection. Normal cohorts remain required.'};
const fd=fs.openSync(root+'/diagnostic-summary.json','wx',0o600);try{fs.writeFileSync(fd,JSON.stringify(result,null,2)+'\n');fs.fsyncSync(fd)}finally{fs.closeSync(fd)}
console.log(JSON.stringify(result,null,2));
