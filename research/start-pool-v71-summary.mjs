import fs from 'node:fs';
import zlib from 'node:zlib';
import crypto from 'node:crypto';

const raw = fs.readFileSync('docs/experiments/mmm-start-pool-v71.json.gz');
const o = JSON.parse(zlib.gunzipSync(raw));
for (const [file, hash] of Object.entries(o.Hashes)) {
  if (crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex') !== hash) throw Error(`hash mismatch: ${file}`);
}
if (o.Records.length !== 10240) throw Error('incomplete artifact');
const names = ['fixed','grid','pooled_fixed','pooled_grid'];
const scenarios = ['symmetric','sparse','boundary_low','boundary_high','dependent','moderate128','moderate256','strong256','negative256','weak256'];
const groups = Map.groupBy(o.Records, r=>`${r.Split}/${r.Scenario}`);
if (groups.size !== 20) throw Error('wrong cell count');
const mean = xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const cpUpper = (k,n,alpha)=>{
  if (k===n) return 1;
  if (k===0) return -Math.expm1(Math.log(alpha)/n);
  let lo=0, hi=1;
  for(let iteration=0;iteration<80;iteration++) {
    const p=(lo+hi)/2;
    let logp=n*Math.log1p(-p),cdf=Math.exp(logp);
    for(let i=1;i<=k;i++) { logp+=Math.log(n-i+1)-Math.log(i)+Math.log(p)-Math.log1p(-p); cdf+=Math.exp(logp); }
    if(cdf>alpha) lo=p; else hi=p;
  }
  return hi;
};
const summaries=[],pass=[true,true,true];
let delayDominanceViolations=0;
for(const [key,rs] of groups) {
  if(rs.length!==512 || new Set(rs.map(r=>r.Seed)).size!==512) throw Error(`incomplete/duplicate cell ${key}`);
  const [split,scenario]=key.split('/');
  const j=scenarios.indexOf(scenario),at=j<5?512:j===5?128:256;
  for(const r of rs) {
    if(r.Differences.length!==512 || r.Differences.some(d=>![-1,0,1].includes(d))) throw Error('invalid tape');
    if(r.First.length!==4 || r.First.some(t=>!Number.isInteger(t)||t < -1||t>=512)) throw Error('invalid first alert');
    for(let arm=0;arm<2;arm++) if(r.First[arm]>=0&&(r.First[arm+2]<0||r.First[arm+2]>r.First[arm])) delayDominanceViolations++;
  }
  const restricted=f=>f<at?512-at:f-at;
  for(let arm=0;arm<4;arm++) {
    const alerts=rs.filter(r=>r.First[arm]>=0).length;
    const premature=j<5?0:rs.filter(r=>r.First[arm]>=0&&r.First[arm]<at).length;
    const detected=j<5?[]:rs.filter(r=>r.First[arm]>=at);
    const missed=j<5?0:512-detected.length;
    const delays=j<5?rs.map(()=>0):rs.map(r=>restricted(r.First[arm]));
    const gains=j<5?rs.map(()=>0):rs.map(r=>restricted(r.First[0])-restricted(r.First[arm]));
    const avg=mean(gains), sd=Math.sqrt(gains.reduce((s,x)=>s+(x-avg)**2,0)/511/512);
    const z=1.959963984540054,p=alerts/512;
    const alertUpper=(p+z*z/1024+z*Math.sqrt(p*(1-p)/512+z*z/(4*512*512)))/(1+z*z/512);
    const s={split,scenario,arm:names[arm],alerts,premature,missed,delay:mean(delays),detectedOnly:detected.length?mean(detected.map(r=>r.First[arm]-at)):-1,gain:avg,lower:avg-3.3*sd,alertUpper};
    const old=o.Summary.find(v=>v.Split===split&&v.Scenario===scenario&&v.Arm===names[arm]);
    for(const [field,value] of Object.entries({Alerts:alerts,Premature:premature,Missed:missed,RestrictedDelay:s.delay,Gain:s.gain,Lower:s.lower,AlertUpper:s.alertUpper})) {
      if(!old||Math.abs(old[field]-value)>1e-9) throw Error(`summary mismatch ${key}/${names[arm]} ${field}`);
    }
    let ok=j<5?alertUpper<=.02:j===5||j===6?avg>=.1*mean(rs.map(r=>restricted(r.First[0])))&&s.lower>0&&premature<=rs.filter(r=>r.First[0]>=0&&r.First[0]<at).length:avg>=-10;
    if(arm>0&&j>=5) {
      const harmful=rs.filter(r=>r.First[arm]<at&&r.First[0]>=at).length;
      const excess=mean(rs.map(r=>Number(r.First[arm]<at)-Number(r.First[0]<at)));
      const upper=cpUpper(harmful,512,.05/30);
      const b=o.MissBounds.find(v=>v.Split===split&&v.Scenario===scenario&&v.Arm===names[arm]);
      if(!b||b.Harmful!==harmful||Math.abs(b.Excess-excess)>1e-12||Math.abs(b.Upper-upper)>1e-9) throw Error('miss bound mismatch');
      Object.assign(s,{harmful,excessMiss:excess,missUpper:upper});
      ok=ok&&excess<=.01&&upper<=.02;
    }
    if(arm>0&&split==='confirmation') pass[arm-1]&&=ok;
    s.cellPass=ok;
    summaries.push(s);
  }
}
if(JSON.stringify(pass)!==JSON.stringify(o.Pass)||delayDominanceViolations) throw Error('pass/dominance mismatch');
console.log(JSON.stringify({sha256:crypto.createHash('sha256').update(raw).digest('hex'),sourceCount:Object.keys(o.Hashes).length,streams:o.Records.length,pass,delayDominanceViolations,summaries},null,2));
