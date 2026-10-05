import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,advice,parent,output]=process.argv.slice(2);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const raw=fs.readFileSync(input),ar=fs.readFileSync(advice),pr=fs.readFileSync(parent);
const parse=b=>b.toString().trim().split('\n').map(JSON.parse);
const [h,...rows]=parse(raw),[, ...as]=parse(ar),[, ...ps]=parse(pr);
assert.equal(hash(ar),h.InputSHA256);assert.equal(hash(pr),h.ParentSHA256);
assert.equal(rows.length,128);assert.equal(as.length,128);assert.equal(ps.length,128);
for(const [p,s]of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),h.Hashes[p]);}
const prior=[.7,.1,.1,.1],clip=x=>Math.max(1e-6,Math.min(1-1e-6,x));
const weights=w=>{const sum=w.reduce((a,b)=>a+b,0);return Math.abs(sum-1)>1e-9?prior:w.map((v,i)=>.998*v/sum+.002*prior[i]);};
let maxDifference=0,checked=0;
const close=(x,y)=>{maxDifference=Math.max(maxDifference,Math.abs(x-y));assert.ok(Math.abs(x-y)<1e-12,`${x} != ${y}`);checked++;};
const records=[];
for(let ri=0;ri<rows.length;ri++)for(const [si,schedule]of ['Immediate','Delayed'].entries()){
  const r=rows[ri],a=as[ri],p=ps[ri];
  for(const k of ['Phase','Case','Index']){assert.equal(r[k],a[k]);assert.equal(r[k],p[k]);}
  assert.equal(r[schedule].length,512);
  const frames=p[schedule].Frames,batches=Array.from({length:544},()=>[]);
  for(let i=0;i<512;i++){const f=frames[i];assert.ok(f.Arrival>=i&&f.Arrival<=i+31);if(!f.Missing)batches[f.Arrival].push(i);}
  let w=[0,0,0,0],updates=0;
  const full=[0,0,0],post=[0,0,0],accuracy=[0,0,0];let mass=0;
  const expert=i=>[a[schedule][i].Original,...Array(3).fill(a[schedule][i].Available)].map(clip);
  for(let clock=0;clock<544;clock++){
    if(clock<512){
      const d=r[schedule][clock],v=weights(w),q=v.reduce((sum,x,i)=>sum+x*expert(clock)[i],0);
      close(q,d.P);w.forEach((x,i)=>close(x,d.Weights[i]));
      const f=frames[clock];close(a[schedule][clock].Original,f.Predictions[2].p);
      for(const [i,q]of [a[schedule][clock].Original,a[schedule][clock].Available,d.P].entries()){
        assert.ok(q>=0&&q<=1);full[i]+=(q-Number(f.Y))**2/512;
        if(clock>=256){post[i]+=(q-Number(f.Y))**2/256;accuracy[i]+=Number((q>=.5)===f.Y)/256;}
      }
      if(clock>=256)mass+=(1-v[0])/256;
    }
    for(const i of batches[clock]){
      const v=weights(w),e=expert(i);w=v.map((x,j)=>x*(frames[i].Y?e[j]:1-e[j]));
      const sum=w.reduce((a,b)=>a+b,0);w=w.map(x=>x/sum);updates++;
    }
  }
  assert.equal(updates,r.Updates[si]);records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,full,post,accuracy,availableMass:mass,updates});
}
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,v)=>s+(v-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
  const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
  const gain=interval(a.map(r=>r.post[0]-r.post[2])),harmFull=interval(a.map(r=>r.full[2]-r.full[0])),harmPost=interval(a.map(r=>r.post[2]-r.post[0]));
  cells.push({phase,case:name,schedule,post:[0,1,2].map(i=>mean(a.map(r=>r.post[i]))),accuracy:[0,1,2].map(i=>mean(a.map(r=>r.accuracy[i]))),availableMass:mean(a.map(r=>r.availableMass)),gain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const targets=cells.filter(c=>c.case==='parity_to_majority');
const result={rawSHA256:hash(raw),inputSHA256:hash(ar),parentSHA256:hash(pr),sources:Object.keys(h.Sources).length,checked,maxDifference,screens:{gains:targets.filter(c=>c.gainPass).length,gainTotal:targets.length,nonharm:cells.filter(c=>c.nonharmPass).length,nonharmTotal:cells.length},cells,records,limits:'Consumed fixed-advice view aggregation, not closed-loop learning, fresh confirmation, or serving performance. Old component models and inner/outer weights remain frozen to tape.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({hash:result.rawSHA256,checked,maxDifference,screens:result.screens,delayed:cells.filter(c=>c.schedule==='Delayed')}));
