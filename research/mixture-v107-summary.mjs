import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import assert from 'node:assert/strict';
const path=process.argv[2]??'docs/experiments/mmm-mixture-v107.json';
const bytes=readFileSync(path),a=JSON.parse(bytes);
const sha=b=>createHash('sha256').update(b).digest('hex');
const near=(a,b,tol=2e-10)=>assert.ok(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=tol,`${a} != ${b}`);
assert.equal(a.Version,'v107');assert.equal(a.ParentSHA256,sha(readFileSync('docs/experiments/mmm-arrival-v106.json')));
for(const [name,hash]of Object.entries(a.Hashes))assert.equal(sha(readFileSync(name)),hash,name);
assert.equal(a.Records.length,128);
const cells=new Map(),seen=new Set();let maxGap=0;
for(const r of a.Records){
 assert.ok(['design','confirmation'].includes(r.Phase));assert.ok(['majority_to_parity','parity_to_majority'].includes(r.Case));
 assert.ok(Number.isInteger(r.Index)&&r.Index>=0&&r.Index<32);
 const id=`${r.Phase}/${r.Case}/${r.Index}`;assert.ok(!seen.has(id));seen.add(id);assert.equal(r.Publications.length,4);
 for(const [index,p]of r.Publications.entries()){
  assert.equal(p.Clock,128+32*index);assert.equal(p.Rows.length,512);assert.equal(p.Truth.length,512);
  const A=Array.from({length:5},()=>Array(5).fill(0)),B=Array(5).fill(0),pure=Array(5).fill(0);let C=0,floor=0;
  for(let x=0;x<512;x++){
   const row=p.Rows[x],q=p.Truth[x];assert.equal(row.length,5);assert.ok(q>=0&&q<=1);near(row[4],.5,0);
   C+=q/512;floor+=q*(1-q)/512;
   for(let i=0;i<5;i++){assert.ok(row[i]>=0&&row[i]<=1);B[i]+=q*row[i]/512;pure[i]+=((row[i]-q)**2+q*(1-q))/512;for(let j=0;j<5;j++)A[i][j]+=row[i]*row[j]/512;}
  }
  near(C,p.Profile.C);near(floor,.0475);near(pure[4],.25);
  for(let i=0;i<5;i++){near(B[i],p.Profile.B[i]);for(let j=0;j<5;j++)near(A[i][j],p.Profile.A[i][j]);}
  for(const [obj,n]of [[p.Raw,4],[p.Neutral,5]]){
   const w=obj.Weights;assert.equal(w.length,5);near(w.reduce((s,x)=>s+x,0),1,1e-12);
   w.forEach((x,j)=>{assert.ok(x>=0&&x<=1);if(j>=n)near(x,0,0);});
   let direct=0;for(let x=0;x<512;x++){const q=p.Truth[x],pred=w.reduce((s,v,j)=>s+v*p.Rows[x][j],0);direct+=((pred-q)**2+q*(1-q))/512;}
   const grad=B.map((b,i)=>-2*b+2*A[i].reduce((s,v,j)=>s+v*w[j],0));
   const gap=Math.max(0,grad.reduce((s,v,j)=>s+v*w[j],0)-Math.min(...grad.slice(0,n)));
   near(direct,obj.Risk);near(gap,obj.Gap);near(obj.Lower,direct-gap);assert.ok(gap<=1e-8);
   assert.ok(direct>=floor-1e-10&&direct<=Math.min(...pure.slice(0,n))+1e-10);maxGap=Math.max(maxGap,gap);
  }
  assert.ok(p.Neutral.Risk<=p.Raw.Risk+1e-10);
  const key=`${r.Phase}/${r.Case}/${p.Clock}`;
  if(!cells.has(key))cells.set(key,[]);
  cells.get(key).push({pure,raw:p.Raw.Risk,neutral:p.Neutral.Risk,neutralWeight:p.Neutral.Weights[4],floor});
 }
}
assert.equal(cells.size,16);
const results=[...cells].map(([key,rows])=>{
 assert.equal(rows.length,32);const avg=f=>rows.reduce((s,r)=>s+f(r),0)/32;
 return {key,n:32,pure:rows[0].pure.map((_,i)=>avg(r=>r.pure[i])),bestPure:avg(r=>Math.min(...r.pure.slice(0,4))),rawOracle:avg(r=>r.raw),neutralOracle:avg(r=>r.neutral),neutralGain:avg(r=>r.raw-r.neutral),neutralWeight:avg(r=>r.neutralWeight),rawWorseThanNeutral:rows.filter(r=>r.raw>.25+1e-8).length,rawAbovePoint24:rows.filter(r=>r.raw>.24).length,floor:avg(r=>r.floor)};
});
console.log(JSON.stringify({version:'v107',status:'CONSUMED_DIAGNOSTIC_NOT_A_QUALITY_GATE',artifactSHA256:sha(bytes),sourceHashes:Object.keys(a.Hashes).length,records:a.Records.length,publications:512,maxNumericalGap:maxGap,results},null,2));
