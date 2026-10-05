import fs from 'node:fs';
import crypto from 'node:crypto';
const n=150,M=27,fields=['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const check=(v,s)=>{if(!v)throw Error(s);};
const near=(a,b,s,tol=1e-10)=>check(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=tol,`${s}: ${a} != ${b}`);
const entropy=p=>p<=0||p>=1?0:-p*Math.log(p)-(1-p)*Math.log1p(-p);
const prior=()=>[.1,.8,...Array(25).fill(.004)];
const table=base=>base.map((b,i)=>[b,(.9*b-.05)/.8,...[.1,.3,.5,.7,.9].flatMap(a=>[-.8,-.4,0,.4,.8].map(c=>Math.max(.02,Math.min(.98,a+c*i/149))))]);
const cost=p=>({random:151,uncertainty:n*(3*M+12)+n,information:n*(6*M+12)+n,risk:6*n*M*M+12*n*M+40*n})[p];
function interval(a){const mean=a.reduce((s,v)=>s+v,0)/a.length,se=Math.sqrt(a.reduce((s,v)=>s+(v-mean)**2,0)/(a.length-1)/a.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};}
function riskScores(t,weights,seen,ys){
  const means=t.map((row,j)=>row.map(p=>seen[j]?(2*p+ys[j])/3:p));
  const q=means.map(row=>weights.reduce((s,w,h)=>s+w*row[h],0));
  const centered=means.map((row,j)=>row.map(p=>p-q[j]));
  const gram=new Float64Array(M*M);
  for(let j=0;j<n;j++){const v=(j<10?3:1)/170,d=centered[j];for(let h=0;h<M;h++)for(let k=0;k<M;k++)gram[h*M+k]+=v*d[h]*d[k];}
  const score=Array(n).fill(-1);
  for(let e=0;e<n;e++)if(!seen[e]){
    const x=weights.map((w,h)=>w*centered[e][h]);let between=0,within=0,gain=0;
    for(let h=0;h<M;h++){between+=weights[h]*centered[e][h]**2;within+=weights[h]*t[e][h]*(1-t[e][h])/3;for(let k=0;k<M;k++)gain+=x[h]*gram[h*M+k]*x[k];}
    gain+=(e<10?3:1)/170*(2*between*within+within*within);score[e]=Math.max(0,gain/(q[e]*(1-q[e])));
  }
  return{score,q,means};
}
function directRisk(e,t,weights,q,means){
  let gain=0;
  for(let j=0;j<n;j++){
    let covariance=0;
    if(j===e){for(let h=0;h<M;h++)covariance+=weights[h]*(2*t[e][h]**2+t[e][h])/3;covariance-=q[e]**2;}
    else{for(let h=0;h<M;h++)covariance+=weights[h]*t[e][h]*means[j][h];covariance-=q[e]*q[j];}
    gain+=(j<10?3:1)/170*covariance**2;
  }
  return gain/(q[e]*(1-q[e]));
}
function metrics(q,rates){
  const packet=Array.from({length:n},(_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10);
  const risk=i=>(q[i]-rates[i])**2+rates[i]*(1-rates[i]);
  return{Packet:packet,Brier:q.reduce((s,_,i)=>s+risk(i)/n,0),PriorityBrier:q.reduce((s,_,i)=>s+(i<10?3:1)*risk(i)/170,0),PacketUsefulness:packet.reduce((s,i)=>s+rates[i]/10,0),PacketBrier:packet.reduce((s,i)=>s+risk(i)/10,0),PacketBias:packet.reduce((s,i)=>s+(q[i]-rates[i])/10,0)};
}
function verifyArm(a,w,t,ent){
  check(a.Forecast.length===n&&a.Weights.length===M,'law dimensions');
  const weights=prior(),seen=Array(n).fill(false),ys=Array(n).fill(0);
  const predict=i=>weights.reduce((s,w,h)=>s+w*(seen[i]?(2*t[i][h]+ys[i])/3:t[i][h]),0);
  if(a.Policy==='local'){
    check(a.Setting==='fixed32'&&a.Trace.length===32,'local control');for(let i=0;i<32;i++){const v=a.Trace[i];check(v.Index===i&&v.Useful===w.Labels[i],'local tape');near(v.Q,w.Base[i],'local forecast before evidence');}
    for(let i=0;i<n;i++)near(a.Forecast[i],i<32?(2*w.Base[i]+Number(w.Labels[i]))/3:w.Base[i],'local forecast');
  }else{
    check(['risk','random','uncertainty','information'].includes(a.Policy),'policy');
    const setup=20*n*M,final=3*n*M+n*Math.ceil(Math.log2(n)),step=cost(a.Policy)+5*M+a.LabelCost;
    check(['fixed32','cost10000','cost100000','cost1000000'].includes(a.Setting),'setting');near(a.LabelCost,a.Setting==='fixed32'?0:Number(a.Setting.slice(4)),'label cost');
    const budget=a.Setting==='fixed32'?0:setup+final+32*(a.LabelCost+5*M+cost('risk'));near(a.Budget,budget,'cap');let used=setup+final;
    for(const v of a.Trace){
      const e=v.Index;check(Number.isInteger(e)&&e>=0&&e<n&&!seen[e]&&v.Useful===w.Labels[e],'distinct selected tape');check(!budget||used+step<=budget,'pre-admission');near(v.Q,predict(e),'forecast before label');
      if(a.Policy==='risk'){
        const r=riskScores(t,weights,seen,ys);near(v.Score,r.score[e],'Gram score');near(v.Score,directRisk(e,t,weights,r.q,r.means),'direct covariance sum');check(r.score[e]>=Math.max(...r.score)-1e-10,'not maximal risk reduction');
      }else if(a.Policy==='uncertainty'||a.Policy==='information'){
        let maximum=-Infinity,selected=0;for(let j=0;j<n;j++)if(!seen[j]){let s=entropy(predict(j));if(a.Policy==='information')s=Math.max(0,s-weights.reduce((z,w,h)=>z+w*ent[j][h],0));maximum=Math.max(maximum,s);if(j===e)selected=s;}near(v.Score,selected,'entropy score');check(selected>=maximum-1e-10,'not maximal acquisition');
      }else near(v.Score,0,'random score');
      let mass=0;for(let h=0;h<M;h++){weights[h]*=v.Useful?t[e][h]:1-t[e][h];mass+=weights[h];}for(let h=0;h<M;h++)weights[h]/=mass;
      seen[e]=true;ys[e]=Number(v.Useful);used+=step;
    }
    check(a.Trace.length===Math.min(n,budget?Math.floor((budget-setup-final)/step):32),'wrong stop');near(a.Used,used,'used cost');near(a.Unused,budget?budget-used:0,'unused cost');
    for(let h=0;h<M;h++)near(a.Weights[h],weights[h],'posterior weights');for(let i=0;i<n;i++)near(a.Forecast[i],predict(i),'posterior predictive');
    for(const k of ['SetupNS','SelectionNS','UpdateNS','FinalNS','TotalNS'])check(Number.isFinite(a[k])&&a[k]>=0,'timing');check(a.TotalNS>=a.SetupNS+a.SelectionNS+a.UpdateNS+a.FinalNS,'timing accounting');
  }
  const m=metrics(a.Forecast,w.Rates);check(JSON.stringify(m.Packet)===JSON.stringify(a.Packet),'packet');for(const k of fields)near(a[k],m[k],k);
}
const output={study:'risk-acquisition-v28',wholeGoalsComplete:false,splits:{}};
for(const split of ['design','confirmation']){
  const path=`docs/experiments/mmm-risk-acquisition-v28-${split}.jsonl`,raw=fs.readFileSync(path),rows=raw.toString().trim().split('\n').map(JSON.parse),manifest=rows.shift();
  check(manifest.Kind==='manifest'&&manifest.Split===split&&manifest.Worlds===384,'manifest');near(manifest.SeedBase,split==='design'?2026102803:2026102804,'base seed');for(const[p,h]of Object.entries(manifest.Sources))check(hash(fs.readFileSync(p))===h,`source changed ${p}`);check(rows.length===384,'world count');
  const grouped={},ids=new Set();let verifiedArms=0;
  for(const w of rows){
    const g=['tight','wide'].indexOf(w.Geometry),r=['independent','aligned','reversed','calibrated','smooth','matched'].indexOf(w.Regime);check(g>=0&&r>=0&&Number.isInteger(w.World)&&w.World>=0&&w.World<32&&w.Kind==='world','world descriptor');near(w.Seed,manifest.SeedBase+g*10000000+r*1000000+w.World*1000,'world seed');const id=`${w.Geometry}/${w.Regime}/${w.World}`;check(!ids.has(id),'duplicate world');ids.add(id);
    const step=g===0?.005:.02,base=[.925,.925,...Array.from({length:198},(_,i)=>.65*(Math.cos(step*(i+1))+1)/2+.275)].sort((a,b)=>b-a).slice(0,n);check(w.Base.length===n&&w.Rates.length===n&&w.Labels.length===n&&w.Labels.every(v=>typeof v==='boolean'),'world dimensions');w.Base.forEach((b,i)=>near(b,base[i],'cosine oracle'));
    const t=table(base),ent=t.map(row=>row.map(entropy));let high=0;
    w.Rates.forEach((p,i)=>{check(Number.isFinite(p)&&p>=0&&p<=1,'rate');let want=p;switch(w.Regime){case'independent':check(p===.2||p===.8,'independent rate');high+=p===.8?1:0;break;case'aligned':want=.9-.8*i/149;break;case'reversed':want=.1+.8*i/149;break;case'calibrated':want=base[i];break;case'smooth':want=.1+.8*Math.sin(Math.PI*i/149)**2;break;case'matched':check(Number.isInteger(w.Theta)&&w.Theta>=0&&w.Theta<M,'matched theta');break;}near(p,want,'truth model');});if(w.Regime==='independent')check(high===75,'balanced rates');
    check(w.Arms.length===17,'arm count');const unique=new Set();for(const a of w.Arms){const key=`${a.Setting}/${a.Policy}`;check(!unique.has(key),'duplicate arm');unique.add(key);verifyArm(a,w,t,ent);verifiedArms++;}(grouped[`${w.Geometry}/${w.Regime}`]??=[]).push(w);
  }
  const result={sha256:hash(raw),worlds:rows.length,verifiedArms,sourceHashes:manifest.Sources,groups:{},rescue:{},observation:{}};
  for(const [key,worlds]of Object.entries(grouped)){
    check(worlds.length===32,'paired cohort');const get=(w,p,s='fixed32')=>w.Arms.find(a=>a.Policy===p&&a.Setting===s);const summaries={};
    for(const a of worlds[0].Arms){const k=`${a.Setting}/${a.Policy}`;summaries[k]=Object.fromEntries(fields.map(f=>[f,interval(worlds.map(w=>get(w,a.Policy,a.Setting)[f]))]));summaries[k].meanLabels=worlds.reduce((s,w)=>s+get(w,a.Policy,a.Setting).Trace.length,0)/32;summaries[k].maximumTotalNS=Math.max(...worlds.map(w=>get(w,a.Policy,a.Setting).TotalNS));}
    result.groups[key]=summaries;const regime=key.split('/')[1],checks={};
    for(const f of ['Brier','PriorityBrier']){const gain=interval(worlds.map(w=>get(w,'local')[f]-get(w,'risk')[f]));if(['independent','reversed'].includes(regime))checks[f]={...gain,pass:gain.mean>=.02&&gain.lower>0};else if(['aligned','calibrated'].includes(regime))checks[f]={...gain,pass:-gain.lower<=.01};}
    if(['aligned','calibrated'].includes(regime)){const gain=interval(worlds.map(w=>get(w,'risk').PacketUsefulness-get(w,'local').PacketUsefulness));checks.PacketUsefulness={...gain,pass:-gain.lower<=.01};}
    const bias=interval(worlds.map(w=>get(w,'risk').PacketBias));checks.bias={...bias,magnitudeUpper:Math.abs(bias.mean)+3.5*bias.se,pass:Math.abs(bias.mean)+3.5*bias.se<=.10};result.rescue[key]={checks,pass:Object.values(checks).every(v=>v.pass)};
    for(const control of ['random','uncertainty']){const checks={};for(const f of ['Brier','PacketUsefulness']){const gain=interval(worlds.map(w=>{const a=get(w,'risk','cost100000'),b=get(w,control,'cost100000');return f==='Brier'?b[f]-a[f]:a[f]-b[f];}));const recovery=['reversed','smooth','matched'].includes(regime);checks[f]={...gain,pass:recovery?gain.mean>=(f==='Brier'?.005:.02)&&gain.lower>0:-gain.lower<=.01};}result.observation[`${key}/${control}`]={checks,pass:Object.values(checks).every(v=>v.pass)};}
  }
  result.rescuePass=Object.values(result.rescue).every(v=>v.pass);result.observationPass=Object.values(result.observation).every(v=>v.pass);output.splits[split]=result;
  console.log(JSON.stringify({split,sha256:result.sha256,worlds:rows.length,verifiedArms,rescuePass:result.rescuePass,observationPass:result.observationPass,failedRescue:Object.entries(result.rescue).filter(([,v])=>!v.pass).map(([k])=>k),failedObservation:Object.entries(result.observation).filter(([,v])=>!v.pass).map(([k])=>k)}));
}
fs.writeFileSync('docs/experiments/mmm-risk-acquisition-v28-summary.json',JSON.stringify(output,null,2)+'\n');
