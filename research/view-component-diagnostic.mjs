import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,parent,output]=process.argv.slice(2);
const ar=fs.readFileSync(input),pr=fs.readFileSync(parent);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(hash(ar),'3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e');
assert.equal(hash(pr),'4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655');
const parse=b=>b.toString().trim().split('\n').slice(1).map(JSON.parse);
const as=parse(ar),ps=parse(pr),cells=[];
for(const phase of ['cohort1','cohort2'])for(const name of [...new Set(as.map(r=>r.Case))])for(const schedule of ['Immediate','Delayed'])for(const [start,end]of [[256,384],[384,512]]){
  const losses=[];
  as.forEach((r,k)=>{
    if(r.Phase!==phase||r.Case!==name)return;
    for(const key of ['Phase','Case','Index'])assert.equal(r[key],ps[k][key]);
    const x=Array(8).fill(0);
    for(let i=start;i<end;i++)r[schedule][i].Experts.flat().forEach((p,j)=>x[j]+=(p-Number(ps[k][schedule].Frames[i].Y))**2/(end-start));
    losses.push(x);
  });
  assert.equal(losses.length,16);
  cells.push({phase,case:name,schedule,start,end,brier:Array.from({length:8},(_,i)=>losses.reduce((sum,x)=>sum+x[i],0)/16)});
}
fs.writeFileSync(output,JSON.stringify({inputSHA256:hash(ar),parentSHA256:hash(pr),order:['narrow base','narrow short-inner','narrow pooled/local','narrow neutral','available base','available short-inner','available pooled/local','available neutral'],cells,limits:'Post-hoc constituent diagnosis, not deployable expert selection or a statistical quality gate. Short-inner still contains the original inner mixture; no model fitting changes.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.schedule==='Delayed'&&c.case.includes('_to_'))));
