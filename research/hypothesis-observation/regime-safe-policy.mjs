import assert from 'node:assert/strict';
import {actualJoint} from './exact-regime.mjs';
const worlds=[.1,.15,.2,.25,.3].flatMap(noise=>Array.from({length:16},(_,mask)=>({noise,mask})));
const slack=1e-12;
export function addRegimeSafePolicy(model){
  const stats={states:0,changed:0,eligibleActions:0,supportChecks:0,branchChecks:0,maxFinalExcess:0,maxAreaSumExcess:0};
  for(const root of model.roots){
    const list=root.states,index=new Map(list.map((s,i)=>[s.counts.join(','),i]));
    const n=list.length,w=worlds.length;
    const mass=new Float64Array(n*w),loss=new Float64Array(n*w);
    const baseF=new Float64Array(n*w),baseA=new Float64Array(n*w),ownF=new Float64Array(n*w),ownA=new Float64Array(n*w);
    const priorCost=new Float64Array(n),risk=list.map(s=>1-s.forecast.reduce((v,p)=>v+p*p,0));
    for(let i=0;i<n;i++)for(let g=0;g<w;g++){
      const s=list[i],joint=actualJoint(root.pattern,s.counts,worlds[g].noise,worlds[g].mask),m=joint.reduce((v,p)=>v+p,0),k=i*w+g;
      mass[k]=m;
      if(m)loss[k]=joint.reduce((v,p,y)=>v+p*s.forecast.reduce((t,q,z)=>t+(q-Number(z===y))**2,0),0)/m;
    }
    const order=Array.from({length:n},(_,i)=>i).sort((a,b)=>list[b].counts.reduce((s,v)=>s+v,0)-list[a].counts.reduce((s,v)=>s+v,0));
    for(const i of order){
      const s=list[i],remaining=6-s.counts.reduce((v,c)=>v+c,0);
      if(!remaining){for(let g=0;g<w;g++){const k=i*w+g;baseF[k]=ownF[k]=loss[k];}continue;}
      stats.states++;
      const children=Array.from({length:4},(_,a)=>[0,1].map(y=>{const c=s.counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);return j;}));
      const candidates=Array.from({length:4},()=>({f:new Float64Array(w),a:new Float64Array(w),eligible:true,cost:0}));
      const reference=s.actions.tie_entropy;
      for(let g=0;g<w;g++){
        const k=i*w+g,m=mass[k];if(!m)continue;
        const expect=(a,values)=>children[a].reduce((v,j)=>v+mass[j*w+g]/m*values[j*w+g],0);
        baseF[k]=expect(reference,baseF);baseA[k]=loss[k]+expect(reference,baseA);
        for(let a=0;a<4;a++){
          const probabilities=children[a].reduce((v,j)=>v+mass[j*w+g]/m,0);
          assert(Math.abs(probabilities-1)<1e-12);stats.branchChecks++;
          const c=candidates[a];c.f[g]=expect(a,ownF);c.a[g]=loss[k]+expect(a,ownA);
          if(c.f[g]>baseF[k]+remaining*slack||c.a[g]>baseA[k]+remaining*slack)c.eligible=false;
        }
      }
      assert(candidates[reference].eligible,'reference must remain feasible under protected continuations');
      for(let a=0;a<4;a++)candidates[a].cost=children[a].reduce((v,j)=>v+list[j].mass/s.mass*(risk[j]+priorCost[j]),0);
      const eligible=candidates.map((c,a)=>a).filter(a=>candidates[a].eligible);
      stats.eligibleActions+=eligible.length;
      const minimum=Math.min(...eligible.map(a=>candidates[a].cost));
      const chosen=eligible.find(a=>candidates[a].cost-minimum<=slack),c=candidates[chosen];
      s.actions.regime_safe=chosen;stats.changed+=Number(chosen!==reference);priorCost[i]=c.cost;
      for(let g=0;g<w;g++){
        const k=i*w+g;if(!mass[k])continue;ownF[k]=c.f[g];ownA[k]=c.a[g];stats.supportChecks++;
        const f=ownF[k]-baseF[k],a=ownA[k]-baseA[k];
        assert(f<=remaining*slack&&a<=remaining*slack);
        stats.maxFinalExcess=Math.max(stats.maxFinalExcess,f);stats.maxAreaSumExcess=Math.max(stats.maxAreaSumExcess,a);
      }
    }
  }
  model.totals.regime_safe={};return stats;
}
