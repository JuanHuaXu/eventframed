import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [rawPath,summaryPath]=process.argv.slice(2),summary=JSON.parse(fs.readFileSync(summaryPath)),[header,...rows]=fs.readFileSync(rawPath,'utf8').trim().split('\n').map(JSON.parse);
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');for(const [p,h]of Object.entries(summary.hashes))assert.equal(sha(p),h);for(const [p,h]of Object.entries(header.Hashes))assert.equal(sha(p),h);
assert.equal(rows.length,5120);assert.equal(new Set(rows.map(r=>`${r.Split}:${r.Scenario}:${r.Index}`)).size,5120);
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-10),average=a=>a.reduce((s,x)=>s+x/a.length,0);
let metrics=0,intervals=0,probabilityChecks=0,pilotPass=true;
for(const c of summary.comparisons){
  const rs=rows.filter(r=>r.Split===c.split&&r.Scenario===c.scenario),at=c.scenario==='recurring'?128:256;assert.equal(rs.length,512);
  const restricted=t=>t<at?512-at:t-at;
  for(let arm=0;arm<5;arm++){
    const s=summary.summaries.find(s=>s.split===c.split&&s.scenario===c.scenario&&s.arm===rs[0].Arms[arm].Arm);assert(s);
    const times=rs.map(r=>r.Arms[arm].SplitAt),detected=times.filter(t=>t>=at);
    assert.equal(s.splits,times.filter(t=>t!==-1).length);assert.equal(s.early,times.filter(t=>t>=0&&t<at).length);assert.equal(s.absent,times.filter(t=>t===-1).length);assert.equal(s.validDetections,detected.length);
    if(detected.length)near(s.detectedOnlyDelay,average(detected.map(t=>t-at)));else assert.equal(s.detectedOnlyDelay,null);
    near(s.restrictedDelay,average(times.map(restricted)));near(s.fullBrier,average(rs.map(r=>r.Arms[arm].Full.Brier/512)));near(s.postBrier,average(rs.map(r=>r.Arms[arm].Post.Brier/(512-at))));near(s.postAccuracy,average(rs.map(r=>r.Arms[arm].Post.Correct/(512-at))));near(s.foreground,average(rs.map(r=>r.Arms[arm].Full.Cost/512)));metrics+=5;
  }
  for(const [field,denom,saved]of [['Full',512,c.fullBrierGain],['Post',512-at,c.postBrierGain]]){
    const gains=rs.map(r=>(r.Arms[2][field].Brier-r.Arms[3][field].Brier)/denom),m=average(gains);let pair=0;
    for(let i=0;i<512;i++)for(let j=i+1;j<512;j++)pair+=(gains[i]-gains[j])**2;
    const se=Math.sqrt(pair/(512*512*511));near(saved.mean,m);near(saved.lower,m-3.5*se);near(saved.upper,m+3.5*se);intervals++;
  }
  const old=average(rs.map(r=>restricted(r.Arms[2].SplitAt))),candidate=average(rs.map(r=>restricted(r.Arms[3].SplitAt)));near(old,c.oldDelay);near(candidate,c.newDelay);
  const early=a=>rs.filter(r=>r.Arms[a].SplitAt>=0&&r.Arms[a].SplitAt<at).length;
  const pass=c.scenario==='member_shift'?(old-candidate>=.1*old&&early(3)<=early(2)&&c.postBrierGain.mean>=.005&&c.postBrierGain.lower>0):(c.fullBrierGain.lower>=-.01&&c.postBrierGain.lower>=-.01);
  assert.equal(c.pilotCellPass,pass);pilotPass&&=pass;
}
for(const b of summary.falseRevocations){
  const rs=rows.filter(r=>r.Split===b.split&&r.Scenario===b.scenario),k=rs.filter(r=>r.Arms.find(a=>a.Arm===b.arm).SplitAt>=0).length;assert.equal(k,b.k);assert.equal(b.n,512);near(b.alpha,.05/12);
  if(k===0)near(b.upper,1-Math.pow(b.alpha,1/512));
  else if(k===512)assert.equal(b.upper,1);
  else {
    const p=b.upper;let term=(1-p)**512,sum=term;for(let i=1;i<=k;i++){term*=((513-i)/i)*p/(1-p);sum+=term;}near(sum,b.alpha);
  }
  assert.equal(b.pass,b.upper<=.02);probabilityChecks++;
}
assert.equal(summary.falseRevocations.length,12);assert.equal(summary.coveragePass,summary.falseRevocations.every(b=>b.pass));assert.equal(summary.pilotPass,pilotPass);
console.log(JSON.stringify({status:'PASS',records:rows.length,metrics,intervals,probabilityChecks,pilotPass,coveragePass:summary.coveragePass,hashes:{[rawPath]:sha(rawPath),[summaryPath]:sha(summaryPath),[import.meta.filename]:sha(import.meta.filename)},scope:'Independent count/metric aggregation, pairwise paired-variance calculation, binomial CDF inversion and unchanged acceptance gates; raw tape replay is a separate Go test'}));
