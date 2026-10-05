import assert from 'node:assert/strict';
import {currentStateInformation} from './current-state-information.mjs';
// Independent joint path enumeration of nominated label and current expert.
function enumerate(rows,prior,origin){
 const joint=[prior.map(()=>0),prior.map(()=>0)];
 for(const y of [0,1]){
  let paths=prior.flatMap((p,k)=>[{k,a:true,w:p/2},{k,a:false,w:p/2}]);
  for(let t=0;t<rows.length;t++){
   const out=[];for(const s of paths){const r=rows[t],label=t===origin?y:r.y,p=r.p[s.k],w=s.w*(label===null?1:label?p:1-p);
    if(!s.a){out.push({...s,w});continue;}
    out.push({...s,w:w*(t+1)/(t+2)});
    prior.forEach((v,k)=>{for(const a of [false,true])out.push({k,a,w:w*v/(2*(t+2))});});
   }paths=out;
  }
  for(const s of paths)joint[y][s.k]+=s.w;
 }
 const z=joint.flat().reduce((a,b)=>a+b,0);joint.forEach(row=>row.forEach((v,k)=>row[k]=v/z));
 const py=joint.map(row=>row.reduce((a,b)=>a+b,0)),pk=prior.map((_,k)=>joint[0][k]+joint[1][k]);
 let mi=0;for(const y of [0,1])for(let k=0;k<prior.length;k++)if(joint[y][k]>0)mi+=joint[y][k]*Math.log(joint[y][k]/(py[y]*pk[k]));
 return{mi,py};
}
let checks=0;
for(let seed=0;seed<16;seed++)for(let origin=0;origin<4;origin++){
 const rows=Array.from({length:4},(_,j)=>({p:[.1+((seed+j)%7)/10,.2+((seed*3+j)%6)/10],y:j===origin?null:(seed+j)%2}));
 const a=currentStateInformation(rows,[.8,.2],origin),b=enumerate(rows,[.8,.2],origin);
 assert(Math.abs(a.information-b.mi)<1e-12);a.probabilities.forEach((p,i)=>assert(Math.abs(p-b.py[i])<1e-12));checks++;
}
assert.equal(currentStateInformation([{p:[.2,.2],y:null}],[.8,.2],0).information,0);
assert.throws(()=>currentStateInformation([{p:[.2],y:1}],[1],0));
console.log(JSON.stringify({checks,status:'PASS'}));
