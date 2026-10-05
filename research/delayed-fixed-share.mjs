import assert from 'node:assert/strict';

const alpha=j=>j===0?0:1/(j+1);
const arrived=(r,j,i)=>j<i&&!r.missing&&j+r.delay<=i;
const update=(w,r)=>{
  const a=w*Math.exp(-((r.c-Number(r.y))**2));
  const b=(1-w)*Math.exp(-((r.b-Number(r.y))**2));
  return a/(a+b);
};

// Re-filter in origin order. Late evidence changes its historical factor,
// then propagates forward through every subsequent transition.
export function delayedShare(rows,share=alpha) {
  return rows.map((_,i)=>{
    let w=.5;
    for(let j=0;j<=i;j++) {
      const a=share(j);assert(a>=0&&a<=1);
      w=(1-a)*w+a*.5;
      if(arrived(rows[j],j,i))w=update(w,rows[j]);
    }
    assert(Number.isFinite(w)&&w>=0&&w<=1);
    return w;
  });
}

function enumeration(rows,i){
  let mass=0,active=0;
  for(let path=0;path<(1<<(i+1));path++){
    let p=.5;
    for(let j=0;j<=i;j++){
      const state=(path>>j)&1;
      if(j>0)p*=alpha(j)*.5+(1-alpha(j))*Number(state===((path>>(j-1))&1));
      if(arrived(rows[j],j,i)){const forecast=state?rows[j].c:rows[j].b;p*=Math.exp(-((forecast-Number(rows[j].y))**2));}
    }
    mass+=p;if((path>>i)&1)active+=p;
  }
  return active/mass;
}

export function testDelayedShare(){
  let count=0,maxError=0;
  for(let bits=0;bits<64;bits++){
    const rows=Array.from({length:6},(_,i)=>({b:.1+.1*(i%3),c:.85-.05*(i%2),y:!!(bits&(1<<i)),delay:i%3,missing:i===1}));
    const ws=delayedShare(rows);
    for(let i=0;i<6;i++){const error=Math.abs(ws[i]-enumeration(rows,i));maxError=Math.max(maxError,error);assert(error<1e-12);count++;}
    const poison=rows.map((r,i)=>({...r,y:i>=3||r.missing||i+r.delay>3?!r.y:r.y}));
    assert.deepEqual(ws.slice(0,4),delayedShare(poison).slice(0,4));
    const staticW=delayedShare(rows,()=>0);
    for(let i=0;i<6;i++){let odds=0;for(let j=0;j<i;j++)if(arrived(rows[j],j,i))odds+=(rows[j].b-rows[j].c)*(rows[j].b+rows[j].c-2*Number(rows[j].y));assert(Math.abs(staticW[i]-1/(1+Math.exp(-odds)))<1e-12);}
  }
  const delayed=Array.from({length:8},()=>({b:.1,c:.9,y:true,delay:20,missing:false}));
  delayed[0].delay=5;
  const correct=delayedShare(delayed);
  assert(correct.slice(0,5).every(w=>w===.5));assert(correct[5]>.5);
  // Applying the old loss directly to the current weight omits intervening
  // switches and therefore is deliberately not the same posterior.
  assert(Math.abs(correct[5]-update(.5,delayed[0]))>.01);
  assert(delayedShare(delayed.map(r=>({...r,missing:true,y:NaN}))).every(w=>w===.5));
  assert(delayedShare(delayed.map(r=>({...r,c:r.b,delay:0}))).every(w=>w===.5));
  const future=delayed.map((r,i)=>({...r,y:i>=5?!r.y:r.y}));
  assert.deepEqual(correct.slice(0,6),delayedShare(future).slice(0,6));
  console.log(`delayed-share exact-path tests PASS: ${count} prefixes, max error ${maxError}`);
}
