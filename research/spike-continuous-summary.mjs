import fs from 'node:fs';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const a=JSON.parse(fs.readFileSync(input,'utf8'));assert.equal(a.results.length,672);
let state=2463534242;
function random(){state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;}
function interval(values){const means=[];for(let b=0;b<10000;b++){let s=0;for(let j=0;j<values.length;j++)s+=values[Math.floor(random()*values.length)];means.push(s/values.length);}means.sort((a,b)=>a-b);return[means[249],means[9749]];}
const measure=(r,period)=>period==='whole'?r.expected:period==='terminal64'?r.expected.map((_,j)=>(r.blocks[6].expected[j]+r.blocks[7].expected[j])/2):r.blocks[Number(period)].expected;
const periods=[];
for(const period of ['whole','terminal64',...Array.from({length:8},(_,i)=>String(i))]){
  const byIndex=Array.from({length:8},()=>[0,0]);let harms=0;
  for(const r of a.results){const index=Number(r.key.split(':')[2]),p=measure(r,period);assert(index>=0&&index<8);byIndex[index][0]+=(p[0]-p[2])/84;byIndex[index][1]+=(p[0]-p[1])/84;harms+=Number(p[0]-p[2]>.01);}
  periods.push({period,versusMarkov:byIndex.reduce((s,x)=>s+x[0],0)/8,versusMarkov95:interval(byIndex.map(x=>x[0])),versusReset32:byIndex.reduce((s,x)=>s+x[1],0)/8,versusReset3295:interval(byIndex.map(x=>x[1])),recordHarmsAbovePoint01:harms});
}
const groups=new Map();
for(const r of a.results){const [phase,scenario,index,schedule]=r.key.split(':').map(Number),key=[phase,scenario,schedule].join(':');if(!groups.has(key))groups.set(key,new Map());assert(!groups.get(key).has(index));groups.get(key).set(index,r);}
assert.equal(groups.size,84);const scenarios=[];
for(const [key,rs]of groups){assert.equal(rs.size,8);for(const period of ['whole','terminal64']){const values=Array.from({length:8},(_,i)=>{const p=measure(rs.get(i),period);return p[0]-p[2];});scenarios.push({key,period,meanDelta:values.reduce((a,b)=>a+b,0)/8,pointwise95:interval(values)});}}
const summary={periods,scenarioMeansHarmedAbovePoint01:scenarios.filter(r=>r.meanDelta>.01),scenarios,limitations:'Eight consumed trajectory-index clusters; pointwise exploratory percentile intervals, not simultaneous safety or untouched confirmation. A whole-stream realized budget does not constrain terminal subwindow excess.'};
fs.writeFileSync(output,JSON.stringify(summary,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({periods,scenarioMeansHarmedAbovePoint01:summary.scenarioMeansHarmedAbovePoint01},null,2));
