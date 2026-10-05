import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const [source,auditPath,output]=process.argv.slice(2);assert(source&&auditPath&&output);
const audit=JSON.parse(fs.readFileSync(auditPath)),groups=new Map();let header=true,count=0;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})) {
  const r=JSON.parse(line);if(header){assert.equal(r.Version,'soft-learners-v120');header=false;continue;}count++;
  const key=[r.Phase,r.Case,r.Schedule].join(':');if(!groups.has(key))groups.set(key,[]);
  const noise=r.Steps.slice(192).reduce((v,s)=>v+s.Q*(1-s.Q),0)/64;
  groups.get(key).push({index:r.Index,noise,risks:r.Metrics.map(m=>m[1].Brier)});
}
assert.equal(count,2688);
let maximumError=0;
const cells=audit.cells.map(c=>{
  const rows=groups.get(c.key);assert.equal(rows.length,32);
  assert.deepEqual(rows.map(r=>r.index).sort((a,b)=>a-b),Array.from({length:32},(_,i)=>i));
  const controlRisk=rows.reduce((v,r)=>v+r.risks[c.control],0)/32;
  const noiseRisk=rows.reduce((v,r)=>v+r.noise,0)/32;
  const ceiling=controlRisk-noiseRisk,error=Math.abs(ceiling-c.oracleMaximumMeanGain);
  maximumError=Math.max(maximumError,error);assert(error<1e-12);
  return {...c,controlRisk,noiseRisk,ceilingFromArchivedRisk:ceiling};
});
const out={checks:cells.length,maximumError,unattainable:cells.filter(c=>c.belowRequiredGain),
  scope:'Independent arithmetic route: archived control Brier minus recomputed Bernoulli noise floor, versus direct squared probability error. Same archived source, not independent outcome evidence.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(out,null,2));
