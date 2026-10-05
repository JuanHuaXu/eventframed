import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,advice,parent,output]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const raw=fs.readFileSync(input),ar=fs.readFileSync(advice),pr=fs.readFileSync(parent),parse=b=>b.toString().trim().split('\n').map(JSON.parse);
const [h,...rows]=parse(raw),[, ...as]=parse(ar),[, ...ps]=parse(pr);
assert.equal(hash(ar),h.InputSHA256);assert.equal(hash(pr),h.ParentSHA256);assert.equal(rows.length,128);
for(const [p,s]of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),h.Hashes[p]);}
const prior=[.7,.1,.1,.1],clip=p=>Math.max(1e-6,Math.min(1-1e-6,p));
const mix=(w,p)=>{const sum=w.reduce((a,b)=>a+b,0),v=Math.abs(sum-1)>1e-9?prior:w.map((x,i)=>.998*x/sum+.002*prior[i]);return v.reduce((s,x,i)=>s+x*clip(p[i]),0);};
let checked=0,maxDifference=0;
const close=(a,b)=>{maxDifference=Math.max(maxDifference,Math.abs(a-b));assert.ok(Math.abs(a-b)<1e-12);checked++;};
const records=[];
for(let k=0;k<128;k++)for(const schedule of ['Immediate','Delayed']){
 const r=rows[k],a=as[k],p=ps[k];for(const key of ['Phase','Case','Index']){assert.equal(r[key],a[key]);assert.equal(r[key],p[key]);}
 const z=r[schedule],frames=p[schedule].Frames,diag=a[schedule];assert.equal(z.Frames.length,512);
 const fit=p[schedule].Fits.filter(f=>f.Clock<512);assert.equal(z.Fits.length,fit.length);
 z.Fits.forEach((f,j)=>{assert.equal(f.Clock,fit[j].Clock);assert.deepEqual(f.LabelIDs,fit[j].Origins.slice(-64));assert.deepEqual(f.EventIDs??[],fit[j].Origins.filter(i=>i>f.Clock-64));for(const i of f.EventIDs??[])assert.ok(frames[i].Audit&&!frames[i].Missing&&frames[i].Arrival<=f.Clock);});
 const full=[0,0,0,0],post=[0,0,0,0],accuracy=[0,0,0,0],rawPost=Array(8).fill(0),rawLate=Array(8).fill(0);let fi=-1;
 for(let i=0;i<512;i++){
  const f=z.Frames[i],d=diag[i],y=Number(frames[i].Y);while(fi+1<fit.length&&fit[fi+1].Clock<i)fi++;
  assert.equal(f.FitClock,fi<0?-1:fit[fi].Clock);
  for(let w=0;w<2;w++)for(let v=0;v<2;v++){
   const raw=f.Raw[w][v];raw.forEach(p=>assert.ok(p>=0&&p<=1));
   close(f.Short[w][v],fi<0?.5:mix(d.InnerWeights,[raw[0],raw[1],raw[1],raw[1]]));
   const e=[...d.Experts[v]];e[1]=f.Short[w][v];close(f.Law[w][v],mix(d.OuterWeights,e));
   if(w===0){close(f.Short[w][v],d.Experts[v][1]);close(f.Law[w][v],v?d.Available:d.Original);}
   const q=f.Law[w][v],j=2*w+v;full[j]+=(q-y)**2/512;if(i>=256){post[j]+=(q-y)**2/256;accuracy[j]+=Number((q>=.5)===frames[i].Y)/256;}
   raw.forEach((q,m)=>{const j=4*w+2*v+m;if(i>=256)rawPost[j]+=(q-y)**2/256;if(i>=384)rawLate[j]+=(q-y)**2/128;});
  }
 }
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,full,post,accuracy,rawPost,rawLate});
}
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length,interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed'])for(let view=0;view<2;view++){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
 const gain=interval(a.map(r=>r.post[view]-r.post[2+view])),harmFull=interval(a.map(r=>r.full[2+view]-r.full[view])),harmPost=interval(a.map(r=>r.post[2+view]-r.post[view]));
 cells.push({phase,case:name,schedule,view,post:[view,2+view].map(i=>mean(a.map(r=>r.post[i]))),accuracy:[view,2+view].map(i=>mean(a.map(r=>r.accuracy[i]))),rawPost:[0,1].flatMap(w=>[0,1].map(m=>mean(a.map(r=>r.rawPost[4*w+2*view+m])))),rawLate:[0,1].flatMap(w=>[0,1].map(m=>mean(a.map(r=>r.rawLate[4*w+2*view+m])))),gain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const screens=[0,1].map(view=>{const a=cells.filter(c=>c.view===view),g=a.filter(c=>c.case==='parity_to_majority');return {view,gains:g.filter(c=>c.gainPass).length,gainTotal:g.length,nonharm:a.filter(c=>c.nonharmPass).length,nonharmTotal:a.length};});
const result={rawSHA256:hash(raw),sources:Object.keys(h.Sources).length,checked,maxDifference,screens,cells,records,limits:'Consumed fixed-weight, fixed-mask model intervention. Reference reconstruction and arithmetic verified; no candidate feedback or adaptive serving test. Raw model order: label count, label subset, event count, event subset.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({rawSHA256:result.rawSHA256,screens,checked,maxDifference,delayed:cells.filter(c=>c.schedule==='Delayed'&&c.view===0)}));
