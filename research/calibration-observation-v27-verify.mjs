import fs from 'node:fs';
import crypto from 'node:crypto';
const M=27,n=150;
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const assert=(p,s)=>{if(!p)throw Error(s);};
const near=(a,b,s,tol=1e-10)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=tol,`${s}: ${a} != ${b}`);
const H=p=>p<=0||p>=1?0:-p*Math.log(p)-(1-p)*Math.log1p(-p);
const prior=()=>[.1,.8,...Array(25).fill(.004)];
const tables=base=>base.map((b,i)=>[b,(.9*b-.05)/.8,...[.1,.3,.5,.7,.9].flatMap(a=>[-.8,-.4,0,.4,.8].map(c=>Math.max(.02,Math.min(.98,a+c*i/(n-1))))) ]);
const cost=p=>({head:n+1,random:n+1,stratified:256,uncertainty:n*(3*M+12)+n,information:n*(6*M+12)+n})[p];
const moments=a=>{const mean=a.reduce((s,x)=>s+x,0)/a.length;const se=Math.sqrt(a.reduce((s,x)=>s+(x-mean)**2,0)/(a.length-1)/a.length);return {mean,se,lower:mean-3.5*se,upper:mean+3.5*se};};
function metrics(q,rates){
  const packet=q.map((_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10);
  const risk=i=>(q[i]-rates[i])**2+rates[i]*(1-rates[i]);
  return {Packet:packet,Brier:q.reduce((s,_,i)=>s+risk(i)/n,0),PriorityBrier:q.reduce((s,_,i)=>s+(i<10?3:1)*risk(i)/(n+20),0),PacketUsefulness:packet.reduce((s,i)=>s+rates[i]/10,0),PacketBrier:packet.reduce((s,i)=>s+risk(i)/10,0),PacketBias:packet.reduce((s,i)=>s+(q[i]-rates[i])/10,0)};
}
function verifyArm(a,w,t){
  const weights=prior(),seen=Array(n).fill(false),ys=Array(n).fill(0);
  const predict=i=>weights.reduce((s,v,h)=>s+v*(seen[i]?(2*t[i][h]+ys[i])/3:t[i][h]),0);
  if(a.Policy==='local'){
    assert(a.Trace.length===32&&a.Setting==='fixed32','local contract');
    for(let k=0;k<32;k++){const v=a.Trace[k];assert(v.Index===k&&v.Useful===w.Labels[k],'local tape');near(v.Q,w.Base[k],'local issued law');}
    const q=w.Base.map((b,i)=>i<32?(2*b+Number(w.Labels[i]))/3:b);
    q.forEach((p,i)=>near(a.Forecast[i],p,'local forecast'));
  }else{
    assert(['head','random','stratified','uncertainty','information'].includes(a.Policy),'policy');
    let used=20*n*M+3*n*M+n*Math.ceil(Math.log2(n));
    const increment=cost(a.Policy)+5*M+a.LabelCost;
    const budget=a.Setting==='fixed32'?0:used+32*(a.LabelCost+5*M+cost('information'));
    near(a.Budget,budget,'budget');
    for(const v of a.Trace){
      const i=v.Index;assert(Number.isInteger(i)&&i>=0&&i<n&&!seen[i]&&v.Useful===w.Labels[i],'unique arrived evidence');
      assert(!budget||used+increment<=budget,'pre-admission');near(v.Q,predict(i),'pre-label law');
      if(a.Policy==='head')assert(i===seen.indexOf(false),'head choice');
      if(a.Policy==='stratified'){
        let wanted=-1;for(let k=0;k<256;k++){let j=0,x=k;for(let z=0;z<8;z++){j=2*j+(x&1);x>>=1;}if(j<n&&!seen[j]){wanted=j;break;}}assert(i===wanted,'stratified choice');
      }
      if(a.Policy==='uncertainty'||a.Policy==='information'){
        let maximum=-Infinity,selected=NaN;
        for(let j=0;j<n;j++)if(!seen[j]){let value=H(predict(j));if(a.Policy==='information')value=Math.max(0,value-weights.reduce((s,v,h)=>s+v*H(t[j][h]),0));maximum=Math.max(maximum,value);if(j===i)selected=value;}
        near(v.Score,selected,'acquisition score');assert(selected>=maximum-1e-10,'not a maximal acquisition');
      }else near(v.Score,0,'non-information score');
      let z=0;for(let h=0;h<M;h++){weights[h]*=v.Useful?t[i][h]:1-t[i][h];z+=weights[h];}for(let h=0;h<M;h++)weights[h]/=z;
      seen[i]=true;ys[i]=Number(v.Useful);used+=increment;
    }
    assert(a.Trace.length===Math.min(n,budget?Math.floor((budget-(20*n*M+3*n*M+n*Math.ceil(Math.log2(n))))/increment):32),'stopped early or late');
    near(a.Used,used,'total modeled cost');near(a.Unused,budget?budget-used:0,'unspent budget');
    for(let h=0;h<M;h++)near(a.Weights[h],weights[h],'posterior weight');
    for(let i=0;i<n;i++)near(a.Forecast[i],predict(i),'posterior predictive');
    for(const field of ['SetupNS','SelectionNS','UpdateNS','FinalNS','TotalNS'])assert(a[field]>=0&&Number.isFinite(a[field]),'timing');
    assert(a.TotalNS>=a.SetupNS+a.SelectionNS+a.UpdateNS+a.FinalNS,'timing accounting');
  }
  const m=metrics(a.Forecast,w.Rates);assert(JSON.stringify(a.Packet)===JSON.stringify(m.Packet),'packed indices');for(const k of ['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'])near(a[k],m[k],k);
}

const directory='docs/experiments/';
const output={study:'calibration-observation-v27',wholeGoalsComplete:false,splits:{}};
for(const split of ['design','confirmation']){
  const path=`${directory}mmm-calibration-observation-v27-${split}.jsonl`,raw=fs.readFileSync(path),rows=raw.toString().trim().split('\n').map(JSON.parse),manifest=rows.shift();
  assert(manifest.Kind==='manifest'&&manifest.Split===split&&manifest.Worlds===384,'manifest');
  for(const [p,h]of Object.entries(manifest.Sources))assert(hash(fs.readFileSync(p))===h,`source changed ${p}`);
  assert(rows.length===384,'cohort count');const ids=new Set();const grouped={};let verifiedArms=0;
  for(const w of rows){
    const g=['tight','wide'].indexOf(w.Geometry),r=['independent','aligned','reversed','calibrated','smooth','matched'].indexOf(w.Regime);
    assert(g>=0&&r>=0&&w.World>=0&&w.World<32&&w.Kind==='world','world');near(w.Seed,manifest.SeedBase+g*10000000+r*1000000+w.World*1000,'seed');
    const id=`${w.Geometry}/${w.Regime}/${w.World}`;assert(!ids.has(id),'duplicate world');ids.add(id);
    const step=g===0?.005:.02;const oracle=[.925,.925,...Array.from({length:198},(_,i)=>.65*(Math.cos(step*(i+1))+1)/2+.275)].sort((a,b)=>b-a).slice(0,n);
    assert(w.Base.length===n&&w.Rates.length===n&&w.Labels.length===n,'frontier');const table=tables(w.Base);
    w.Base.forEach((b,i)=>near(b,oracle[i],'cosine baseline'));assert(w.Labels.every(v=>typeof v==='boolean'),'label domain');
    let high=0;
    w.Rates.forEach((p,i)=>{assert(Number.isFinite(p)&&p>=0&&p<=1,'truth domain');let want=p;switch(w.Regime){case'independent':assert(p===.2||p===.8,'independent rates');high+=p===.8?1:0;break;case'aligned':want=.9-.8*i/149;break;case'reversed':want=.1+.8*i/149;break;case'calibrated':want=w.Base[i];break;case'smooth':want=.1+.8*Math.sin(Math.PI*i/149)**2;break;case'matched':assert(Number.isInteger(w.Theta)&&w.Theta>=0&&w.Theta<M,'theta draw');break;}near(p,want,'truth specification');});if(w.Regime==='independent')assert(high===75,'balanced truth');
    assert(w.Arms.length===21,'arm count');const unique=new Set();for(const a of w.Arms){const k=`${a.Setting}/${a.Policy}`;assert(!unique.has(k),'duplicate arm');unique.add(k);verifyArm(a,w,table);verifiedArms++;}
    (grouped[`${w.Geometry}/${w.Regime}`]??=[]).push(w);
  }
  const result={sha256:hash(raw),worlds:rows.length,verifiedArms,sourceHashes:manifest.Sources,groups:{},rescue:{},observation:{}};
  for(const [name,worlds]of Object.entries(grouped)){
    assert(worlds.length===32,'paired cohort');const get=(w,policy,setting='fixed32')=>w.Arms.find(a=>a.Policy===policy&&a.Setting===setting);
    const summaries={};for(const a of worlds[0].Arms){const k=`${a.Setting}/${a.Policy}`;summaries[k]={};for(const field of ['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'])summaries[k][field]=moments(worlds.map(w=>get(w,a.Policy,a.Setting)[field]));summaries[k].meanLabels=worlds.reduce((s,w)=>s+get(w,a.Policy,a.Setting).Trace.length,0)/32;summaries[k].maximumTotalNS=Math.max(...worlds.map(w=>get(w,a.Policy,a.Setting).TotalNS));}
    result.groups[name]=summaries;
    const regime=name.split('/')[1];
    for(const policy of ['stratified','information']){
      const checks={};
      for(const field of ['Brier','PriorityBrier']){
        const gain=moments(worlds.map(w=>get(w,'local')[field]-get(w,policy)[field]));const recovery=['independent','reversed'].includes(regime);if(recovery)checks[field]={...gain,pass:gain.mean>=.02&&gain.lower>0};else if(['aligned','calibrated'].includes(regime))checks[field]={...gain,pass:-gain.lower<=.01};
      }
      if(['aligned','calibrated'].includes(regime)){const gain=moments(worlds.map(w=>get(w,policy).PacketUsefulness-get(w,'local').PacketUsefulness));checks.PacketUsefulness={...gain,pass:-gain.lower<=.01};}
      const bias=moments(worlds.map(w=>get(w,policy).PacketBias));checks.bias={...bias,magnitudeUpper:Math.abs(bias.mean)+3.5*bias.se,pass:Math.abs(bias.mean)+3.5*bias.se<=.10};
      result.rescue[`${name}/${policy}`]={checks,pass:Object.values(checks).every(c=>c.pass)};
    }
    for(const control of ['random','uncertainty']){
      const checks={};for(const field of ['Brier','PacketUsefulness']){
        const gain=moments(worlds.map(w=>{const a=get(w,'information','cost100000'),b=get(w,control,'cost100000');return field==='Brier'?b[field]-a[field]:a[field]-b[field];}));const recovery=['reversed','smooth','matched'].includes(regime),minimum=field==='Brier'?.005:.02;checks[field]={...gain,pass:recovery?gain.mean>=minimum&&gain.lower>0:-gain.lower<=.01};
      }
      result.observation[`${name}/${control}`]={checks,pass:Object.values(checks).every(c=>c.pass)};
    }
  }
  result.rescuePass=Object.values(result.rescue).every(v=>v.pass);result.observationPass=Object.values(result.observation).every(v=>v.pass);output.splits[split]=result;
}
const path=`${directory}mmm-calibration-observation-v27-summary.json`;
fs.writeFileSync(path,JSON.stringify(output,null,2)+'\n');
for(const [split,r]of Object.entries(output.splits))console.log(JSON.stringify({split,sha256:r.sha256,worlds:r.worlds,verifiedArms:r.verifiedArms,rescuePass:r.rescuePass,observationPass:r.observationPass,failedRescue:Object.entries(r.rescue).filter(([,v])=>!v.pass).map(([k])=>k),failedObservation:Object.entries(r.observation).filter(([,v])=>!v.pass).map(([k])=>k)}));
