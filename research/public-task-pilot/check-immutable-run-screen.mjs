import fs from 'node:fs';
import crypto from 'node:crypto';
const p=process.argv[2]??'research/public-task-pilot/immutable-run-screen-results.json';
const out=JSON.parse(fs.readFileSync(p));
const assert=(v,m)=>{if(!v)throw Error(m)};
for(const [p,h] of Object.entries(out.Hashes))assert(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')===h,`hash ${p}`);
function vector(seed){const v=[];for(let b=0;b<24;b++)for(const x of crypto.createHash('sha256').update(`${seed}/block-${b}`).digest())v.push(x-127.5);return v;}
function score(a,b){let d=0,aa=0,bb=0;for(let i=0;i<a.length;i++){d+=a[i]*b[i];aa+=a[i]*a[i];bb+=b[i]*b[i];}return d/Math.sqrt(aa*bb);}
function order(a,b){return b.Score-a.Score||(a.ID<b.ID?-1:a.ID>b.ID?1:0);}
assert(out.Stages.length===16,'stage count');
let allPass=true;
for(let repeat=0;repeat<2;repeat++){
 const state=new Map();for(let i=0;i<800;i++)state.set(`seed-${i}`,vector(`seed-${i}`));
 for(let step=0;step<8;step++){
  for(let i=0;i<32;i++){
   const id=i<6?`seed-${step*6+i}`:`new-${step}-${i}`;
   if(i>=4&&i<6)state.delete(id);else state.set(id,vector(i<6?`update-${step}-${i}`:id));
  }
  const s=out.Stages[repeat*8+step];assert(s.Repeat===repeat&&s.Step===step&&s.Records===state.size&&s.Runs===step+2,'stage identity');
  assert(s.Queries.length===32,'query count');let rh=0,fh=0;const rt=[],ft=[];
  for(let probe=0;probe<32;probe++){
   const row=s.Queries[probe];assert(row.Seed===`probe-${step}-${probe}`,'probe seed');const q=vector(row.Seed);
   const oracle=[...state].map(([ID,v])=>({ID,Score:score(q,v)})).sort(order).slice(0,10);
   assert(row.Oracle.length===10,'oracle length');
   for(let i=0;i<10;i++)assert(row.Oracle[i].ID===oracle[i].ID&&Math.abs(row.Oracle[i].Score-oracle[i].Score)<1e-14,'oracle mismatch');
   const ids=new Set(oracle.map(x=>x.ID));
   for(const kind of ['Runs','Full']){
    const cs=row[kind];assert(cs.length===10&&new Set(cs.map(x=>x.ID)).size===10,'candidate count');
    cs.forEach((c,i)=>{assert(state.has(c.ID),'stale/deleted ID');assert(Math.abs(c.Score-score(q,state.get(c.ID)))<1e-14,'stale vector/score');if(i)assert(order(cs[i-1],c)<=0,'order');});
    const hits=cs.filter(c=>ids.has(c.ID)).length;if(kind==='Runs')rh+=hits;else fh+=hits;
    assert(row[kind+'NS']>0,'query time');(kind==='Runs'?rt:ft).push(row[kind+'NS']/1e6);
   }
  }
  rt.sort((a,b)=>a-b);ft.sort((a,b)=>a-b);
  const runMS=(s.RunBuildNS+s.RunPlanNS)/1e6,fullMS=(s.FullBuildNS+s.FullPlanNS)/1e6;
  assert(runMS>0&&fullMS>0&&s.CloseNS>0,'build/close times');
  const pass=runMS<fullMS&&rh/320>=.99&&rh/320>=fh/320-.005&&rt.at(-1)<100&&ft.at(-1)<100;
  allPass&&=pass;
  console.log(JSON.stringify({repeat,step,runs:s.Runs,runMS,fullMS,runRecall:rh/320,fullRecall:fh/320,runMedianMS:rt[16],fullMedianMS:ft[16],runMaxMS:rt.at(-1),fullMaxMS:ft.at(-1),pass}));
 }
}
assert(out.FinalCloseNS>0,'final drain');
console.log(JSON.stringify({allPass,finalCloseMS:out.FinalCloseNS/1e6}));
