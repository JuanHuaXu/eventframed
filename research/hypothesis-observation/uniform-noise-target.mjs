import assert from 'node:assert/strict';

// Every degree-d Bernstein basis term integrates to 1/(d+1) on [0,1].
// Integrate joint class/history masses BEFORE normalization; averaging
// normalized conditional forecasts would use the wrong noise posterior.
export function uniformNoiseTarget(models) {
  const genuine=models.find(m=>m.mask===0);assert(genuine);
  const mean=c=>c.reduce((s,x)=>s+x,0)/c.length;
  const mass=mean(genuine.total);assert(Number.isFinite(mass)&&mass>0);
  const forecast=genuine.classCoefficients.map(c=>mean(c)/mass);
  assert(forecast.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
  assert(Math.abs(forecast.reduce((s,p)=>s+p,0)-1)<1e-12);
  return forecast;
}

