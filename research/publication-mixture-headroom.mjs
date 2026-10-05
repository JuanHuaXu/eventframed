import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

function optimum(rows) {
 let cross=0,square=0,baseline=0;
 for(const [p,q,y] of rows){assert([p,q,y].every(Number.isFinite));assert(p>=0&&p<=1&&q>=0&&q<=1&&(y===0||y===1));cross+=(p-y)*(q-p);square+=(q-p)**2;baseline+=(p-y)**2;}
 const alpha=square===0?0:Math.max(0,Math.min(1,-cross/square));
 const gain=(-2*alpha*cross-alpha*alpha*square)/rows.length;
 const direct=rows.reduce((s,[p,q,y])=>s+(p-y)**2-(p+alpha*(q-p)-y)**2,0)/rows.length;
 assert(Math.abs(gain-direct)<1e-12);assert(gain>=-1e-12);
 return {alpha,gain,baseline:baseline/rows.length};
}
assert(optimum([[.2,.2,1]]).gain===0);
assert(Math.abs(optimum([[.2,.8,1]]).gain-.6)<1e-12);
assert(Math.abs(optimum([[.2,.8,0],[.2,.8,1]]).alpha-.5)<1e-12);
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),rows=bytes.toString().trim().split('\n').slice(1).map(JSON.parse);
assert.equal(rows.length,192);
const results=rows.flatMap(r=>['Immediate','Delayed'].map(schedule=>({phase:r.Phase,case:r.Case,index:r.Index,schedule,...optimum(r[schedule].Frames.slice(256).map(f=>[f.Predictions[0].p,f.Predictions[2].p,Number(f.Y)]))})));
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const cells=[];
for(const phase of ['design','confirmation'])for(const name of ['copy_bit','copy_xor2','noise10_bit','reverses_xor2','stable_noise10','null_uniform'])for(const schedule of ['Immediate','Delayed']){
 const rs=results.filter(r=>r.phase===phase&&r.case===name&&r.schedule===schedule);assert.equal(rs.length,16);
 cells.push({phase,case:name,schedule,meanOracleGain:mean(rs.map(r=>r.gain)),meanAlpha:mean(rs.map(r=>r.alpha))});
}
fs.writeFileSync(output,JSON.stringify({sourceHash:crypto.createHash('sha256').update(bytes).digest('hex'),results,cells,limits:'Consumed hindsight upper bound for a constant convex weight chosen separately for each trajectory using its scored labels. Not deployable, not fresh confirmation, not a bound on time-varying policies. Ignores extra observation/computation cost, making the bound optimistic.'},null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify(cells.filter(c=>c.schedule==='Delayed')));
