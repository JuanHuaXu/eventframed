import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const [source,output]=process.argv.slice(2);assert(source&&output);
const groups=new Map();let header=true,count=0;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})) {
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  const changing=r.Case<9?r.Case%3!==0:r.Case>=19;count++;if(!changing)continue;
  const key=[r.Phase,r.Case,r.Schedule].join(':');
  if(!groups.has(key))groups.set(key,[]);
  const limits=[0,12,13].map(a=>r.Steps.slice(192).reduce((v,s)=>v+(s.P[a]-s.Q)**2/64,0));
  groups.get(key).push({index:r.Index,limits});
}
assert.equal(count,672);assert.equal(groups.size,32);
const cells=[];
for(const [key,rows]of groups) {
  assert.deepEqual(rows.map(r=>r.index).sort((a,b)=>a-b),[0,1,2,3,4,5,6,7]);
  for(let a=0;a<3;a++) {
    const upperGain=rows.reduce((v,r)=>v+r.limits[a]/8,0);
    cells.push({key,control:[0,12,13][a],oracleMaximumMeanGain:upperGain,belowRequiredGain:upperGain<.005-1e-12});
  }
}
const out={scope:'Consumed fixed-forecast conditional-risk ceiling, Q evaluator only',
  derivation:'Expected scalar Brier is (p-Q)^2+Q*(1-Q). Its minimum over any p is the noise floor at p=Q. The maximum gain over a fixed control is therefore (control-Q)^2. No guard or forecast-family restriction.',
  cells,belowRequiredGain:cells.filter(c=>c.belowRequiredGain).length,total:cells.length,
  limitations:'Exact bound on these measured records, not a population/fresh-cohort impossibility, not new confirmation, and not authority to change frozen thresholds.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({...out,cells:cells.filter(c=>c.belowRequiredGain)},null,2));
