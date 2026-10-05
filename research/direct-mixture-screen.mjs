import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {directMixture,testDirectMixture} from './direct-mixture.mjs';
import {pointwiseGuard,testPointwiseGuard} from './pointwise-brier-guard.mjs';
testDirectMixture();testPointwiseGuard();
const[source,input,output]=process.argv.slice(2);assert(source&&input&&output);
const old=JSON.parse(fs.readFileSync(input,'utf8'));assert.equal(old.summary.cohort,'spike-independent-v1');
const byKey=new Map(old.results.map(r=>[r.key,r])),results=[];let header=true;
const variants=[{}, {window:64},{currentCovariance:true},{window:64,currentCovariance:true}];
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const src=JSON.parse(line);if(header){assert.equal(src.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[src.Phase,src.Case,src.Index,src.Schedule].join(':'),old=byKey.get(key);assert(old);byKey.delete(key);
  const rows=src.Steps.map((s,i)=>({b:s.P[12],c:old.armPredictions[6][i],y:s.Y,delay:s.Delay,missing:s.Missing}));
  const weights=variants.map(v=>directMixture(rows,v)),expected=Array(11).fill(0),realized=Array(11).fill(0),blocks=Array.from({length:8},()=>({expected:Array(11).fill(0),realized:Array(11).fill(0)}));
  for(let i=0;i<256;i++){
    const{b,c}=rows[i],q=src.Steps[i].Q,y=Number(rows[i].y);
    const ps=weights.map(ws=>pointwiseGuard(b,c,ws[i].w).p);
    ps.push(pointwiseGuard(b,c,old.forecasts[i].proposed).p,old.forecasts[i].p,b,...weights.map(ws=>b+ws[i].w*(c-b)));
    for(let a=0;a<11;a++){
      const e=(ps[a]-q)**2+q*(1-q),l=(ps[a]-y)**2;
      expected[a]+=e/256;realized[a]+=l/256;blocks[Math.floor(i/32)].expected[a]+=e/32;blocks[Math.floor(i/32)].realized[a]+=l/32;
      if(a<5)assert((ps[a]-b)*(ps[a]+b-2*q)<=.01+1e-12);
    }
  }
  results.push({key,expected,realized,blocks});
}
assert.equal(results.length,672);assert.equal(byKey.size,0);
const names=['ridgeAll','ridge64','currentCovAll','currentCov64','sharePointwise','shareLocal','Markov','ridgeAllRaw','ridge64Raw','currentCovAllRaw','currentCov64Raw'];
const summary={cohort:'consumed-direct-mixture-v1',records:672,arms:names.map((arm,a)=>({arm,expected:results.reduce((s,r)=>s+r.expected[a]/672,0),terminal:results.reduce((s,r)=>s+(r.blocks[6].expected[a]+r.blocks[7].expected[a])/1344,0),windowHarms:results.reduce((s,r)=>s+r.blocks.filter(b=>b.expected[a]-b.expected[6]>.01+1e-12).length,0)})),limitations:'Consumed-data screen. As-of centered ridge/current-covariance adaptations, not an inherited AAR regret result or proof of faster adaptation.'};
fs.writeFileSync(output,JSON.stringify({summary,results},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
