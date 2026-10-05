// Independent flat integrated-model reconstruction for the frozen V30 study.
import fs from 'node:fs';
import crypto from 'node:crypto';
const n=150,fields=['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'];
const regimes=['independent','aligned','reversed','calibrated','curved','shifted_peak','alternating','phase_alternating','permuted_curved','baseline_matched','tree_matched','narrow_peak'];
const expectedArms=['baseline/head','local/stratified_random','old/stratified_random','old/random','old/uncertainty','old/information','partition/stratified_random','blend/stratified_random','blend/random','blend/uncertainty','blend/family_information'];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const check=(p,s)=>{if(!p)throw Error(s);};
const near=(a,b,s,t=1e-10)=>check(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=t,`${s}: ${a} != ${b}`);
const H=p=>p<=0||p>=1?0:-p*Math.log(p)-(1-p)*Math.log1p(-p);
function trees(lo=0,hi=8,depth=0,node=0){
  const p={nodes:Array(8).fill(0),prior:depth===3?1:.5};for(let i=lo;i<hi;i++)p.nodes[i]=node;
  if(depth===3)return[p];const out=[p],mid=(lo+hi)/2;
  for(const l of trees(lo,mid,depth+1,2*node+1))for(const r of trees(mid,hi,depth+1,2*node+2)){
    const q={nodes:Array(8).fill(0),prior:.5*l.prior*r.prior};for(let i=lo;i<mid;i++)q.nodes[i]=l.nodes[i];for(let i=mid;i<hi;i++)q.nodes[i]=r.nodes[i];out.push(q);
  }return out;
}
const partitions=trees();check(partitions.length===26,'tree count');near(partitions.reduce((z,p)=>z+p.prior,0),1,'tree prior');
function hypotheses(base,coordinate,model){
  const out=[],b=base.slice(),push=(family,prior,p,nodes=null)=>out.push({family,prior,p,nodes});
  const affine=scale=>{push(1,scale*.1,b);push(1,scale*.8,b.map(v=>(.9*v-.05)/.8));for(const a of [.1,.3,.5,.7,.9])for(const c of [-.8,-.4,0,.4,.8])push(1,scale*.004,b.map((_,i)=>Math.max(.02,Math.min(.98,a+c*i/149))));};
  const partition=scale=>{push(2,scale*.99,b);for(const tree of partitions)push(2,scale*.01*tree.prior,null,coordinate.map(r=>tree.nodes[Math.min(7,Math.floor(8*r))]));};
  if(model==='blend'){push(0,.98,b);affine(.01);partition(.01);}
  else if(model==='old')affine(1);
  else if(model==='partition')partition(1);
  else if(model==='local'||model==='baseline')push(0,1,b);
  else throw Error('unknown model');
  near(out.reduce((z,h)=>z+h.prior,0),1,'model prior');return out;
}
function predictions(hyp,weights,evidence){
  const conditional=hyp.map(h=>{
    const q=Array(n),s=Array(15).fill(0),f=Array(15).fill(0);
    if(h.nodes)for(const[i,y]of evidence)(y?s:f)[h.nodes[i]]++;
    for(let i=0;i<n;i++){let p=h.nodes?(1+s[h.nodes[i]])/(2+s[h.nodes[i]]+f[h.nodes[i]]):h.p[i];if(evidence.has(i))p=(2*p+Number(evidence.get(i)))/3;q[i]=p;}
    return q;
  });
  const q=Array(n).fill(0),family=Array(3).fill(0),familyMeans=Array.from({length:3},()=>Array(n).fill(0));
  weights.forEach((w,k)=>{family[hyp[k].family]+=w;for(let i=0;i<n;i++){q[i]+=w*conditional[k][i];familyMeans[hyp[k].family][i]+=w*conditional[k][i];}});
  for(let m=0;m<3;m++)if(family[m]>0)for(let i=0;i<n;i++)familyMeans[m][i]/=family[m];
  return{q,conditional,family,familyMeans};
}
const logFactorial=k=>{let v=0;for(let j=2;j<=k;j++)v+=Math.log(j);return v;};
function verifyIntegrated(hyp,weights,evidence){
  const logs=hyp.map(h=>{
    let l=Math.log(h.prior);if(!h.nodes){for(const[i,y]of evidence)l+=Math.log(y?h.p[i]:1-h.p[i]);return l;}
    const s=Array(15).fill(0),f=Array(15).fill(0);for(const[i,y]of evidence)(y?s:f)[h.nodes[i]]++;
    for(let node=0;node<15;node++)l+=logFactorial(s[node])+logFactorial(f[node])-logFactorial(s[node]+f[node]+1);return l;
  });
  const maximum=Math.max(...logs),relative=logs.map(v=>Math.exp(v-maximum)),sum=relative.reduce((z,v)=>z+v,0);
  weights.forEach((w,k)=>near(w,relative[k]/sum,'flat integrated likelihood'));
}
function metrics(q,p){
  const packet=Array.from({length:n},(_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10),risk=i=>(q[i]-p[i])**2+p[i]*(1-p[i]);
  return{Packet:packet,Brier:q.reduce((z,_,i)=>z+risk(i)/n,0),PriorityBrier:q.reduce((z,_,i)=>z+(i<10?3:1)*risk(i)/170,0),PacketUsefulness:packet.reduce((z,i)=>z+p[i]/10,0),PacketBrier:packet.reduce((z,i)=>z+risk(i)/10,0),PacketBias:packet.reduce((z,i)=>z+(q[i]-p[i])/10,0)};
}
function verifyArm(a,w){
  check(expectedArms.includes(`${a.Model}/${a.Policy}`),'model/policy');check(Array.isArray(a.Trace)&&a.Trace.length===(a.Model==='baseline'?0:32),'trace length');check(a.Forecast.length===n&&a.Weights.length===3,'law dimensions');
  const hyp=hypotheses(w.Base,w.Coordinates,a.Model),weights=hyp.map(h=>h.prior),evidence=new Map();
  for(let step=0;step<a.Trace.length;step++){
    const t=a.Trace[step],i=t.Index;check(Number.isInteger(i)&&i>=0&&i<n&&!evidence.has(i)&&t.Useful===w.Labels[i],'distinct available evidence');
    const p=predictions(hyp,weights,evidence);near(t.Q,p.q[i],'issued pre-label law');check(t.Probability>0&&t.Probability<=1,'nomination probability');
    if(a.Policy==='stratified_random'){
      let bucket=0,x=step;for(let bit=0;bit<5;bit++){bucket=2*bucket+(x&1);x>>=1;}
      const lo=Math.floor(bucket*n/32),hi=Math.floor((bucket+1)*n/32);check(i>=lo&&i<hi,'scheduled stratum');near(t.Probability,1/(hi-lo),'stratum probability');near(t.Score,0,'stratum score');
    }else if(a.Policy==='random'){near(t.Probability,1/(n-step),'uniform probability');near(t.Score,0,'random score');}
    else if(['uncertainty','information','family_information'].includes(a.Policy)){
      let maximum=-Infinity,best=-1,selected=NaN;
      for(let j=0;j<n;j++)if(!evidence.has(j)){
        let value=H(p.q[j]);if(a.Policy==='information')value-=weights.reduce((z,v,k)=>z+v*H(p.conditional[k][j]),0);
        if(a.Policy==='family_information')value-=p.family.reduce((z,v,k)=>z+v*H(p.familyMeans[k][j]),0);
        value=Math.max(0,value);if(value>maximum){maximum=value;best=j;}if(j===i)selected=value;
      }
      near(t.Score,maximum,'selection maximum');
      if(a.Policy==='family_information'){
        const floor=.2/(n-step);check(Math.abs(t.Probability-floor)<=1e-12||Math.abs(t.Probability-floor-.8)<=1e-12,'exploration mixture probability');
        if(t.Probability>floor+.4)near(selected,maximum,'family exploit maximum');
        if(i===best&&Math.abs(selected-maximum)>1e-12)throw Error('bad maximizing index');
      }else{near(selected,maximum,'selected maximum');near(t.Probability,1,'deterministic probability');}
    }else throw Error('unsupported nomination');
    let sum=0;for(let k=0;k<hyp.length;k++){weights[k]*=t.Useful?p.conditional[k][i]:1-p.conditional[k][i];sum+=weights[k];}weights.forEach((v,k)=>weights[k]=v/sum);evidence.set(i,t.Useful);
  }
  verifyIntegrated(hyp,weights,evidence);const p=predictions(hyp,weights,evidence);for(let i=0;i<n;i++)near(a.Forecast[i],p.q[i],'final predictive');
  if(a.Model==='blend')a.Weights.forEach((v,k)=>near(v,p.family[k],'family posterior'));else check(a.Weights.every(v=>v===0),'opaque child weight marker');
  const m=metrics(a.Forecast,w.Rates);check(JSON.stringify(m.Packet)===JSON.stringify(a.Packet),'packet');for(const f of fields)near(a[f],m[f],f);
  for(const f of ['SetupNS','SelectionNS','UpdateNS','FinalNS','TotalNS'])check(Number.isFinite(a[f])&&a[f]>=0,'timing');check(a.TotalNS>=a.SetupNS+a.SelectionNS+a.UpdateNS+a.FinalNS,'timing accounting');
}
function interval(values){const mean=values.reduce((z,v)=>z+v,0)/values.length,se=Math.sqrt(values.reduce((z,v)=>z+(v-mean)**2,0)/(values.length-1)/values.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};}
const result={study:'blend-v30',wholeGoalsComplete:false,goal7EqualTotalCost:false,splits:{}};
for(const split of ['design','confirmation']){
  const raw=fs.readFileSync(`docs/experiments/mmm-blend-v30-${split}.jsonl`),rows=raw.toString().trim().split('\n').map(JSON.parse),manifest=rows.shift();check(manifest.Kind==='manifest'&&manifest.Split===split&&manifest.Worlds===768,'manifest');near(manifest.SeedBase,split==='design'?2026103003:2026103004,'seed base');check(rows.length===768,'world count');
  for(const[p,h]of Object.entries(manifest.Sources))check(hash(fs.readFileSync(p))===h,`source changed ${p}`);
  const groups={},ids=new Set();let verifiedArms=0;
  for(const w of rows){
    const g=['tight','wide'].indexOf(w.Geometry),r=regimes.indexOf(w.Regime);check(g>=0&&r>=0&&Number.isInteger(w.World)&&w.World>=0&&w.World<32&&w.Kind==='world','world domain');near(w.Seed,manifest.SeedBase+g*10000000+r*1000000+w.World*1000,'world seed');const id=`${w.Geometry}/${w.Regime}/${w.World}`;check(!ids.has(id),'duplicate world');ids.add(id);
    const step=g===0?.005:.02,base=[.925,.925,...Array.from({length:198},(_,i)=>.65*(Math.cos(step*(i+1))+1)/2+.275)].sort((a,b)=>b-a).slice(0,n);
    check(w.Base.length===n&&w.Coordinates.length===n&&w.Rates.length===n&&w.Labels.length===n&&w.Labels.every(v=>typeof v==='boolean'),'world dimensions');let high=0;
    w.Rates.forEach((p,i)=>{near(w.Base[i],base[i],'cosine baseline');near(w.Coordinates[i],i/149,'declared coordinate');check(Number.isFinite(p)&&p>=0&&p<=1,'true probability');let want=p,x=i/149;
      switch(w.Regime){case'independent':check(p===.2||p===.8,'independent rates');high+=p===.8?1:0;break;case'aligned':want=.9-.8*x;break;case'reversed':want=.1+.8*x;break;case'calibrated':want=base[i];break;case'curved':want=.1+.8*Math.sin(Math.PI*x)**2;break;case'shifted_peak':case'narrow_peak':{const width=w.Regime==='narrow_peak'?.04:.14;want=.1+.8*Math.exp(-(((x-(.25+.5*(w.World%8)/7))/width)**2));break;}case'alternating':want=i%2===0?.8:.2;break;case'phase_alternating':want=i%2===0?.2:.8;break;case'tree_matched':check(Number.isInteger(w.Partition)&&w.Partition>=0&&w.Partition<26&&w.LeafMeans.length===15,'tree draw');{const v=w.LeafMeans[partitions[w.Partition].nodes[Math.min(7,Math.floor(8*x))]];check(v>0&&v<1,'leaf mean');}break;}near(p,want,'true rate formula');
    });if(w.Regime==='independent')check(high===75,'balanced population');if(w.Regime==='permuted_curved'){const sorted=w.Rates.slice().sort((a,b)=>a-b),want=Array.from({length:n},(_,i)=>.1+.8*Math.sin(Math.PI*i/149)**2).sort((a,b)=>a-b);sorted.forEach((v,i)=>near(v,want[i],'permuted curve multiset'));}
    check(w.Arms.length===expectedArms.length,'arm count');const unique=new Set();for(const a of w.Arms){const key=`${a.Model}/${a.Policy}`;check(!unique.has(key),'duplicate arm');unique.add(key);verifyArm(a,w);verifiedArms++;}
    const primary=w.Arms.find(a=>a.Model==='blend'&&a.Policy==='stratified_random');for(const model of ['local','old','partition']){const control=w.Arms.find(a=>a.Model===model&&a.Policy==='stratified_random');check(JSON.stringify(primary.Trace.map(t=>[t.Index,t.Useful,t.Probability]))===JSON.stringify(control.Trace.map(t=>[t.Index,t.Useful,t.Probability])),'matched evidence differs');}
    (groups[`${w.Geometry}/${w.Regime}`]??=[]).push(w);
  }
  const summary={sha256:hash(raw),worlds:rows.length,verifiedArms,sourceHashes:manifest.Sources,groups:{},gates:{}};
  for(const[name,worlds]of Object.entries(groups)){
    check(worlds.length===32,'paired sample size');const get=(w,m,p)=>w.Arms.find(a=>a.Model===m&&a.Policy===p),stats={};
    for(const a of worlds[0].Arms){const key=`${a.Model}/${a.Policy}`;stats[key]=Object.fromEntries(fields.map(f=>[f,interval(worlds.map(w=>get(w,a.Model,a.Policy)[f]))]));stats[key].maximumTotalNS=Math.max(...worlds.map(w=>get(w,a.Model,a.Policy).TotalNS));stats[key].meanOddSelections=worlds.reduce((z,w)=>z+get(w,a.Model,a.Policy).Trace.filter(t=>t.Index%2===1).length,0)/32;}
    summary.groups[name]=stats;const regime=name.split('/')[1],checks={};
    for(const control of ['local','old'])for(const f of ['Brier','PriorityBrier','PacketUsefulness']){
      const values=worlds.map(w=>{const a=get(w,'blend','stratified_random'),b=get(w,control,'stratified_random');return f==='PacketUsefulness'?a[f]-b[f]:b[f]-a[f];}),v=interval(values),improve=control==='old'&&['curved','shifted_peak'].includes(regime),threshold=improve?(f==='PacketUsefulness'?.02:.01):-.01;
      checks[`${control}/${f}`]={...v,threshold,pass:improve?v.mean>=threshold&&v.lower>0:v.lower>=threshold};
    }
    const bias=interval(worlds.map(w=>get(w,'blend','stratified_random').PacketBias));checks.bias={...bias,magnitudeUpper:Math.abs(bias.mean)+3.5*bias.se,pass:Math.abs(bias.mean)+3.5*bias.se<=.10};
    const maximum=Math.max(...worlds.flatMap(w=>w.Arms.map(a=>a.TotalNS)));checks.modelTime={maximumNS:maximum,pass:maximum<=10000000};summary.gates[name]={checks,pass:Object.values(checks).every(c=>c.pass)};
  }
  summary.pass=Object.values(summary.gates).every(g=>g.pass);result.splits[split]=summary;console.log(JSON.stringify({split,worlds:rows.length,verifiedArms,pass:summary.pass,failed:Object.entries(summary.gates).filter(([,v])=>!v.pass).map(([name])=>name)}));
}
fs.writeFileSync('docs/experiments/mmm-blend-v30-summary.json',JSON.stringify(result,null,2)+'\n');
