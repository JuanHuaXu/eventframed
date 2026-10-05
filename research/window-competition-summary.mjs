import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,window,advice,parent,output]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const raw=fs.readFileSync(input),wr=fs.readFileSync(window),ar=fs.readFileSync(advice),pr=fs.readFileSync(parent),parse=b=>b.toString().trim().split('\n').map(JSON.parse);
const [h,...rows]=parse(raw),[, ...ws]=parse(wr),[, ...as]=parse(ar),[, ...ps]=parse(pr);
assert.deepEqual([wr,ar,pr].map(hash),h.InputSHA256);for(const a of [rows,ws,as,ps])assert.equal(a.length,128);
for(const [p,s]of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),h.Hashes[p]);}
const prior=[.7,.1,.1,.1],clip=p=>Math.max(1e-6,Math.min(1-1e-6,p)),sum=a=>a.reduce((s,x)=>s+x,0);
const weights=w=>Math.abs(sum(w)-1)>1e-9?prior:w.map((v,i)=>.998*v/sum(w)+.002*prior[i]);
const forecast=(w,p)=>weights(w).reduce((s,x,i)=>s+x*clip(p[i]),0);
const observe=(w,p,y)=>{const v=weights(w).map((x,i)=>x*(y?clip(p[i]):1-clip(p[i])));return v.map(x=>x/sum(v));};
const revoke=w=>{const v=w.every(x=>x===0)?[...prior]:[...w];v[2]=Math.min(.1,v[2]);const n=sum(v);return v.map(x=>x/n);};
let checked=0,maxDifference=0;const close=(a,b)=>{checked++;maxDifference=Math.max(maxDifference,Math.abs(a-b));assert.ok(Math.abs(a-b)<1e-12);};
const records=[];
for(let k=0;k<128;k++)for(const schedule of ['Immediate','Delayed']){
 const r=rows[k],a=as[k],p=ps[k],w=ws[k];for(const key of ['Phase','Case','Index'])for(const x of [a,p,w])assert.equal(r[key],x[key]);
 assert.equal(r[schedule].length,512);
 const frames=p[schedule].Frames,batches=Array.from({length:544},()=>[]);
 frames.forEach((f,i)=>{assert.ok(f.Arrival>=i&&f.Arrival<=i+31);if(!f.Missing)batches[f.Arrival].push(i);});
 let control=[0,0,0,0],inner=Array.from({length:2},()=>[0,0,0,0]),outer=Array.from({length:2},()=>[0,0,0,0]),split=-1;
 const metrics=Array.from({length:2},()=>({full:[0,0],post:[0,0],accuracy:[0,0],newWindowWeight:0}));
 for(let clock=0;clock<544;clock++){
  if(clock<512){
   const d=a[schedule][clock],z=r[schedule][clock],y=Number(frames[clock].Y);
   (control.every(x=>x===0)?prior:control).forEach((x,i)=>close(x,d.OuterWeights[i]));
   for(let v=0;v<2;v++){
    const x=z.Views[v],raw=w[schedule].Frames[clock].Raw[1][v],ia=[d.Experts[v][1],raw[0],raw[1],raw[1]];
    ia.forEach((q,i)=>close(q,x.InnerAdvice[i]));inner[v].forEach((q,i)=>close(q,x.InnerWeights[i]));outer[v].forEach((q,i)=>close(q,x.OuterWeights[i]));
    const short=forecast(inner[v],ia),oa=[...d.Experts[v]];oa[1]=short;close(short,x.Short);oa.forEach((q,i)=>close(q,x.OuterAdvice[i]));close(forecast(outer[v],oa),x.P);
    const old=v?d.Available:d.Original;close(forecast(control,d.Experts[v]),old);
    [old,x.P].forEach((q,i)=>{metrics[v].full[i]+=(q-y)**2/512;if(clock>=256){metrics[v].post[i]+=(q-y)**2/256;metrics[v].accuracy[i]+=Number((q>=.5)===frames[clock].Y)/256;}});
    if(clock>=256)metrics[v].newWindowWeight+=(1-weights(inner[v])[0])/256;
   }
  }
  for(const i of batches[clock])if(split<0||i>split){
   control=observe(control,a[schedule][i].Experts[0],frames[i].Y);
   for(let v=0;v<2;v++){const x=r[schedule][i].Views[v];inner[v]=observe(inner[v],x.InnerAdvice,frames[i].Y);outer[v]=observe(outer[v],x.OuterAdvice,frames[i].Y);}
  }
  if(clock===p[schedule].Arms[2].SplitAt){split=clock;control=revoke(control);outer=outer.map(revoke);}
 }
 metrics.forEach((m,view)=>records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,view,...m}));
}
const mean=a=>sum(a)/a.length,interval=a=>{const m=mean(a),se=Math.sqrt(sum(a.map(x=>(x-m)**2))/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed'])for(let view=0;view<2;view++){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule&&r.view===view);assert.equal(a.length,16);
 const gain=interval(a.map(r=>r.post[0]-r.post[1])),harmFull=interval(a.map(r=>r.full[1]-r.full[0])),harmPost=interval(a.map(r=>r.post[1]-r.post[0]));
 cells.push({phase,case:name,schedule,view,post:[0,1].map(i=>mean(a.map(r=>r.post[i]))),accuracy:[0,1].map(i=>mean(a.map(r=>r.accuracy[i]))),newWindowWeight:mean(a.map(r=>r.newWindowWeight)),gain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const screens=[0,1].map(view=>{const a=cells.filter(c=>c.view===view),g=a.filter(c=>c.case==='parity_to_majority');return {view,gains:g.filter(c=>c.gainPass).length,gainTotal:g.length,nonharm:a.filter(c=>c.nonharmPass).length,nonharmTotal:a.length};});
const result={rawSHA256:hash(raw),sources:Object.keys(h.Sources).length,checked,maxDifference,screens,cells,records,limits:'Consumed fixed-acquisition nested window competition. Original model fits and gates remain frozen. Neither view changes acquisition; available control retains narrow-trained original weights. No fresh confirmation or serving claim.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({rawSHA256:result.rawSHA256,screens,checked,maxDifference,requested:cells.filter(c=>c.view===0)}));
