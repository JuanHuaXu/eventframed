// Post-collection serialization-only repair. Frozen checker/data remain unchanged.
import fs from 'node:fs';
import crypto from 'node:crypto';
const n=150,M=27,fields=['Brier','PriorityBrier','PacketUsefulness','PacketBrier','PacketBias'];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const check=(p,s)=>{if(!p)throw Error(s);};
const near=(a,b,s,tol=1e-10)=>check(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<=tol,`${s}: ${a} != ${b}`);
const entropy=p=>p<=0||p>=1?0:-p*Math.log(p)-(1-p)*Math.log1p(-p);
const oldPrior=()=>[.1,.8,...Array(25).fill(.004)];
const oldTable=b=>b.map((p,i)=>[p,(.9*p-.05)/.8,...[.1,.3,.5,.7,.9].flatMap(a=>[-.8,-.4,0,.4,.8].map(c=>Math.max(.02,Math.min(.98,a+c*i/149))))]);
function partitions(lo=0,hi=8,depth=0,node=0){
  const stop={nodes:Array(8).fill(0),prior:depth===3?1:.5};for(let b=lo;b<hi;b++)stop.nodes[b]=node;
  if(depth===3)return[stop];const out=[stop],mid=(lo+hi)/2;
  for(const a of partitions(lo,mid,depth+1,2*node+1))for(const b of partitions(mid,hi,depth+1,2*node+2)){
    const p={nodes:Array(8).fill(0),prior:.5*a.prior*b.prior};for(let j=lo;j<mid;j++)p.nodes[j]=a.nodes[j];for(let j=mid;j<hi;j++)p.nodes[j]=b.nodes[j];out.push(p);
  }return out;
}
const parts=partitions();check(parts.length===26,'partition enumeration');near(parts.reduce((s,p)=>s+p.prior,0),1,'partition prior');
const logFactorial=k=>{let s=0;for(let j=2;j<=k;j++)s+=Math.log(j);return s;};
const cost=p=>({head:151,random:151,stratified:2048,uncertainty:n*(3*M+12)+n,information:n*(6*M+12)+n})[p];
function interval(a){const mean=a.reduce((s,x)=>s+x,0)/a.length,se=Math.sqrt(a.reduce((s,x)=>s+(x-mean)**2,0)/(a.length-1)/a.length);return{mean,se,lower:mean-3.5*se,upper:mean+3.5*se};}
function metrics(q,rates){const packet=Array.from({length:n},(_,i)=>i).sort((a,b)=>q[b]-q[a]||a-b).slice(0,10),risk=i=>(q[i]-rates[i])**2+rates[i]*(1-rates[i]);return{Packet:packet,Brier:q.reduce((s,_,i)=>s+risk(i)/n,0),PriorityBrier:q.reduce((s,_,i)=>s+(i<10?3:1)*risk(i)/170,0),PacketUsefulness:packet.reduce((s,i)=>s+rates[i]/10,0),PacketBrier:packet.reduce((s,i)=>s+risk(i)/10,0),PacketBias:packet.reduce((s,i)=>s+(q[i]-rates[i])/10,0)};}
function bitChoice(seen){for(let k=0;k<256;k++){let j=0,x=k;for(let bit=0;bit<8;bit++){j=2*j+(x&1);x>>=1;}if(j<n&&!seen[j])return j;}throw Error('exhausted bit choice');}
function verifyArm(a,w,table,ent){
  check(a.Trace===null||Array.isArray(a.Trace),'trace domain');
  const trace=a.Trace??[];
  check(a.Forecast.length===n&&a.Weights.length===M,'law dimensions');
  if(a.Model==='baseline'||a.Model==='local'){
    check(a.Setting==='fixed32'&&a.Policy==='head'&&trace.length===(a.Model==='local'?32:0),'control');
    for(let i=0;i<trace.length;i++){const v=trace[i];check(v.Index===i&&v.Useful===w.Labels[i],'local evidence');near(v.Q,w.Base[i],'issued local law');}
    for(let i=0;i<n;i++)near(a.Forecast[i],a.Model==='local'&&i<32?(2*w.Base[i]+Number(w.Labels[i]))/3:w.Base[i],'control final law');
  }else{
    check(['partition','old'].includes(a.Model),'model');const seen=Array(n).fill(false),ys=Array(n).fill(0),s=Array(15).fill(0),f=Array(15).fill(0),bins=w.Coordinates.map(r=>Math.min(7,Math.floor(8*r)));
    const weights=a.Model==='old'?oldPrior():[.99,...parts.map(p=>.01*p.prior)];
    const conditional=(i,k)=>{let p=a.Model==='old'?table[i][k]:k===0?w.Base[i]:(1+s[parts[k-1].nodes[bins[i]]])/(2+s[parts[k-1].nodes[bins[i]]]+f[parts[k-1].nodes[bins[i]]]);return seen[i]?(2*p+ys[i])/3:p;};
    const predict=i=>weights.reduce((z,v,k)=>z+v*conditional(i,k),0);
    check(['fixed32','cost100000','cost1000000'].includes(a.Setting),'setting');near(a.LabelCost,a.Setting==='fixed32'?0:Number(a.Setting.slice(4)),'label cost');
    const setup=20*n*M,final=3*n*M+n*Math.ceil(Math.log2(n)),step=cost(a.Policy)+5*M+16+a.LabelCost,budget=a.Setting==='fixed32'?0:setup+final+32*(a.LabelCost+5*M+16+cost('information'));near(a.Budget,budget,'cap');let used=setup+final;
    check(a.Model==='partition'?['head','random','stratified','uncertainty'].includes(a.Policy):['stratified','information'].includes(a.Policy),'model policy');
    for(const v of trace){const i=v.Index;check(Number.isInteger(i)&&i>=0&&i<n&&!seen[i]&&v.Useful===w.Labels[i],'selected evidence');check(!budget||used+step<=budget,'pre-admission');near(v.Q,predict(i),'pre-label prediction');
      if(a.Policy==='head')check(i===seen.indexOf(false),'head choice');
      if(a.Policy==='stratified')check(i===bitChoice(seen),'stratified choice');
      if(['uncertainty','information'].includes(a.Policy)){let maximum=-Infinity,value=NaN;for(let j=0;j<n;j++)if(!seen[j]){let score=entropy(predict(j));if(a.Policy==='information')score=Math.max(0,score-weights.reduce((z,v,h)=>z+v*ent[j][h],0));maximum=Math.max(maximum,score);if(j===i)value=score;}check(value>=maximum-1e-10,'not maximal selection');}
      let mass=0;for(let k=0;k<M;k++){const p=conditional(i,k);weights[k]*=v.Useful?p:1-p;mass+=weights[k];}for(let k=0;k<M;k++)weights[k]/=mass;
      let node=bins[i]+7;while(true){if(v.Useful)s[node]++;else f[node]++;if(node===0)break;node=Math.floor((node-1)/2);}
      seen[i]=true;ys[i]=Number(v.Useful);used+=step;
    }
    check(trace.length===Math.min(n,budget?Math.floor((budget-setup-final)/step):32),'stopping');near(a.Used,used,'used cost');near(a.Unused,budget?budget-used:0,'unused cost');for(let i=0;i<n;i++)near(a.Forecast[i],predict(i),'final predictive');
    if(a.Model==='partition'){
      for(let k=0;k<M;k++)near(a.Weights[k],weights[k],'posterior partition weight');
      const logs=[Math.log(.99)];for(let i=0;i<n;i++)if(seen[i])logs[0]+=Math.log(ys[i]?w.Base[i]:1-w.Base[i]);
      for(const p of parts){let l=Math.log(.01*p.prior);for(const node of new Set(p.nodes))l+=logFactorial(s[node])+logFactorial(f[node])-logFactorial(s[node]+f[node]+1);logs.push(l);}
      const maximum=Math.max(...logs),relative=logs.map(v=>Math.exp(v-maximum)),mass=relative.reduce((z,v)=>z+v,0);relative.forEach((v,k)=>near(v/mass,weights[k],'independent integrated leaf likelihood'));
    }else check(a.Weights.every(v=>v===0),'old opaque weights marker');
    for(const k of ['SetupNS','SelectionNS','UpdateNS','FinalNS','TotalNS'])check(Number.isFinite(a[k])&&a[k]>=0,'timing');check(a.TotalNS>=a.SetupNS+a.SelectionNS+a.UpdateNS+a.FinalNS,'timing accounting');
  }
  const m=metrics(a.Forecast,w.Rates);check(JSON.stringify(a.Packet)===JSON.stringify(m.Packet),'packet');for(const k of fields)near(a[k],m[k],k);
}
const output={study:'partition-v29',wholeGoalsComplete:false,goal7Screen:false,audit:{repair:'Accept serialized null as empty baseline trace; gates unchanged',executedCheckerSha256:hash(fs.readFileSync(new URL(import.meta.url))),frozenCheckerSha256:hash(fs.readFileSync('research/partition-v29-verify.mjs'))},splits:{}};
for(const split of ['design','confirmation']){
  const path=`docs/experiments/mmm-partition-v29-${split}.jsonl`,raw=fs.readFileSync(path),rows=raw.toString().trim().split('\n').map(JSON.parse),manifest=rows.shift();check(manifest.Kind==='manifest'&&manifest.Split===split&&manifest.Worlds===576,'manifest');near(manifest.SeedBase,split==='design'?2026102903:2026102904,'base seed');for(const[p,h]of Object.entries(manifest.Sources))check(hash(fs.readFileSync(p))===h,`source changed ${p}`);check(rows.length===576,'world count');
  const groups={},ids=new Set();let verifiedArms=0;
  for(const w of rows){const g=['tight','wide'].indexOf(w.Geometry),r=['independent','aligned','reversed','calibrated','curved','shifted_peak','alternating','baseline_matched','tree_matched'].indexOf(w.Regime);check(g>=0&&r>=0&&Number.isInteger(w.World)&&w.World>=0&&w.World<32&&w.Kind==='world','world');near(w.Seed,manifest.SeedBase+g*10000000+r*1000000+w.World*1000,'seed');const id=`${w.Geometry}/${w.Regime}/${w.World}`;check(!ids.has(id),'duplicate world');ids.add(id);
    const step=g===0?.005:.02,base=[.925,.925,...Array.from({length:198},(_,i)=>.65*(Math.cos(step*(i+1))+1)/2+.275)].sort((a,b)=>b-a).slice(0,n);check(w.Base.length===n&&w.Coordinates.length===n&&w.Rates.length===n&&w.Labels.length===n&&w.Labels.every(v=>typeof v==='boolean'),'dimensions');for(let i=0;i<n;i++){near(w.Base[i],base[i],'cosine baseline');near(w.Coordinates[i],i/149,'predeclared coordinate');}
    let high=0;w.Rates.forEach((p,i)=>{check(Number.isFinite(p)&&p>=0&&p<=1,'true rate');let want=p,x=i/149;switch(w.Regime){case'independent':check(p===.2||p===.8,'independent rates');high+=p===.8?1:0;break;case'aligned':want=.9-.8*x;break;case'reversed':want=.1+.8*x;break;case'calibrated':want=base[i];break;case'curved':want=.1+.8*Math.sin(Math.PI*x)**2;break;case'shifted_peak':want=.1+.8*Math.exp(-(((x-(.25+.5*(w.World%8)/7))/.14)**2));break;case'alternating':want=i%2===0?.8:.2;break;case'tree_matched':check(Number.isInteger(w.Partition)&&w.Partition>=0&&w.Partition<26&&w.LeafMeans.length===15,'tree branch');const node=parts[w.Partition].nodes[Math.min(7,Math.floor(8*x))];check(w.LeafMeans[node]>0&&w.LeafMeans[node]<1,'leaf mean');break;}near(p,want,'truth formula');});if(w.Regime==='independent')check(high===75,'balanced rates');
    const table=oldTable(base),ent=table.map(row=>row.map(entropy));check(w.Arms.length===20,'arm count');const unique=new Set();for(const a of w.Arms){const key=`${a.Setting}/${a.Model}/${a.Policy}`;check(!unique.has(key),'duplicate arm');unique.add(key);verifyArm(a,w,table,ent);verifiedArms++;}
    const part=w.Arms.find(a=>a.Setting==='fixed32'&&a.Model==='partition'&&a.Policy==='stratified'),old=w.Arms.find(a=>a.Setting==='fixed32'&&a.Model==='old'&&a.Policy==='stratified');check(JSON.stringify(part.Trace.map(v=>[v.Index,v.Useful]))===JSON.stringify(old.Trace.map(v=>[v.Index,v.Useful])),'different evidence for model comparison');(groups[`${w.Geometry}/${w.Regime}`]??=[]).push(w);
  }
  const result={sha256:hash(raw),worlds:rows.length,verifiedArms,sourceHashes:manifest.Sources,groups:{},gates:{}};
  for(const [name,worlds]of Object.entries(groups)){check(worlds.length===32,'paired group');const get=(w,model,policy,setting='fixed32')=>w.Arms.find(a=>a.Model===model&&a.Policy===policy&&a.Setting===setting),summaries={};
    for(const a of worlds[0].Arms){const key=`${a.Setting}/${a.Model}/${a.Policy}`;summaries[key]=Object.fromEntries(fields.map(f=>[f,interval(worlds.map(w=>get(w,a.Model,a.Policy,a.Setting)[f]))]));summaries[key].meanLabels=worlds.reduce((z,w)=>z+(get(w,a.Model,a.Policy,a.Setting).Trace??[]).length,0)/32;summaries[key].maximumTotalNS=Math.max(...worlds.map(w=>get(w,a.Model,a.Policy,a.Setting).TotalNS));}
    result.groups[name]=summaries;const regime=name.split('/')[1],checks={};for(const control of ['local','old']){
      for(const field of ['Brier','PriorityBrier','PacketUsefulness']){const gains=interval(worlds.map(w=>{const a=get(w,'partition','stratified'),b=get(w,control,control==='old'?'stratified':'head');return field==='PacketUsefulness'?a[field]-b[field]:b[field]-a[field];}));let pass,required;
        if(['independent','reversed'].includes(regime)){if(field==='PacketUsefulness')continue;required=.02;pass=gains.mean>=required&&gains.lower>0;}
        else if(['curved','shifted_peak'].includes(regime)&&control==='old'){required=field==='PacketUsefulness'?.02:.01;pass=gains.mean>=required&&gains.lower>0;}
        else{required=-.01;pass=gains.lower>=required;}
        checks[`${control}/${field}`]={...gains,required,pass};
      }
    }
    const bias=interval(worlds.map(w=>get(w,'partition','stratified').PacketBias));checks.bias={...bias,magnitudeUpper:Math.abs(bias.mean)+3.5*bias.se,pass:Math.abs(bias.mean)+3.5*bias.se<=.10};result.gates[name]={checks,pass:Object.values(checks).every(v=>v.pass)};
  }
  result.pass=Object.values(result.gates).every(v=>v.pass);output.splits[split]=result;console.log(JSON.stringify({split,sha256:result.sha256,worlds:rows.length,verifiedArms,pass:result.pass,failed:Object.entries(result.gates).filter(([,v])=>!v.pass).map(([k])=>k)}));
}
fs.writeFileSync('docs/experiments/mmm-partition-v29-summary.json',JSON.stringify(output,null,2)+'\n');
