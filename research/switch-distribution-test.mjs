import assert from 'node:assert/strict';
import {switchWeights} from './switch-distribution.mjs';
// Explicit paths over stopped/active expert states, without recursive merging.
function enumerate(rows,prior) {
  let paths=prior.flatMap((p,i)=>[{i,active:true,w:p/2},{i,active:false,w:p/2}]);
  for(let t=0;t<rows.length;t++) {
    const next=[];
    for(const path of paths) {
      const row=rows[t],p=Math.max(1e-12,Math.min(1-1e-12,row.p[path.i]));
      const w=path.w*(row.y===null?1:row.y?p:1-p);
      if(!path.active){next.push({...path,w});continue;}
      next.push({...path,w:w*(t+1)/(t+2)});
      for(let i=0;i<prior.length;i++)for(const active of [false,true])
        next.push({i,active,w:w*prior[i]/(2*(t+2))});
    }
    paths=next;
  }
  const z=paths.reduce((a,b)=>a+b.w,0),weights=prior.map((_,i)=>paths.filter(p=>p.i===i).reduce((a,b)=>a+b.w,0)/z);
  return {weights,logEvidence:Math.log(z)};
}
let checks=0;
for(let seed=0;seed<32;seed++) {
  const prior=[.7,.2,.1],rows=Array.from({length:5},(_,j)=>({p:prior.map((_,i)=>.05+((seed*7+j*11+i*13)%90)/100),y:(seed+j)%3===0?null:(seed+j)%2}));
  for(let n=0;n<=5;n++) {
    const a=switchWeights(rows.slice(0,n),prior),b=enumerate(rows.slice(0,n),prior);
    a.weights.forEach((v,i)=>assert(Math.abs(v-b.weights[i])<1e-12));
    assert(Math.abs(a.logEvidence-b.logEvidence)<1e-12);
    const fixed=switchWeights(rows.slice(0,n),prior,false);
    assert(a.logEvidence>=fixed.logEvidence-Math.log(2)-1e-12);checks++;
  }
}
assert.deepEqual(switchWeights([{p:[.2],y:1}],[1]).weights,[1]);
assert.throws(()=>switchWeights([],[NaN]));
assert.throws(()=>switchWeights([{p:[2],y:1}],[1]));
console.log(JSON.stringify({checks,status:'PASS'}));
