import assert from 'node:assert/strict';

const probability=x=>assert(Number.isFinite(x)&&x>=0&&x<=1);
function distribution(xs){assert(Array.isArray(xs)&&xs.length>0);xs.forEach(probability);assert(Math.abs(xs.reduce((a,b)=>a+b,0)-1)<1e-12);}

// Construct before drawing the logged action. The regression need not be
// correct, but logging probabilities must be correct and cover both policies.
export function prepareLoggedGain(logging,candidate,control,prediction){
  [logging,candidate,control].forEach(distribution);
  assert(Array.isArray(prediction)&&[candidate,control,prediction].every(x=>x.length===logging.length));
  prediction.forEach(probability);
  const p=[...logging],m=[...prediction],d=control.map((x,i)=>x-candidate[i]);
  const offset=d.reduce((s,x,i)=>s+x*m[i],0);
  let lower=Infinity,upper=-Infinity;
  for(let a=0;a<p.length;a++){
    assert(p[a]>0 || (candidate[a]===0 && control[a]===0),'unsupported policy action');
    if(!p[a])continue;
    for(const loss of [0,1]){const value=offset+d[a]/p[a]*(loss-m[a]);lower=Math.min(lower,value);upper=Math.max(upper,value);}
  }
  assert(Number.isFinite(lower)&&Number.isFinite(upper));
  return Object.freeze({lower,upper,observe(action,loss){
    assert(Number.isInteger(action)&&action>=0&&action<p.length&&p[action]>0);probability(loss);
    const value=offset+d[action]/p[action]*(loss-m[action]);assert(Number.isFinite(value));
    return value;
  }});
}

const logMeanExp=xs=>{const max=Math.max(...xs);return max+Math.log(xs.reduce((s,x)=>s+Math.exp(x-max),0)/xs.length);};

// Our bounded Hoeffding-mixture reference, not the source paper's tighter
// empirical-Bernstein construction. Range bounds must be predictable.
export class LoggedGainCS {
  #n=0; #sum=0; #v=0; #alpha; #rates;
  constructor(alpha=.05,rates=[.01,.02,.05,.1,.2,.5,1,2]){
    assert(Number.isFinite(alpha)&&alpha>0&&alpha<1);
    assert(Array.isArray(rates)&&rates.length>0&&rates.length<=64&&rates.every(x=>Number.isFinite(x)&&x>=1e-6&&x<=2));
    this.#alpha=alpha;this.#rates=[...rates];
  }
  add(index,prepared,action,loss){
    assert(index===this.#n,'duplicate or out-of-order evidence');
    const {lower,upper}=prepared;
    assert(Number.isFinite(lower)&&Number.isFinite(upper)&&lower<=upper);
    const value=prepared.observe(action,loss),width=upper-lower;
    assert(Number.isFinite(value)&&value>=lower-1e-12&&value<=upper+1e-12);
    const conservativeWidth=width===0?0:Math.max(width,1e-12);
    const sum=this.#sum+value,v=this.#v+conservativeWidth*conservativeWidth;
    assert(Number.isFinite(sum)&&Number.isFinite(v));
    this.#sum=sum;this.#v=v;this.#n++;
    return this.interval();
  }
  interval(){
    if(!this.#n)return {n:0,mean:null,lower:-1,upper:1,widthSum:0};
    const mean=this.#sum/this.#n;
    if(this.#v===0)return {n:this.#n,mean,lower:mean,upper:mean,widthSum:0};
    const threshold=Math.log(2)-Math.log(this.#alpha);
    const e=b=>logMeanExp(this.#rates.map(l=>l*b-(l*(l/8))*this.#v));
    let lo=0,hi=1;while(e(hi)<threshold){hi*=2;assert(Number.isFinite(hi));}
    for(let i=0;i<80;i++){const mid=(lo+hi)/2;if(e(mid)<threshold)lo=mid;else hi=mid;}
    const radius=hi/this.#n;
    // Do not intersect successive intervals: the average target may change.
    return {n:this.#n,mean,lower:Math.max(-1,mean-radius),upper:Math.min(1,mean+radius),widthSum:this.#v};
  }
}
