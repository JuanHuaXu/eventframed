import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,output]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const raw=fs.readFileSync(input),pr=fs.readFileSync(parent);
assert.equal(hash(raw),'ab3756fc535db2a1cc47d3c97629d5e2e291452284dd1238702e7cebf840dda5');
assert.equal(hash(pr),'4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655');
const parse=b=>b.toString().trim().split('\n').slice(1).map(JSON.parse),rows=parse(raw),ps=parse(pr),records=[];
for(let k=0;k<128;k++)for(const schedule of ['Immediate','Delayed'])for(const [start,end]of [[256,384],[384,512]]){
 const r=rows[k],p=ps[k];for(const key of ['Phase','Case','Index'])assert.equal(r[key],p[key]);
 const target=p.Masks[r.Case.includes('_to_')?1:0],coverage=[0,0],loss=[0,0],guides={};
 let lost=0,gained=0;
 for(let i=start;i<end;i++){
  const arms=r[schedule].Frames[i].Arms,y=Number(p[schedule].Frames[i].Y),covered=arms.map(a=>(a.Mask&target)===target);
  for(let a=0;a<2;a++){coverage[a]+=Number(covered[a])/(end-start);loss[a]+=(arms[a].P-y)**2/(end-start);}
  lost+=Number(covered[0]&&!covered[1]);gained+=Number(!covered[0]&&covered[1]);guides[arms[1].Guide]=(guides[arms[1].Guide]??0)+1;
 }
 records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule,start,end,coverage,loss,lost,gained,guides});
}
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length,cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(records.map(r=>r.case))])for(const schedule of ['Immediate','Delayed'])for(const start of [256,384]){
 const a=records.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule&&r.start===start);assert.equal(a.length,16);
 const guides={};for(const r of a)for(const [g,n]of Object.entries(r.guides))guides[g]=(guides[g]??0)+n/(16*128);
 cells.push({phase,case:name,schedule,start,end:start+128,coverage:[0,1].map(i=>mean(a.map(r=>r.coverage[i]))),loss:[0,1].map(i=>mean(a.map(r=>r.loss[i]))),lost:mean(a.map(r=>r.lost/128)),gained:mean(a.map(r=>r.gained/128)),guides});
}
fs.writeFileSync(output,JSON.stringify({rawSHA256:hash(raw),parentSHA256:hash(pr),scriptSHA256:hash(fs.readFileSync(new URL(import.meta.url))),cells,records,limits:'Retrospective simulator-rule coverage, NEVER a learner input or permitted guide-selection signal. Complete rule coverage is not itself a prediction or calibration guarantee. Majority rules can be partly informative without full coverage. No causal attribution from subgroup averages.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.case==='majority_to_parity'&&c.schedule==='Immediate')));
