import assert from 'node:assert/strict';
export const TIE_TOLERANCE=1e-12;
export function tieChoice(costs){
  assert(costs.length===4&&costs.every(Number.isFinite));
  const minimum=Math.min(...costs);
  return costs.findIndex(c=>c-minimum<=TIE_TOLERANCE);
}
// Reuse frozen forecasts. Only the decision rule changes, not the posterior.
export function addTiePolicies(model){
  const stats={states:0,branchChecks:0,maxFullCostExcess:0,changes:{one:0,two:0,full:0,entropy:0}};
  for(const root of model.roots){
    const states=new Map(root.states.map(s=>[s.counts.join(','),s])),memo=new Map();
    const risk=s=>1-s.forecast.reduce((a,p)=>a+p*p,0);
    const branches=(s,a)=>[0,1].map(y=>{
      const counts=s.counts.slice();counts[2*a+y]++;
      const child=states.get(counts.join(','));assert(child);
      return {child,p:child.mass/s.mass};
    });
    const plan=(s,depth)=>{
      const d=Math.min(depth,6-s.counts.reduce((a,n)=>a+n,0));
      if(d===0)return {cost:0,action:null};
      const key=d+':'+s.counts.join(',');
      if(!memo.has(key)){
        const costs=Array.from({length:4},(_,a)=>branches(s,a).reduce((v,c)=>v+c.p*(risk(c.child)+plan(c.child,d-1).cost),0));
        const action=tieChoice(costs);memo.set(key,{action,cost:costs[action]});
      }
      return memo.get(key);
    };
    for(const s of root.states){
      const remaining=6-s.counts.reduce((a,n)=>a+n,0);if(!remaining)continue;
      stats.states++;
      const entropy=Array.from({length:4},(_,a)=>{
        const children=branches(s,a);assert(Math.abs(children.reduce((v,c)=>v+c.p,0)-1)<1e-12);stats.branchChecks++;
        return children.reduce((v,c)=>v+c.p*Math.log(c.p),0);
      });
      for(const name of ['one','two','full','entropy']){
        const action=name==='entropy'?tieChoice(entropy):plan(s,name==='one'?1:name==='two'?2:remaining).action;
        s.actions['tie_'+name]=action;stats.changes[name]+=Number(action!==s.actions[name]);
      }
      const excess=plan(s,remaining).cost-s.fullCost;
      assert(excess>=-1e-12&&excess<=remaining*TIE_TOLERANCE+1e-12);
      stats.maxFullCostExcess=Math.max(stats.maxFullCostExcess,excess);
    }
  }
  for(const name of ['one','two','full','entropy'])model.totals['tie_'+name]={};
  return stats;
}
