import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {pointwiseGuard} from './pointwise-brier-guard.mjs';
const[source,input,output]=process.argv.slice(2);assert(source&&input&&output);
const old=JSON.parse(fs.readFileSync(input)),map=new Map(old.results.map(r=>[r.key,r]));let header=true,count=0;
const scores=Array(4).fill(0),terminal=Array(4).fill(0),byRegime={stationary:{count:0,scores:Array(4).fill(0)},changing:{count:0,scores:Array(4).fill(0)}};
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const s=JSON.parse(line);if(header){assert.equal(s.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[s.Phase,s.Case,s.Index,s.Schedule].join(':'),r=map.get(key);assert(r);map.delete(key);
  const changing=s.Case<9?s.Case%3!==0:s.Case>=19,regime=byRegime[changing?'changing':'stationary'];
  for(let i=0;i<256;i++){
    const b=s.Steps[i].P[12],c=r.armPredictions[6][i],q=s.Steps[i].Q,d=c-b,cap=pointwiseGuard(b,c,1).w;
    const w=d===0?0:Math.max(0,Math.min(cap,(q-b)/d)),p=b+w*d;
    for(const test of [0,cap,cap/2])assert((p-q)**2<=(b+test*d-q)**2+1e-12);
    const ps=[p,pointwiseGuard(b,c,r.forecasts[i].proposed).p,r.forecasts[i].p,b];
    for(let a=0;a<4;a++){const risk=(ps[a]-q)**2+q*(1-q);scores[a]+=risk;regime.scores[a]+=risk;if(i>=192)terminal[a]+=risk;}
    count++;regime.count++;
  }
}
assert.equal(map.size,0);assert.equal(count,172032);
for(const r of Object.values(byRegime))r.scores=r.scores.map(x=>x/r.count);
const summary={arms:['infeasibleOracleWithinPointwiseBound','sharePointwise','shareLocal','Markov'],count,scores:scores.map(x=>x/count),terminal:terminal.map(x=>x/(672*64)),byRegime,
  limitations:'Evaluator-only exact conditional minimizer along the fixed baseline/challenger segment under the fixed pointwise bound. Uses inaccessible oracle q. Not a deployable algorithm, causal claim, or new learner validation.'};
fs.writeFileSync(output,JSON.stringify(summary,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
