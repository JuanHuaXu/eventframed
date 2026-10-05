import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const [source,artifact,frozenPath,output,mode,clockArg]=process.argv.slice(2);
const start=clockArg===undefined?128:Number(clockArg);assert(Number.isInteger(start)&&start>=0&&start<=224&&start%32===0);
assert(mode===undefined||mode==='eight');
const indices=mode==='eight'?8:1,expectedCount=84*indices;
assert(source&&artifact&&frozenPath&&output);
const read=p=>fs.readFileSync(p,'utf8').trim().split('\n').map(JSON.parse);
const rows=read(artifact),frozen=frozenPath==='-'?[]:read(frozenPath).filter(r=>mode!=='eight'||r.Clock===start),key=r=>[r.Phase,r.Case,r.Index,r.Schedule].join(':');
assert(rows.every(r=>(r.Clock??128)===start));
assert.equal(rows.length,expectedCount);assert.equal(frozen.length,frozenPath==='-'?0:expectedCount);
const byKey=new Map(rows.map(r=>[key(r),r])),oldByKey=new Map(frozen.map(r=>[key(r),r]));
assert.equal(byKey.size,expectedCount);assert.equal(oldByKey.size,frozen.length);
const masks=Array.from({length:511},(_,i)=>i+1).filter(m=>m.toString(2).replaceAll('0','').length<=4);
const phi=(x,m)=>(m&~x).toString(2).replaceAll('0','').length%2?-1:1;
const close=(a,b)=>{assert(Number.isFinite(a)&&Number.isFinite(b));assert(Math.abs(a-b)<1e-9,`${a} != ${b}`);};
let header=true,checked=0,fitCount=0,maxMeanError=0,maxVarianceError=0,maxIterations=0,maxEvaluations=0;
const stops={},cells={},paired=[];
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})) {
  const s=JSON.parse(line);
  if(header){assert.equal(s.Version,'soft-learners-v120');header=false;continue;}
  const r=byKey.get(key(s));if(!r)continue;byKey.delete(key(s));checked++;
  const old=oldByKey.get(key(s));assert(old||frozenPath==='-');assert(r.Index>=0&&r.Index<indices);
  assert(r.Fits.length>=1&&r.Fits.length<=32);
  if(old){for(const name of ['LogOdds','Mean','Variance','Xi','Trace','Motion','Stop','Origins']) assert.deepEqual(r.Fits[0][name],old[name]);
  assert.equal(r.P[0],old.P[0][0]);}
  let fitIndex=-1,previous=null,prepared=null;
  const row={phase:r.Phase,scenario:r.Case,index:r.Index,schedule:r.Schedule,fits:r.Fits.length,expected:[0,0,0,0,0],realized:[0,0,0,0,0]};
  const cell=cells[`${r.Phase}:${r.Schedule}`]??={forecasts:0,fits:0,expected:[0,0,0,0,0],realized:[0,0,0,0,0]};
  cell.fits+=r.Fits.length;
  for(let t=0;t<32;t++) {
    const clock=start+t;
    let origins=[];
    for(let i=-16;i<clock;i++)if(i<0||(!s.Steps[i].Missing&&i+s.Steps[i].Delay<=clock))origins.push(i);
    origins=origins.slice(-64);
    if(JSON.stringify(origins)!==JSON.stringify(previous)) {
      fitIndex++;fitCount++;
      const f=r.Fits[fitIndex];assert(f);assert.equal(f.Clock,clock);assert.deepEqual(f.Origins,origins);
      if(t===0)assert.deepEqual(origins,s.Fits[start/32].Origins[0]);
      const samples=origins.map(i=>i<0?s.Initial[i+16]:{Bits:s.Steps[i].X,Outcome:s.Steps[i].Y});
      assert.equal(f.Xi.length,samples.length);
      for(const name of ['LogOdds','Mean','Variance'])assert.equal(f[name].length,255);
      const n=f.Trace.length;assert(n>0&&n<=1024);assert.equal(n,f.Motion.length);
      maxIterations=Math.max(maxIterations,n);
      for(let j=0;j<n;j++) {
        const [a,b,c]=f.Trace[j];assert([a,b,c,f.Motion[j]].every(Number.isFinite));assert(b>=a-1e-9&&c>=b-1e-9);
        if(j)close(a,f.Trace[j-1][2]);
      }
      assert(['bound-and-state','iteration-cap'].includes(f.Stop));
      if(f.Stop==='bound-and-state'){assert(Math.abs(f.Trace[n-1][2]-f.Trace[n-1][0])<=1e-6);assert(f.Motion[n-1]<=1e-6);}else assert.equal(n,1024);
      stops[f.Stop]=(stops[f.Stop]??0)+1;
      const w=f.Xi.map(x=>{assert(Number.isFinite(x)&&x>0);return Math.tanh(x/2)/(2*x);});
      const v0=1/(1+w.reduce((a,b)=>a+b,0)),ks=samples.reduce((a,s)=>a+(s.Outcome?.5:-.5),0);
      const ts=masks.map(m=>samples.reduce((a,s,i)=>a+w[i]*phi(s.Bits,m),0));
      const gs=[],hs=[];
      for(let j=0;j<255;j++) {
        const o=f.LogOdds[j],m=f.Mean[j],v=f.Variance[j];assert(Number.isFinite(o)&&Number.isFinite(m)&&Number.isFinite(v)&&v>0);
        const u=Math.exp(-Math.abs(o)),small=u/(1+u),large=1/(1+u);gs.push(o>=0?large:small);hs.push(o>=0?small:large);
      }
      prepared={f,v0,ks,ts,gs,hs};previous=origins;
    }
    assert.equal(r.Used[t],fitIndex);
    const step=s.Steps[clock];assert.equal(r.X[t],step.X);assert.equal(r.Q[t],step.Q);assert.equal(r.Y[t],step.Y);assert.deepEqual(r.Control[t],step.P);
    const {f,v0,ks,ts,gs,hs}=prepared;
    let mean=v0*ks,variance=v0;
    for(let j=0;j<255;j++){const a=phi(step.X,masks[j])-v0*ts[j],g=gs[j],h=hs[j],m=f.Mean[j],v=f.Variance[j];mean+=a*g*m;variance+=a*a*(g*v+g*h*m*m);}
    maxMeanError=Math.max(maxMeanError,Math.abs(mean-r.LogitMean[t]));maxVarianceError=Math.max(maxVarianceError,Math.abs(variance-r.LogitVariance[t]));
    close(mean,r.LogitMean[t]);close(variance,r.LogitVariance[t]);
    const st=r.Stats[t];assert(st.Evaluations>0&&st.Evaluations<=65537&&st.EstimatedError<=1e-10&&st.TailBound<=1.45e-14);maxEvaluations=Math.max(maxEvaluations,st.Evaluations);
    const ps=[r.P[t],old?old.P[t][0]:null,step.P[0],step.P[1],step.P[12]];
    ps.forEach((p,j)=>{if(p===null)return;assert(Number.isFinite(p)&&p>0&&p<1);const e=(p-step.Q)**2+step.Q*(1-step.Q),b=(p-Number(step.Y))**2;row.expected[j]+=e/32;row.realized[j]+=b/32;cell.expected[j]+=e;cell.realized[j]+=b;});cell.forecasts++;
  }
  assert.equal(fitIndex+1,r.Fits.length);paired.push(row);
}
assert.equal(checked,expectedCount);assert.equal(byKey.size,0);
for(const c of Object.values(cells))for(const name of ['expected','realized'])c[name]=c[name].map(x=>x/c.forecasts);
if(frozenPath==='-')for(const r of [...paired,...Object.values(cells)])for(const name of ['expected','realized'])r[name][1]=null;
const report={checked,fitCount,forecasts:checked*32,stops,maxIterations,maxEvaluations,maxMeanError,maxVarianceError,
  arms:['arrivalRefit','frozen32','generic64','Boolean64','Markov'],cells,paired,
  limitations:'Consumed-data immediate-publication diagnostic; compute charged but not equal-compute or asynchronous serving. Independent moments, not independent full-mixture integration.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
const {paired:details,...summary}=report;console.log(JSON.stringify(summary,null,2));
