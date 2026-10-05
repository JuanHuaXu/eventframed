import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {pointwiseGuard,testPointwiseGuard} from './pointwise-brier-guard.mjs';
testPointwiseGuard();
const[source,input,output]=process.argv.slice(2);assert(source&&input&&output);
const old=JSON.parse(fs.readFileSync(input,'utf8'));assert.equal(old.summary.cohort,'spike-independent-v1');
const map=new Map(old.results.map(r=>[r.key,r])),results=[];let header=true,changes=0,maxExcess=-Infinity;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const s=JSON.parse(line);if(header){assert.equal(s.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[s.Phase,s.Case,s.Index,s.Schedule].join(':'),r=map.get(key);assert(r);map.delete(key);
  const expected=Array(7).fill(0),realized=Array(7).fill(0),blocks=Array.from({length:8},()=>({expected:Array(7).fill(0),realized:Array(7).fill(0)}));
  let weight=0,clipped=0;
  for(let i=0;i<256;i++){
    const b=s.Steps[i].P[12],c=r.armPredictions[6][i],q=s.Steps[i].Q,y=Number(s.Steps[i].Y);
    const f=pointwiseGuard(b,c,r.forecasts[i].proposed),half=pointwiseGuard(b,c,.5),full=pointwiseGuard(b,c,1);
    const ps=[f.p,r.forecasts[i].p,b,half.p,full.p,(b+c)/2,c];
    changes+=Number(f.p!==b);weight+=f.w/256;clipped+=Number(f.clipped);
    for(let a=0;a<7;a++){
      const e=(ps[a]-q)**2+q*(1-q),l=(ps[a]-y)**2;
      expected[a]+=e/256;realized[a]+=l/256;blocks[Math.floor(i/32)].expected[a]+=e/32;blocks[Math.floor(i/32)].realized[a]+=l/32;
      if([0,3,4].includes(a)){const excess=(ps[a]-b)*(ps[a]+b-2*q);assert(excess<=.01+1e-12);if(a===0)maxExcess=Math.max(maxExcess,excess);}
    }
  }
  results.push({key,expected,realized,blocks,meanWeight:weight,clipped});
}
assert.equal(results.length,672);assert.equal(map.size,0);
const names=['pointwiseShare','localShare','Markov','pointwiseHalf','pointwiseFull','unclippedHalf','raw'];
const summary={cohort:'consumed-pointwise-screen-v1',records:672,changes,maxExcess,arms:names.map((arm,a)=>({arm,expected:results.reduce((s,r)=>s+r.expected[a]/672,0),terminal:results.reduce((s,r)=>s+(r.blocks[6].expected[a]+r.blocks[7].expected[a])/1344,0),windowHarms:results.reduce((s,r)=>s+r.blocks.filter(b=>b.expected[a]-b.expected[2]>.01+1e-12).length,0)})),
  meanWeight:results.reduce((s,r)=>s+r.meanWeight/672,0),limitations:'Consumed-data diagnostic safety/utility screen. Endpoint theorem concerns binary Brier excess only; no fresh quality or adaptation evidence.'};
fs.writeFileSync(output,JSON.stringify({summary,results},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
