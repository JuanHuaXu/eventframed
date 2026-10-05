import assert from 'node:assert/strict';

export function fitQueryShrinkage(rows){
  assert(rows.length>0);let numerator=0,denominator=0,weight=0;
  for(const r of rows){
    assert(Number.isFinite(r.p)&&r.p>=0&&r.p<=1&&Number.isFinite(r.y)&&r.y>=0&&r.y<=1&&Number.isFinite(r.weight)&&r.weight>0);
    const x=r.p-.5;numerator+=r.weight*x*(r.y-.5);denominator+=r.weight*x*x;weight+=r.weight;
  }
  return {alpha:denominator>0?Math.max(0,Math.min(1,numerator/denominator)):0,numerator,denominator,weight,rows:rows.length};
}

export function shrinkQueryMass(p,alpha){
  assert(Number.isFinite(p)&&p>=0&&p<=1&&Number.isFinite(alpha)&&alpha>=0&&alpha<=1);
  if(alpha===1)return p;
  return .5+alpha*(p-.5);
}

// Conditional forecasts are held fixed. This is a query-order heuristic, not a
// claim that recalibrating weights leaves the incumbent predictive law intact.
export function queryMixtureScore(conditional,p){
  assert(Array.isArray(conditional)&&conditional.length===2&&conditional[0].length>0&&conditional[0].length===conditional[1].length);
  assert(Number.isFinite(p)&&p>=0&&p<=1);let sensitivity=0;
  for(let j=0;j<conditional[0].length;j++){
    const a=conditional[0][j],b=conditional[1][j];assert(Number.isFinite(a)&&a>=0&&a<=1&&Number.isFinite(b)&&b>=0&&b<=1);sensitivity+=(b-a)**2/conditional[0].length;
  }
  return p*(1-p)*sensitivity;
}

export function chooseShrinkage(values,alpha){
  if(values.length===0)return {origin:-1,scores:[]};
  assert(new Set(values.map(v=>v.Origin)).size===values.length);
  const scores=values.map(v=>queryMixtureScore(v.Conditional,shrinkQueryMass(v.Mass[1],alpha))),max=Math.max(...scores);
  const origin=Math.min(...values.filter((_,j)=>max-scores[j]<=1e-10).map(v=>v.Origin));
  return {origin,scores};
}
