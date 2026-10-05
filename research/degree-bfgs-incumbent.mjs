import assert from 'node:assert/strict';
import {createReadStream} from 'node:fs';
import {createInterface} from 'node:readline';

export function useBFGS(oldObjective,newObjective) {
 if(!Number.isFinite(oldObjective)||!Number.isFinite(newObjective))throw new Error('finite objectives required');
 return newObjective<=oldObjective;
}
assert.equal(useBFGS(10,9),true);assert.equal(useBFGS(9,10),false);assert.equal(useBFGS(9,9),true);
assert.equal(useBFGS(-10,-11),true);assert.throws(()=>useBFGS(0,NaN));
const poisoned={OldStats:{Final:2},Stats:{Final:1},get Q(){throw new Error('future access');},get Y(){throw new Error('future access');}};
assert.equal(useBFGS(poisoned.OldStats.Final,poisoned.Stats.Final),true);
let n=0,converged=0,fallbacks=0,objectiveNonharm=0;
const groups=Array.from({length:4},()=>({n:0,converged:0,fallbacks:0,meanEvaluations:0,brierGain:0,realizedGain:0}));
for await(const line of createInterface({input:createReadStream('docs/experiments/mmm-degree-bfgs-v1.jsonl'),crlfDelay:Infinity}))if(line.trim()){
 const r=JSON.parse(line),use=useBFGS(r.OldStats.Final,r.Stats.Final);
 const stats=use?r.Stats:r.OldStats,p=use?r.Predictions:r.OldPredictions;
 const ok=stats.Stop==='projected-gradient'&&stats.ProjectedGradient<=1e-6;
 const group=groups[Number(r.LearnNoise)*2+r.Window];group.n++;n++;
 objectiveNonharm+=Number(stats.Final<=r.OldStats.Final);converged+=Number(ok);fallbacks+=Number(!use);
 group.converged+=Number(ok);group.fallbacks+=Number(!use);group.meanEvaluations+=r.Stats.Evaluations+r.OldStats.Evaluations;
 for(let i=0;i<32;i++){group.brierGain+=((r.OldPredictions[i]-r.Q[i])**2-(p[i]-r.Q[i])**2)/32;group.realizedGain+=((r.OldPredictions[i]-Number(r.Y[i]))**2-(p[i]-Number(r.Y[i]))**2)/32;}
}
assert.equal(n,1008);assert.equal(objectiveNonharm,n);
for(const g of groups)for(const k of ['meanEvaluations','brierGain','realizedGain'])g[k]/=g.n;
console.log(JSON.stringify({n,converged,fallbacks,objectiveNonharm,optimizationScreenPass:converged/n>=.95&&objectiveNonharm===n,groups},null,2));
