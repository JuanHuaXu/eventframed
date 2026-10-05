import assert from 'node:assert/strict';
import crypto from 'node:crypto';

const prior=[.95,.05/3,.05/3,.05/3];
function step(w,ps,y){const z=w.reduce((sum,v,k)=>sum+v*(y===undefined?1:y?ps[k]:1-ps[k]),0);assert(z>0);return w.map((v,k)=>.999*v*(y===undefined?1:y?ps[k]:1-ps[k])/z+.001*prior[k]);}

export function auditCadence(r,raw) {
 assert(!r.Error,r.Error);assert.deepEqual([r.Phase,r.Case,r.Index,r.Schedule],[raw.Phase,raw.Case,raw.Index,raw.Schedule]);
 const out=r.Result;assert.equal((out.Queries??[]).length,0);assert.equal((out.Arrivals??[]).length,0);assert.equal(out.Predictions.length,256);assert.equal(out.Fits.length,32);
 for(const s of raw.Steps){assert(Number.isFinite(s.Q)&&s.Q>=0&&s.Q<=1);assert(Number.isInteger(s.Delay)&&s.Delay>=0&&s.Delay<=31);}
 for(let i=0;i<32;i++) {
  const fit=out.Fits[i];assert.equal(fit.Clock,i*8);
  const eligible=[];for(let j=-16;j<fit.Clock;j++)if(j<0||(!raw.Steps[j].Missing&&j+raw.Steps[j].Delay<=fit.Clock))eligible.push(j);
  for(let w=0;w<2;w++)assert.deepEqual(fit.Origins[w],eligible.slice(-(w===0?64:32)));
 }
 const brier=Array.from({length:2},()=>Array.from({length:5},()=>[0,0]));
 for(let cadence=0;cadence<2;cadence++) {
  const predictions=cadence===0?raw.Steps.map(s=>[0,1,2,3,12].map(a=>s.P[a])):out.Predictions;
  let base=0,checkpoint=[...prior];
  for(let t=0;t<256;t++) {
   const ps=predictions[t];assert.equal(ps.length,5);
   const forward=(w,j)=>{const s=raw.Steps[j];return step(w,predictions[j],!s.Missing&&j+s.Delay<=t?s.Y:undefined);};
   while(base<=t-32){checkpoint=forward(checkpoint,base);base++;}
   let weights=[...checkpoint];for(let j=base;j<t;j++)weights=forward(weights,j);
   const served=weights.reduce((sum,w,k)=>sum+w*ps[k],0);assert(Math.abs(served-ps[4])<1e-10);
   ps.forEach((p,a)=>{assert(Number.isFinite(p)&&p>0&&p<1);const q=raw.Steps[t].Q,loss=(p-q)**2+q*(1-q);brier[cadence][a][0]+=loss/256;if(t>=192)brier[cadence][a][1]+=loss/64;});
   if(t<8)ps.forEach((p,a)=>assert(Math.abs(p-raw.Steps[t].P[[0,1,2,3,12][a]])<1e-12));
  }
 }
 const signature=crypto.createHash('sha256').update(JSON.stringify([raw.Initial,raw.Steps.map(s=>[s.X,s.Y,s.Q])])).digest('hex');
 return {phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,brier,signature};
}
