import assert from 'node:assert/strict';

const rates=[1/128,1/64,1/32,1/16,1/8,1/4,1/2];
const psi=rates.map(l=>-Math.log1p(-l)-l);
const logMeanExp=xs=>{const m=Math.max(...xs);return m+Math.log(xs.reduce((s,x)=>s+Math.exp(x-m),0)/xs.length);};

// Fixed-grid adaptation of the time-varying EB martingale. The scale is fixed
// across time; adapting it would change the unweighted average target.
export class LoggedGainEBCS {
  #n=0; #sum=0; #v=0; #alpha; #allExact=true;
  constructor(alpha=.05){assert(Number.isFinite(alpha)&&alpha>0&&alpha<1);this.#alpha=alpha;}
  add(index,prepared,action,loss){
    assert(index===this.#n&&Number.isSafeInteger(index),'duplicate or out-of-order evidence');
    assert(Number.isFinite(prepared.lower)&&Number.isFinite(prepared.upper)&&prepared.lower>=-3&&prepared.upper<=3&&prepared.lower<=prepared.upper,'unsupported predictable range');
    const value=prepared.observe(action,loss);
    assert(Number.isFinite(value)&&value>=prepared.lower&&value<=prepared.upper);
    const center=this.#n?Math.max(-1,Math.min(1,this.#sum/this.#n)):0;
    const sum=this.#sum+value,v=this.#v+((value-center)/4)**2;
    assert(Number.isFinite(sum)&&Number.isFinite(v));
    this.#sum=sum;this.#v=v;this.#n++;
    this.#allExact &&= prepared.lower===prepared.upper;
    return this.interval();
  }
  interval(){
    if(!this.#n)return {n:0,mean:null,lower:-1,upper:1,residualSquares:0,scale:4};
    const mean=this.#sum/this.#n;
    if(this.#allExact)return {n:this.#n,mean,lower:mean,upper:mean,residualSquares:this.#v,scale:4};
    const threshold=Math.log(2)-Math.log(this.#alpha),e=b=>logMeanExp(rates.map((l,i)=>l*b-psi[i]*this.#v));
    let lo=0,hi=1;while(e(hi)<threshold){hi*=2;assert(Number.isFinite(hi));}
    for(let i=0;i<80;i++){const mid=(lo+hi)/2;if(e(mid)<threshold)lo=mid;else hi=mid;}
    const radius=4*hi/this.#n;
    return {n:this.#n,mean,lower:Math.max(-1,mean-radius),upper:Math.min(1,mean+radius),residualSquares:this.#v,scale:4};
  }
}
