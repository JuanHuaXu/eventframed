import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const hash=bytes=>createHash('sha256').update(bytes).digest('hex');
const average=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const clamp=x=>Math.min(1,Math.max(0,x));
function exactUtility(monitored, fillers, k) {
  assert.ok(fillers.length>=k);
  let distribution=Array(k+1).fill(0);
  distribution[0]=1;
  let expected=0;
  for(const p of monitored) {
    assert.ok(p>=0 && p<=1);
    expected+=p*p*distribution.slice(0,k).reduce((a,b)=>a+b,0);
    const next=Array(k+1).fill(0);
    for(let successes=0;successes<=k;successes++) {
      next[successes]+=distribution[successes]*(1-p);
      next[Math.min(k,successes+1)]+=distribution[successes]*p;
    }
    distribution=next;
  }
  for(let successes=0;successes<k;successes++) {
    expected+=distribution[successes]*fillers.slice(0,k-successes).reduce((a,b)=>a+b,0);
  }
  assert.ok(Math.abs(distribution.reduce((a,b)=>a+b,0)-1)<1e-12);
  return expected/k;
}
function enumeratedUtility(monitored, fillers, k) {
  let total=0;
  for(let mask=0;mask<2**monitored.length;mask++) {
    let weight=1;
    const selected=[];
    for(let i=0;i<monitored.length;i++) {
      const useful=Boolean(mask&(1<<i));
      weight*=useful ? monitored[i] : 1-monitored[i];
      if(useful && selected.length<k) selected.push(monitored[i]);
    }
    selected.push(...fillers.slice(0,k-selected.length));
    total+=weight*average(selected);
  }
  return total;
}
for(const probabilities of [[.2,.7,.9,.1],[0,0,0,0],[1,1,1,1],[.01,.02,.03,.04]]) {
  for(const k of [1,2,3]) {
    const fillers=[.4,.6,.8];
    assert.ok(Math.abs(exactUtility(probabilities,fillers,k)-enumeratedUtility(probabilities,fillers,k))<1e-12);
  }
}
const recorded=JSON.parse(readFileSync(resolve(root,'docs/experiments/mmm-packet-innovation-v24-summary.json'),'utf8'));
const report={kind:'post-hoc exact conditional model diagnostic', exhaustiveSmallControls:12, splits:{}};
for(const split of ['design','confirmation']) {
  const raw=readFileSync(resolve(root,`docs/experiments/mmm-packet-innovation-v24-${split}.jsonl`));
  assert.equal(hash(raw),recorded[split].sha256);
  const rows=raw.toString('utf8').trim().split('\n').map(JSON.parse).slice(1);
  const regimes={};
  for(const row of rows) {
    const members=row.Candidates.slice(0,32), fillers=row.Candidates.slice(32);
    assert.equal(members.length,32);
    const floor=fillers[9].Base;
    const ceiling=Math.max(...fillers.map(c=>c.Base));
    for(let i=0;i<row.Candidates.length-1;i++) assert.ok(row.Candidates[i].Base>=row.Candidates[i+1].Base-1e-12);
    for(const c of members) {
      const positiveFlat=clamp(c.Base+.1*(2/3-.5));
      const negativeFlat=clamp(c.Base+.1*(1/3-.5));
      const positiveContext=clamp(c.Base+.1*((2*c.Base+1)/3-c.Base));
      const negativeContext=clamp(c.Base+.1*(2*c.Base/3-c.Base));
      assert.ok(positiveFlat>ceiling && positiveContext>ceiling,'positive separation');
      assert.ok(negativeFlat<floor && negativeContext<floor,'negative separation');
      for(const posterior of [1/3,2/3]) {
        for(const scale of [.5,2.5]) {
          const score=clamp(c.Base+.1*(posterior-c.Base)*scale);
          assert.ok(score<floor,'current separation throughout elastic range');
        }
      }
    }
    const baseline=average(row.Candidates.slice(0,10).map(c=>c.Probability));
    const current=average(fillers.slice(0,10).map(c=>c.Probability));
    const proposal=exactUtility(members.map(c=>c.Probability),fillers.map(c=>c.Probability),10);
    const item={world:row.World,baseline,current,proposal,gain:proposal-baseline,harm:baseline-proposal};
    (regimes[row.Regime]??=[]).push(item);
  }
  report.splits[split]={};
  for(const [regime,worlds] of Object.entries(regimes)) {
    assert.equal(worlds.length,8);
    report.splits[split][regime]={baseline:average(worlds.map(w=>w.baseline)),current:average(worlds.map(w=>w.current)),proposal:average(worlds.map(w=>w.proposal)),proposalGain:average(worlds.map(w=>w.gain)),proposalHarm:average(worlds.map(w=>w.harm)),worlds};
  }
}
const outIndex=process.argv.indexOf('--out');
if(outIndex>=0) writeFileSync(resolve(root,process.argv[outIndex+1]),JSON.stringify(report,null,2)+'\n',{flag:'wx'});
process.stdout.write(JSON.stringify(report,null,2)+'\n');
