// Independent integrated-likelihood LOO reconstruction. Go removes conditional
// likelihoods from child masses; this audit explicitly changes Beta integrals.
import fs from 'node:fs';
import crypto from 'node:crypto';
const n=150,fields=['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'];
const regimes=['independent','aligned','reversed','calibrated','curved','shifted_peak','alternating','phase_alternating','permuted_curved','baseline_matched','tree_matched','narrow_peak'];
const arms=['baseline/head','local/stratified_random','old/stratified_random','old/random','old/uncertainty','old/information','partition/stratified_random','blend/stratified_random','blend/random','blend/uncertainty','blend/family_information','stack/stratified_random','stack/random','stack/uncertainty','stack/disagreement','crossfit/stratified_random','crossfit/random'];
const check=(p,s)=>{if(!p)throw Error(s);},near=(a,b,s,t=1e-10)=>check(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=t,`${s}: ${a} != ${b}`);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const factorial=[0];for(let j=1;j<=201;j++)factorial[j]=factorial[j-1]+Math.log(j);
const beta=(s,f)=>factorial[s]+factorial[f]-factorial[s+f+1];
function trees(lo=0,hi=8,depth=0,node=0){const p={nodes:Array(8).fill(0),prior:depth===3?1:.5};for(let i=lo;i<hi;i++)p.nodes[i]=node;if(depth===3)return[p];const out=[p],mid=(lo+hi)/2;for(const l of trees(lo,mid,depth+1,node*2+1))for(const r of trees(mid,hi,depth+1,node*2+2)){const q={nodes:Array(8).fill(0),prior:.5*l.prior*r.prior};for(let i=lo;i<mid;i++)q.nodes[i]=l.nodes[i];for(let i=mid;i<hi;i++)q.nodes[i]=r.nodes[i];out.push(q);}return out;}
const partitions=trees();check(partitions.length===26,'partition count');near(partitions.reduce((z,p)=>z+p.prior,0),1,'partition mass');
function norm(logs){const max=Math.max(...logs),v=logs.map(x=>Math.exp(x-max)),sum=v.reduce((z,x)=>z+x,0);return{w:v.map(x=>x/sum),z:max+Math.log(sum)};}
function components(base){const out=[{prior:.1,p:base},{prior:.8,p:base.map(b=>(.9*b-.05)/.8)}];for(const a of [.1,.3,.5,.7,.9])for(const c of [-.8,-.4,0,.4,.8])out.push({prior:.004,p:base.map((_,i)=>Math.max(.02,Math.min(.98,a+c*i/149)))});return out;}
function joint(base,coordinate,evidence,affine){
  const s=Array(15).fill(0),f=Array(15).fill(0);let baseLL=0;
  const affineLogs=affine.map(h=>{let v=Math.log(h.prior);for(const[i,y]of evidence)v+=Math.log(y?h.p[i]:1-h.p[i]);return v;});
  for(const[i,y]of evidence){baseLL+=Math.log(y?base[i]:1-base[i]);let node=Math.min(7,Math.floor(8*coordinate[i]))+7;while(true){(y?s:f)[node]++;if(node===0)break;node=Math.floor((node-1)/2);}}
  const partLogs=[Math.log(.99)+baseLL,...partitions.map(p=>{let v=Math.log(.01*p.prior);for(const node of new Set(p.nodes))v+=beta(s[node],f[node]);return v;})];
  const af=norm(affineLogs),pt=norm(partLogs),family=norm([Math.log(.98)+baseLL,Math.log(.01)+af.z,Math.log(.01)+pt.z]).w;
  const future=i=>{let q=[base[i],af.w.reduce((z,v,k)=>z+v*affine[k].p[i],0),pt.w[0]*base[i]];for(let k=0;k<26;k++){const node=partitions[k].nodes[Math.min(7,Math.floor(8*coordinate[i]))];q[2]+=pt.w[k+1]*(1+s[node])/(2+s[node]+f[node]);}if(evidence.has(i))q=q.map(v=>(2*v+Number(evidence.get(i)))/3);return q;};
  const loo=i=>{check(evidence.has(i),'LOO member observed');const y=evidence.get(i),removed=Number(y),a=norm(affineLogs.map((v,k)=>v-Math.log(y?affine[k].p[i]:1-affine[k].p[i]))),logs=partLogs.slice();logs[0]-=Math.log(y?base[i]:1-base[i]);for(let k=0;k<26;k++){const node=partitions[k].nodes[Math.min(7,Math.floor(8*coordinate[i]))];logs[k+1]+=beta(s[node]-removed,f[node]-(1-removed))-beta(s[node],f[node]);}const p=norm(logs);let part=p.w[0]*base[i];for(let k=0;k<26;k++){const node=partitions[k].nodes[Math.min(7,Math.floor(8*coordinate[i]))];part+=p.w[k+1]*(1+s[node]-removed)/(1+s[node]+f[node]);}return[base[i],a.w.reduce((z,v,k)=>z+v*affine[k].p[i],0),part];};
  return{family,future,loo};
}
const gram0=()=>[[1,0,0],[0,1,0],[0,0,1]],dot=(a,b)=>a.reduce((z,v,k)=>z+v*b[k],0);
function ridge(a,c){const options=[[1,0,0],[0,1,0],[0,0,1]],cost=w=>w.reduce((z,v,i)=>z-2*c[i]*v+w.reduce((x,u,j)=>x+v*a[i][j]*u,0),0);for(const[i,j]of [[0,1],[0,2],[1,2]]){const t=Math.max(0,Math.min(1,(c[i]-c[j]-a[i][j]+a[j][j])/(a[i][i]+a[j][j]-2*a[i][j]))),w=[0,0,0];w[i]=t;w[j]=1-t;options.push(w);}const aa=a[0][0]+a[2][2]-2*a[0][2],bb=a[0][1]-a[0][2]-a[1][2]+a[2][2],cc=a[1][1]+a[2][2]-2*a[1][2],u=c[0]-c[2]-a[0][2]+a[2][2],v=c[1]-c[2]-a[1][2]+a[2][2],det=aa*cc-bb*bb;check(det>0,'positive ridge');const x=(cc*u-bb*v)/det,y=(aa*v-bb*u)/det;if(x>=0&&y>=0&&x+y<=1)options.push([x,y,1-x-y]);options.sort((x,y)=>cost(x)-cost(y));return options[0];}
function add(a,c,row,y){for(let i=0;i<3;i++){c[i]+=row[i]*Number(y);for(let j=0;j<3;j++)a[i][j]+=row[i]*row[j];}}
function fit(state,evidence){const a=gram0(),c=[.98,.01,.01];for(const[i,y]of evidence)add(a,c,state.loo(i),y);return ridge(a,c);}
function metrics(q,p){const Packet=Array.from({length:n},(_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10),risk=i=>(q[i]-p[i])**2+p[i]*(1-p[i]);return{Packet,Brier:q.reduce((z,_,i)=>z+risk(i)/150,0),PriorityBrier:q.reduce((z,_,i)=>z+(i<10?3:1)*risk(i)/170,0),PacketUsefulness:Packet.reduce((z,i)=>z+p[i]/10,0),PacketBrier:Packet.reduce((z,i)=>z+risk(i)/10,0),PacketBias:Packet.reduce((z,i)=>z+(q[i]-p[i])/10,0)};}
function verifyArm(a,w,affine){
  const model=a.Model,key=`${model}/${a.Policy}`;check(arms.includes(key),'model policy');check(a.Trace.length===(model==='baseline'?0:32),'trace size');check(a.Issued.length===(model==='stack'?32:0),'issued size');check(a.LOO.length===(model==='crossfit'?32:0),'LOO size');check(a.Weights.length===3&&a.Forecast.length===150,'prediction dimensions');
  const evidence=new Map(),gram=gram0(),response=[.98,.01,.01];let state=joint(w.Base,w.Coordinates,evidence,affine),weights=[.98,.01,.01];
  const predict=i=>{const p=state.future(i);switch(model){case'baseline':case'local':return p[0];case'old':return p[1];case'partition':return p[2];case'blend':return dot(state.family,p);case'stack':case'crossfit':return dot(weights,p);default:throw Error('model');}};
  for(let step=0;step<a.Trace.length;step++){
    const t=a.Trace[step],i=t.Index;check(Number.isInteger(i)&&i>=0&&i<n&&!evidence.has(i)&&t.Useful===w.Labels[i],'distinct available evidence');near(t.Q,predict(i),'pre-outcome law');check(Number.isFinite(t.Score)&&t.Score>=0&&t.Probability>0&&t.Probability<=1,'nomination domain');
    if(a.Policy==='stratified_random'){let bucket=0,x=step;for(let bit=0;bit<5;bit++){bucket=2*bucket+(x&1);x>>=1;}const lo=Math.floor(bucket*150/32),hi=Math.floor((bucket+1)*150/32);check(i>=lo&&i<hi,'stratum');near(t.Probability,1/(hi-lo),'stratum probability');near(t.Score,0,'stratum score');}else if(a.Policy==='random'){near(t.Probability,1/(150-step),'uniform probability');near(t.Score,0,'uniform score');}else if(a.Policy==='family_information'||a.Policy==='disagreement'){const floor=.2/(150-step);check(Math.abs(t.Probability-floor)<1e-12||Math.abs(t.Probability-floor-.8)<1e-12,'old exploration form');}else near(t.Probability,1,'old deterministic nomination');
    if(model==='stack'){const row=state.future(i);check(a.Issued[step].length===3,'issued row dimensions');row.forEach((v,k)=>near(v,a.Issued[step][k],'sealed issued child'));add(gram,response,row,t.Useful);weights=ridge(gram,response);}
    evidence.set(i,t.Useful);state=joint(w.Base,w.Coordinates,evidence,affine);if(model==='crossfit')weights=fit(state,evidence);
  }
  a.Forecast.forEach((v,i)=>near(v,predict(i),'final law'));if(model==='blend')a.Weights.forEach((v,k)=>near(v,state.family[k],'family weights'));else if(model==='stack'||model==='crossfit')a.Weights.forEach((v,k)=>near(v,weights[k],'predictive fit weights'));else check(a.Weights.every(v=>v===0),'child-only weights');
  if(model==='crossfit')a.Trace.forEach((t,j)=>{const row=state.loo(t.Index);check(a.LOO[j].length===3,'LOO row dimensions');row.forEach((v,k)=>near(v,a.LOO[j][k],'integrated omitted-member row'));});
  const m=metrics(a.Forecast,w.Rates);check(JSON.stringify(m.Packet)===JSON.stringify(a.Packet),'packet');for(const f of fields)near(a[f],m[f],f);for(const f of ['SetupNS','SelectionNS','UpdateNS','FinalNS','TotalNS'])check(Number.isFinite(a[f])&&a[f]>=0,'timing');check(a.TotalNS>=a.SetupNS+a.SelectionNS+a.UpdateNS+a.FinalNS,'elapsed accounting');
}
const interval=a=>{const mean=a.reduce((z,v)=>z+v,0)/a.length,se=Math.sqrt(a.reduce((z,v)=>z+(v-mean)**2,0)/(a.length-1)/a.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};};
const result={study:'crossfit-v32',wholeGoalsComplete:false,goal7EqualTotalCost:false,splits:{}};
if(process.argv.includes('--self-test')){
  const base=Array.from({length:n},(_,i)=>.925-.675*i/149),coordinate=Array.from({length:n},(_,i)=>i/149),affine=components(base);
  let checks=0;
  for(const count of [1,16,32,150]){
    const evidence=new Map(Array.from({length:count},(_,i)=>[i,i%3!==0])),state=joint(base,coordinate,evidence,affine);
    for(const i of new Set([0,Math.floor((count-1)/2),count-1])){
      const removed=new Map(evidence);removed.delete(i);const want=joint(base,coordinate,removed,affine).future(i),got=state.loo(i),flipped=new Map(evidence);flipped.set(i,!flipped.get(i));const other=joint(base,coordinate,flipped,affine).loo(i);
      got.forEach((v,k)=>{near(v,want[k],'independent audit full omission',2e-13);near(v,other[k],'audit own-label flip',2e-13);checks+=2;});
    }
    const w=fit(state,evidence);near(w.reduce((a,b)=>a+b,0),1,'fit simplex');check(w.every(v=>v>=0),'fit positivity');
  }
  console.log(JSON.stringify({independentAuditSelfTest:true,checks}));process.exit(0);
}
for(const split of ['design','confirmation']){
  const raw=fs.readFileSync(`docs/experiments/mmm-crossfit-v32-${split}.jsonl`),rows=raw.toString().trim().split('\n').map(JSON.parse),manifest=rows.shift();check(manifest.Kind==='manifest'&&manifest.Split===split&&manifest.Worlds===768&&rows.length===768,'manifest');near(manifest.SeedBase,split==='design'?2026103203:2026103204,'seed');check(Object.keys(manifest.Sources).length===21,'source count');for(const[p,h]of Object.entries(manifest.Sources))check(hash(fs.readFileSync(p))===h,`source changed ${p}`);
  const groups={},ids=new Set();let verifiedArms=0;
  for(const w of rows){
    const g=['tight','wide'].indexOf(w.Geometry),r=regimes.indexOf(w.Regime);check(g>=0&&r>=0&&w.Kind==='world'&&Number.isInteger(w.World)&&w.World>=0&&w.World<32,'world');near(w.Seed,manifest.SeedBase+g*10000000+r*1000000+w.World*1000,'world seed');const id=`${w.Geometry}/${w.Regime}/${w.World}`;check(!ids.has(id),'world duplicate');ids.add(id);
    check(w.Base.length===150&&w.Coordinates.length===150&&w.Rates.length===150&&w.Labels.length===150&&w.Labels.every(v=>typeof v==='boolean'),'world dimensions');const step=g===0?.005:.02,base=[.925,.925,...Array.from({length:198},(_,i)=>.65*(Math.cos(step*(i+1))+1)/2+.275)].sort((a,b)=>b-a).slice(0,150);for(let i=0;i<n;i++){near(w.Base[i],base[i],'cosine baseline');near(w.Coordinates[i],i/149,'coordinate');check(Number.isFinite(w.Rates[i])&&w.Rates[i]>=0&&w.Rates[i]<=1,'rate domain');}
    const affine=components(w.Base),unique=new Set();check(w.Arms.length===17,'arm count');for(const a of w.Arms){const key=`${a.Model}/${a.Policy}`;check(!unique.has(key),'duplicate arm');unique.add(key);verifyArm(a,w,affine);verifiedArms++;}
    const primary=w.Arms.find(a=>a.Model==='crossfit'&&a.Policy==='stratified_random');for(const m of ['local','old','partition','blend','stack']){const control=w.Arms.find(a=>a.Model===m&&a.Policy==='stratified_random');check(JSON.stringify(primary.Trace.map(t=>[t.Index,t.Useful,t.Probability]))===JSON.stringify(control.Trace.map(t=>[t.Index,t.Useful,t.Probability])),'matched stratum tape');}(groups[`${w.Geometry}/${w.Regime}`]??=[]).push(w);
  }
  const summary={sha256:hash(raw),worlds:rows.length,verifiedArms,sourceHashes:manifest.Sources,groups:{},gates:{}};
  for(const[name,worlds]of Object.entries(groups)){
    check(worlds.length===32,'cell size');const get=(w,m,p)=>w.Arms.find(a=>a.Model===m&&a.Policy===p),stats={};for(const a of worlds[0].Arms){const key=`${a.Model}/${a.Policy}`;stats[key]=Object.fromEntries(fields.map(f=>[f,interval(worlds.map(w=>get(w,a.Model,a.Policy)[f]))]));stats[key].maximumTotalNS=Math.max(...worlds.map(w=>get(w,a.Model,a.Policy).TotalNS));stats[key].meanWeights=[0,1,2].map(k=>worlds.reduce((z,w)=>z+get(w,a.Model,a.Policy).Weights[k]/32,0));}summary.groups[name]=stats;
    const[geometry,regime]=name.split('/'),checks={};for(const control of ['local','old','blend','stack'])for(const f of ['Brier','PriorityBrier','PacketUsefulness']){const values=worlds.map(w=>{const a=get(w,'crossfit','stratified_random'),b=get(w,control,'stratified_random');return f==='PacketUsefulness'?a[f]-b[f]:b[f]-a[f];}),v=interval(values),improve=control==='old'&&['curved','shifted_peak'].includes(regime)||control==='blend'&&geometry==='wide'&&['independent','curved'].includes(regime),threshold=improve?(f==='PacketUsefulness'?.02:.01):-.01;checks[`${control}/${f}`]={...v,threshold,pass:improve?v.mean>=threshold&&v.lower>0:v.lower>=threshold};}
    const bias=interval(worlds.map(w=>get(w,'crossfit','stratified_random').PacketBias));checks.bias={...bias,magnitudeUpper:Math.abs(bias.mean)+3.5*bias.se,pass:Math.abs(bias.mean)+3.5*bias.se<=.10};const maximum=Math.max(...worlds.flatMap(w=>w.Arms.map(a=>a.TotalNS)));checks.modelTime={maximumNS:maximum,pass:maximum<=10000000};summary.gates[name]={checks,pass:Object.values(checks).every(v=>v.pass)};
  }
  summary.pass=Object.values(summary.gates).every(v=>v.pass);result.splits[split]=summary;console.log(JSON.stringify({split,worlds:rows.length,verifiedArms,pass:summary.pass,failed:Object.entries(summary.gates).filter(([,v])=>!v.pass).map(([n])=>n)}));
}
fs.writeFileSync('docs/experiments/mmm-crossfit-v32-summary.json',JSON.stringify(result,null,2)+'\n');
