import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {directMixture} from './direct-mixture.mjs';
import {pointwiseGuard} from './pointwise-brier-guard.mjs';
const[source,input,output]=process.argv.slice(2);assert(source&&input&&output);
const old=JSON.parse(fs.readFileSync(input)),map=new Map(old.results.map(r=>[r.key,r]));
const configs=[{}, {window:64},{currentCovariance:true},{window:64,currentCovariance:true}];
const stats=configs.map(()=>({forecasts:0,clipped:0,sameAsShareWithin1e12:0,sameAfterBothClipped:0,absoluteProposalDifference:0,absoluteForecastDifference:0}));let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const s=JSON.parse(line);if(header){assert.equal(s.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[s.Phase,s.Case,s.Index,s.Schedule].join(':'),r=map.get(key);assert(r);map.delete(key);
  const rows=s.Steps.map((s,i)=>({b:s.P[12],c:r.armPredictions[6][i],y:s.Y,delay:s.Delay,missing:s.Missing}));
  configs.forEach((config,a)=>{
    const ws=directMixture(rows,config),stat=stats[a];
    rows.forEach((row,i)=>{
      const f=pointwiseGuard(row.b,row.c,ws[i].w),reference=pointwiseGuard(row.b,row.c,r.forecasts[i].proposed);
      const same=Math.abs(f.p-reference.p)<1e-12;
      stat.forecasts++;stat.clipped+=Number(f.clipped);stat.sameAsShareWithin1e12+=Number(same);stat.sameAfterBothClipped+=Number(same&&f.clipped&&reference.clipped);
      stat.absoluteProposalDifference+=Math.abs(ws[i].w-r.forecasts[i].proposed);stat.absoluteForecastDifference+=Math.abs(f.p-reference.p);
    });
  });
}
assert.equal(map.size,0);
for(const s of stats){assert.equal(s.forecasts,172032);s.absoluteProposalDifference/=s.forecasts;s.absoluteForecastDifference/=s.forecasts;}
fs.writeFileSync(output,JSON.stringify({order:['ridgeAll','ridge64','currentCovAll','currentCov64'],stats,limitations:'Consumed-data deterministic saturation diagnosis, not a causal performance experiment.'},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(stats,null,2));
