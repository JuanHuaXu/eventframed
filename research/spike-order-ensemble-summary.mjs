import fs from 'node:fs';
import assert from 'node:assert/strict';
const [forwardPath,reversePath,output]=process.argv.slice(2);
assert(forwardPath && reversePath && output);
const read=p=>fs.readFileSync(p,'utf8').trim().split('\n').map(JSON.parse);
const forward=read(forwardPath),reverse=read(reversePath);
assert.equal(forward.length,84);assert.equal(reverse.length,84);
const cells={},paired=[];
let maxOrderDifference=0;
for(let i=0;i<84;i++) {
  const f=forward[i],r=reverse[i];assert.equal(r.Order,'descending');
  for(const key of ['Phase','Case','Index','Schedule','Origins','X','Q','Y','Control']) assert.deepEqual(f[key],r[key]);
  const cell=cells[`${f.Phase}:${f.Schedule}`]??={n:0,expected:[0,0,0,0,0,0],realized:[0,0,0,0,0,0]};
  const row={phase:f.Phase,scenario:f.Case,schedule:f.Schedule,expected:[0,0,0,0,0,0]};
  for(let t=0;t<32;t++) {
    const p=f.P[t][0],q=r.P[t][0];assert(p>0&&p<1&&q>0&&q<1);
    maxOrderDifference=Math.max(maxOrderDifference,Math.abs(p-q));
    const ps=[p,q,(p+q)/2,f.Control[t][0],f.Control[t][1],f.Control[t][12]];
    ps.forEach((v,j)=>{
      const expected=(v-f.Q[t])**2+f.Q[t]*(1-f.Q[t]);
      cell.expected[j]+=expected;cell.realized[j]+=(v-Number(f.Y[t]))**2;row.expected[j]+=expected/32;
    });cell.n++;
  }
  paired.push(row);
}
for(const c of Object.values(cells))for(const field of ['expected','realized'])c[field]=c[field].map(x=>x/c.n);
const report={arms:['ascending','descending','equalEnsemble','generic64','Boolean64','Markov'],maxOrderDifference,cells,paired,
  limitations:'Consumed one-index pilot. Equal averaging is a bounded predictive ensemble, not exact Bayesian model averaging. No full-goal validation.'};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n',{flag:'wx'});
const {paired:details,...summary}=report;console.log(JSON.stringify(summary,null,2));
