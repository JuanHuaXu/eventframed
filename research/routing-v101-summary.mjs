import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const path = 'docs/experiments/mmm-routing-v101.json';
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const raw = fs.readFileSync(path), a = JSON.parse(raw);
assert.equal(a.Protocol, 'mmm-routing-v101');
assert.equal(Object.keys(a.Hashes).length, 21);
for (const [p, hash] of Object.entries(a.Hashes)) assert.equal(sha(fs.readFileSync(p)), hash, p);
const names = ['parity1','parity2','parity3','parity4','complement4','majority3','mux3','constant','null','dependent4','majority_to_parity','parity_to_majority'];
const phases = ['design','confirmation'];
assert.equal(a.Records.length, 768);
const seen = new Set(), tapes = new Set();
for (const r of a.Records) {
  assert(names.includes(r.Case) && phases.includes(r.Phase));
  assert(Number.isInteger(r.Index) && r.Index >= 0 && r.Index < 32);
  const key = `${r.Phase}/${r.Case}/${r.Index}`; assert(!seen.has(key)); seen.add(key);
  assert(/^[0-9a-f]{64}$/.test(r.Tape)); assert(!tapes.has(r.Tape)); tapes.add(r.Tape);
  assert(r.Rules.length === 2 && r.Rules.every(x => Number.isInteger(x) && x >= 0 && x < 512));
  assert.equal(r.Metrics.length, 12); assert.equal(r.Realized.length, 12);
  for (const row of r.Metrics) {assert.equal(row.length, 2); for (const view of row) {
    assert.equal(view.length, 2); for (const m of view) for (const k of ['Brier','Accuracy']) assert(Number.isFinite(m[k]) && m[k] >= 0 && m[k] <= 1);
  }}
  for (const row of r.Realized) {assert.equal(row.length, 2); assert(row.every(v => Number.isFinite(v) && v >= 0 && v <= 256));}
  assert.equal(r.Masked.length,2);assert.equal(r.Neutral.length,2);assert.equal(r.RejectAt.length,8);
  for(let v=0;v<8;v++){assert.equal(r.RejectAt[v].length,2);for(const row of r.RejectAt[v]){assert.equal(row.length,4);assert(row.every(t=>Number.isInteger(t)&&(t===-1||(t>=v*32&&t<(v+1)*32))));}}
  for(let view=0;view<2;view++){
    let masked=0,neutral=0;
    for(let step=0;step<256;step++){const rejected=r.RejectAt[Math.floor(step/32)][view].map(t=>t>=0&&t<step);masked+=Number(rejected.some(Boolean));neutral+=Number(rejected.every(Boolean));}
    assert.equal(r.Masked[view],masked);assert.equal(r.Neutral[view],neutral);
  }
  assert.equal(r.ComparativeMasked.length,2);assert.equal(r.ComparativeNeutral.length,2);assert.equal(r.ComparativeRejectAt.length,8);
  for(let v=0;v<8;v++){assert.equal(r.ComparativeRejectAt[v].length,2);for(const row of r.ComparativeRejectAt[v]){assert.equal(row.length,4);assert(row.every(t=>Number.isInteger(t)&&(t===-1||(t>=v*32&&t<(v+1)*32))));}}
  for(let view=0;view<2;view++){
    let masked=0,neutral=0;
    for(let step=0;step<256;step++){const rejected=r.ComparativeRejectAt[Math.floor(step/32)][view].map(t=>t>=0&&t<step);masked+=Number(rejected.some(Boolean));neutral+=Number(rejected.every(Boolean));}
    assert.equal(r.ComparativeMasked[view],masked);assert.equal(r.ComparativeNeutral[view],neutral);
  }
  assert.equal(r.BlockMetrics.length,12);
  for(let arm=0;arm<12;arm++){
    assert.equal(r.BlockMetrics[arm].length,8);
    for(const block of r.BlockMetrics[arm]){
      assert.equal(block.length,2);
      for(const m of block)for(const key of ['Brier','Accuracy'])assert(Number.isFinite(m[key])&&m[key]>=0&&m[key]<=1);
    }
    for(let view=0;view<2;view++)for(const key of ['Brier','Accuracy']){
      const all=r.BlockMetrics[arm].reduce((sum,b)=>sum+b[view][key],0)/8;
      const late=r.BlockMetrics[arm].slice(4).reduce((sum,b)=>sum+b[view][key],0)/4;
      assert(Math.abs(all-r.Metrics[arm][view][0][key])<1e-12);
      assert(Math.abs(late-r.Metrics[arm][view][1][key])<1e-12);
    }
  }
  for(const [field,n]of [['RoutedWeights',5],['RawBlockWeights',4]]){
    assert.equal(r[field].length,8);
    for(const block of r[field]){
      assert.equal(block.length,2);
      for(const w of block){assert.equal(w.length,n);assert(w.every(v=>Number.isFinite(v)&&v>=0&&v<=1+1e-12));assert(Math.abs(w.reduce((s,v)=>s+v,0)-1)<1e-12);}
    }
  }
  assert.deepEqual(r.RoutedNeutral,r.ComparativeNeutral);
  assert.equal(r.BankViolation.length, 2);
  assert(r.BankViolation.every(x => Number.isFinite(x) && x >= 0 && x <= 1e-8));
  assert.equal(r.Publications.length, 8);
  for (const [j,p] of r.Publications.entries()) {
    const step=j*32; assert.equal(p.Step,step);
    const counts=[Math.min(step+16,64),Math.min(step+16,32)];
    assert.deepEqual(p.Counts,counts);
    assert.deepEqual(p.Origins,counts.map(n => [step-n,step-1]));
    assert.equal(p.Weights.length,2);
    for (const w of p.Weights) {assert.equal(w.length,4); assert(w.every(x => Number.isFinite(x) && x>=0 && x<=1));assert(Math.abs(w.reduce((s,x)=>s+x,0)-1)<1e-12);}
  }
  assert.equal(r.MaxViolation.length, 2);
  for (const row of r.MaxViolation) {assert.equal(row.length, 2); assert(row.every(v => Number.isFinite(v) && v >= 0 && v <= 1e-8));}
}
const interval = values => {
  const n = values.length, mean = values.reduce((s,x) => s+x, 0)/n;
  const se = Math.sqrt(values.reduce((s,x) => s+(x-mean)**2, 0)/(n*(n-1)));
  return {mean, lower:mean-3.5*se, upper:mean+3.5*se};
};
const cells = [], gates = [], publications = [], rejectionSummary = [], blocks = [];
for (const phase of phases) for (const scenario of names) {
  const rows = a.Records.filter(r => r.Phase === phase && r.Case === scenario); assert.equal(rows.length,32);
  const mean = f => rows.reduce((s,r) => s+f(r),0)/rows.length;
  for(let view=0;view<2;view++) rejectionSummary.push({phase,scenario,view,method:'neutral',meanMaskedFrames:mean(r=>r.Masked[view]),meanNeutralFrames:mean(r=>r.Neutral[view]),meanRejectedVersions:mean(r=>r.RejectAt.reduce((n,v)=>n+v[view].filter(t=>t>=0).length,0))});
  for(let view=0;view<2;view++) rejectionSummary.push({phase,scenario,view,method:'comparative',meanMaskedFrames:mean(r=>r.ComparativeMasked[view]),meanNeutralFrames:mean(r=>r.ComparativeNeutral[view]),meanRejectedVersions:mean(r=>r.ComparativeRejectAt.reduce((n,v)=>n+v[view].filter(t=>t>=0).length,0))});
  for(let block=0;block<8;block++)for(let view=0;view<2;view++)blocks.push({
    phase,scenario,view,step:block*32,
    rawWeights:Array.from({length:4},(_,k)=>mean(r=>r.RawBlockWeights[block][view][k])),
    routedWeights:Array.from({length:5},(_,k)=>mean(r=>r.RoutedWeights[block][view][k])),
    arms:Array.from({length:12},(_,arm)=>({brier:mean(r=>r.BlockMetrics[arm][block][view].Brier),accuracy:mean(r=>r.BlockMetrics[arm][block][view].Accuracy)}))
  });
  for (let j=0;j<8;j++) for (let view=0;view<2;view++) publications.push({phase,scenario,view,step:j*32,counts:rows[0].Publications[j].Counts,origins:rows[0].Publications[j].Origins,meanWeights:Array.from({length:4},(_,k)=>mean(r=>r.Publications[j].Weights[view][k]))});
  for (let view = 0; view < 2; view++) for (let segment = 0; segment < 2; segment++) {
    const arms = Array.from({length:12}, (_,arm) => ({brier:mean(r => r.Metrics[arm][view][segment].Brier), accuracy:mean(r => r.Metrics[arm][view][segment].Accuracy),
      gain:interval(rows.map(r => r.Metrics[0][view][segment].Brier-r.Metrics[arm][view][segment].Brier))}));
    const label = {phase,scenario,view,segment};
    cells.push({...label, arms});
    const primary = arms[11].gain;
    gates.push({...label,kind:segment === 0 ? 'nonharm' : 'late_nonharm',pass:-primary.lower <= .01,value:-primary.lower});
    if (view === 0 && segment === 0 && ['parity3','parity4','complement4'].includes(scenario)) gates.push({...label,kind:'interaction_gain',pass:primary.lower > 0 && primary.mean >= .005,...primary});
    if (view === 0 && segment === 1 && scenario.includes('_to_')) gates.push({...label,kind:'recovery_gain',pass:primary.lower > 0 && primary.mean >= .005,...primary});
  }
}
assert.equal(gates.length,106);
assert.equal(a.Nulls.length,2);
const nulls=a.Nulls.flatMap((counts,mode)=>{
 assert.equal(counts.length,2);
 return counts.map((count,method)=>{
  assert(Number.isInteger(count)&&count>=0&&count<=4096);
  const n=4096,z=1.959963984540054,p=count/n;
  const upper=(p+z*z/(2*n)+z*Math.sqrt(p*(1-p)/n+z*z/(4*n*n)))/(1+z*z/n);
  return {mode:['independent','shared'][mode],method:['neutral','comparative'][method],families:n,rejected:count,rate:p,wilson95Upper:upper,pass:upper<=.01};
 });
});
const result = {protocol:a.Protocol,artifactSHA256:sha(raw),evaluatorSHA256:sha(fs.readFileSync(import.meta.filename)),runtime:a.Runtime,
  streams:768,stepsPerStream:256,armOrder:['generic','parity','skeptical_BMA','online_no_share','online_share','interval','generic32','Boolean32','bank','neutral_gated','comparative_gated','evidence_routed'],
  caveat:'Approximate paired trajectory z=3.5 intervals, not anytime confidence sequences; complete full-frame training feedback in both views.',
  passed:gates.filter(g => g.pass).length,total:106,verdict:gates.every(g => g.pass)&&nulls.every(n=>n.pass)?'PASS':'FAIL',
  maxBoundViolation:Math.max(...a.Records.flatMap(r => r.MaxViolation.flat())),bankMaxViolation:Math.max(...a.Records.flatMap(r => r.BankViolation)),gates,cells,publications,nulls,rejectionSummary,blocks};
const output = JSON.stringify(result,null,2)+'\n';
if (process.argv[2]) fs.writeFileSync(process.argv[2],output,{flag:'wx',mode:0o600}); else process.stdout.write(output);
