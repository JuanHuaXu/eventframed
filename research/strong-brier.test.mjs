import assert from 'node:assert/strict';
import {strongBrier,brierWeights} from './strong-brier.mjs';
import {delayedShare} from './delayed-fixed-share.mjs';
let checks=0,maxDefect=0,linearDefect=0,maxPathError=0;
for(let bi=0;bi<=10;bi++)for(let ci=0;ci<=10;ci++)for(let wi=0;wi<=10;wi++) {
  const b=bi/10,c=ci/10,w=wi/10,p=strongBrier(b,c,w);
  const g=[0,1].map(y=>-Math.log((1-w)*Math.exp(-2*(b-y)**2)+w*Math.exp(-2*(c-y)**2)));
  // Literal paper substitution: solve sum_y (s-g_y)_+ = 2.
  let lo=Math.min(...g),hi=Math.max(...g)+2;
  for(let i=0;i<70;i++){const s=(lo+hi)/2;if(g.reduce((v,x)=>v+Math.max(0,s-x),0)>2)hi=s;else lo=s;}
  assert(Math.abs(p-Math.max(0,((lo+hi)/2-g[1])/2))<1e-12);
  for(const y of [0,1]) {
    const defect=2*(p-y)**2-g[y];maxDefect=Math.max(maxDefect,defect);assert(defect<1e-12);
    linearDefect=Math.max(linearDefect,2*((1-w)*b+w*c-y)**2-g[y]);checks++;
  }
  assert(Math.abs(strongBrier(1-b,1-c,w)-(1-p))<1e-12);
  if(b===c)assert(Math.abs(p-b)<1e-12);
}
assert(linearDefect>.01,'negative control must detect invalid linear substitution');
for(let bits=0;bits<256;bits++) {
  const rows=Array.from({length:8},(_,j)=>({b:.1+j*.04,c:.85-j*.03,y:!!(bits&(1<<j)),delay:j%3,missing:j===2}));
  assert.deepEqual(brierWeights(rows,1),delayedShare(rows));
  const w=brierWeights(rows),poison=rows.map((r,j)=>({...r,y:j>=4||r.missing||j+r.delay>4?!r.y:r.y}));
  for(let t=0;t<rows.length;t++) {
    let total=0,active=0;
    for(let path=0;path<2**(t+1);path++) {
      let mass=.5;
      for(let j=0;j<=t;j++) {
        const state=(path>>j)&1,r=rows[j];
        if(j>0){const a=1/(j+1);mass*=a*.5+(1-a)*Number(state===((path>>(j-1))&1));}
        if(j<t&&!r.missing&&j+r.delay<=t)mass*=Math.exp(-2*((state?r.c:r.b)-Number(r.y))**2);
      }
      total+=mass;if((path>>t)&1)active+=mass;
    }
    maxPathError=Math.max(maxPathError,Math.abs(w[t]-active/total));assert(maxPathError<1e-12);checks++;
  }
  assert.deepEqual(w.slice(0,5),brierWeights(poison).slice(0,5));
  const immediate=rows.map(r=>({...r,delay:0,missing:false})),ws=brierWeights(immediate,2,false);
  let loss=0,lb=0,lc=0;
  immediate.forEach((r,j)=>{const y=Number(r.y);loss+=(strongBrier(r.b,r.c,ws[j])-y)**2;lb+=(r.b-y)**2;lc+=(r.c-y)**2;assert(loss-Math.min(lb,lc)<=Math.log(2)/2+1e-12);checks++;});
}
for(const x of [[NaN,.5,.5],[0,1,2],[-.1,0,.5]])assert.throws(()=>strongBrier(...x));
console.log(JSON.stringify({checks,maxDefect,linearDefect,maxPathError,asOfPaths:256,limits:'Finite component tests. Static immediate-feedback bound only; no delayed/missing/sharing theorem.'}));
