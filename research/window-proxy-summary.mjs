import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,competition,advice,parent,original,output]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const raw=fs.readFileSync(input),cr=fs.readFileSync(competition),ar=fs.readFileSync(advice),pr=fs.readFileSync(parent),orr=fs.readFileSync(original),parse=b=>b.toString().trim().split('\n').map(JSON.parse);
const [h,...rows]=parse(raw),[, ...cs]=parse(cr),[, ...as]=parse(ar),[, ...ps]=parse(pr),[, ...olds]=parse(orr);
assert.deepEqual([cr,ar,pr,orr].map(hash),h.InputSHA256);for(const a of [rows,cs,as,ps,olds])assert.equal(a.length,128);
for(const [p,s]of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),h.Hashes[p]);}
const prior=[.7,.1,.1,.1],clip=p=>Math.max(1e-6,Math.min(1-1e-6,p)),sum=a=>a.reduce((s,x)=>s+x,0),pop=x=>{let n=0;for(;x;x&=x-1)n++;return n;};
const weights=w=>Math.abs(sum(w)-1)>1e-9?prior:w.map((x,i)=>.998*x/sum(w)+.002*prior[i]);
const forecast=(w,p)=>weights(w).reduce((s,x,i)=>s+x*clip(p[i]),0);
const observe=(w,p,y)=>{const v=weights(w).map((x,i)=>x*(y?clip(p[i]):1-clip(p[i])));return v.map(x=>x/sum(v));};
const revoke=w=>{const v=w.every(x=>x===0)?[...prior]:[...w];v[2]=Math.min(.1,v[2]);const n=sum(v);return v.map(x=>x/n);};
let checked=0,maxDifference=0;const close=(a,b)=>{checked++;maxDifference=Math.max(maxDifference,Math.abs(a-b));assert.ok(Math.abs(a-b)<1e-12,`${a} != ${b}`);};
const records=[];
for(let k=0;k<128;k++)for(const [si,schedule]of ['Immediate','Delayed'].entries()){
 const r=rows[k],a=as[k],p=ps[k],c=cs[k];for(const key of ['Phase','Case','Index'])for(const x of [a,p,c,olds[k]])assert.equal(r[key],x[key]);
 const z=r[schedule],parent=p[schedule],frames=parent.Frames;assert.equal(z.Frames.length,512);assert.equal(z.Fits.length,parent.Fits.length);
 z.Fits.forEach((f,j)=>{const q=parent.Fits[j];assert.equal(f.Clock,q.Clock);assert.deepEqual(f.LabelIDs,q.Origins.slice(-64));assert.deepEqual(f.EventIDs??[],q.Origins.filter(i=>i>q.Clock-64));for(const i of q.Origins)assert.ok(frames[i].Audit&&!frames[i].Missing&&frames[i].Arrival<=q.Clock);});
 const batches=Array.from({length:544},()=>[]);frames.forEach((f,i)=>{assert.ok(f.Arrival>=i&&f.Arrival<=i+31);if(!f.Missing)batches[f.Arrival].push(i);});
 let previousFull=0,previousPost=0,previousAccuracy=0,previousIncremental=0;
 let state=Array.from({length:2},()=>Array.from({length:3},()=>[0,0,0,0])),split=-1,fi=-1,changedMasks=0;
 const full=[0,0,0],post=[0,0,0],accuracy=[0,0,0],requested=[0,0],incremental=[0,0],guides={};
 for(let clock=0;clock<544;clock++){
  if(clock<512){
   const f=z.Frames[clock],d=a[schedule][clock],y=Number(frames[clock].Y),ref=c[schedule][clock].Views[0],previous=olds[k][schedule].Frames[clock];
   assert.deepEqual(f.Arms[0],previous.Arms[0]);assert.deepEqual(f.Weights[0],previous.Weights[0]);assert.equal(f.Incremental[0],previous.Incremental[0]);
   const previousP=previous.Arms[1].P;previousFull+=(previousP-y)**2/512;
   if(clock>=256){previousPost+=(previousP-y)**2/256;previousAccuracy+=Number((previousP>=.5)===frames[clock].Y)/256;}
   assert.equal(previous.Incremental[1],pop(previous.Arms[1].Mask&~d.MonitorMask));previousIncremental+=previous.Incremental[1];
   while(fi+1<parent.Fits.length&&parent.Fits[fi+1].Clock<clock)fi++;
   assert.equal(f.FitClock,fi<0?-1:parent.Fits[fi].Clock);assert.equal(f.OldAvailable,fi>=0);
   f.Arms.forEach((arm,j)=>{
    state[j].forEach((w,l)=>w.forEach((x,i)=>close(x,f.Weights[j][l][i])));
    assert.ok(arm.Cost<=6&&arm.Cost===pop(arm.Mask)&&arm.Mask<512);assert.equal(arm.Values,frames[clock].X&arm.Mask);
    assert.equal(f.Incremental[j],pop(arm.Mask&~d.MonitorMask));
    const old=f.OldAvailable?forecast(state[j][0],arm.OldInner):arm.OldInner[0];close(old,arm.NewInner[0]);
    close(forecast(state[j][1],arm.NewInner),arm.Outer[1]);close(forecast(state[j][2],arm.Outer),arm.P);
    [...arm.OldInner,...arm.NewInner,...arm.Outer].forEach(q=>assert.ok(q>=0&&q<=1));
    requested[j]+=arm.Cost;incremental[j]+=f.Incremental[j];
   });
   assert.equal(f.Arms[0].Mask,d.RequestedMask);close(f.Arms[0].P,ref.P);
   f.Arms[0].NewInner.forEach((x,i)=>close(x,ref.InnerAdvice[i]));f.Arms[0].Outer.forEach((x,i)=>close(x,ref.OuterAdvice[i]));
   if(f.Arms[0].Mask!==f.Arms[1].Mask)changedMasks++;
   guides[f.Arms[1].Guide]=(guides[f.Arms[1].Guide]??0)+1;
   [d.Original,f.Arms[0].P,f.Arms[1].P].forEach((q,i)=>{full[i]+=(q-y)**2/512;if(clock>=256){post[i]+=(q-y)**2/256;accuracy[i]+=Number((q>=.5)===frames[clock].Y)/256;}});
  }
  for(const origin of batches[clock])if(split<0||origin>split){
   const f=z.Frames[origin];for(let j=0;j<2;j++){
    if(f.OldAvailable)state[j][0]=observe(state[j][0],f.Arms[j].OldInner,frames[origin].Y);
    state[j][1]=observe(state[j][1],f.Arms[j].NewInner,frames[origin].Y);state[j][2]=observe(state[j][2],f.Arms[j].Outer,frames[origin].Y);
   }
  }
  if(clock===parent.Arms[2].SplitAt){split=clock;state.forEach(s=>s[2]=revoke(s[2]));}
 }
 const monitoring=sum(p.Acquisition[si].map(x=>x.Charge));assert.equal(monitoring,parent.MonitorCost+parent.AuditCost);
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,full,post,accuracy,requested,incremental,previousFull,previousPost,previousAccuracy,previousTotal:monitoring+previousIncremental,total:incremental.map(x=>x+monitoring),changedMasks,guides});
}
const mean=a=>sum(a)/a.length,interval=a=>{const m=mean(a),se=Math.sqrt(sum(a.map(x=>(x-m)**2))/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
 const proxyGain=interval(a.map(r=>r.previousPost-r.post[2]));
 const gain=interval(a.map(r=>r.post[0]-r.post[2])),couplingGain=interval(a.map(r=>r.post[1]-r.post[2])),harmFull=interval(a.map(r=>r.full[2]-r.full[0])),harmPost=interval(a.map(r=>r.post[2]-r.post[0]));
 cells.push({phase,case:name,schedule,post:[0,1,2].map(i=>mean(a.map(r=>r.post[i]))),accuracy:[0,1,2].map(i=>mean(a.map(r=>r.accuracy[i]))),totalPerFrame:[0,1].map(i=>mean(a.map(r=>r.total[i]/512))),requestedPerFrame:[0,1].map(i=>mean(a.map(r=>r.requested[i]/512))),changedMaskFraction:mean(a.map(r=>r.changedMasks/512)),previousPost:mean(a.map(r=>r.previousPost)),previousTotalPerFrame:mean(a.map(r=>r.previousTotal/512)),proxyGain,gain,couplingGain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const targets=cells.filter(c=>c.case==='parity_to_majority');
const result={rawSHA256:hash(raw),sources:Object.keys(h.Sources).length,checked,maxDifference,screens:{gains:targets.filter(c=>c.gainPass).length,gainTotal:targets.length,nonharm:cells.filter(c=>c.nonharmPass).length,nonharmTotal:cells.length},costVersusPrevious:{more:records.filter(r=>r.total[1]>r.previousTotal).length,equal:records.filter(r=>r.total[1]===r.previousTotal).length,less:records.filter(r=>r.total[1]<r.previousTotal).length},cost:{more:records.filter(r=>r.total[1]>r.total[0]).length,equal:records.filter(r=>r.total[1]===r.total[0]).length,less:records.filter(r=>r.total[1]<r.total[0]).length},cells,records,limits:'Consumed event-count-to-subset acquisition proxy ablation; count remains scored. Previous coupled and fixed controls retained. Random-audit fits and external monitoring are exogenous. Coordinate costs are not bytes, I/O requests or CPU. No fresh confirmation, agent utility or serving claim.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({rawSHA256:result.rawSHA256,sources:result.sources,checked,maxDifference,screens:result.screens,cost:result.cost,costVersusPrevious:result.costVersusPrevious,cells}));

