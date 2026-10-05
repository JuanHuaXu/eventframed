import fs from 'node:fs';
import crypto from 'node:crypto';
const sha = data => crypto.createHash('sha256').update(data).digest('hex');
const raw = fs.readFileSync('docs/experiments/mmm-delay-diagnostic-v86.jsonl');
const parentRaw = fs.readFileSync('docs/experiments/mmm-delayed-learning-v85.jsonl');
const [header, ...rows] = raw.toString().trim().split('\n').map(JSON.parse);
const [parentHeader, ...parent] = parentRaw.toString().trim().split('\n').map(JSON.parse);
if (header.Version !== 'v86' || header.ParentSHA256 !== sha(parentRaw) || rows.length !== 2560 || parent.length !== rows.length) throw Error('artifact provenance');
for (const [p, h] of Object.entries({ ...parentHeader.Hashes, ...header.Hashes })) if (sha(fs.readFileSync(p)) !== h) throw Error(`source changed ${p}`);
const close = (a, b) => Math.abs(a - b) < 1e-9;
for (let i = 0; i < rows.length; i++) {
  const r = rows[i];
  if (JSON.stringify(r.Original) !== JSON.stringify(parent[i]) || r.Windows.length !== 8) throw Error('original parity');
  let brier = 0, applied = 0, stale = 0;
  for (const w of r.Windows) {
    const ints = ['N', 'Available', 'SubsetGuide', 'ModelAge', 'NewestAge', 'OldestAge', 'Support', 'Applied', 'Stale', 'AppliedAge', 'StaleAge'];
    if (!ints.every(k => Number.isInteger(w[k]) && w[k] >= 0) || w.N !== 64 || w.Available > w.N || w.SubsetGuide > w.Available ||
        w.Guide.reduce((a,b)=>a+b,0) !== w.N || w.ModelAge > 512*w.Available || w.OldestAge > 512*w.Available || w.NewestAge > w.OldestAge ||
        w.Support > 64*w.Available || w.CurrentRegimeFraction < 0 || w.CurrentRegimeFraction > w.Available) throw Error('window accounting');
    if (![w.MixtureBrier, ...w.ExpertBrier].every(x => Number.isFinite(x) && x >= 0 && x <= w.N) ||
        !w.InnerBrier.every(x => Number.isFinite(x) && x >= 0 && x <= w.Available) ||
        !close(w.Weights.reduce((a,b)=>a+b,0), w.N) || !close(w.InnerWeights.reduce((a,b)=>a+b,0), w.Available)) throw Error('window metrics');
    brier += w.MixtureBrier; applied += w.Applied; stale += w.Stale;
  }
  if (!close(brier, r.Original.Arms[1].Full.Brier) || applied !== r.Original.Stats[1].Applied || stale !== r.Original.Stats[1].Stale) throw Error('aggregate parity');
}
const summaries = [];
for (const split of ['design','confirmation']) for (const scenario of ['stable','member_shift','common_shift','recurring','null'])
for (const schedule of ['immediate','delay16','missing20','jitter31_missing20']) {
  const rs = rows.filter(r=>r.Original.Split===split&&r.Original.Scenario===scenario&&r.Original.Schedule.Name===schedule);
  if(rs.length!==64)throw Error('cell count');
  for(let window=0;window<8;window++){
    const ws=rs.map(r=>r.Windows[window]);
    const sum=fn=>ws.reduce((s,w)=>s+fn(w),0), n=sum(w=>w.N), av=sum(w=>w.Available);
    const conditional=fn=>av ? sum(fn)/av : null;
    const applied=sum(w=>w.Applied),stale=sum(w=>w.Stale);
    summaries.push({split,scenario,schedule,window,start:window*64,end:window*64+63,observations:n,available:av,
      mixtureBrier:sum(w=>w.MixtureBrier)/n,
      expertBrier:Array.from({length:4},(_,i)=>sum(w=>w.ExpertBrier[i])/n),
      outerWeights:Array.from({length:4},(_,i)=>sum(w=>w.Weights[i])/n),
      innerBrier:[0,1].map(i=>conditional(w=>w.InnerBrier[i])),
      innerWeights:[0,1].map(i=>conditional(w=>w.InnerWeights[i])),
      guideFraction:[0,1,2].map(i=>sum(w=>w.Guide[i])/n),subsetGuideFraction:sum(w=>w.SubsetGuide)/n,
      modelAge:conditional(w=>w.ModelAge),newestTrainingAge:conditional(w=>w.NewestAge),oldestTrainingAge:conditional(w=>w.OldestAge),
      support:conditional(w=>w.Support),currentRegimeFraction:conditional(w=>w.CurrentRegimeFraction),
      applied,stale,appliedAge:applied?sum(w=>w.AppliedAge)/applied:null,staleAge:stale?sum(w=>w.StaleAge)/stale:null,
    });
  }
}
const result=JSON.stringify({sha256:sha(raw),parentSHA256:header.ParentSHA256,sourceCount:Object.keys(header.Hashes).length,
  evaluatorSHA256:sha(fs.readFileSync('research/delay-diagnostic-v86-summary.mjs')),classification:'post-hoc diagnostic; not confirmation',
  scheduleRuns:rows.length,originalParity:true,summaries},null,2)+'\n';
if(process.argv[2])fs.writeFileSync(process.argv[2],result,{flag:'wx',mode:0o600});else process.stdout.write(result);
