import fs from 'node:fs';
import assert from 'node:assert/strict';
const[input,output]=process.argv.slice(2);assert(input&&output);
const data=JSON.parse(fs.readFileSync(input));assert.equal(data.summary.cohort,'consumed-segment-guard-v1');assert.equal(data.results.length,672);
let state=2463534242;const rnd=()=>{state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;};const mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
function ci(v){const m=[];for(let i=0;i<10000;i++)m.push(mean(v.map(()=>v[Math.floor(rnd()*v.length)])));m.sort((a,b)=>a-b);return[m[249],m[9749]];}
const comparisons=[];
for(const regime of ['all','stationary','changing'])for(const period of ['whole','terminal64']){
  const rows=data.results.filter(r=>{const c=Number(r.key.split(':')[1]),ch=c<9?c%3!==0:c>=19;return regime==='all'||ch===(regime==='changing');});
  const metric=r=>period==='whole'?r.expected:r.expected.map((_,a)=>(r.blocks[6].expected[a]+r.blocks[7].expected[a])/2);
  for(const[a,ref]of [[0,1],[0,2],[3,4],[3,2]]){
    const clusters=Array.from({length:8},(_,i)=>rows.filter(r=>Number(r.key.split(':')[2])===i));assert(clusters.every(c=>c.length===clusters[0].length));
    const v=clusters.map(c=>mean(c.map(r=>metric(r)[a]-metric(r)[ref])));
    comparisons.push({regime,period,arm:data.summary.arms[a].arm,reference:data.summary.arms[ref].arm,delta:mean(v),pointwise95:ci(v)});
  }
}
const groups=new Map();for(const r of data.results){const[p,c,,s]=r.key.split(':'),key=[p,c,s].join(':');if(!groups.has(key))groups.set(key,[]);groups.get(key).push(r);}
const recoveryMeans=[];
for(const[key,rows]of groups){assert.equal(rows.length,8);const c=Number(key.split(':')[1]);if(!(c<9?c%3!==0:c>=19))continue;
  for(const[a,ref]of [[0,1],[0,2],[3,4],[3,2]]){const gain=mean(rows.map(r=>(r.blocks[6].expected[ref]+r.blocks[7].expected[ref]-r.blocks[6].expected[a]-r.blocks[7].expected[a])/2));recoveryMeans.push({key,arm:data.summary.arms[a].arm,reference:data.summary.arms[ref].arm,gain,meetsOriginalPoint005Mean:gain>=.005});}
}
const out={comparisons,recoveryMeans,limitations:'Consumed-data pointwise8-index intervals. Recovery mean screen retains original .005 effect requirement descriptively, not the original32-index confidence protocol. Not full validation of768 criteria.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({comparisons,recoveryMeansMeetingPoint005:recoveryMeans.filter(r=>r.meetsOriginalPoint005Mean).length,recoveryMeansTotal:recoveryMeans.length},null,2));
