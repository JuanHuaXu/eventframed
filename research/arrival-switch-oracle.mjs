import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input, output] = process.argv.slice(2);
const raw = fs.readFileSync(input);
const [, ...rows] = raw.toString().trim().split('\n').map(JSON.parse);
const pop = x => { let n=0; for(;x;x&=x-1)n++; return n; };
const truth = (r, x, step) => {
  let majority = r.Case==='stable_majority3'||r.Case==='majority_to_parity';
  let m=r.Masks[0];
  if (!r.Case.startsWith('stable_') && step>=256) { majority=!majority;m=r.Masks[1]; }
  const n=pop(x&m);
  return (majority?n>=2:n%2===1)?.95:.05;
};
// This oracle knows latent truth at each step and can pick different weights
// each time. It is an unattainable lower bound, never a deployable selector.
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of ['majority_to_parity','parity_to_majority'])for(const schedule of ['Immediate','Delayed'])for(const start of [256,384]){
  const a=rows.filter(r=>r.Phase===phase&&r.Case===name);assert.equal(a.length,16);
  let actual=0,oracle=0,neutralOnly=0,n=0,split=0;
  for(const r of a){
    split+=Number(r[schedule].Arms[2].SplitAt>=0);
    for(let i=start;i<start+128;i++){
      const f=r[schedule].Frames[i],p=f.Predictions[2],q=truth(r,f.X,i);
      assert.ok(p.experts.length===4&&p.experts.every(v=>v>=0&&v<=1));
      assert.equal(p.experts[3],.5);
      const lo=Math.min(...p.experts),hi=Math.max(...p.experts);
      const best=Math.max(lo,Math.min(hi,q));
      actual+=(p.p-q)**2+q*(1-q);
      oracle+=(best-q)**2+q*(1-q);
      neutralOnly+=Number(best===.5);n++;
    }
  }
  cells.push({phase,case:name,schedule,start,n,actualExpectedBrier:actual/n,oracleExpectedBrier:oracle/n,neutralOptimalFraction:neutralOnly/n,splitTrajectories:split,trajectories:a.length});
}
const result={rawSHA256:crypto.createHash('sha256').update(raw).digest('hex'),limits:'Post-hoc truth-informed per-frame convex-hull oracle on issued views. No attainable learning or counterfactual acquisition claim. Neutral is already an expert. Correlated frames are not independent replicates.',cells};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.phase==='cohort2'&&c.schedule==='Delayed')));
