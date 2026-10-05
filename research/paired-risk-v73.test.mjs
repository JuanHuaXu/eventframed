import assert from 'node:assert/strict';
import test from 'node:test';
import {binomialUpper,pairedUpper} from './paired-risk-v73.mjs';

test('analytic special cases, input rejection and nonzero uncertainty',()=>{
  assert.ok(Math.abs(binomialUpper(1,2,.05)-Math.sqrt(.95))<1e-12);
  assert.ok(Math.abs(binomialUpper(0,512,.05)-(1-Math.pow(.05,1/512)))<1e-14);
  assert.equal(binomialUpper(3,3,.05),1);
  assert.ok(pairedUpper(0,0,512,.05).upper>0);
  assert.ok(pairedUpper(0,512,512,.05).upper<0);
  for(const args of [[0,0,.05],[-1,3,.05],[4,3,.05],[1,3,0],[1,3,1],[1,3,NaN],[1.5,3,.05]])assert.throws(()=>binomialUpper(...args));
  for(const args of [[-1,0,4,.05],[3,2,4,.05],[.5,0,4,.05]])assert.throws(()=>pairedUpper(...args));
  for(const alpha of [0,1,1.5,2,NaN,Infinity])assert.throws(()=>pairedUpper(0,0,4,alpha));
});

function multinomial(n,h,b,ph,pb) {
  const counts=[h,b,n-h-b],probs=[ph,pb,Math.max(0,1-ph-pb)];
  let lp=0;
  for(let i=1;i<=n;i++)lp+=Math.log(i);
  for(let j=0;j<3;j++) {
    for(let i=1;i<=counts[j];i++)lp-=Math.log(i);
    if(counts[j]) { if(probs[j]===0)return 0;lp+=counts[j]*Math.log(probs[j]); }
  }
  return Math.exp(lp);
}

test('exhaustive outcome enumeration across frozen parameter grid',()=>{
  let checked=0,worstRatio=0;
  const grid=[0,.01,.05,.1,.2,.4,.6,.8,.95,1];
  for(const n of [1,4,8,16,32])for(const alpha of [.05,.01]) {
    const outcomes=[];
    for(let h=0;h<=n;h++)for(let b=0;b<=n-h;b++) {
      const bound=pairedUpper(h,b,n,alpha);
      assert.ok(bound.upper>=-1&&bound.upper<=1&&bound.sLower<=bound.sUpper+1e-12);
      outcomes.push({h,b,upper:bound.upper});
    }
    for(const ph of grid)for(const pb of grid) {
      if(ph+pb>1+1e-14)continue;
      let total=0,noncoverage=0;
      for(const {h,b,upper}of outcomes) {
        const probability=multinomial(n,h,b,ph,pb);total+=probability;
        if(upper<ph-pb-1e-12)noncoverage+=probability;
      }
      assert.ok(Math.abs(total-1)<1e-10,`mass ${n} ${ph} ${pb}: ${total}`);
      assert.ok(noncoverage<=alpha+1e-10,`undercoverage ${n} ${ph} ${pb}: ${noncoverage}`);
      worstRatio=Math.max(worstRatio,noncoverage/alpha);checked++;
    }
  }
  console.log(JSON.stringify({parameterCells:checked,worstNoncoverageToAlpha:worstRatio}));
});
