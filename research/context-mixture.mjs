import assert from 'node:assert/strict';
import {directMixture} from './direct-mixture.mjs';

function features(x,context){
  assert(Number.isInteger(x)&&x>=0&&x<512);
  return [1,...Array.from({length:9},(_,j)=>context?((x>>j)&1?1:-1)/3:0)];
}
function solve(a,b){
  const n=b.length,l=Array.from({length:n},()=>Array(n).fill(0));
  for(let i=0;i<n;i++)for(let j=0;j<=i;j++){
    let s=a[i][j];for(let k=0;k<j;k++)s-=l[i][k]*l[j][k];
    if(i===j){assert(s>0&&Number.isFinite(s));l[i][j]=Math.sqrt(s);}else l[i][j]=s/l[j][j];
  }
  const y=Array(n).fill(0),x=Array(n).fill(0);
  for(let i=0;i<n;i++){let s=b[i];for(let j=0;j<i;j++)s-=l[i][j]*y[j];y[i]=s/l[i][i];}
  for(let i=n-1;i>=0;i--){let s=y[i];for(let j=i+1;j<n;j++)s-=l[j][i]*x[j];x[i]=s/l[i][i];}
  return x;
}

export function contextMixture(rows,{context=true,trace=false}={}){
  const phis=rows.map(r=>features(r.x,context));
  return rows.map((r,t)=>{
    const available=[];for(let j=0;j<t;j++)if(!rows[j].missing&&j+rows[j].delay<=t)available.push(j);
    const origins=available.slice(-64),a=Array.from({length:10},(_,i)=>Array.from({length:10},(_,j)=>Number(i===j))),b=Array(10).fill(0);
    for(const j of origins){const s=rows[j],d=s.c-s.b,target=Number(s.y)-(s.b+s.c)/2,z=phis[j].map(x=>x*d);
      for(let k=0;k<10;k++){b[k]+=z[k]*target;for(let l=0;l<10;l++)a[k][l]+=z[k]*z[l];}
    }
    const theta=solve(a,b),w=Math.max(0,Math.min(1,.5+phis[t].reduce((s,x,j)=>s+x*theta[j],0)));
    const out={w,origins};if(trace)Object.assign(out,{a,b,theta});return out;
  });
}

export function testContextMixture(){
  const rows=Array.from({length:96},(_,i)=>({x:(i*71)%512,b:.1+(i%7)*.08,c:.9-(i%5)*.1,y:i%3===0,delay:i%11,missing:i%7===0}));
  let maxResidual=0,maxReference=0;
  for(const context of [false,true]){
    const out=contextMixture(rows,{context,trace:true});
    const poison=rows.map((r,i)=>({...r,y:i>=40||r.missing||i+r.delay>40?!r.y:r.y,x:i>40?r.x^511:r.x}));
    assert.deepEqual(out.slice(0,41),contextMixture(poison,{context,trace:true}).slice(0,41));
    const scalar=directMixture(rows,{window:64});
    for(let t=0;t<rows.length;t++){
      const f=out[t],want=Array.from({length:t},(_,j)=>j).filter(j=>!rows[j].missing&&j+rows[j].delay<=t).slice(-64);assert.deepEqual(f.origins,want);
      if(!context)assert(Math.abs(f.w-scalar[t].w)<1e-12);
      for(let i=0;i<10;i++)maxResidual=Math.max(maxResidual,Math.abs(f.a[i].reduce((s,x,j)=>s+x*f.theta[j],0)-f.b[i]));
      // Independent Gauss-Seidel solve, not the Cholesky factorization above.
      const beta=Array(10).fill(0);
      for(let it=0;it<1000;it++)for(let i=0;i<10;i++){let s=f.b[i];for(let j=0;j<10;j++)if(j!==i)s-=f.a[i][j]*beta[j];beta[i]=s/f.a[i][i];}
      for(let i=0;i<10;i++)maxReference=Math.max(maxReference,Math.abs(beta[i]-f.theta[i]));
    }
  }
  assert(maxResidual<1e-11&&maxReference<1e-11);
  const positive=Array.from({length:96},(_,i)=>({x:i%2?511:0,b:0,c:1,y:!!(i%2),delay:0,missing:false}));
  const ctx=contextMixture(positive),global=contextMixture(positive,{context:false});
  const loss=p=>p.slice(64).reduce((s,r,i)=>s+(r.w-Number(positive[64+i].y))**2,0)/32;
  assert(loss(ctx)<.01&&loss(global)>.2);
  assert(contextMixture(rows.map(r=>({...r,c:r.b}))).every(f=>f.w===.5));
  assert(contextMixture(rows.map(r=>({...r,missing:true,y:NaN}))).every(f=>f.w===.5));
  console.log(JSON.stringify({tests:'context objective/as-of/global-equivalence/positive-control PASS',maxResidual,maxReference,positiveContext:loss(ctx),positiveGlobal:loss(global)}));
}
