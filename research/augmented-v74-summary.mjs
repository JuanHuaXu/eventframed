import fs from 'node:fs';
import crypto from 'node:crypto';
import {pairedUpper} from './paired-risk-v73.mjs';

const raw=fs.readFileSync('docs/experiments/mmm-augmented-evidence-v74.jsonl');
const [header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
if(header.Version!=='v74'||rows.length!==10240)throw Error('incomplete artifact');
for(const [p,h]of Object.entries(header.Hashes))if(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')!==h)throw Error(`source changed ${p}`);
const scenarios=['symmetric','sparse_null','heterogeneous_boundary','cancelling','homogeneous128','sparse128','sparse256','negative256','sparse384','weak256'];
const names=['uniform','old_ipw','variance_ipw','augmented'];
const mean=xs=>xs.reduce((a,b)=>a+b,0)/xs.length;
const summaries=[];let pass=true;
for(let phase=0;phase<2;phase++)for(let j=0;j<10;j++){
  const split=['design','confirmation'][phase],scenario=scenarios[j],rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario);
  if(rs.length!==512)throw Error('bad cell');
  for(let i=0;i<512;i++){
    const r=rs[i];if(r.Seed!==(2026117401+phase)*1e6+j*1000+i||!/^[a-f0-9]{64}$/.test(r.TapeSHA256))throw Error('seed/tape');
    if(r.First.length!==4||r.First.some(t=>!Number.isInteger(t)||t < -1||t>=512))throw Error('first alarm');
    if(r.Counts.length!==3||r.Counts.some(cs=>cs.length!==4||cs.some(c=>!Number.isInteger(c)||c<0)||cs.reduce((a,b)=>a+b,0)!==512))throw Error('query count');
    if(!Number.isInteger(r.Clipped)||r.Clipped<0||r.Clipped>512)throw Error('clipping count');
  }
  const at=j<4?512:j===4||j===5?128:j===8?384:256,delay=f=>f<at?512-at:f-at;
  for(let arm=0;arm<4;arm++){
    const alerts=rs.filter(r=>r.First[arm]>=0).length,p=alerts/512,z=1.959963984540054;
    const nullUpper=(p+z*z/1024+z*Math.sqrt(p*(1-p)/512+z*z/(4*512*512)))/(1+z*z/512);
    const s={split,scenario,arm:names[arm],alerts,nullUpper,meanClippedSteps:mean(rs.map(r=>r.Clipped))};let ok=nullUpper<=.02;
    if(j>=4){
      const gains=rs.map(r=>delay(r.First[0])-delay(r.First[arm])),g=mean(gains),radius=3.3*Math.sqrt(gains.reduce((a,v)=>a+(v-g)**2,0)/511/512);
      const control=mean(rs.map(r=>delay(r.First[0]))),premature=rs.filter(r=>r.First[arm]>=0&&r.First[arm]<at).length,detected=rs.filter(r=>r.First[arm]>=at);
      const harmful=rs.filter(r=>r.First[arm]<at&&r.First[0]>=at).length,beneficial=rs.filter(r=>r.First[0]<at&&r.First[arm]>=at).length;
      Object.assign(s,{premature,missed:512-detected.length,restrictedDelay:mean(rs.map(r=>delay(r.First[arm]))),detectedOnlyDelay:detected.length?mean(detected.map(r=>r.First[arm]-at)):-1,gain:g,gainLower:g-radius,gainPercent:100*g/control,harmful,beneficial,excessMiss:(harmful-beneficial)/512});
      ok=j===5||j===6?g>=.1*control&&g-radius>0&&premature<=rs.filter(r=>r.First[0]>=0&&r.First[0]<at).length:g>=-10;
      if(arm===3){s.missUpper=pairedUpper(harmful,beneficial,512,.05/12).upper;ok=ok&&s.excessMiss<=.01&&s.missUpper<=.02;}
    }
    if(arm===3){s.cellPass=ok;if(phase===1)pass&&=ok;}
    summaries.push(s);
  }
}
console.log(JSON.stringify({sha256:crypto.createHash('sha256').update(raw).digest('hex'),sourceCount:Object.keys(header.Hashes).length,streams:rows.length,primaryScreen:pass?'PASS':'FAIL',summaries},null,2));
