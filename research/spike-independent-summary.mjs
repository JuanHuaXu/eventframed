import fs from 'node:fs';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const a=JSON.parse(fs.readFileSync(input,'utf8'));
assert.equal(a.summary.cohort,'spike-independent-v1');assert.equal(a.results.length,672);
let state=2463534242;
function random(){state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;}
function interval(values){const means=[];for(let b=0;b<10000;b++){let s=0;for(let j=0;j<values.length;j++)s+=values[Math.floor(random()*values.length)];means.push(s/values.length);}means.sort((a,b)=>a-b);return[means[249],means[9749]];}
const measure=(r,period,field='expected')=>period==='whole'?r[field]:period==='terminal64'?r[field].map((_,j)=>(r.blocks[6][field][j]+r.blocks[7][field][j])/2):r.blocks[Number(period)][field];
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const ids=new Set();
for(const r of a.results){assert(!ids.has(r.key));ids.add(r.key);const[p,c,i,s]=r.key.split(':').map(Number);assert([p,c,i,s].every(Number.isInteger)&&p>=0&&p<2&&c>=0&&c<21&&i>=0&&i<8&&s>=0&&s<2);assert(r.expected.length===7&&r.realized.length===7&&r.blocks.length===8);}
function evaluate(records,period){
  const clusters=Array.from({length:8},()=>[]);
  for(const r of records)clusters[Number(r.key.split(':')[2])].push(r);
  assert(clusters.every(c=>c.length===clusters[0].length&&c.length>0));
  const delta=(r,ref,field='expected')=>measure(r,period,field)[0]-measure(r,period,field)[ref];
  return {period,records:records.length,expected:Array.from({length:7},(_,j)=>mean(records.map(r=>measure(r,period)[j]))),
    realized:Array.from({length:7},(_,j)=>mean(records.map(r=>measure(r,period,'realized')[j]))),
    versusMarkov:mean(records.map(r=>delta(r,2))),versusMarkov95:interval(clusters.map(c=>mean(c.map(r=>delta(r,2))))),
    versusReset32:mean(records.map(r=>delta(r,1))),versusReset3295:interval(clusters.map(c=>mean(c.map(r=>delta(r,1))))),
    harmsAbovePoint01:records.filter(r=>delta(r,2)>.01).map(r=>({key:r.key,delta:delta(r,2)})),
    maximumHarm:Math.max(...records.map(r=>delta(r,2)))};
}
const periods=['whole','terminal64',...Array.from({length:8},(_,i)=>String(i))].map(p=>evaluate(a.results,p));
const groups=new Map();
for(const r of a.results){const[p,c,,s]=r.key.split(':');const k=[p,c,s].join(':');if(!groups.has(k))groups.set(k,[]);groups.get(k).push(r);}
assert.equal(groups.size,84);
const scenarios=[];
for(const[key,rows]of groups){assert.equal(rows.length,8);for(const p of ['whole','terminal64'])scenarios.push({key,...evaluate(rows,p)});}
const changing=r=>{const c=Number(r.key.split(':')[1]);return c<9?c%3!==0:c>=19;};
const regimes=[];
for(const isChanging of [false,true])for(const period of ['whole','terminal64'])regimes.push({regime:isChanging?'changing':'stationary',...evaluate(a.results.filter(r=>changing(r)===isChanging),period)});
const windowScenarioHarms=[];
for(const[key,rows]of groups)for(let b=0;b<8;b++){const d=mean(rows.map(r=>r.blocks[b].expected[0]-r.blocks[b].expected[2]));if(d>.01)windowScenarioHarms.push({key,clock:32*b,delta:d});}
const summary={cohort:a.summary.cohort,arms:a.summary.arms,counts:a.summary,periods,regimes,scenarios,windowScenarioHarms,
  limitations:'Frozen-policy independent synthetic replication. Eight trajectory-index clusters, paired schedules; percentile bootstrap intervals are pointwise, not simultaneous coverage. No real-agent-task or serving-latency claim. The realized loss ledger does not certify conditional expected loss or arbitrary windows.'};
fs.writeFileSync(output,JSON.stringify(summary,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({counts:a.summary,periods:periods.map(({harmsAbovePoint01,...r})=>({...r,harms:harmsAbovePoint01.length})),regimes:regimes.map(({harmsAbovePoint01,...r})=>({...r,harms:harmsAbovePoint01.length})),windowScenarioHarms},null,2));
