import assert from 'node:assert/strict';

// Hindsight quantities are intentionally absent: this validates branches only.
export function auditEnvelope(r, source) {
  assert(!r.Error); assert.equal(source.Steps.length,256);
  for(const k of ['Phase','Case','Index','Schedule'])assert.equal(r[k],source[k]);
  const pool=Array.from({length:8},(_,i)=>152+i).filter(j=>source.Steps[j].Missing||j+source.Steps[j].Delay>160);
  const choices=[-1,...pool]; assert.deepEqual(r.Choices,choices);
  assert.equal(r.Bundles.length,Math.ceil(choices.length/4));
  const natural=Array.from({length:16},(_,i)=>i-16);
  for(let j=0;j<161;j++)if(!source.Steps[j].Missing&&j+source.Steps[j].Delay<=161)natural.push(j);
  let fits=0,checks=0; const branches=[];
  const allSupports=new Map();
  for(let b=0;b<r.Bundles.length;b++){
    const bundle=r.Bundles[b];assert(!bundle.Error);
    for(const k of ['Phase','Case','Index','Schedule'])assert.equal(bundle[k],source[k]);
    assert.equal(bundle.Decision.Clock,160);
    for(const key of ['Origins','Redundant','AtPublication','Predictions','LogEvidence'])assert.equal(bundle[key].length,4);
    const supports=new Set();
    for(let arm=0;arm<4;arm++){
      const selected=choices[Math.min(b*4+arm,choices.length-1)];
      assert.equal(bundle.Decision.Selected[arm],selected);assert.equal(bundle.Decision.Costs[arm],Number(selected>=0));
      const support=[...new Set([...natural,...(selected<0?[]:[selected])])].sort((a,b)=>a-b).slice(-64);
      assert.deepEqual(bundle.Origins[arm],support);
      assert.equal(bundle.Redundant[arm],selected>=0&&natural.includes(selected));
      assert(Number.isFinite(bundle.LogEvidence[arm]));
      const at=bundle.AtPublication[arm],pred=bundle.Predictions[arm];assert.equal(at.length,31);assert.equal(pred.length,31);
      for(let i=0;i<31;i++){
        assert(Number.isFinite(at[i])&&at[i]>0&&at[i]<1);assert(Number.isFinite(pred[i])&&pred[i]>0&&pred[i]<1);
        assert(Math.abs(pred[i]-(.5+.99**i*(at[i]-.5)))<=1e-14);checks++;
      }
      const item={selected,cost:Number(selected>=0),support,redundant:bundle.Redundant[arm],at,pred,logEvidence:bundle.LogEvidence[arm]};
      const key=support.join(',');supports.add(key);
      if(allSupports.has(key)){const old=allSupports.get(key);assert.deepEqual(at,old.at);assert.deepEqual(pred,old.pred);assert.equal(item.logEvidence,old.logEvidence);}
      allSupports.set(key,item);
      if(b*4+arm<choices.length)branches.push(item);
    }
    assert.equal(bundle.ActualFits,supports.size);fits+=bundle.ActualFits;
  }
  assert.equal(r.ActualFits,fits);assert.equal(branches.length,choices.length);
  if(source.Schedule===0)assert.deepEqual(choices,[-1]);
  return {branches,fits,checks};
}
