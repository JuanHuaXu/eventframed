import assert from 'node:assert/strict';
import {prepareLoggedGain,LoggedGainCS} from './logged-gain.mjs';
const near=(a,b)=>assert(Math.abs(a-b)<1e-11,`${a} != ${b}`);
let identities=0,mgfs=0;
for(const p0 of [.05,.2,.5,.9])for(const q0 of [0,.2,.7,1])for(const q1 of [0,.1,.8,1])for(const m0 of [0,.3,1])for(const m1 of [0,.6,1]){
  const p=[p0,1-p0],q=[q0,q1],m=[m0,m1],a=prepareLoggedGain(p,[0,1],[1,0],m),truth=q0-q1;
  let mean=0;
  for(let j=0;j<2;j++)for(let y=0;y<2;y++)mean+=p[j]*(y?q[j]:1-q[j])*a.observe(j,y);
  near(mean,truth);identities++;
  for(const lambda of [.01,.1,.5,1])for(const sign of [-1,1]){
    let expectation=0;
    for(let j=0;j<2;j++)for(let y=0;y<2;y++)expectation+=p[j]*(y?q[j]:1-q[j])*Math.exp(sign*lambda*(a.observe(j,y)-truth)-lambda*lambda*(a.upper-a.lower)**2/8);
    assert(expectation<=1+1e-12);mgfs++;
  }
}
const logging=[.5,.5],candidate=[0,1],control=[1,0],prediction=[.1,.8];
const owned=prepareLoggedGain(logging,candidate,control,prediction),before=owned.observe(0,1);
logging[0]=0;candidate[1]=0;control[0]=0;prediction[0]=1;near(owned.observe(0,1),before);
assert.throws(()=>prepareLoggedGain([1,0],[0,1],[1,0],[.5,.5]));
assert.throws(()=>prepareLoggedGain([.5,.5],[0,1],[1,0],[NaN,.5]));
assert.throws(()=>owned.observe(2,.1));assert.throws(()=>owned.observe(0,1.1));
const cs=new LoggedGainCS(),prepared=prepareLoggedGain([.5,.5],[0,1],[1,0],[.5,.5]);
cs.add(0,prepared,0,.5);const snapshot=cs.interval();
assert.throws(()=>cs.add(0,prepared,0,.5));assert.deepEqual(cs.interval(),snapshot);
assert.throws(()=>cs.add(1,prepared,0,NaN));assert.deepEqual(cs.interval(),snapshot);
const zero=new LoggedGainCS();assert.deepEqual(zero.add(0,prepareLoggedGain([1],[1],[1],[.4]),0,1),{n:1,mean:0,lower:0,upper:0,widthSum:0});
for(const rates of [[1e308],[1e-308],[],Array(65).fill(.1)])assert.throws(()=>new LoggedGainCS(.05,rates));
const tinyAlpha=new LoggedGainCS(Number.MIN_VALUE);assert.deepEqual(tinyAlpha.add(0,prepared,0,.5),{n:1,mean:0,lower:-1,upper:1,widthSum:4});
// Exhaustive adaptive two-action experiment, including every stopping time
// on a six-step tree. Truth varies with past observations, not just with time.
let leaves=0,noncoverage=0;
function tree(history,weight,rows,gainSum,ever){
  if(history.length===6){leaves++;if(ever)noncoverage+=weight;return;}
  const h=history.length,last=h?history.at(-1)%2:0,p=h%2?[.8,.2]:[.3,.7];
  const q=last?[.8,.2]:[.2,.6],m=[(h+1)/8,(7-h)/8],prep=prepareLoggedGain(p,[0,1],[1,0],m);
  for(let a=0;a<2;a++)for(let y=0;y<2;y++){
    const newRows=[...rows,{prep,a,y}],state=new LoggedGainCS(.2);
    newRows.forEach((r,i)=>state.add(i,r.prep,r.a,r.y));
    const interval=state.interval(),g=gainSum+q[0]-q[1],truth=g/(h+1);
    tree([...history,2*a+y],weight*p[a]*(y?q[a]:1-q[a]),newRows,g,ever||truth<interval.lower-1e-12||truth>interval.upper+1e-12);
  }
}
tree([],1,[],0,false);assert(noncoverage<=.2+1e-12);
console.log(JSON.stringify({status:'PASS',identities,mgfs,adaptiveTreeLeaves:leaves,anytimeNoncoverage:noncoverage,scope:'Exact DR expectation and Hoeffding MGF enumeration, adaptive finite-tree check, ownership/positivity/atomicity; not policy efficacy'}));
