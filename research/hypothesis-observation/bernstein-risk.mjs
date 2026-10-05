import assert from 'node:assert/strict';
export function split(c,t){
  assert(c.length>0&&c.every(Number.isFinite)&&t>=0&&t<=1);
  let row=c.slice();const left=[row[0]],right=[row.at(-1)];
  while(row.length>1){row=row.slice(0,-1).map((v,i)=>(1-t)*v+t*row[i+1]);left.push(row[0]);right.push(row.at(-1));}
  return [left,right.reverse()];
}
export const evaluate=(c,t)=>split(c,t)[0].at(-1);
export function restrict(c,a,b){
  assert(a>=0&&a<=b&&b<=1);
  if(a===b)return c.map(()=>evaluate(c,a));
  const left=b===1?c:split(c,b)[0];return a===0?left:split(left,a/b)[1];
}
export function bounds(c,depth=10){
  assert(Number.isInteger(depth)&&depth>=0);
  if(!depth)return {lower:Math.min(...c),upper:Math.max(...c)};
  const [a,b]=split(c,.5),x=bounds(a,depth-1),y=bounds(b,depth-1);
  return {lower:Math.min(x.lower,y.lower),upper:Math.max(x.upper,y.upper)};
}
