import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const [input,output]=process.argv.slice(2);assert(input&&output);
const bytes=fs.readFileSync(input),[header,...rows]=bytes.toString().trim().split('\n').map(JSON.parse);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert.equal(header.Kind,'header');assert.equal(rows.length,160);
assert.equal(hash(fs.readFileSync('docs/experiments/mmm-subset-gate-pair-v1.jsonl')),header.ArchiveHash);
for(const [name,digest]of Object.entries(header.Hashes)){
  assert.equal(hash(header.Sources[name]),digest);assert.equal(hash(fs.readFileSync(name)),digest);
}
const groups=new Map();let checks=0,maxAffineError=0;
const clip=p=>Math.max(1e-6,Math.min(1-1e-6,p));
const add=(key,frame,y)=>{
  let g=groups.get(key);
  if(!g){g={n:0,weight:0,absMotion:0,maxMotion:0,rawSlotGain:0,directGain:0,guide:[0,0,0]};groups.set(key,g);}
  const motion=Math.abs(frame.WithLocal-frame.WithPool);
  g.n++;g.weight+=frame.Weight;g.absMotion+=motion;g.maxMotion=Math.max(g.maxMotion,motion);
  g.rawSlotGain+=(frame.Pool-y)**2-(frame.Local-y)**2;
  g.directGain+=(frame.WithPool-y)**2-(frame.WithLocal-y)**2;g.guide[frame.Guide]++;
};
for(const row of rows){
  const a=row.Original,o=a.Original;assert.equal(row.Steps.length,512);
  const sums={Old:{full:0,post:0},Mixture:{full:0,post:0}};
  for(const [i,s]of row.Steps.entries()){
    assert.equal(s.Step,i);const y=Number(s.Y);
    for(const arm of ['Old','Mixture']){
      const f=s[arm];assert(f.P>=0&&f.P<=1&&f.Weight>=0&&f.Weight<=1);assert([0,1,2].includes(f.Guide));
      sums[arm].full+=(f.P-y)**2;if(i>=(o.Scenario==='recurring'?128:256))sums[arm].post+=(f.P-y)**2;
      if(!f.Available)continue;
      const error=Math.abs(f.WithLocal-f.WithPool-f.Weight*(clip(f.Local)-clip(f.Pool)));
      assert(error<1e-14);maxAffineError=Math.max(maxAffineError,error);checks++;
      assert(Math.abs(f.P-(f.Split?f.WithLocal:f.WithPool))<1e-14);
      add([o.Split,o.Scenario,arm,'window'+Math.floor(i/64)].join('/'),f,y);
      if(s.Mixture.Split&&!s.Old.Split)add([o.Split,o.Scenario,arm,'only-new-split'].join('/'),f,y);
    }
  }
  for(const arm of ['Old','Mixture']){
    const expected=arm==='Old'?a.OldSubset:a.MixtureSubset;
    assert(Math.abs(sums[arm].full-expected.Full.Brier)<1e-12);
    assert(Math.abs(sums[arm].post-expected.Post.Brier)<1e-12);
  }
}
const summaries=[...groups].sort(([a],[b])=>a.localeCompare(b)).map(([key,g])=>({key,n:g.n,
  meanWeight:g.weight/g.n,meanAbsMotion:g.absMotion/g.n,maxMotion:g.maxMotion,
  rawSlotBrierGain:g.rawSlotGain/g.n,directMixtureBrierGain:g.directGain/g.n,
  guideFractions:g.guide.map(n=>n/g.n)}));
const out={sourceHash:hash(bytes),archiveHash:header.ArchiveHash,capturedSources:Object.keys(header.Hashes).length,
  trajectories:rows.length,frames:rows.length*512,affineChecks:checks,maxAffineError,summaries,
  limits:'Read-only fixed-mask substitution on consumed trajectories; no alternate acquisition/learning trajectory or new quality guarantee. Aggregates are descriptive, not independent-frame confidence intervals.'};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n',{flag:'wx'});
console.log(JSON.stringify({...out,summaries:summaries.filter(s=>s.key.includes('only-new-split'))},null,2));
