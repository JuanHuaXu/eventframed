import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';

const [source, artifact, output, mode='pilot'] = process.argv.slice(2);
assert(source && artifact && output);
assert(['pilot','screen'].includes(mode));
const expectedFits=mode==='screen'?2016:84;
const clocks=mode==='screen'?[0,128,224]:[128];
const rows = fs.readFileSync(artifact, 'utf8').trim().split('\n').map(JSON.parse);
assert.equal(rows.length, expectedFits);
const key = r => [r.Phase, r.Case, r.Index, r.Schedule,r.Clock??128].join(':');
const byKey = new Map(rows.map(r => [key(r), r]));
assert.equal(byKey.size, expectedFits);
for(const r of rows) {
  assert([0,1].includes(r.Phase) && [0,1].includes(r.Schedule));
  assert(Number.isInteger(r.Case) && r.Case>=0 && r.Case<21);
  assert(Number.isInteger(r.Index) && r.Index>=0 && r.Index<(mode==='screen'?8:1));
  assert(clocks.includes(r.Clock??128));
}
const masks = Array.from({length: 511}, (_, i) => i + 1)
  .filter(m => m.toString(2).replaceAll('0', '').length <= 4);
assert.equal(masks.length, 255);
const phi = (x, m) => ((m & ~x).toString(2).replaceAll('0', '').length % 2) ? -1 : 1;
const close = (a, b, tol=1e-9) => { assert(Number.isFinite(a) && Number.isFinite(b)); assert(Math.abs(a-b) <= tol, `${a} != ${b}`); };
let header=true, checked=0, maxMeanError=0, maxVarianceError=0, maxEvaluations=0, clamps=0;
const cells = {}, stops = {}, iterations=[], paired=[];
for await (const line of readline.createInterface({input:fs.createReadStream(source), crlfDelay:Infinity})) {
  const s=JSON.parse(line);
  if (header) { assert.equal(s.Version,'soft-learners-v120'); header=false; continue; }
  for(const clock of clocks) {
  const sourceKey=key({...s,Clock:clock});
  const r=byKey.get(sourceKey);
  if (!r) continue;
  assert(r.Order===undefined || r.Order==='descending');
  const rowMasks=r.Order==='descending'?[...masks].reverse():masks;
  byKey.delete(sourceKey); checked++;
  const origins=[];
  for (let i=-16;i<clock;i++) if(i<0 || (!s.Steps[i].Missing && i+s.Steps[i].Delay<=clock)) origins.push(i);
  const admitted=origins.slice(-64);
  assert.deepEqual(r.Origins,admitted);
  assert.deepEqual(admitted,s.Fits[clock/32].Origins[0]);
  const samples=admitted.map(i=>i<0?s.Initial[i+16]:{Bits:s.Steps[i].X,Outcome:s.Steps[i].Y});
  assert.equal(r.Xi.length,samples.length);
  for(const a of [r.Inclusion,r.Exclusion,r.LogOdds,r.Mean,r.Variance]) assert.equal(a.length,255);
  const w=r.Xi.map(x=>{assert(x>0 && Number.isFinite(x));return Math.tanh(x/2)/(2*x);});
  const v0=1/(1+w.reduce((a,b)=>a+b,0));
  const ks=samples.reduce((a,s)=>a+(s.Outcome?.5:-.5),0);
  const t=rowMasks.map(m=>samples.reduce((a,s,i)=>a+w[i]*phi(s.Bits,m),0));
  const n=r.Trace.length;
  assert(n>0 && n<=1024); assert.equal(r.Motion.length,n);
  for(let i=0;i<n;i++) {
    const [a,b,c]=r.Trace[i]; assert([a,b,c,r.Motion[i]].every(Number.isFinite));
    assert(b>=a-1e-9 && c>=b-1e-9);
    if(i) close(a,r.Trace[i-1][2]);
  }
  stops[r.Stop]=(stops[r.Stop]??0)+1; iterations.push(n);
  assert(['bound-and-state','iteration-cap'].includes(r.Stop));
  if(r.Stop==='bound-and-state') { assert(Math.abs(r.Trace[n-1][2]-r.Trace[n-1][0])<=1e-6); assert(r.Motion[n-1]<=1e-6); }
  else assert.equal(n,1024);
  const cellKey=mode==='screen'?`${clock}:${r.Phase}:${r.Schedule}`:`${r.Phase}:${r.Schedule}`;
  const cell=cells[cellKey]??={fits:0, forecasts:0, expected:[0,0,0,0,0],realized:[0,0,0,0,0]};
  const local={Phase:r.Phase,Case:r.Case,Index:r.Index,Schedule:r.Schedule,Clock:clock,expected:[0,0,0,0,0],realized:[0,0,0,0,0]};
  cell.fits++;
  for(let i=0;i<32;i++) {
    const step=s.Steps[clock+i];
    assert.equal(r.X[i],step.X); assert.equal(r.Q[i],step.Q); assert.equal(r.Y[i],step.Y); assert.deepEqual(r.Control[i],step.P);
    let mean=v0*ks,variance=v0;
    for(let j=0;j<255;j++) {
      const g=r.Inclusion[j],h=r.Exclusion[j],o=r.LogOdds[j],m=r.Mean[j],v=r.Variance[j],a=phi(step.X,rowMasks[j])-v0*t[j];
      assert(g>=0 && g<=1 && h>=0 && h<=1 && v>0 && Number.isFinite(m) && Number.isFinite(o));
      const u=Math.exp(-Math.abs(o)),small=u/(1+u),large=1/(1+u);
      close(g,o>=0?large:small,3e-16);close(h,o>=0?small:large,3e-16);
      mean+=a*g*m; variance+=a*a*(g*v+g*h*m*m);
    }
    maxMeanError=Math.max(maxMeanError,Math.abs(mean-r.LogitMean[i]));
    maxVarianceError=Math.max(maxVarianceError,Math.abs(variance-r.LogitVariance[i]));
    close(mean,r.LogitMean[i]);close(variance,r.LogitVariance[i]);
    close(r.P[i][1],Math.max(1e-12,Math.min(1-1e-12,1/(1+Math.exp(-mean)))));
    const st=r.Stats[i];
    assert(st.Evaluations>0 && st.Evaluations<=65537 && st.EstimatedError<=1e-10 && st.TailBound<=1.45e-14);
    maxEvaluations=Math.max(maxEvaluations,st.Evaluations); clamps+=Number(st.RangeClamped);
    const ps=[...r.P[i],step.P[0],step.P[1],step.P[12]];
    ps.forEach((p,j)=>{
      assert(p>=0 && p<=1);
      const expected=(p-step.Q)**2+step.Q*(1-step.Q),realized=(p-Number(step.Y))**2;
      cell.expected[j]+=expected;cell.realized[j]+=realized;
      local.expected[j]+=expected/32;local.realized[j]+=realized/32;
    });
    cell.forecasts++;
  }
  paired.push(local);
  }
}
assert.equal(checked,expectedFits);assert.equal(byKey.size,0);
for(const cell of Object.values(cells)) for(const field of ['expected','realized']) cell[field]=cell[field].map(x=>x/cell.forecasts);
const report={checked,forecasts:checked*32,arms:['mixture','plugin','generic64','Boolean64','Markov'],stops,
 iterations:{min:Math.min(...iterations),max:Math.max(...iterations)},maxMeanError,maxVarianceError,maxEvaluations,clamps,cells,
 limitations:'Consumed exploratory tape; moment reconstruction is independent, not an independent full-mixture probability audit. Quadrature estimates are not uniform certificates.'};
if(mode==='screen') report.paired=paired;
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
const {paired:details,...summary}=report;
console.log(JSON.stringify(summary,null,2));
