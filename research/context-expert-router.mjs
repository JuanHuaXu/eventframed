import assert from 'node:assert/strict';

const tables=new Map();
function indexTable(d){
  if(tables.has(d))return tables.get(d);
  const size=1<<d,index=new Uint16Array(size*size);
  for(let x=0;x<size;x++)for(let m=0;m<size;m++){
    let c=0,power=1;for(let j=0;j<d;j++){if(m&(1<<j))c+=power*((x&(1<<j))?2:1);power*=3;}
    index[x*size+m]=c;
  }
  tables.set(d,index);return index;
}
export function contextMaskPrior(d,contextual=true){
  assert(Number.isInteger(d)&&d>=0&&d<=9);
  if(!contextual||d===0)return [1];
  const zero=(2/3)**d;
  return Array.from({length:1<<d},(_,m)=>{
    if(m===0)return .95;
    let k=0;for(let v=m;v;v>>=1)k+=v&1;
    return .05*(1/3)**k*(2/3)**(d-k)/(1-zero);
  });
}

// Finite joint model: one latent input mask; an independent latent expert
// per projected cell. Every admitted outcome contributes once per alternative
// mask model, NOT multiple observations inside one model. Issued expert laws
// remain fixed when feedback arrives. This is a research-only bounded router.
export function createContextRouter(d,prior,contextual=true,horizon=256){
  assert(Number.isInteger(d)&&d>=0&&d<=9&&typeof contextual==='boolean');
  assert(Number.isInteger(horizon)&&horizon>=1&&horizon<=256);
  assert(Array.isArray(prior)&&prior.length>=2&&prior.length<=8&&prior.every(x=>Number.isFinite(x)&&x>=1e-12));
  assert(Math.abs(prior.reduce((s,x)=>s+x,0)-1)<1e-12);
  const totalPrior=prior.reduce((s,x)=>s+x,0),pi=prior.map(x=>x/totalPrior),K=pi.length,size=1<<d,maskPrior=contextMaskPrior(d,contextual),M=maskPrior.length;
  const cells=contextual?3**d:1,index=contextual?indexTable(d):null,logPi=pi.map(Math.log);
  const logWeights=new Float64Array(cells*K),posterior=new Float64Array(cells*K),cellZ=new Float64Array(cells),seen=new Uint8Array(cells),maskZ=new Float64Array(M);
  let maskWeights=maskPrior.slice(),evidence=0,next=0,accepted=0;const journal=[];
  function validate(x,ps){assert(Number.isInteger(x)&&x>=0&&x<size);assert(Array.isArray(ps)&&ps.length===K&&ps.every(p=>Number.isFinite(p)&&p>0&&p<1));}
  const cell=(x,m)=>contextual?index[x*size+m]:0;
  function predict(x,ps){
    validate(x,ps);const weights=Array(K).fill(0);
    for(let m=0;m<M;m++){
      const c=cell(x,m),base=c*K;
      for(let k=0;k<K;k++)weights[k]+=maskWeights[m]*(seen[c]?posterior[base+k]:pi[k]);
    }
    const z=weights.reduce((s,w)=>s+w,0);for(let k=0;k<K;k++)weights[k]/=z;
    const p=weights.reduce((s,w,k)=>s+w*ps[k],0);
    return {p:Math.max(Math.min(...ps),Math.min(Math.max(...ps),p)),weights};
  }
  function issue(x,ps){
    assert(next<horizon);const result=predict(x,ps),id=next++;
    journal.push({x,ps:ps.slice(),status:'pending'});return {id,...result};
  }
  function deliver(id,y){
    assert(Number.isInteger(id)&&id>=0&&id<next&&typeof y==='boolean');const row=journal[id];
    if(row.status==='delivered'){assert.equal(row.y,y,'conflicting outcome');return false;}
    assert.equal(row.status,'pending','expired outcome');
    const logL=row.ps.map(p=>y?Math.log(p):Math.log1p(-p));
    for(let m=0;m<M;m++){
      const c=cell(row.x,m),base=c*K;let max=-Infinity;
      for(let k=0;k<K;k++){logWeights[base+k]+=logL[k];max=Math.max(max,logPi[k]+logWeights[base+k]);}
      let z=0;for(let k=0;k<K;k++){const v=Math.exp(logPi[k]+logWeights[base+k]-max);posterior[base+k]=v;z+=v;}
      const newZ=max+Math.log(z);maskZ[m]+=newZ-cellZ[c];cellZ[c]=newZ;seen[c]=1;
      // Log likelihoods remain intact even if a normalized mass underflows;
      // later contradictory evidence can revive the expert.
      for(let k=0;k<K;k++)posterior[base+k]/=z;
    }
    let max=-Infinity;for(let m=0;m<M;m++)max=Math.max(max,Math.log(maskPrior[m])+maskZ[m]);
    let z=0;maskWeights=Array.from(maskZ,(v,m)=>{const w=Math.exp(Math.log(maskPrior[m])+v-max);z+=w;return w;});
    evidence=max+Math.log(z);for(let m=0;m<M;m++)maskWeights[m]/=z;
    row.status='delivered';row.y=y;delete row.ps;accepted++;return true;
  }
  function expireBefore(cutoff){assert(Number.isInteger(cutoff)&&cutoff>=0&&cutoff<=next);for(let j=0;j<cutoff;j++)if(journal[j].status==='pending')journal[j]={status:'expired'};}
  return {predict,issue,deliver,expireBefore,summary:()=>({next,accepted,logEvidence:evidence,globalMass:maskWeights[0]}),inspect:()=>({next,accepted,logEvidence:evidence,maskWeights:maskWeights.slice(),journal:structuredClone(journal)})};
}
