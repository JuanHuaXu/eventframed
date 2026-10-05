// Descriptive post-hoc mechanism analysis; not a new acceptance gate.
import fs from 'node:fs';
import crypto from 'node:crypto';
const raw=fs.readFileSync('docs/experiments/mmm-post-split-v81.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v81-posthoc'||rows.length!==640)throw Error('incomplete diagnostic');
for(const[p,h]of Object.entries(header.Hashes))if(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')!==h)throw Error(`source changed ${p}`);
const summaries=[];
for(const split of ['design','confirmation'])for(const scenario of ['stable','member_shift','common_shift','recurring','null']){
  const rs=rows.filter(r=>r.Original.Split===split&&r.Original.Scenario===scenario);if(rs.length!==64)throw Error('cell count');
  for(let i=0;i<64;i++){
    const r=rs[i];if(r.Original.Index!==i||r.Windows.length!==8)throw Error('identity');
    if(Math.abs(r.Windows.reduce((s,w)=>s+w.MixtureBrier,0)-r.Original.Arms[3].Full.Brier)>1e-10)throw Error('scored law drift');
    for(const w of r.Windows){
      if(w.N!==64||w.Guide.length!==3||w.Guide.reduce((s,x)=>s+x,0)!==64||['Split','Bit2','Parity'].some(k=>!Number.isInteger(w[k])||w[k]<0||w[k]>64))throw Error('window counts');
      for(const k of ['Weights','ExpertBrier','Available','Support'])if(w[k].length!==4||w[k].some(x=>!Number.isFinite(x)||x<0))throw Error('invalid aggregate');
      if(w.Available.some(x=>!Number.isInteger(x)||x>64)||Math.abs(w.Weights.reduce((s,x)=>s+x,0)-64)>1e-8)throw Error('availability/weight');
      if(w.ModelBrier.length!==4||w.ModelBrier.some((xs,j)=>xs.length!==4||xs.some(x=>!Number.isFinite(x)||x<0||x>w.Available[j])))throw Error('model score');
    }
  }
  for(let k=0;k<8;k++){
    const ws=rs.map(r=>r.Windows[k]),n=4096;
    const sum=fn=>ws.reduce((s,w)=>s+fn(w),0),average=key=>sum(w=>w[key])/n;
    const s={split,scenario,steps:[k*64,k*64+63],splitFraction:average('Split'),bit2Fraction:average('Bit2'),parityFraction:average('Parity'),guide:[0,1,2].map(j=>sum(w=>w.Guide[j])/n),weights:[0,1,2,3].map(j=>sum(w=>w.Weights[j])/n),expertBrier:[0,1,2,3].map(j=>sum(w=>w.ExpertBrier[j])/n),mixtureBrier:average('MixtureBrier'),models:[]};
    for(let i=0;i<4;i++){
      const count=sum(w=>w.Available[i]);
      s.models.push({model:['base','short','local','pooled'][i],available:count,meanSupport:count?sum(w=>w.Support[i])/count:null,masks:['actual','all','bit2','parity'],brier:count?[0,1,2,3].map(j=>sum(w=>w.ModelBrier[i][j])/count):null});
    }
    summaries.push(s);
  }
}
console.log(JSON.stringify({status:'post-hoc diagnostic, not a rescue',sha256:crypto.createHash('sha256').update(raw).digest('hex'),trajectories:rows.length,summaries},null,2));
