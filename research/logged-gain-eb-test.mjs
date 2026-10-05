import assert from 'node:assert/strict';
import {prepareLoggedGain} from './logged-gain.mjs';
import {LoggedGainEBCS} from './logged-gain-eb.mjs';
let mgfs=0;
for(const p0 of [1/3,.5,2/3])for(const q0 of [0,.2,.7,1])for(const q1 of [0,.1,.8,1])for(const m of [[0,0],[.1,.9],[1,1]])for(const center of [-1,0,1]){
  const p=[p0,1-p0],q=[q0,q1],prepared=prepareLoggedGain(p,[0,1],[1,0],m),truth=q0-q1;
  for(const lambda of [1/128,1/32,1/8,1/2])for(const sign of [-1,1]){
    let expectation=0;
    for(let a=0;a<2;a++)for(let y=0;y<2;y++){
      const value=prepared.observe(a,y),residual=(value-center)/4;
      assert(sign*residual>=-1);
      expectation+=p[a]*(y?q[a]:1-q[a])*Math.exp(sign*lambda*(value-truth)/4-(-Math.log1p(-lambda)-lambda)*residual**2);
    }
    assert(expectation<=1+1e-12);mgfs++;
  }
}
const uncertain=prepareLoggedGain([.5,.5],[0,1],[1,0],[.5,.5]),cs=new LoggedGainEBCS();
const first=cs.add(0,uncertain,0,.5);assert(first.lower<0&&first.upper>0&&first.residualSquares===0);
assert.throws(()=>cs.add(0,uncertain,0,.5));assert.deepEqual(first,cs.interval());
assert.throws(()=>cs.add(1,uncertain,0,NaN));assert.deepEqual(first,cs.interval());
assert.throws(()=>cs.add(1,prepareLoggedGain([.01,.99],[0,1],[1,0],[0,0]),0,1));assert.deepEqual(first,cs.interval());
const zero=new LoggedGainEBCS();assert.equal(zero.add(0,prepareLoggedGain([1],[1],[1],[.4]),0,1).upper,0);
// A positive signal must eventually be detectable; zero residual alone must
// never create certainty when the declared support is not a singleton.
const power=new LoggedGainEBCS(.0125),positive=prepareLoggedGain([.5,.5],[0,1],[1,0],[1,0]);
for(let i=0;i<1000;i++)power.add(i,positive,i%2,i%2?0:1);
assert(power.interval().lower>0);assert(power.interval().lower<1);
let leaves=0,miss=0,mass=0;
function tree(rows,weight,truthSum,ever){
  if(rows.length===6){leaves++;mass+=weight;if(ever)miss+=weight;return;}
  const last=rows.length?rows.at(-1).y:0,p=last?[.6,.4]:[.4,.6],q=last?[.8,.2]:[.1,.7];
  const prep=prepareLoggedGain(p,[0,1],[1,0],[rows.length/8,1-rows.length/8]);
  for(let a=0;a<2;a++)for(let y=0;y<2;y++){
    const next=[...rows,{prep,a,y}],state=new LoggedGainEBCS(.2);next.forEach((r,i)=>state.add(i,r.prep,r.a,r.y));
    const interval=state.interval(),sum=truthSum+q[0]-q[1],truth=sum/next.length;
    tree(next,weight*p[a]*(y?q[a]:1-q[a]),sum,ever||truth<interval.lower||truth>interval.upper);
  }
}
tree([],1,0,false);assert(Math.abs(mass-1)<1e-12);assert(miss<=.2+1e-12);
console.log(JSON.stringify({status:'PASS',mgfs,adaptiveLeaves:leaves,anytimeMissProbability:miss,positiveControl:power.interval(),scope:'Finite EB MGF/coverage checks, predictable range/center, non-vacuous power and sequence/atomicity; not policy efficacy'}));
