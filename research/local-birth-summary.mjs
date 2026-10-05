import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind,'header');assert.equal(header.PerCell,16);assert.equal(rows.length,160);
assert.deepEqual(header.Seeds,[2026092121,2026092122]);
for(const [path,digest]of Object.entries(header.Hashes)){assert.equal(hash(header.Sources[path]),digest);assert.equal(hash(fs.readFileSync(path)),digest);}
assert.equal(new Set(rows.map(r=>r.FitSeed)).size,160);assert.equal(new Set(rows.map(r=>r.FitHash)).size,160);
const mean=a=>a.reduce((s,v)=>s+v,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((s,v)=>s+(v-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const armNames=['Old','Mixture','BirthOld','BirthMixture'];
const summaries=[];
for(const phase of ['design','confirmation'])for(const scenario of ['stable','member_shift','common_shift','recurring','null']){
  const rs=rows.filter(r=>r.Original.Split===phase&&r.Original.Scenario===scenario);assert.equal(rs.length,16);
  assert.deepEqual(rs.map(r=>r.Original.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
  for(const r of rs){
    assert.equal(r.Old.SplitAt,r.Original.Arms[2].SplitAt);assert.equal(r.Mixture.SplitAt,r.Original.Arms[3].SplitAt);
    assert.equal(r.BirthOld.SplitAt,r.Old.SplitAt);assert.equal(r.BirthMixture.SplitAt,r.Mixture.SplitAt);assert.equal(r.Fits*3,r.Original.Fits);
    for(const [a,time]of [['BirthOld','ActivationOld'],['BirthMixture','ActivationMixture']]){assert(r[time]===-1||(r[time]>r[a].SplitAt&&r[a].SplitAt>=0&&r[time]<512));}
    for(const a of armNames){assert.equal(r[a].Full.N,512);assert.equal(r[a].Post.N,scenario==='recurring'?384:256);assert(r[a].Full.Cost<=3072);}
  }
  const loss=(r,a,period)=>r[a][period].Brier/r[a][period].N;
  const arms=Object.fromEntries(armNames.map(a=>[a,{postBrier:mean(rs.map(r=>loss(r,a,'Post'))),fullBrier:mean(rs.map(r=>loss(r,a,'Full'))),cost:mean(rs.map(r=>r[a].Full.Cost/512)),splits:rs.filter(r=>r[a].SplitAt>=0).length}]));
  const comparisons={};
  for(const [name,a,b]of [['actionMixture','Mixture','BirthMixture'],['actionOld','Old','BirthOld'],['gateBirth','BirthOld','BirthMixture'],['gateInherited','Old','Mixture']]){
    const post=interval(rs.map(r=>loss(r,a,'Post')-loss(r,b,'Post'))),full=interval(rs.map(r=>loss(r,a,'Full')-loss(r,b,'Full')));
    comparisons[name]={post,full,gainScreen:post.mean>=.005&&post.lower>0,nonHarmScreen:post.lower>=-.01&&full.lower>=-.01};
  }
  summaries.push({phase,scenario,n:16,arms,comparisons});
}
const out={sourceHash:hash(bytes),capturedSources:Object.keys(header.Hashes).length,trajectories:160,summaries,
  limits:'Small exploratory factorial pilot, mean +/-3.5SE screens; no rare-error coverage, delayed/dependent-data, calibration or daemon performance proof.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({...out,summaries:summaries.map(s=>({phase:s.phase,scenario:s.scenario,arms:s.arms,primary:s.comparisons.actionMixture.post,primaryGainScreen:s.comparisons.actionMixture.gainScreen,gateWithinBirth:s.comparisons.gateBirth.post}))},null,2));
