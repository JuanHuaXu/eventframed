import assert from 'node:assert/strict';
import {acquisitionBelief,acquisitionChoice,TYPES} from './interval-acquisition.mjs';
function legendre(z){let a=1,b=z;for(let n=2;n<=6;n++){const c=((2*n-1)*z*b-(n-1)*a)/n;a=b;b=c;}return[b,6*(z*b-a)/(z*z-1)];}
const quadrature=Array.from({length:6},(_,i)=>{let z=Math.cos(Math.PI*(i+.75)/6.5);for(let k=0;k<50;k++){const[p,d]=legendre(z),step=p/d;z-=step;if(Math.abs(step)<1e-15)break;}const[,d]=legendre(z);return{noise:.1+.1*(z+1),weight:1/((1-z*z)*d*d)};});
let checks=0,maxError=0,choices=0;
for(let bits=0;bits<16;bits++){
  const history=TYPES.map((t,i)=>({action:[0,t,0],outcome:!!(bits&(1<<i))}));
  for(let step=0;step<=6;step++){
    const expected=Array(4).fill(0);
    for(const {noise,weight} of quadrature)for(let mask=0;mask<16;mask++)for(let h=0;h<16;h++){
      let w=weight/16*(.5/16+(mask===0||mask===15?.25:0));const roots={};
      for(const {action:[kind,t],outcome} of history){
        if(kind===1&&(mask&(1<<TYPES.indexOf(t)))){w*=Number(outcome===roots[t]);continue;}
        const truth=t===7?((h%2)^Math.floor(h/4)%2):Math.floor(h/2**t)%2;
        const error=t===2?.01:noise,p=truth?1-error:error;w*=outcome?p:1-p;
        if(kind===0)roots[t]=outcome;
      }
      expected[h%4]+=w;
    }
    const total=expected.reduce((s,v)=>s+v,0),b=acquisitionBelief(history);
    assert(Math.abs(b.mass-total)<1e-12);
    expected.forEach((v,i)=>{const e=Math.abs(v/total-b.forecast[i]);assert(e<1e-12);maxError=Math.max(maxError,e);checks++;});
    if(step===6)break;
    const before=JSON.stringify(history),cache=new Map();
    for(const policy of ['target','entropy']){
      const noRandom=()=>{throw Error('Unexpected generator access');};
      assert.deepEqual(acquisitionChoice(history,policy,noRandom,cache),acquisitionChoice(history,policy,noRandom));choices++;
    }
    assert.equal(JSON.stringify(history),before);
    const t=[0,7,2,1,0,7][step],slot=history.filter(s=>s.action[0]===1&&s.action[1]===t).length;
    history.push({action:[1,t,slot],outcome:!!((bits+step)%2)});
  }
}
console.log(JSON.stringify({forecastChecks:checks,maxError,cachedChoiceChecks:choices}));

