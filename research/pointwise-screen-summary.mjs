import fs from 'node:fs';
import assert from 'node:assert/strict';
const[input,output]=process.argv.slice(2);assert(input&&output);
const x=JSON.parse(fs.readFileSync(input));assert.equal(x.summary.cohort,'consumed-pointwise-screen-v1');assert.equal(x.results.length,672);
let state=2463534242;
const rnd=()=>{state^=state<<13;state^=state>>>17;state^=state<<5;return(state>>>0)/4294967296;};
const mean=x=>x.reduce((a,b)=>a+b,0)/x.length;
function ci(v){const m=[];for(let i=0;i<10000;i++)m.push(mean(v.map(()=>v[Math.floor(rnd()*v.length)])));m.sort((a,b)=>a-b);return[m[249],m[9749]];}
const out=[];
for(const regime of ['all','stationary','changing']){
  const rows=x.results.filter(r=>{const c=Number(r.key.split(':')[1]),ch=c<9?c%3!==0:c>=19;return regime==='all'||ch===(regime==='changing');});
  for(const period of ['whole','terminal64']){
    const metric=r=>period==='whole'?r.expected:r.expected.map((_,a)=>(r.blocks[6].expected[a]+r.blocks[7].expected[a])/2);
    const paired=[];
    for(const ref of [2,1,3,4]){
      const groups=Array.from({length:8},(_,i)=>rows.filter(r=>Number(r.key.split(':')[2])===i));
      assert(groups.every(g=>g.length===groups[0].length&&g.length>0));
      const values=groups.map(g=>mean(g.map(r=>metric(r)[0]-metric(r)[ref])));
      paired.push({reference:x.summary.arms[ref].arm,delta:mean(values),pointwise95:ci(values)});
    }
    out.push({regime,period,records:rows.length,paired});
  }
}
const doc={analysis:'consumed-data exploratory pointwise intervals',comparisons:out,gainRetained:(x.summary.arms[2].expected-x.summary.arms[0].expected)/(x.summary.arms[2].expected-x.summary.arms[1].expected)};
fs.writeFileSync(output,JSON.stringify(doc,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({comparisons:out.length,gainRetained:doc.gainRetained}));
