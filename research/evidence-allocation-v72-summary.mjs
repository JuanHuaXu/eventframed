import fs from 'node:fs';
import crypto from 'node:crypto';

const raw=fs.readFileSync('docs/experiments/mmm-evidence-allocation-v72.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v72'||rows.length!==10240) throw Error('incomplete artifact');
for(const [file,hash] of Object.entries(header.Hashes)) {
  if(crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex')!==hash) throw Error(`changed source ${file}`);
}
const scenarios=['symmetric','sparse_null','heterogeneous_boundary','cancelling','homogeneous128','sparse128','sparse256','negative256','sparse384','weak256'];
const names=['uniform','naive_targeting','corrected_targeting'];
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const cp=(k,n,alpha)=>{
  if(k===n)return 1;
  if(k===0)return -Math.expm1(Math.log(alpha)/n);
  let lo=0,hi=1;
  for(let it=0;it<80;it++) {
    const p=(lo+hi)/2;let lp=n*Math.log1p(-p),cdf=Math.exp(lp);
    for(let i=1;i<=k;i++){lp+=Math.log(n-i+1)-Math.log(i)+Math.log(p)-Math.log1p(-p);cdf+=Math.exp(lp);}
    if(cdf>alpha)lo=p;else hi=p;
  }
  return hi;
};
const summaries=[];let passed=true;
for(let phase=0;phase<2;phase++)for(let j=0;j<10;j++) {
  const split=['design','confirmation'][phase],scenario=scenarios[j];
  const rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario);
  if(rs.length!==512)throw Error('wrong cell size');
  for(let i=0;i<rs.length;i++) {
    const r=rs[i];
    if(r.Seed!==(2026117201+phase)*1000000+j*1000+i||!/^[a-f0-9]{64}$/.test(r.TapeSHA256))throw Error('bad seed/tape');
    if(r.First.length!==3||r.First.some(t=>!Number.isInteger(t)||t < -1||t>=512))throw Error('bad first alarm');
    if(r.Counts.length!==2||r.Counts.some(cs=>cs.length!==4||cs.some(c=>!Number.isInteger(c)||c<0)||cs.reduce((a,b)=>a+b,0)!==512))throw Error('bad cost');
  }
  const at=j<4?512:(j===4||j===5)?128:j===8?384:256;
  const delay=f=>f<at?512-at:f-at;
  for(let arm=0;arm<3;arm++) {
    const alerts=rs.filter(r=>r.First[arm]>=0).length;
    const p=alerts/512,z=1.959963984540054;
    const nullUpper=(p+z*z/1024+z*Math.sqrt(p*(1-p)/512+z*z/(4*512*512)))/(1+z*z/512);
    const s={split,scenario,arm:names[arm],alerts,observedMean:mean(rs.map(r=>r.Means[arm])),nullUpper};
    let ok=nullUpper<=.02;
    if(j>=4) {
      const ds=rs.map(r=>delay(r.First[arm])),gain=rs.map(r=>delay(r.First[0])-delay(r.First[arm]));
      const g=mean(gain),radius=3.3*Math.sqrt(gain.reduce((a,v)=>a+(v-g)**2,0)/511/512);
      const premature=rs.filter(r=>r.First[arm]>=0&&r.First[arm]<at).length;
      const detected=rs.filter(r=>r.First[arm]>=at);
      const missed=512-detected.length;
      const harmful=rs.filter(r=>r.First[arm]<at&&r.First[0]>=at).length;
      const excess=mean(rs.map(r=>Number(r.First[arm]<at)-Number(r.First[0]<at)));
      const upper=cp(harmful,512,.05/12),control=mean(rs.map(r=>delay(r.First[0])));
      Object.assign(s,{premature,missed,restrictedDelay:mean(ds),detectedOnlyDelay:detected.length?mean(detected.map(r=>r.First[arm]-at)):-1,gain:g,gainLower:g-radius,gainPercent:100*g/control,harmful,excessMiss:excess,missUpper:upper});
      ok=(j===5||j===6)?g>=.1*control&&g-radius>0&&premature<=rs.filter(r=>r.First[0]>=0&&r.First[0]<at).length:g>=-10;
      ok=ok&&excess<=.01&&upper<=.02;
    }
    s.cellPass=ok;
    if(phase===1&&arm===2)passed&&=ok;
    summaries.push(s);
  }
}
console.log(JSON.stringify({sha256:crypto.createHash('sha256').update(raw).digest('hex'),sources:Object.keys(header.Hashes).length,streams:rows.length,correctedScreen:passed?'PASS':'FAIL',summaries},null,2));
