import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {performance} from 'node:perf_hooks';
import {poolWeights,strongPool} from './strong-brier-pool.mjs';
import {incrementalPoolWeights} from './incremental-pool.mjs';
const [source,screenPath,output]=process.argv.slice(2);assert(source&&screenPath&&output);
const old=new Map(JSON.parse(fs.readFileSync(screenPath)).results.map(r=>[r.key,r]));
const tapes=[];let header=true;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})) {
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[r.Phase,r.Case,r.Index,r.Schedule].join(':'),ref=old.get(key);assert(ref);old.delete(key);
  const rows=r.Steps.map(s=>({p:[12,0,1,2,3,10].map(a=>s.P[a]),y:s.Y,delay:s.Delay,missing:s.Missing}));
  tapes.push({rows,schedule:r.Schedule,reference:ref.forecasts[0]});
}
assert.equal(tapes.length,672);assert.equal(old.size,0);
const prior=[.5,.1,.1,.1,.1,.1];let verified=0,transitions=0;
// Validation/warm-up is outside timed rounds; all served forecasts must match.
for(const t of tapes) {
  const fast=incrementalPoolWeights(t.rows,prior),slow=poolWeights(t.rows,prior);
  assert.deepEqual(fast.weights,slow);transitions+=fast.transitions;
  t.rows.forEach((r,i)=>{assert.equal(strongPool(r.p,fast.weights[i]),t.reference[i]);verified++;});
}
const rounds=[];let checksum=0;
for(let repeat=0;repeat<3;repeat++) {
  // Alternate order to avoid assigning all first-position costs to one method.
  for(const name of repeat%2===0?['reference','incremental']:['incremental','reference']) {
    const ms=[0,0];
    for(const t of tapes) {
      const start=performance.now();
      const w=name==='reference'?poolWeights(t.rows,prior):incrementalPoolWeights(t.rows,prior).weights;
      ms[t.schedule]+=performance.now()-start;checksum+=w[255][0];
    }
    rounds.push({repeat,name,millisecondsBySchedule:ms,microsecondsPerForecastBySchedule:ms.map(v=>v*1000/(336*256))});
  }
}
const out={verifiedForecasts:verified,referenceTransitions:672*256*257/2,incrementalTransitions:transitions,rounds,checksum,
  complexity:'O(T*K*(D+1)) for maximum delivery lag D (D=0 still next-issue update), O(T*K) batch cache; unbounded delays retain O(T^2*K) worst case.',
  limitations:'Same-source warm weights-only comparison, excludes aggregation, I/O, model fitting, scoring, persistence and concurrent serving. No runtime/production integration. All172032 strong forecasts checked bit-identical.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(out,null,2));
