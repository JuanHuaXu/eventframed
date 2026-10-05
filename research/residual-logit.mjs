import assert from 'node:assert/strict';

export const probabilityFloor=1e-12;
const sigmoid=z=>z>=0?1/(1+Math.exp(-z)):Math.exp(z)/(1+Math.exp(z));
const dot=(a,b)=>a.reduce((s,x,i)=>s+x*b[i],0);
const nll=(z,y)=>{const a=y?-z:z;return Math.max(a,0)+Math.log1p(Math.exp(-Math.abs(a)));};
function features(ps,full){
  assert(Array.isArray(ps)&&ps.length===8&&ps.every(p=>Number.isFinite(p)&&p>=0&&p<=1));
  const u=ps.map(p=>{p=Math.max(probabilityFloor,Math.min(1-probabilityFloor,p));return Math.log(p)-Math.log1p(-p);});
  return {offset:u[0],x:full?[1,u[0],...u.slice(1).map(v=>v-u[0])]:[1,u[0]]};
}
function prepare(rows,full){
  assert(typeof full==='boolean'&&Array.isArray(rows)&&rows.length<=256);
  return rows.map(r=>{assert(typeof r.y==='boolean');return {...features(r.ps,full),y:r.y};});
}
function objective(rows,beta){return dot(beta,beta)/2+rows.reduce((s,r)=>s+nll(r.offset+dot(r.x,beta),r.y),0);}
function derivatives(rows,beta){
  const d=beta.length,g=beta.slice(),h=Array.from({length:d},(_,i)=>Array.from({length:d},(_,j)=>i===j?1:0));
  for(const r of rows){
    const a=r.offset+dot(r.x,beta),p=sigmoid(a),complement=sigmoid(-a),residual=r.y?-complement:p,w=p*complement;
    for(let i=0;i<d;i++){g[i]+=residual*r.x[i];for(let j=0;j<=i;j++)h[i][j]+=w*r.x[i]*r.x[j];}
  }
  for(let i=0;i<d;i++)for(let j=0;j<i;j++)h[j][i]=h[i][j];
  return {g,h};
}
export function solvePositive(h,g){
  const d=g.length,L=Array.from({length:d},()=>Array(d).fill(0)),y=Array(d).fill(0),out=Array(d).fill(0);
  for(let i=0;i<d;i++){
    for(let j=0;j<=i;j++){
      let v=h[i][j];for(let k=0;k<j;k++)v-=L[i][k]*L[j][k];
      if(i===j){assert(v>0&&Number.isFinite(v),'non-positive Hessian');L[i][j]=Math.sqrt(v);}else L[i][j]=v/L[j][j];
    }
    let v=g[i];for(let j=0;j<i;j++)v-=L[i][j]*y[j];y[i]=v/L[i][i];
  }
  for(let i=d-1;i>=0;i--){let v=y[i];for(let j=i+1;j<d;j++)v-=L[j][i]*out[j];out[i]=v/L[i][i];assert(Number.isFinite(out[i]));}
  return out;
}
function objectiveDelta(rows,beta,delta){
  let sum=0,compensation=0;
  const add=v=>{const y=v-compensation,t=sum+y;compensation=(t-sum)-y;sum=t;};
  delta.forEach((d,i)=>add(beta[i]*d+d*d/2));
  for(const r of rows){
    let a=r.offset+dot(r.x,beta),da=dot(r.x,delta);
    if(Math.abs(da)<=1){if(r.y){a=-a;da=-da;}add(Math.log1p(sigmoid(a)*Math.expm1(da)));}
    else add(nll(a+da,r.y)-nll(a,r.y));
  }
  return sum;
}
export function residualLogitDerivatives(rows,beta,full=true){
  const prepared=prepare(rows,full);assert(beta.length===(full?9:2)&&beta.every(Number.isFinite));
  return {...derivatives(prepared,beta),objective:objective(prepared,beta)};
}
// Port the repo ridge fitter's stable line-search rule, adding an immutable
// baseline offset and forecast-disagreement features. This is a penalized
// likelihood plug-in, not an integrated Gaussian posterior predictive.
export function fitResidualLogit(rows,full=true){
  const prepared=prepare(rows,full);let beta=Array(full?9:2).fill(0),evaluations=0;
  for(let iteration=0;iteration<=64;iteration++){
    const {g,h}=derivatives(prepared,beta),gradientInf=Math.max(...g.map(Math.abs));
    assert(Number.isFinite(gradientInf));
    if(gradientInf<=1e-8){
      const fitted=beta.slice();
      return {beta:fitted.slice(),iterations:iteration,evaluations,gradientInf,objective:objective(prepared,fitted),predict:ps=>{
        const f=features(ps,full),p=sigmoid(f.offset+dot(f.x,fitted));
        return Math.max(probabilityFloor,Math.min(1-probabilityFloor,p));
      }};
    }
    assert(iteration<64,'iteration cap');
    const direction=solvePositive(h,g),descent=dot(g,direction);assert(descent>0&&Number.isFinite(descent));
    let accepted=false,step=1;
    for(let backtrack=0;backtrack<24;backtrack++){
      const delta=direction.map(v=>-step*v),trial=beta.map((v,i)=>v+delta[i]),change=objectiveDelta(prepared,beta,delta);evaluations++;
      if(Number.isFinite(change)&&change<=-1e-4*step*descent&&trial.every(Number.isFinite)){beta=trial;accepted=true;break;}
      step/=2;
    }
    assert(accepted,`line search exhausted: gradient=${gradientInf}`);
  }
  throw Error('unreachable');
}
