import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind,'header');assert.equal(rows.length,160);
assert.equal(hash(fs.readFileSync('docs/experiments/mmm-local-birth-v1.jsonl')),header.ArchiveHash);
for(const [name,digest]of Object.entries(header.Hashes)){assert.equal(hash(header.Sources[name]),digest);assert.equal(hash(fs.readFileSync(name)),digest);}
const mean=a=>a.reduce((n,x)=>n+x,0)/a.length;
const interval=a=>{const m=mean(a),se=Math.sqrt(a.reduce((n,x)=>n+(x-m)**2,0)/(a.length-1)/a.length);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};
const summaries=[];
for(const phase of ['design','confirmation'])for(const scenario of ['stable','member_shift','common_shift','recurring','null']){
  const rs=rows.filter(r=>r.Original.Original.Split===phase&&r.Original.Original.Scenario===scenario);assert.equal(rs.length,16);
  assert.deepEqual(rs.map(r=>r.Original.Original.Index).sort((a,b)=>a-b),Array.from({length:16},(_,i)=>i));
  for(const r of rs)for(const [fixed,control]of [['FixedOld','Old'],['FixedMixture','Mixture']]){
    assert.equal(r[fixed].Full.N,512);assert.equal(r[fixed].Post.N,scenario==='recurring'?384:256);
    assert.equal(r[fixed].SplitAt,r.Original[control].SplitAt);
    for(const p of ['Full','Post'])assert.equal(r[fixed][p].Cost,r.Original[control][p].Cost);
  }
  const get=(r,a)=>r[a]??r.Original[a];
  const loss=(r,a,p)=>get(r,a)[p].Brier/get(r,a)[p].N;
  const arms=Object.fromEntries(['Old','Mixture','BirthOld','BirthMixture','FixedOld','FixedMixture'].map(a=>[a,mean(rs.map(r=>loss(r,a,'Post')))]));
  const contrasts={};
  for(const [name,a,b]of [['fixedVsInherited','Mixture','FixedMixture'],['fixedVsCoupled','BirthMixture','FixedMixture'],['fixedOldVsInherited','Old','FixedOld'],['gateWithinFixed','FixedOld','FixedMixture']]){
    const post=interval(rs.map(r=>loss(r,a,'Post')-loss(r,b,'Post'))),full=interval(rs.map(r=>loss(r,a,'Full')-loss(r,b,'Full')));
    contrasts[name]={post,full,gainScreen:post.mean>=.005&&post.lower>0,nonHarmScreen:post.lower>=-.01&&full.lower>=-.01};
  }
  summaries.push({phase,scenario,n:16,arms,contrasts});
}
const out={sourceHash:hash(bytes),archiveHash:header.ArchiveHash,capturedSources:Object.keys(header.Hashes).length,
  trajectories:160,summaries,limits:'Consumed fixed-observer diagnostic, not fresh confirmation; paired trajectory mean +/-3.5SE, no rare-error or broad deployment claim.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({...out,summaries:summaries.filter(s=>['member_shift','recurring'].includes(s.scenario))},null,2));
