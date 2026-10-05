import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const[source,output]=process.argv.slice(2);assert(source&&output);
const names=['generic64','Boolean64','generic32','Boolean32','ridge64','ridge32','tree64','tree32','variational64','variational32','segment64','segment32','Markov','static64','static32'];
const cells=new Map(),records=[];let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  const changing=r.Case<9?r.Case%3!==0:r.Case>=19,whole=Array(15).fill(0),terminal=Array(15).fill(0);
  for(let i=0;i<256;i++)for(let a=0;a<15;a++){
    const s=r.Steps[i],p=s.P[a],risk=(p-s.Q)**2+s.Q*(1-s.Q);whole[a]+=risk/256;if(i>=192)terminal[a]+=risk/64;
  }
  records.push({key:[r.Phase,r.Case,r.Index,r.Schedule].join(':'),whole,terminal});
  for(const cell of ['all',changing?'changing':'stationary',`case${r.Case}`]){
    if(!cells.has(cell))cells.set(cell,{records:0,whole:Array(15).fill(0),terminal:Array(15).fill(0)});
    const c=cells.get(cell);c.records++;for(let a=0;a<15;a++){c.whole[a]+=whole[a];c.terminal[a]+=terminal[a];}
  }
}
assert.equal(records.length,672);
const summaries=[];
for(const[cell,c]of cells)summaries.push({cell,records:c.records,heads:names.map((name,a)=>({name,whole:c.whole[a]/c.records,terminal:c.terminal[a]/c.records})).sort((a,b)=>a.whole-b.whole)});
const out={summaries,records,limitations:'Post-hoc ranking of existing as-of model heads on consumed synthetic data. No refits or oracle routing; no selected-model confirmation. All model construction costs remain charged to original source collection.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summaries.filter(s=>['all','stationary','changing'].includes(s.cell)),null,2));
