import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {contextMixture,testContextMixture} from './context-mixture.mjs';
import {pointwiseGuard} from './pointwise-brier-guard.mjs';
testContextMixture();
const[source,input,output]=process.argv.slice(2);assert(source&&input&&output);
const old=JSON.parse(fs.readFileSync(input));assert.equal(old.summary.cohort,'spike-independent-v1');
const map=new Map(old.results.map(r=>[r.key,r])),results=[];let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const src=JSON.parse(line);if(header){assert.equal(src.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[src.Phase,src.Case,src.Index,src.Schedule].join(':'),r=map.get(key);assert(r);map.delete(key);
  const rows=src.Steps.map((s,i)=>({x:s.X,b:s.P[12],c:r.armPredictions[6][i],y:s.Y,delay:s.Delay,missing:s.Missing}));
  const ctx=contextMixture(rows),global=contextMixture(rows,{context:false}),expected=Array(7).fill(0),realized=Array(7).fill(0),blocks=Array.from({length:8},()=>({expected:Array(7).fill(0),realized:Array(7).fill(0)}));
  for(let i=0;i<256;i++){
    const{b,c}=rows[i],q=src.Steps[i].Q,y=Number(rows[i].y);
    const ps=[pointwiseGuard(b,c,ctx[i].w).p,pointwiseGuard(b,c,global[i].w).p,pointwiseGuard(b,c,r.forecasts[i].proposed).p,r.forecasts[i].p,b,b+ctx[i].w*(c-b),b+global[i].w*(c-b)];
    for(let a=0;a<7;a++){
      const e=(ps[a]-q)**2+q*(1-q),l=(ps[a]-y)**2;expected[a]+=e/256;realized[a]+=l/256;blocks[Math.floor(i/32)].expected[a]+=e/32;blocks[Math.floor(i/32)].realized[a]+=l/32;
      if(a<3)assert((ps[a]-b)*(ps[a]+b-2*q)<=.01+1e-12);
    }
  }
  results.push({key,expected,realized,blocks});
}
assert.equal(results.length,672);assert.equal(map.size,0);
const arms=['contextGuard','globalGuard','sharePointwise','shareLocal','Markov','contextRaw','globalRaw'];
const summary={cohort:'consumed-context-mixture-v1',records:672,arms:arms.map((arm,a)=>({arm,expected:results.reduce((s,r)=>s+r.expected[a]/672,0),terminal:results.reduce((s,r)=>s+(r.blocks[6].expected[a]+r.blocks[7].expected[a])/1344,0),windowHarms:results.reduce((s,r)=>s+r.blocks.filter(b=>b.expected[a]-b.expected[4]>.01+1e-12).length,0)})),limitations:'Consumed-data contextual ridge screen. No fresh or real-agent quality claim, no inherited regret guarantee.'};
fs.writeFileSync(output,JSON.stringify({summary,results},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
