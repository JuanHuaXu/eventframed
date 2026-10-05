import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
const [source, screenPath, output] = process.argv.slice(2);
assert(source && screenPath && output);
const screen=JSON.parse(fs.readFileSync(screenPath)), lookup=new Map(screen.results.map(r=>[r.key,r]));
const records=[];let header=true,checks=0,maxError=0;
const project=(q,ps)=>Math.max(Math.min(...ps),Math.min(Math.max(...ps),q));
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})) {
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  const key=[r.Phase,r.Case,r.Index,r.Schedule].join(':'),old=lookup.get(key);assert(old);lookup.delete(key);
  const periods=Array.from({length:2},()=>Array(7).fill(0));
  for(let t=0;t<256;t++) {
    const s=r.Steps[t],q=s.Q,b=s.P[12],p=old.forecasts[0][t];
    // Q is inaccessible generator truth. These are upper-bound diagnostics,
    // not a policy, conditional-risk estimate, or candidate training signal.
    const h2=project(q,[b,s.P[10]]),h6=project(q,[b,...[0,1,2,3,10].map(a=>s.P[a])]);
    const excess=(u,v)=>(u-q)**2-(v-q)**2;
    const values=[excess(p,b),excess(p,s.P[1]),excess(p,h2),excess(h2,h6),excess(h6,q),excess(q,b),excess(q,s.P[1])];
    for(const c of [0,1]) {
      const error=Math.abs(values[c]-values[2]-values[3]-values[4]-values[5+c]);
      maxError=Math.max(maxError,error);assert(error<1e-12);checks++;
    }
    for(const k of [2,3,4])assert(values[k]>=-1e-12);
    values.forEach((v,k)=>{periods[0][k]+=v/256;if(t>=192)periods[1][k]+=v/64;});
  }
  records.push({key,periods});
}
assert.equal(lookup.size,0);assert.equal(records.length,672);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
const groups=[];
for(let phase=0;phase<2;phase++)for(let c=0;c<21;c++)for(let schedule=0;schedule<2;schedule++)for(let period=0;period<2;period++){
  const rows=records.filter(r=>{const[p,k,,s]=r.key.split(':').map(Number);return p===phase&&k===c&&s===schedule;});assert.equal(rows.length,8);
  const metrics=Array.from({length:7},(_,a)=>mean(rows.map(r=>r.periods[period][a])));
  const fail=screen.gates.filter(g=>g.arm===0&&g.key===[phase,c,schedule].join(':')&&g.period===(period===0?'whole':'terminal64')&&g.type==='nonharm'&&!g.pass);
  groups.push({key:[phase,c,schedule].join(':'),period,metrics,failedControls:fail.map(g=>g.control),rangeVsMarkov:[Math.min(...rows.map(r=>r.periods[period][0])),Math.max(...rows.map(r=>r.periods[period][0]))]});
}
const out={checks,maxError,names:['excessVsMarkov','excessVsBoolean','twoHeadSelectionGap','missingHeadGap','sixHeadApproximationGap','noiseMinusMarkov','noiseMinusBoolean'],groups,
  limitations:'Consumed fixed-forecast oracle projection. Exact decomposition, not causal attribution or evidence that oracle routing is learnable. Includes all cases, not only failures.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({checks,maxError,markovFailures:groups.filter(g=>g.failedControls.includes(6)),booleanFailures:groups.filter(g=>g.failedControls.includes(5)).length},null,2));
