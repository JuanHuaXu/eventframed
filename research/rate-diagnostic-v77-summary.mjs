// Descriptive post-hoc analysis, not a new acceptance test or fresh validation.
import fs from 'node:fs';
import crypto from 'node:crypto';
const raw=fs.readFileSync('docs/experiments/mmm-rate-diagnostic-v77.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v77-posthoc'||rows.length!==10240)throw Error('incomplete diagnostic');
for(const[p,h]of Object.entries(header.Hashes))if(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')!==h)throw Error(`changed source ${p}`);
const names=['symmetric','sparse_null','heterogeneous_boundary','cancelling','homogeneous128','sparse128','sparse256','negative256','sparse384','weak256'];
const summaries=[];
for(let phase=0;phase<2;phase++)for(let j=0;j<10;j++){
  const split=['design','confirmation'][phase],scenario=names[j],rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario);
  if(rs.length!==512)throw Error('cell count');
  const at=j<4?512:j===4||j===5?128:j===8?384:256;
  for(let i=0;i<512;i++){
    const r=rs[i];if(r.Seed!==(2026117601+phase)*1e6+j*1000+i||!/^[a-f0-9]{64}$/.test(r.Tape))throw Error('identity');
    if(r.First.length!==3||r.First.some(f=>!Number.isInteger(f)||f < -1||f>=512)||r.Windows.length!==6)throw Error('shape');
    const sizes=[64,32,32,64,128,128];
    for(let k=0;k<6;k++){
      const w=r.Windows[k],want=j<4?(k===0?64:0):k===5?(at===128?128:0):k===4?(at<=256?128:0):sizes[k];
      if(w.N!==want||['Rate','Growth','Zero'].some(key=>w[key].length!==3||w[key].some(x=>!Number.isFinite(x))))throw Error('window');
      if(w.Rate.some(x=>x<0||x>.8*w.N+1e-10)||w.Zero.some(x=>!Number.isInteger(x)||x<0||x>w.N))throw Error('window range');
    }
  }
  const windows=Array.from({length:6},(_,k)=>{
    const n=rs.reduce((s,r)=>s+r.Windows[k].N,0);
    const average=key=>n?[0,1,2].map(a=>rs.reduce((s,r)=>s+r.Windows[k][key][a],0)/n):null;
    return {window:k,samples:n,meanRate:average('Rate'),meanExpectedLogFactor:average('Growth'),zeroFraction:average('Zero')};
  });
  const result={split,scenario,arms:['fixed','actual','oracle'],windows};
  if(j>=4){result.restrictedDelay=[0,1,2].map(a=>rs.reduce((s,r)=>s+(r.First[a]<at?512-at:r.First[a]-at),0)/512);result.missed=[0,1,2].map(a=>rs.filter(r=>r.First[a]<at).length);}
  summaries.push(result);
}
console.log(JSON.stringify({interpretation:'post-hoc simulator diagnostic, not implementable oracle or new confirmation',sha256:crypto.createHash('sha256').update(raw).digest('hex'),streams:rows.length,summaries},null,2));
