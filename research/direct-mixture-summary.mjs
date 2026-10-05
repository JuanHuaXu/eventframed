import fs from 'node:fs';
import assert from 'node:assert/strict';
const[input,output]=process.argv.slice(2);assert(input&&output);
const x=JSON.parse(fs.readFileSync(input));assert.equal(x.summary.cohort,'consumed-direct-mixture-v1');assert.equal(x.results.length,672);
let state=2463534242;const rnd=()=>{state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;};
const mean=v=>v.reduce((a,b)=>a+b,0)/v.length;
function ci(v){const m=[];for(let i=0;i<10000;i++)m.push(mean(v.map(()=>v[Math.floor(rnd()*v.length)])));m.sort((a,b)=>a-b);return[m[249],m[9749]];}
const comparisons=[];
for(const regime of ['all','stationary','changing'])for(const period of ['whole','terminal64']){
  const rows=x.results.filter(r=>{const c=Number(r.key.split(':')[1]),changing=c<9?c%3!==0:c>=19;return regime==='all'||changing===(regime==='changing');});
  const metric=r=>period==='whole'?r.expected:r.expected.map((_,a)=>(r.blocks[6].expected[a]+r.blocks[7].expected[a])/2);
  for(let a=0;a<4;a++)for(const ref of [4,6]){
    const clusters=Array.from({length:8},(_,i)=>rows.filter(r=>Number(r.key.split(':')[2])===i));assert(clusters.every(c=>c.length===clusters[0].length));
    const v=clusters.map(c=>mean(c.map(r=>metric(r)[a]-metric(r)[ref])));
    comparisons.push({regime,period,arm:x.summary.arms[a].arm,reference:x.summary.arms[ref].arm,delta:mean(v),pointwise95:ci(v)});
  }
}
const out={comparisons,limitations:'Consumed-data paired eight-index cluster bootstrap. Pointwise intervals, not simultaneous or selection-adjusted. Report all four variants.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(comparisons.filter(c=>c.reference==='sharePointwise'),null,2));
