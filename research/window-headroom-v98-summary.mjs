import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const raw = fs.readFileSync('docs/experiments/mmm-window-headroom-v98.json');
const a = JSON.parse(raw), parentRaw = fs.readFileSync('docs/experiments/mmm-window-bank-v97.json');
assert.equal(a.Protocol,'mmm-window-headroom-v98');
assert.equal(a.ParentSHA256,'6b4559a63d7b6c1e1b2603513a90dcf4c82281b27e48430405522a1fb48380e5');
assert.equal(sha(parentRaw),a.ParentSHA256);
assert.equal(Object.keys(a.Hashes).length,18);
for (const [p,h] of Object.entries(a.Hashes)) assert.equal(sha(fs.readFileSync(p)),h,p);
const parent=JSON.parse(parentRaw), phases=['design','confirmation'], names=['majority_to_parity','parity_to_majority'];
const key=r=>`${r.Phase}/${r.Case}/${r.Index}`;
const parents=new Map(parent.Records.map(r=>[key(r),r]));
assert.equal(a.Records.length,128);
const seen=new Set(), rows=[];
const near=(x,y,tol=1e-11)=>assert(Number.isFinite(x)&&Number.isFinite(y)&&Math.abs(x-y)<=tol,`${x} != ${y}`);
const probability=x=>assert(Number.isFinite(x)&&x>=0&&x<=1);
const loss=(p,q)=>(p-q)**2+q*(1-q);
const pop=x=>{let n=0;while(x){n+=x&1;x>>=1;}return n;};
for(const r of a.Records){
 const o=r.Original,k=key(o);assert(phases.includes(o.Phase)&&names.includes(o.Case));assert(!seen.has(k));seen.add(k);
 assert.deepEqual(o,parents.get(k));assert.equal(r.Trace.length,256);assert.equal(r.Blocks.length,8);
 for(const [i,s] of r.Trace.entries()){
  assert.equal(s.Step,i);assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512);assert.equal(typeof s.Y,'boolean');
  let majority=o.Case==='majority_to_parity';if(i>=128)majority=!majority;
  const count=pop(s.X&o.Rules[i<128?0:1]);const truth=majority?count>=2:count%2===1;
  near(s.Q,truth?.95:.05);
  assert.equal(s.P.length,9);for(const p of s.P){assert.equal(p.length,2);p.forEach(probability);}
  assert.equal(s.Weights.length,2);
  for(let view=0;view<2;view++){const w=s.Weights[view];assert.equal(w.length,4);w.forEach(probability);near(w.reduce((s,x)=>s+x,0),1);
   near(s.P[8][view],w.reduce((sum,x,j)=>sum+x*s.P[[0,1,6,7][j]][view],0));
   if(i%32===0)for(let j=0;j<4;j++)near(w[j],o.Publications[i/32].Weights[view][j]);}
 }
 // Verify original aggregate scores independently from all recorded forecasts.
 for(let arm=0;arm<9;arm++)for(let view=0;view<2;view++){
  near(o.Realized[arm][view],r.Trace.reduce((sum,s)=>sum+(s.P[arm][view]-Number(s.Y))**2,0),1e-9);
  for(let segment=0;segment<2;segment++){const samples=r.Trace.slice(segment*128),n=samples.length;
   near(o.Metrics[arm][view][segment].Brier,samples.reduce((sum,s)=>sum+loss(s.P[arm][view],s.Q)/n,0));
   near(o.Metrics[arm][view][segment].Accuracy,samples.reduce((sum,s)=>sum+(s.P[arm][view]>=.5?s.Q:1-s.Q)/n,0));
  }
 }
 for(let block=0;block<8;block++)for(let view=0;view<2;view++){
  const samples=r.Trace.slice(block*32,(block+1)*32);let numerator=0,denominator=0;
  for(const s of samples){const d=s.P[6][view]-s.P[0][view];numerator+=d*(s.Q-s.P[0][view]);denominator+=d*d;}
  const alpha=denominator===0?0:Math.max(0,Math.min(1,numerator/denominator));
  const m={Generic:0,Short:0,Bank:0,Hull:0,Pair:0,Alpha:alpha},w=[0,0,0,0];
  for(const s of samples){const p=[0,1,6,7].map(k=>s.P[k][view]);
   m.Generic+=loss(p[0],s.Q)/32;m.Short+=loss(p[2],s.Q)/32;m.Bank+=loss(s.P[8][view],s.Q)/32;
   m.Hull+=loss(Math.max(Math.min(...p),Math.min(Math.max(...p),s.Q)),s.Q)/32;
   m.Pair+=loss(p[0]+alpha*(p[2]-p[0]),s.Q)/32;
   for(let j=0;j<4;j++)w[j]+=s.Weights[view][j]/32;
  }
  for(const [field,v]of Object.entries(m))near(r.Blocks[block][view][field],v);
  assert(m.Hull<=m.Pair+1e-12&&m.Hull<=m.Bank+1e-12&&m.Pair<=m.Generic+1e-12&&m.Pair<=m.Short+1e-12);
  rows.push({phase:o.Phase,scenario:o.Case,index:o.Index,block,view,m,w});
 }
}
const interval=values=>{const n=values.length,mean=values.reduce((s,x)=>s+x,0)/n,se=Math.sqrt(values.reduce((s,x)=>s+(x-mean)**2,0)/(n*(n-1)));return {mean,lower:mean-3.5*se,upper:mean+3.5*se};};
const cells=[];
for(const phase of phases)for(const scenario of names)for(let view=0;view<2;view++)for(let block=0;block<8;block++){
 const group=rows.filter(r=>r.phase===phase&&r.scenario===scenario&&r.view===view&&r.block===block);assert.equal(group.length,32);
 const mean=f=>group.reduce((s,r)=>s+f(r),0)/32;
 cells.push({phase,scenario,view,start:block*32,end:block*32+31,
  metrics:Object.fromEntries(['Generic','Short','Bank','Hull','Pair','Alpha'].map(k=>[k,mean(r=>r.m[k])])),
  meanWeights:[0,1,2,3].map(j=>mean(r=>r.w[j])),
  gains:Object.fromEntries(['Short','Bank','Hull','Pair'].map(k=>[k,interval(group.map(r=>r.m.Generic-r.m[k]))]))});
}
const result={protocol:a.Protocol,runtime:a.Runtime,parentSHA256:a.ParentSHA256,artifactSHA256:sha(raw),evaluatorSHA256:sha(fs.readFileSync(import.meta.filename)),streams:128,
 caveat:'Consumed switch-case diagnosis only; simulator oracle and block-optimal weights use hindsight, never learning. Approximate paired z=3.5 summaries are not new confirmation.',cells};
const output=JSON.stringify(result,null,2)+'\n';if(process.argv[2])fs.writeFileSync(process.argv[2],output,{flag:'wx',mode:0o600});else process.stdout.write(output);
