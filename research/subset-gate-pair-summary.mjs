import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind,'header');assert.equal(header.PerCell,16);assert.equal(rows.length,160);
assert.deepEqual(header.Seeds,[2026092111,2026092112]);
for(const [path,digest]of Object.entries(header.Hashes)){
  assert.equal(hash(header.Sources[path]),digest);assert.equal(hash(fs.readFileSync(path)),digest);
}
const names=['stable','member_shift','common_shift','recurring','null'];
assert.equal(new Set(rows.map(r=>[r.Original.Split,r.Original.Scenario,r.Original.Index].join(':'))).size,160);
assert.equal(new Set(rows.map(r=>r.FitSeed)).size,160);assert.equal(new Set(rows.map(r=>r.FitHash)).size,160);
const mean=a=>a.reduce((x,y)=>x+y,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((n,x)=>n+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const summaries=[];
for(const phase of ['design','confirmation'])for(const scenario of names){
  const rs=rows.filter(r=>r.Original.Split===phase&&r.Original.Scenario===scenario);assert.equal(rs.length,16);
  assert.deepEqual(rs.map(r=>r.Original.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
  for(const r of rs){
    assert.equal(r.OldSubset.SplitAt,r.Original.Arms[2].SplitAt);assert.equal(r.MixtureSubset.SplitAt,r.Original.Arms[3].SplitAt);
    assert.equal(r.Fits*3,r.Original.Fits);
    for(const a of [r.OldSubset,r.MixtureSubset]){assert.equal(a.Full.N,512);assert.equal(a.Post.N,scenario==='recurring'?384:256);assert(a.Full.Cost<=6*512);}
  }
  const loss=(r,arm,period)=>r[arm][period].Brier/r[arm][period].N;
  const postGain=interval(rs.map(r=>loss(r,'OldSubset','Post')-loss(r,'MixtureSubset','Post')));
  const fullGain=interval(rs.map(r=>loss(r,'OldSubset','Full')-loss(r,'MixtureSubset','Full')));
  const changed=['member_shift','recurring'].includes(scenario),start=scenario==='recurring'?128:256;
  const delay=(r,arm)=>{const at=r[arm].SplitAt;return at>=start?at-start:512-start;};
  summaries.push({phase,scenario,n:16,
    oldPostBrier:mean(rs.map(r=>loss(r,'OldSubset','Post'))),mixturePostBrier:mean(rs.map(r=>loss(r,'MixtureSubset','Post'))),postGain,fullGain,
    oldCountPostBrier:mean(rs.map(r=>r.Original.Arms[2].Post.Brier/r.Original.Arms[2].Post.N)),
    mixtureCountPostBrier:mean(rs.map(r=>r.Original.Arms[3].Post.Brier/r.Original.Arms[3].Post.N)),
    oldSplits:rs.filter(r=>r.OldSubset.SplitAt>=0).length,mixtureSplits:rs.filter(r=>r.MixtureSubset.SplitAt>=0).length,
    oldRestrictedDelay:changed?mean(rs.map(r=>delay(r,'OldSubset'))):null,
    mixtureRestrictedDelay:changed?mean(rs.map(r=>delay(r,'MixtureSubset'))):null,
    oldCost:mean(rs.map(r=>r.OldSubset.Full.Cost/512)),mixtureCost:mean(rs.map(r=>r.MixtureSubset.Full.Cost/512)),
    positiveGainScreen:postGain.mean>=.005&&postGain.lower>0,
    nonHarmScreen:postGain.lower>=-.01&&fullGain.lower>=-.01});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,trajectories:160,summaries,
  limits:'Exploratory n16 per cell; mean +/-3.5SE screens, no anytime or rare-event guarantee; fixed representation/noise, immediate feedback; no complete goal validation.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(out,null,2));
