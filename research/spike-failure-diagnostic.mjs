import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const [source,resultFile,output]=process.argv.slice(2);assert(source&&resultFile&&output);
const result=JSON.parse(fs.readFileSync(resultFile,'utf8'));
assert.equal(result.summary.cohort,'spike-independent-v1');
const byKey=new Map(result.results.map(r=>[r.key,r]));assert.equal(byKey.size,672);
const delta=(p,b,q)=>(p-b)*(p+b-2*q);
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const close=(a,b)=>assert(Math.abs(a-b)<1e-12,`${a} != ${b}`);
// The excess is affine in q, so endpoint checks also bound every q in [0,1].
for(const [b,c,q,w]of [[.1,.8,.3,.4],[.5,.5,.7,1],[.9,.2,0,.8]]){
  const p=b+w*(c-b),raw=delta(c,b,q);
  close(delta(p,b,q),w*raw-w*(1-w)*(c-b)**2);
}
const windows=[],counterexample=[];let header=true,pointwise=0,forecastCount=0;
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const src=JSON.parse(line);if(header){assert.equal(src.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[src.Phase,src.Case,src.Index,src.Schedule].join(':'),r=byKey.get(key);assert(r);byKey.delete(key);
  const trace=src.Steps.map((s,i)=>{
    const b=s.P[12],c=r.armPredictions[6][i],f=r.forecasts[i],q=s.Q,d=c-b;
    const raw=delta(c,b,q),actual=delta(f.p,b,q),unclippedP=b+f.proposed*d,unclipped=delta(unclippedP,b,q);
    close(f.p,b+f.w*d);close(actual,f.w*raw-f.w*(1-f.w)*d*d);
    const oracleWeight=d===0?0:Math.max(0,Math.min(1,(q-b)/d));
    const oracleDelta=delta(b+oracleWeight*d,b,q);assert(oracleDelta<=1e-12);
    const worst=Math.max(delta(f.p,b,0),delta(f.p,b,1));assert(actual<=worst+1e-12);
    forecastCount++;pointwise+=Number(actual>.01);
    return{clock:i,b,c,q,y:s.Y,missing:s.Missing,delay:s.Delay,weight:f.w,proposed:f.proposed,clipped:f.clipped,ledger:f.ledger,reserved:f.reserved,
      raw,actual,unclipped,clipEffect:actual-unclipped,weightedRaw:f.w*raw,mixtureBenefit:-f.w*(1-f.w)*d*d,
      oracleWeight,oracleDelta,worst,realized:delta(f.p,b,Number(s.Y))};
  });
  for(let block=0;block<8;block++){
    const rows=trace.slice(block*32,block*32+32);
    const sum={key,clock:block*32,change:src.Change};
    for(const field of ['raw','actual','unclipped','clipEffect','weightedRaw','mixtureBenefit','oracleDelta','weight','proposed','realized','worst'])sum[field]=mean(rows.map(r=>r[field]));
    sum.clipped=rows.filter(r=>r.clipped).length;sum.rawHarmFrames=rows.filter(r=>r.raw>0).length;
    close(sum.actual,r.blocks[block].expected[0]-r.blocks[block].expected[2]);
    close(sum.actual,sum.weightedRaw+sum.mixtureBenefit);
    assert(sum.realized<=.01+1e-12);
    windows.push(sum);
  }
  if(key==='0:20:6:0')counterexample.push(...trace);
}
assert.equal(byKey.size,0);assert.equal(windows.length,5376);assert.equal(counterexample.length,256);
function aggregate(rows){return{windows:rows.length,...Object.fromEntries(['raw','actual','unclipped','clipEffect','weightedRaw','mixtureBenefit','oracleDelta','weight','proposed'].map(f=>[f,mean(rows.map(r=>r[f]))])),clippedFrames:rows.reduce((s,r)=>s+r.clipped,0),clipImprovedWindows:rows.filter(r=>r.clipEffect<0).length,clipWorsenedWindows:rows.filter(r=>r.clipEffect>0).length};}
const harmed=windows.filter(r=>r.actual>.01),safe=windows.filter(r=>r.actual<=.01);
const summary={forecastCount,pointwiseHarmsAbovePoint01:pointwise,all:aggregate(windows),harmed:aggregate(harmed),other:aggregate(safe),harmedWithRawImprovement:harmed.filter(r=>r.raw<0).length,harmedWithUnclippedImprovement:harmed.filter(r=>r.unclipped<0).length,
  limitations:'Post-hoc diagnosis on consumed synthetic data. Oracle q is evaluator-only, never an available model input. Algebraic decomposition is not a causal intervention experiment or new quality validation. Realized fixed-window protection does not imply expected-score protection.'};
fs.writeFileSync(output,JSON.stringify({summary,harmed,windows,counterexample},null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(summary,null,2));
