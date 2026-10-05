import assert from 'node:assert/strict';
const kernel=Array.from({length:512},(_,x)=>{let bits=0;for(let v=x;v;v&=v-1)bits++;return Math.exp(-bits/2);});
const input=xs=>assert(Array.isArray(xs)&&xs.length>0&&xs.every(x=>Number.isInteger(x)&&x>=0&&x<512));
const checked=d=>{assert(Number.isFinite(d)&&d>=-1e-12);return Math.max(0,d);};
export function mmdDesignView(origins,pool,source){
  assert(Array.isArray(origins)&&origins.length<=63&&new Set(origins).size===origins.length);
  assert(Array.isArray(pool)&&pool.length<=8&&new Set(pool).size===pool.length);
  const history=origins.map(j=>{
    assert(Number.isInteger(j)&&j>=-16&&j<160);
    if(j<0)return source.Initial[j+16].Bits;
    const s=source.Steps[j];assert(!s.Missing&&Number.isInteger(s.Delay)&&s.Delay>=0&&j+s.Delay<=160);return s.X;
  });
  const target=Array.from({length:161},(_,i)=>source.Steps[i].X);
  const candidates=pool.map(j=>{assert(Number.isInteger(j)&&j>=152&&j<160);const s=source.Steps[j];assert(s.Missing||j+s.Delay>160);return s.X;});
  input(history);input(target);if(candidates.length)input(candidates);
  return {history,target,candidates};
}
export function mmdState(history,target){
  input(history);input(target);assert(history.length<=64&&target.length<=161);
  // Own the design multisets so caller mutation cannot alter a cached state.
  const h=history.slice(),t=target.slice(),n=h.length,m=t.length;
  let hh=0,ht=0,tt=0;
  for(const x of h){for(const y of h)hh+=kernel[x^y];for(const y of t)ht+=kernel[x^y];}
  for(const x of t)for(const y of t)tt+=kernel[x^y];
  const baseline=checked(hh/(n*n)+tt/(m*m)-2*ht/(n*m));
  return {baseline,append(x){
    assert(Number.isInteger(x)&&x>=0&&x<512);let hx=0,tx=0;for(const y of h)hx+=kernel[x^y];for(const y of t)tx+=kernel[x^y];
    const after=checked((hh+2*hx+1)/((n+1)**2)+tt/(m*m)-2*(ht+tx)/((n+1)*m));
    return {after,factor:baseline<=1e-12?1:1-.5*Math.sqrt(after/baseline)};
  }};
}
