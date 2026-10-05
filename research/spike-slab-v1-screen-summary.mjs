import fs from 'node:fs';
import assert from 'node:assert/strict';

const [path,output]=process.argv.slice(2);
assert(path && output);
const audit=JSON.parse(fs.readFileSync(path,'utf8'));
assert.equal(audit.checked,2016);assert.equal(audit.paired.length,2016);
const groups=new Map();
const names=['additive-stationary','additive-abrupt','additive-gradual',
 'hierarchy-stationary','hierarchy-abrupt','hierarchy-gradual',
 'local-table-stationary','local-table-abrupt','local-table-gradual',
 'parity1','parity2','parity3','parity4','complement4','majority3','mux3',
 'constant','null','dependent4','majority-to-parity','parity-to-majority'];
for(const r of audit.paired) {
  const key=[r.Clock,r.Phase,r.Case,r.Schedule].join(':');
  if(!groups.has(key))groups.set(key,[]);
  groups.get(key).push(r);
}
const cases=[];
for(const [key,rs] of groups) {
  assert.equal(rs.length,8);assert.equal(new Set(rs.map(r=>r.Index)).size,8);
  const [clock,phase,scenario,schedule]=key.split(':').map(Number);
  for(const [name,arm] of [['plugin',1],['generic64',2],['Boolean64',3],['Markov',4]]) {
    const xs=rs.map(r=>r.expected[0]-r.expected[arm]);
    const mean=xs.reduce((a,b)=>a+b,0)/8;
    const se=Math.sqrt(xs.reduce((a,x)=>a+(x-mean)**2,0)/(7*8));
    cases.push({clock,phase,scenario,caseName:names[scenario],schedule,control:name,mean,se,indices:rs.map(r=>({index:r.Index,delta:r.expected[0]-r.expected[arm]}))});
  }
}
const counts=[];
for(const clock of [0,128,224])for(const control of ['plugin','generic64','Boolean64','Markov']) {
  const rows=cases.filter(r=>r.clock===clock&&r.control===control);
  assert.equal(rows.length,84);
  counts.push({clock,control,cells:84,improved:rows.filter(r=>r.mean<0).length,
    worseByMoreThanPoint01:rows.filter(r=>r.mean>.01).length,
    averageDelta:rows.reduce((a,r)=>a+r.mean,0)/84,
    worst:rows.reduce((a,r)=>r.mean>a.mean?r:a)});
}
const report={counts,cases,limitations:'Exploratory eight-index paired means and standard errors. No simultaneous coverage claim or completed 32-index all-clock gates. Positive delta means harm.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(counts,null,2));
