import assert from 'node:assert/strict';

// Multiplication by a nonnegative affine likelihood in the Bernstein basis.
// Degree elevation coefficients preserve positivity across the whole interval.
export function multiplyAffine(coefficients, left, right) {
  assert(coefficients.length>0 && coefficients.every(x=>Number.isFinite(x)&&x>=0));
  assert([left,right].every(x=>Number.isFinite(x)&&x>=0&&x<=1));
  const n=coefficients.length;
  return Array.from({length:n+1},(_,k)=>
    (k<n?(n-k)/n*coefficients[k]*left:0)
    +(k>0?k/n*coefficients[k-1]*right:0));
}

export function reportProbability(type,h,noise) {
  assert([0,1,2,7].includes(type));
  const truth=type===7?((h&1)^((h>>2)&1)):((h>>type)&1);
  const error=type===2?.01:noise;
  return truth?1-error:error;
}

// This reference supports only the frozen four-source acquisition contract.
// Copied renewals reuse the ordinary root; no unobserved report becomes evidence.
export function noiseEnvelope(history,types=[0,1,2,7],lower=.1,upper=.3) {
  assert.deepEqual(types,[0,1,2,7]);
  assert(0<lower&&lower<=upper&&upper<.5);
  assert(history.length>=4&&history.length<=10);
  const roots=new Map(),slots=new Map();
  history.forEach(({action,outcome},i)=>{
    assert.equal(typeof outcome,'boolean');
    const [kind,t,slot]=action;assert.equal(action.length,3);assert(types.includes(t));
    if(i<4){assert.equal(kind,0);assert.equal(t,types[i]);assert.equal(slot,0);roots.set(t,outcome);}
    else {assert.equal(kind,1);assert.equal(slot,slots.get(t)||0);slots.set(t,slot+1);}
  });
  const models=[];
  for(let mask=0;mask<16;mask++){
    const copied=s=>s.action[0]===1&&(mask&(1<<types.indexOf(s.action[1])));
    if(history.some(s=>copied(s)&&s.outcome!==roots.get(s.action[1])))continue;
    const coefficients=Array.from({length:16},(_,h)=>{
      let c=[1/16];
      for(const s of history){
        if(copied(s))continue;
        const probabilities=[lower,upper].map(n=>reportProbability(s.action[1],h,n));
        c=multiplyAffine(c,...probabilities.map(p=>s.outcome?p:1-p));
      }
      return c;
    });
    const total=coefficients[0].map((_,k)=>coefficients.reduce((sum,c)=>sum+c[k],0));
    const classCoefficients=Array.from({length:4},(_,y)=>total.map((_,k)=>coefficients.reduce((sum,c,h)=>sum+(h%4===y?c[k]:0),0)));
    const laws=total.map((mass,k)=>{assert(mass>0);return classCoefficients.map(c=>c[k]/mass);});
    models.push({mask,coefficients,total,classCoefficients,laws});
  }
  assert(models.length>0);
  return models;
}

