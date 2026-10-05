import fs from 'node:fs';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);
assert(input&&output);
const s=JSON.parse(fs.readFileSync(input,'utf8'));assert.equal(s.checked,84);assert.equal(s.paired.length,84);
const comparisons=[];
for(const [name,k] of [['frozen32',1],['generic64',2],['Boolean64',3],['Markov',4]]) {
  const rows=s.paired.map(r=>({phase:r.phase,scenario:r.scenario,schedule:r.schedule,delta:r.expected[0]-r.expected[k]}));
  comparisons.push({control:name,improved:rows.filter(r=>r.delta<0).length,harmGreaterThanPoint01:rows.filter(r=>r.delta>.01).length,
    meanDelta:rows.reduce((a,r)=>a+r.delta,0)/84,rows});
}
fs.writeFileSync(output,JSON.stringify({comparisons,limitations:'One-index descriptive differences, not confidence-certified non-harm or fresh confirmation.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(comparisons.map(({rows,...summary})=>summary),null,2));
