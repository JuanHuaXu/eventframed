import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,shadow,output]=process.argv.slice(2),raw=fs.readFileSync(input),praw=fs.readFileSync(parent),sraw=fs.readFileSync(shadow);
const [h,...rows]=raw.toString().trim().split('\n').map(JSON.parse),[, ...parents]=praw.toString().trim().split('\n').map(JSON.parse),sh=JSON.parse(sraw);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex'),key=r=>[r.Phase,r.Case,r.Index].join('/');
assert.equal(hash(praw),h.ParentSHA256);assert.equal(rows.length,128);assert.equal(h.Consumed,true);
for(const [p,s]of Object.entries(h.Sources)){assert.equal(hash(s),h.Hashes[p]);assert.equal(hash(fs.readFileSync(p)),h.Hashes[p]);}
const pm=new Map(parents.map(r=>[key(r),r])),sm=new Map(sh.Records.map(r=>[key(r)+'/'+r.Schedule,r]));
const prior=[.7,.1,.1,.1],clip=x=>Math.max(1e-6,Math.min(1-1e-6,x));
const forecast=(w,p)=>{const n=w.reduce((s,v)=>s+v,0),q=n>0?w.map(v=>v/n):prior;return q.reduce((s,v,i)=>s+(.998*v+.002*prior[i])*clip(p[i]),0);};
const close=(x,y)=>assert.ok(Math.abs(x-y)<1e-12,`${x} != ${y}`),records=[];
for(const r of rows)for(const schedule of ['Immediate','Delayed']){
  const p=pm.get(key(r))[schedule],m=sm.get(key(r)+'/'+schedule);assert.equal(r[schedule].length,512);
  const full=[0,0],post=[0,0],correct=[0,0];let enlarged=0,additionalBits=0;
  const pop=x=>{let n=0;for(;x;x&=x-1)n++;return n;};
  for(let i=0;i<512;i++){
    const d=r[schedule][i],f=p.Frames[i],issued=f.Predictions[2];assert.equal(d.Origin,i);
    assert.equal(d.RequestedMask,issued.mask);assert.equal(d.RequestedValues,issued.values);
    assert.equal(d.MonitorMask,m.Masks[i][0]);assert.equal(d.MonitorValues,m.Values[i][0]);
    assert.equal(d.ConsumedMask,d.RequestedMask|d.MonitorMask);assert.equal(d.ConsumedValues,d.RequestedValues|d.MonitorValues);
    assert.equal(d.ConsumedValues,f.X&d.ConsumedMask);assert.deepEqual(d.OuterWeights,issued.weights_before_share);
    d.Experts.forEach(p=>assert.ok(p.every(v=>v>=0&&v<=1)));assert.deepEqual(d.Experts[0],issued.experts);
    close(d.Original,issued.p);close(d.Original,forecast(d.OuterWeights,d.Experts[0]));close(d.Available,forecast(d.OuterWeights,d.Experts[1]));
    if(d.ConsumedMask===d.RequestedMask){close(d.Original,d.Available);}else{enlarged++;additionalBits+=pop(d.ConsumedMask&~d.RequestedMask);}
    for(const [a,q]of [d.Original,d.Available].entries()){full[a]+=(q-Number(f.Y))**2/512;if(i>=256){post[a]+=(q-Number(f.Y))**2/256;correct[a]+=Number((q>=.5)===f.Y)/256;}}
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,full,post,accuracy:correct,enlargedFraction:enlarged/512,extraConsumedBits:additionalBits/512});
}
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length,interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,v)=>s+(v-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed']){
  const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(a.length,16);
  const gain=interval(a.map(r=>r.post[0]-r.post[1])),harmFull=interval(a.map(r=>r.full[1]-r.full[0])),harmPost=interval(a.map(r=>r.post[1]-r.post[0]));
  cells.push({phase,case:name,schedule,post:[0,1].map(i=>mean(a.map(r=>r.post[i]))),accuracy:[0,1].map(i=>mean(a.map(r=>r.accuracy[i]))),enlargedFraction:mean(a.map(r=>r.enlargedFraction)),extraConsumedBits:mean(a.map(r=>r.extraConsumedBits)),gain,harmFull,harmPost,gainPass:gain.mean>=.005&&gain.lower>0,nonharmPass:harmFull.upper<=.01&&harmPost.upper<=.01});
}
const targets=cells.filter(c=>c.case==='parity_to_majority');
const result={rawSHA256:hash(raw),parentSHA256:hash(praw),shadowSHA256:hash(sraw),sources:Object.keys(h.Sources).length,screens:{gains:targets.filter(c=>c.gainPass).length,gainTotal:targets.length,nonharm:cells.filter(c=>c.nonharmPass).length,nonharmTotal:cells.length},cells,records,limits:'Consumed fixed-state same-weight diagnostic. No extra acquisitions or future labels, but no available-view feedback or closed-loop policy. Extra consumed bits were already paid, not extra reads.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify({hash:result.rawSHA256,sources:result.sources,screens:result.screens,delayed:cells.filter(c=>c.schedule==='Delayed')}));
