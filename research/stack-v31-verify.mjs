// Independent flat child models + analytic edge/interior ridge optimum.
import fs from 'node:fs';
import crypto from 'node:crypto';
const n=150,fields=['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'];
const regimes=['independent','aligned','reversed','calibrated','curved','shifted_peak','alternating','phase_alternating','permuted_curved','baseline_matched','tree_matched','narrow_peak'];
const arms=['baseline/head','local/stratified_random','old/stratified_random','old/random','old/uncertainty','old/information','partition/stratified_random','blend/stratified_random','blend/random','blend/uncertainty','blend/family_information','stack/stratified_random','stack/random','stack/uncertainty','stack/disagreement'];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),check=(p,s)=>{if(!p)throw Error(s);};
const near=(a,b,s,t=1e-10)=>check(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=t,`${s}: ${a} != ${b}`);
const H=p=>p<=0||p>=1?0:-p*Math.log(p)-(1-p)*Math.log1p(-p),clamp=v=>Math.max(0,Math.min(1,v));
function trees(lo=0,hi=8,depth=0,node=0){
  const p={nodes:Array(8).fill(0),prior:depth===3?1:.5};for(let i=lo;i<hi;i++)p.nodes[i]=node;
  if(depth===3)return[p];const out=[p],mid=(lo+hi)/2;
  for(const l of trees(lo,mid,depth+1,2*node+1))for(const r of trees(mid,hi,depth+1,2*node+2)){const q={nodes:Array(8).fill(0),prior:.5*l.prior*r.prior};for(let i=lo;i<mid;i++)q.nodes[i]=l.nodes[i];for(let i=mid;i<hi;i++)q.nodes[i]=r.nodes[i];out.push(q);}return out;
}
const partitions=trees();check(partitions.length===26,'tree count');near(partitions.reduce((z,t)=>z+t.prior,0),1,'tree prior');
function hypotheses(base,coordinate,model){
  const out=[],push=(family,prior,p,nodes=null)=>out.push({family,prior,p,nodes});
  const affine=scale=>{push(1,scale*.1,base);push(1,scale*.8,base.map(b=>(.9*b-.05)/.8));for(const a of [.1,.3,.5,.7,.9])for(const c of [-.8,-.4,0,.4,.8])push(1,scale*.004,base.map((_,i)=>Math.max(.02,Math.min(.98,a+c*i/149))));};
  const partition=scale=>{push(2,scale*.99,base);for(const t of partitions)push(2,scale*.01*t.prior,null,coordinate.map(r=>t.nodes[Math.min(7,Math.floor(8*r))]));};
  if(model==='blend'||model==='stack'){push(0,.98,base);affine(.01);partition(.01);}else if(model==='old')affine(1);else if(model==='partition')partition(1);else if(model==='local'||model==='baseline')push(0,1,base);else throw Error('model');
  near(out.reduce((z,h)=>z+h.prior,0),1,'hypothesis prior');return out;
}
function predictions(hyp,weights,evidence){
  const conditional=hyp.map(h=>{const s=Array(15).fill(0),f=Array(15).fill(0);if(h.nodes)for(const[i,y]of evidence)(y?s:f)[h.nodes[i]]++;return Array.from({length:n},(_,i)=>{let q=h.nodes?(1+s[h.nodes[i]])/(2+s[h.nodes[i]]+f[h.nodes[i]]):h.p[i];if(evidence.has(i))q=(2*q+Number(evidence.get(i)))/3;return q;});});
  const q=Array(n).fill(0),family=Array(3).fill(0),means=Array.from({length:3},()=>Array(n).fill(0));
  weights.forEach((v,k)=>{family[hyp[k].family]+=v;for(let i=0;i<n;i++){q[i]+=v*conditional[k][i];means[hyp[k].family][i]+=v*conditional[k][i];}});for(let k=0;k<3;k++)if(family[k]>0)for(let i=0;i<n;i++)means[k][i]/=family[k];
  return{q,family,means,conditional};
}
const logFactorial=k=>{let s=0;for(let j=2;j<=k;j++)s+=Math.log(j);return s;};
function integrated(hyp,weights,evidence){
  const logs=hyp.map(h=>{let v=Math.log(h.prior);if(!h.nodes){for(const[i,y]of evidence)v+=Math.log(y?h.p[i]:1-h.p[i]);return v;}const s=Array(15).fill(0),f=Array(15).fill(0);for(const[i,y]of evidence)(y?s:f)[h.nodes[i]]++;for(let j=0;j<15;j++)v+=logFactorial(s[j])+logFactorial(f[j])-logFactorial(s[j]+f[j]+1);return v;});const max=Math.max(...logs),relative=logs.map(v=>Math.exp(v-max)),sum=relative.reduce((z,v)=>z+v,0);weights.forEach((v,k)=>near(v,relative[k]/sum,'integrated child likelihood'));
}
// Unlike Go's seven-face KKT elimination, explicitly minimize each edge and
// solve the reduced 2D interior normal equations on w2=1-w0-w1.
function ridge(a,c){
  const candidates=[[1,0,0],[0,1,0],[0,0,1]],objective=w=>w.reduce((z,v,i)=>z-2*c[i]*v+w.reduce((s,u,j)=>s+v*a[i][j]*u,0),0);
  for(const[i,j]of [[0,1],[0,2],[1,2]]){const t=clamp((c[i]-c[j]-a[i][j]+a[j][j])/(a[i][i]+a[j][j]-2*a[i][j])),w=[0,0,0];w[i]=t;w[j]=1-t;candidates.push(w);}
  const aa=a[0][0]+a[2][2]-2*a[0][2],bb=a[0][1]-a[0][2]-a[1][2]+a[2][2],cc=a[1][1]+a[2][2]-2*a[1][2],u=c[0]-c[2]-a[0][2]+a[2][2],v=c[1]-c[2]-a[1][2]+a[2][2],det=aa*cc-bb*bb;
  check(det>0&&Number.isFinite(det),'positive ridge Hessian');const x=(cc*u-bb*v)/det,y=(aa*v-bb*u)/det;if(x>=0&&y>=0&&x+y<=1)candidates.push([x,y,1-x-y]);candidates.sort((x,y)=>objective(x)-objective(y));return candidates[0];
}
const initialGram=()=>[[1,0,0],[0,1,0],[0,0,1]];
ridge(initialGram(),[.98,.01,.01]).forEach((v,k)=>near(v,[.98,.01,.01][k],'cold ridge'));
function metrics(q,p){const packet=Array.from({length:n},(_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10),risk=i=>(q[i]-p[i])**2+p[i]*(1-p[i]);return{Packet:packet,Brier:q.reduce((z,_,i)=>z+risk(i)/n,0),PriorityBrier:q.reduce((z,_,i)=>z+(i<10?3:1)*risk(i)/170,0),PacketUsefulness:packet.reduce((z,i)=>z+p[i]/10,0),PacketBrier:packet.reduce((z,i)=>z+risk(i)/10,0),PacketBias:packet.reduce((z,i)=>z+(q[i]-p[i])/10,0)};}

function verifyArm(a,w){
  check(arms.includes(`${a.Model}/${a.Policy}`),'model policy');check(Array.isArray(a.Trace)&&a.Trace.length===(a.Model==='baseline'?0:32),'trace length');check(Array.isArray(a.Issued)&&a.Issued.length===(a.Model==='stack'?32:0),'issued-row count');check(a.Forecast.length===n&&a.Weights.length===3,'law dimensions');
  const hyp=hypotheses(w.Base,w.Coordinates,a.Model),weights=hyp.map(h=>h.prior),evidence=new Map(),gram=initialGram(),response=[.98,.01,.01];let fit=[.98,.01,.01];
  for(let step=0;step<a.Trace.length;step++){
    const t=a.Trace[step],i=t.Index;check(Number.isInteger(i)&&i>=0&&i<n&&!evidence.has(i)&&t.Useful===w.Labels[i],'distinct available evidence');const p=predictions(hyp,weights,evidence),law=a.Model==='stack'?Array.from({length:n},(_,j)=>fit.reduce((z,v,k)=>z+v*p.means[k][j],0)):p.q;near(t.Q,law[i],'pre-outcome law');check(t.Probability>0&&t.Probability<=1,'nomination probability');
    if(a.Policy==='stratified_random'){let bucket=0,x=step;for(let bit=0;bit<5;bit++){bucket=2*bucket+(x&1);x>>=1;}const lo=Math.floor(bucket*n/32),hi=Math.floor((bucket+1)*n/32);check(i>=lo&&i<hi,'scheduled stratum');near(t.Probability,1/(hi-lo),'stratum probability');near(t.Score,0,'stratum score');}
    else if(a.Policy==='random'){near(t.Probability,1/(n-step),'random probability');near(t.Score,0,'random score');}
    else{
      check(['uncertainty','information','family_information','disagreement'].includes(a.Policy),'unknown selection');let max=-Infinity,selected=NaN;
      for(let j=0;j<n;j++)if(!evidence.has(j)){let score=H(law[j]);if(a.Policy==='information')score-=weights.reduce((z,v,k)=>z+v*H(p.conditional[k][j]),0);if(a.Policy==='family_information')score-=p.family.reduce((z,v,k)=>z+v*H(p.means[k][j]),0);if(a.Policy==='disagreement')score=fit.reduce((z,v,k)=>z+v*(p.means[k][j]-law[j])**2,0);score=Math.max(0,score);max=Math.max(max,score);if(j===i)selected=score;}
      near(t.Score,max,'selection maximum');if(a.Policy==='family_information'||a.Policy==='disagreement'){const floor=.2/(n-step);check(Math.abs(t.Probability-floor)<=1e-12||Math.abs(t.Probability-floor-.8)<=1e-12,'exploration mixture form');if(t.Probability>floor+.4)near(selected,max,'exploitation maximizer');}else{near(selected,max,'selected maximizer');near(t.Probability,1,'deterministic probability');}
    }
    if(a.Model==='stack'){const row=a.Issued[step];check(Array.isArray(row)&&row.length===3,'issued-row dimensions');row.forEach((v,k)=>near(v,p.means[k][i],'sealed pre-outcome child'));for(let j=0;j<3;j++){response[j]+=row[j]*Number(t.Useful);for(let k=0;k<3;k++)gram[j][k]+=row[j]*row[k];}fit=ridge(gram,response);}
    let mass=0;weights.forEach((v,k)=>{weights[k]*=t.Useful?p.conditional[k][i]:1-p.conditional[k][i];mass+=weights[k];});weights.forEach((v,k)=>weights[k]=v/mass);evidence.set(i,t.Useful);
  }
  integrated(hyp,weights,evidence);const p=predictions(hyp,weights,evidence),law=a.Model==='stack'?Array.from({length:n},(_,i)=>fit.reduce((z,v,k)=>z+v*p.means[k][i],0)):p.q;a.Forecast.forEach((v,i)=>near(v,law[i],'final predictive'));if(a.Model==='stack')a.Weights.forEach((v,k)=>near(v,fit[k],'predictive fit weights'));else if(a.Model==='blend')a.Weights.forEach((v,k)=>near(v,p.family[k],'posterior family weights'));else check(a.Weights.every(v=>v===0),'opaque child weights');
  const m=metrics(a.Forecast,w.Rates);check(JSON.stringify(m.Packet)===JSON.stringify(a.Packet),'packet');for(const f of fields)near(a[f],m[f],f);for(const f of ['SetupNS','SelectionNS','UpdateNS','FinalNS','TotalNS'])check(Number.isFinite(a[f])&&a[f]>=0,'timing');check(a.TotalNS>=a.SetupNS+a.SelectionNS+a.UpdateNS+a.FinalNS,'timing accounting');
}
const interval=a=>{const mean=a.reduce((z,v)=>z+v,0)/a.length,se=Math.sqrt(a.reduce((z,v)=>z+(v-mean)**2,0)/(a.length-1)/a.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};};
const result={study:'stack-v31',wholeGoalsComplete:false,goal7EqualTotalCost:false,splits:{}};
for(const split of ['design','confirmation']){
  const raw=fs.readFileSync(`docs/experiments/mmm-stack-v31-${split}.jsonl`),rows=raw.toString().trim().split('\n').map(JSON.parse),manifest=rows.shift();check(manifest.Kind==='manifest'&&manifest.Split===split&&manifest.Worlds===768,'manifest');near(manifest.SeedBase,split==='design'?2026103103:2026103104,'seed base');check(rows.length===768&&Object.keys(manifest.Sources).length===11,'world/source count');for(const[p,h]of Object.entries(manifest.Sources))check(hash(fs.readFileSync(p))===h,`source changed ${p}`);
  const groups={},ids=new Set();let verifiedArms=0;
  for(const w of rows){
    const g=['tight','wide'].indexOf(w.Geometry),r=regimes.indexOf(w.Regime);check(g>=0&&r>=0&&Number.isInteger(w.World)&&w.World>=0&&w.World<32&&w.Kind==='world','world domain');near(w.Seed,manifest.SeedBase+g*10000000+r*1000000+w.World*1000,'world seed');const id=`${w.Geometry}/${w.Regime}/${w.World}`;check(!ids.has(id),'duplicate world');ids.add(id);
    const step=g===0?.005:.02,base=[.925,.925,...Array.from({length:198},(_,i)=>.65*(Math.cos(step*(i+1))+1)/2+.275)].sort((a,b)=>b-a).slice(0,n);check(w.Base.length===n&&w.Coordinates.length===n&&w.Rates.length===n&&w.Labels.length===n&&w.Labels.every(v=>typeof v==='boolean'),'dimensions');let high=0;
    w.Rates.forEach((p,i)=>{near(w.Base[i],base[i],'baseline');near(w.Coordinates[i],i/149,'coordinate');check(Number.isFinite(p)&&p>=0&&p<=1,'true rate');let want=p,x=i/149;switch(w.Regime){case'independent':check(p===.2||p===.8,'independent rate');high+=p===.8?1:0;break;case'aligned':want=.9-.8*x;break;case'reversed':want=.1+.8*x;break;case'calibrated':want=base[i];break;case'curved':want=.1+.8*Math.sin(Math.PI*x)**2;break;case'shifted_peak':case'narrow_peak':{const width=w.Regime==='narrow_peak'?.04:.14;want=.1+.8*Math.exp(-(((x-(.25+.5*(w.World%8)/7))/width)**2));break;}case'alternating':want=i%2===0?.8:.2;break;case'phase_alternating':want=i%2===0?.2:.8;break;case'tree_matched':check(Number.isInteger(w.Partition)&&w.Partition>=0&&w.Partition<26&&w.LeafMeans.length===15,'tree draw');{const v=w.LeafMeans[partitions[w.Partition].nodes[Math.min(7,Math.floor(8*x))]];check(v>0&&v<1,'leaf mean');}break;}near(p,want,'true-rate formula');});if(w.Regime==='independent')check(high===75,'balanced rates');if(w.Regime==='permuted_curved'){const sorted=w.Rates.slice().sort((a,b)=>a-b),want=Array.from({length:n},(_,i)=>.1+.8*Math.sin(Math.PI*i/149)**2).sort((a,b)=>a-b);sorted.forEach((v,i)=>near(v,want[i],'permuted curve'));}
    check(w.Arms.length===15,'arm count');const unique=new Set();for(const a of w.Arms){const key=`${a.Model}/${a.Policy}`;check(!unique.has(key),'duplicate arm');unique.add(key);verifyArm(a,w);verifiedArms++;}
    const primary=w.Arms.find(a=>a.Model==='stack'&&a.Policy==='stratified_random');for(const m of ['local','old','partition','blend']){const c=w.Arms.find(a=>a.Model===m&&a.Policy==='stratified_random');check(JSON.stringify(primary.Trace.map(t=>[t.Index,t.Useful,t.Probability]))===JSON.stringify(c.Trace.map(t=>[t.Index,t.Useful,t.Probability])),'matched nominee tape');}(groups[`${w.Geometry}/${w.Regime}`]??=[]).push(w);
  }
  const summary={sha256:hash(raw),worlds:rows.length,verifiedArms,sourceHashes:manifest.Sources,groups:{},gates:{}};
  for(const[name,worlds]of Object.entries(groups)){
    check(worlds.length===32,'sample size');const get=(w,m,p)=>w.Arms.find(a=>a.Model===m&&a.Policy===p),stats={};for(const a of worlds[0].Arms){const key=`${a.Model}/${a.Policy}`;stats[key]=Object.fromEntries(fields.map(f=>[f,interval(worlds.map(w=>get(w,a.Model,a.Policy)[f]))]));stats[key].maximumTotalNS=Math.max(...worlds.map(w=>get(w,a.Model,a.Policy).TotalNS));stats[key].meanWeights=[0,1,2].map(k=>worlds.reduce((z,w)=>z+get(w,a.Model,a.Policy).Weights[k]/32,0));}
    summary.groups[name]=stats;const regime=name.split('/')[1],geometry=name.split('/')[0],checks={};
    for(const control of ['local','old','blend'])for(const f of ['Brier','PriorityBrier','PacketUsefulness']){
      const values=worlds.map(w=>{const a=get(w,'stack','stratified_random'),b=get(w,control,'stratified_random');return f==='PacketUsefulness'?a[f]-b[f]:b[f]-a[f];}),v=interval(values),improve=control==='old'&&['curved','shifted_peak'].includes(regime)||control==='blend'&&geometry==='wide'&&['independent','curved'].includes(regime),threshold=improve?(f==='PacketUsefulness'?.02:.01):-.01;
      checks[`${control}/${f}`]={...v,threshold,pass:improve?v.mean>=threshold&&v.lower>0:v.lower>=threshold};
    }
    const bias=interval(worlds.map(w=>get(w,'stack','stratified_random').PacketBias));checks.bias={...bias,magnitudeUpper:Math.abs(bias.mean)+3.5*bias.se,pass:Math.abs(bias.mean)+3.5*bias.se<=.10};const maximum=Math.max(...worlds.flatMap(w=>w.Arms.map(a=>a.TotalNS)));checks.modelTime={maximumNS:maximum,pass:maximum<=10000000};summary.gates[name]={checks,pass:Object.values(checks).every(c=>c.pass)};
  }
  summary.pass=Object.values(summary.gates).every(g=>g.pass);result.splits[split]=summary;console.log(JSON.stringify({split,worlds:rows.length,verifiedArms,pass:summary.pass,failed:Object.entries(summary.gates).filter(([,v])=>!v.pass).map(([name])=>name)}));
}
fs.writeFileSync('docs/experiments/mmm-stack-v31-summary.json',JSON.stringify(result,null,2)+'\n');
