import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,output]=process.argv.slice(2),raw=fs.readFileSync(input),praw=fs.readFileSync(parent);
const [, ...rows]=raw.toString().trim().split('\n').map(JSON.parse),[, ...parents]=praw.toString().trim().split('\n').map(JSON.parse);
const key=r=>[r.Phase,r.Case,r.Index].join('/'),pm=new Map(parents.map(r=>[key(r),r]));
const prior=[.7,.1,.1,.1],cells=[];
for(const phase of ['cohort1','cohort2'])for(const schedule of ['Immediate','Delayed']){
  let n=0,weights=[0,0,0,0],movement=0,changedViews=0;
  for(const r of rows.filter(r=>r.Phase===phase&&r.Case==='parity_to_majority')){
    const p=pm.get(key(r))[schedule];
    for(let i=256;i<512;i++){
      const a=r[schedule].Frames[i].Predictions[2],b=p.Frames[i].Predictions[2];
      if(!a.split||b.split)continue;
      let w=a.weights_before_share;const sum=w.reduce((s,v)=>s+v,0);w=sum>0?w.map(v=>v/sum):prior;
      const effective=w.map((v,j)=>.998*v+.002*prior[j]);
      const pred=effective.reduce((s,v,j)=>s+v*Math.max(1e-6,Math.min(1-1e-6,a.experts[j])),0);
      assert.ok(Math.abs(pred-a.p)<1e-12);
      weights=weights.map((v,j)=>v+effective[j]);movement+=Math.abs(a.p-b.p);changedViews+=Number(a.mask!==b.mask);n++;
    }
  }
  cells.push({phase,schedule,earlierSplitFrames:n,meanEffectiveWeights:n?weights.map(v=>v/n):null,meanAbsForecastMovement:n?movement/n:null,changedViewFraction:n?changedViews/n:null});
}
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
fs.writeFileSync(output,JSON.stringify({rawSHA256:hash(raw),parentSHA256:hash(praw),cells,limits:'Consumed earlier-split interval description; frames are dependent, no causal mediation decomposition or counterfactual substitution.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells));
