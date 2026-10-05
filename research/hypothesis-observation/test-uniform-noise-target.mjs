import assert from 'node:assert/strict';
import {noiseEnvelope} from './noise-envelope.mjs';
import {uniformNoiseTarget} from './uniform-noise-target.mjs';
function legendre(z) {
  let a=1,b=z;for(let n=2;n<=6;n++){const next=((2*n-1)*z*b-(n-1)*a)/n;a=b;b=next;}
  return [b,6*(z*b-a)/(z*z-1)];
}
const quadrature=Array.from({length:6},(_,i)=>{
  let z=Math.cos(Math.PI*(i+.75)/6.5);
  for(let k=0;k<50;k++){const [p,d]=legendre(z),step=p/d;z-=step;if(Math.abs(step)<1e-15)break;}
  const [p,d]=legendre(z);assert(Math.abs(p)<1e-14);
  return {u:(z+1)/2,w:1/((1-z*z)*d*d)};
});
let polynomialError=0;
for(let k=0;k<=11;k++){const e=Math.abs(quadrature.reduce((s,{u,w})=>s+w*u**k,0)-1/(k+1));assert(e<5e-15);polynomialError=Math.max(polynomialError,e);}
const types=[0,1,2,7];let checks=0,maxError=0;
for(const counts of [[1,1,1,3],[2,2,1,1],[3,1,1,1]]){
  const schedule=types.map(t=>[0,t,0]);types.forEach((t,i)=>{for(let j=0;j<counts[i];j++)schedule.push([1,t,j]);});
  for(let bits=0;bits<1024;bits+=17){
    const history=schedule.map((action,i)=>({action,outcome:!!(bits&(1<<(9-i)))}));
    const target=uniformNoiseTarget(noiseEnvelope(history)),masses=Array(4).fill(0);
    for(const {u,w} of quadrature)for(let h=0;h<16;h++){
      let mass=w/16;const noise=.1+.2*u;
      for(const {action:[,t],outcome} of history){
        const truth=t===7?((h%2)^Math.floor(h/4)%2):Math.floor(h/2**t)%2;
        const error=t===2?.01:noise,p=truth?1-error:error;mass*=outcome?p:1-p;
      }
      masses[h%4]+=mass;
    }
    const total=masses.reduce((s,m)=>s+m,0);
    target.forEach((p,i)=>{const e=Math.abs(p-masses[i]/total);assert(e<1e-12);maxError=Math.max(maxError,e);checks++;});
  }
}
assert.throws(()=>uniformNoiseTarget([]));
assert.throws(()=>uniformNoiseTarget([{mask:0,total:[0],classCoefficients:[[0],[0],[0],[0]]}]));
console.log(JSON.stringify({quadratureMonomials:12,polynomialError,forecastChecks:checks,maxError,negativeTests:2}));

