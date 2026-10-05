import fs from 'node:fs';
import crypto from 'node:crypto';
const sha=data=>crypto.createHash('sha256').update(data).digest('hex');
const raw=fs.readFileSync('docs/experiments/mmm-interaction-diagnostic-v89.jsonl');
const pr=fs.readFileSync('docs/experiments/mmm-age-breadth-v88.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
const [ph,...parent]=pr.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v89'||header.ParentSHA256!==sha(pr)||rows.length!==2560||parent.length!==rows.length)throw Error('provenance/count');
for(const[p,h]of Object.entries({...ph.Hashes,...header.Hashes}))if(sha(fs.readFileSync(p))!==h)throw Error(`source changed ${p}`);
const close=(a,b)=>Math.abs(a-b)<1e-9;
const integer=(x,a,b)=>Number.isInteger(x)&&x>=a&&x<=b;
for(let i=0;i<rows.length;i++){
 const r=rows[i];if(JSON.stringify(r.Original)!==JSON.stringify(parent[i])||r.Windows.length!==8)throw Error('original parity');
 let total=0;
 for(const w of r.Windows){
  if(w.N!==64||!['RequiredFrames','Covered','SubsetGuide','AgeGuide','InnerAvailable'].every(k=>integer(w[k],0,w.N))||w.Covered>w.RequiredFrames||w.AgeGuide>w.SubsetGuide||w.SubsetGuide>w.InnerAvailable)throw Error('counts');
  if(!w.Guide.every(x=>integer(x,0,w.N))||w.Guide.reduce((a,b)=>a+b,0)!==w.N||w.Available.length!==5||!w.Available.every(x=>integer(x,0,w.N)))throw Error('availability');
  if(![w.MixtureBrier,w.FullMixtureBrier].every(x=>Number.isFinite(x)&&x>=0&&x<=w.N)||!w.Brier.every((ps,j)=>ps.length===2&&ps.every(x=>Number.isFinite(x)&&x>=0&&x<=w.Available[j])))throw Error('Brier bounds');
  if(!close(w.Weights.reduce((a,b)=>a+b,0),w.N)||!close(w.InnerWeights.reduce((a,b)=>a+b,0),w.InnerAvailable))throw Error('weight totals');
  total+=w.MixtureBrier;
 }
 if(!close(total,r.Original.Arms[1].Full.Brier))throw Error('mixture score parity');
}
const scenarios=[...new Set(parent.map(r=>r.Scenario))],summaries=[];
for(const split of ['design','confirmation'])for(const scenario of scenarios)for(const schedule of ['immediate','jitter31_missing20']){
 const rs=rows.filter(r=>r.Original.Split===split&&r.Original.Scenario===scenario&&r.Original.Schedule.Name===schedule);
 if(rs.length!==64)throw Error('cell size');
 for(let window=0;window<8;window++){
  const ws=rs.map(r=>r.Windows[window]),sum=f=>ws.reduce((s,w)=>s+f(w),0),n=sum(w=>w.N),req=sum(w=>w.RequiredFrames),inner=sum(w=>w.InnerAvailable);
  summaries.push({split,scenario,schedule,window,start:64*window,end:64*window+63,forecasts:n,requiredFrames:req,
   requiredCoverage:req?sum(w=>w.Covered)/req:null,mixtureBrier:sum(w=>w.MixtureBrier)/n,fullInputMixtureBrier:sum(w=>w.FullMixtureBrier)/n,
   outerWeights:[0,1,2,3].map(j=>sum(w=>w.Weights[j])/n),innerWeights:[0,1,2].map(j=>inner?sum(w=>w.InnerWeights[j])/inner:null),
   guideFraction:[0,1,2].map(j=>sum(w=>w.Guide[j])/n),subsetGuideFraction:sum(w=>w.SubsetGuide)/n,ageGuideFraction:sum(w=>w.AgeGuide)/n,
   models:['base','count_short','retained_subset','age_subset','long'].map((name,j)=>{const available=sum(w=>w.Available[j]);return{name,available,actualBrier:available?sum(w=>w.Brier[j][0])/available:null,fullInputBrier:available?sum(w=>w.Brier[j][1])/available:null}}),
  });
 }
}
const output=JSON.stringify({sha256:sha(raw),parentSHA256:header.ParentSHA256,sourceCount:Object.keys(header.Hashes).length,evaluatorSHA256:sha(fs.readFileSync('research/interaction-diagnostic-v89-summary.mjs')),classification:'post-hoc diagnostic; not a rescue or confirmation',originalParity:true,scheduleRuns:rows.length,underlyingTrajectories:1280,summaries},null,2)+'\n';
if(process.argv[2])fs.writeFileSync(process.argv[2],output,{flag:'wx',mode:0o600});else process.stdout.write(output);
