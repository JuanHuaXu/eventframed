import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const raw=fs.readFileSync(process.argv[2]),a=JSON.parse(raw),root=path.resolve(import.meta.dirname,'..');
const assert=(x,m)=>{if(!x)throw Error(m)},eq=(x,y)=>JSON.stringify(x)===JSON.stringify(y);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
assert(a.Version==='transfer-v116'&&a.SeedBase===2180111700&&a.PerCase===32&&a.Records.length===1152,'contract');
for(const [p,h] of Object.entries(a.Hashes))assert(hash(fs.readFileSync(path.join(root,p)))===h,`source ${p}`);
const pop=x=>x.toString(2).replaceAll('0','').length;
const families=['additive','hierarchy','local_table'],modes=['stationary','abrupt','gradual'];
function endpoint(p,f,x){
 const bit=c=>(x>>c)&1;
 if(f===0){let z=p.Bias;for(let i=0;i<5;i++)z+=p.Coefficients[i]*(bit(p.Coordinates[i])?1:-1);return .5+.4*Math.tanh(z/2);}
 if(f===1){const aa=bit(p.Branches[0]),b=bit(p.Branches[1+aa]),c=bit(p.Branches[3+2*aa+b]);return p.Probabilities[4*aa+2*b+c];}
 let k=0;for(let i=0;i<4;i++)k+=bit(p.Coordinates[i])*(2**i);return p.Probabilities[k];
}
function truth(d,x,t){let w=0;if(d.Spec.Mode!==0&&t>=d.Change)w=d.Spec.Mode===1?1:Math.min(1,(t-d.Change+1)/32);return (1-w)*endpoint(d.Models[0],d.Spec.Family,x)+w*endpoint(d.Models[1],d.Spec.Family,x);}
const seen=new Set(),seeds=new Set(),pairs=new Map();
for(const r of a.Records){
 const d=r.Teacher,spec=d.Spec,{Phase:ph,Family:f,Mode:m,Index:i}=spec;
 assert(spec.SeedBase===a.SeedBase&&[0,1].includes(ph)&&[0,1,2].includes(f)&&[0,1,2].includes(m)&&Number.isInteger(i)&&i>=0&&i<32&&[0,1].includes(r.Schedule),'identity');
 assert(d.Change>=97&&d.Change<=159&&d.Change%2===1,'change clock');
 const latent=[ph,f,m,i].join('/'),key=latent+'/'+r.Schedule;assert(!seen.has(key),'duplicate run');seen.add(key);
 if(!pairs.has(latent))pairs.set(latent,[]);pairs.get(latent)[r.Schedule]=r;
 for(let role=0;role<5;role++){const seed=a.SeedBase+ph*1e6+f*1e5+m*1e4+i*10+role;assert(d.Seeds[role]===seed,'declared seed');if(r.Schedule===0){assert(!seeds.has(seed%2147483647),'seed collision');seeds.add(seed%2147483647);}}
 assert(r.Steps.length===256&&r.Fits.length===8&&r.Final.length===2,'sizes');
 const b=Array.from({length:3},()=>[0,0]),ac=Array.from({length:3},()=>[0,0]),logs=Array.from({length:3},()=>[0,0]),real=[0,0,0],rl=[0,0,0];
 for(const [t,s] of r.Steps.entries()){
  assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512&&typeof s.Y==='boolean'&&Math.abs(s.Q-truth(d,s.X,t))<1e-13,'teacher probability');
  assert(s.Q>=.1&&s.Q<=.9&&Number.isInteger(s.Delay)&&s.Delay>=0&&s.Delay<=31&&typeof s.Missing==='boolean','packet');
  if(r.Schedule===0)assert(s.Delay===0&&!s.Missing,'immediate');
  for(let arm=0;arm<3;arm++){
   const p=s.P[arm],mask=s.Mask[arm];assert(Number.isFinite(p)&&p>0&&p<1,'forecast');
   assert(Number.isInteger(mask)&&mask>0&&mask<512&&(mask&1)===1&&s.Cost[arm]===pop(mask)&&s.Cost[arm]<=6,'charged acquisition');
   for(let scope=0;scope<3;scope++)assert([0,1,3,7].includes((mask>>(3*scope))&7),'prefix mask');
   const loss=(p-s.Q)**2+s.Q*(1-s.Q),acc=p>=.5?s.Q:1-s.Q,ll=-s.Q*Math.log(p)-(1-s.Q)*Math.log1p(-p);
   for(const seg of t>=192?[0,1]:[0]){const n=seg?64:256;b[arm][seg]+=loss/n;ac[arm][seg]+=acc/n;logs[arm][seg]+=ll/n;}
   real[arm]+=(p-Number(s.Y))**2/256;rl[arm]-=(s.Y?Math.log(p):Math.log1p(-p))/256;
  }
  if(r.Schedule===0)assert(s.P[1]===s.P[2]&&s.Mask[1]===s.Mask[2],'immediate log/Markov identity');
 }
 for(let arm=0;arm<3;arm++){
  assert(Math.abs(real[arm]-r.Realized[arm])<1e-12&&Math.abs(rl[arm]-r.RealizedLog[arm])<1e-12,'realized scores');
  for(let seg=0;seg<2;seg++)assert(Math.abs(b[arm][seg]-r.Metrics[arm][seg].Brier)<1e-12&&Math.abs(ac[arm][seg]-r.Metrics[arm][seg].Accuracy)<1e-12&&Math.abs(logs[arm][seg]-r.LogLoss[arm][seg])<1e-12,'expected scores');
 }
 // Rebuild origin-order gate accounting without using the Go queue/counters.
 const state=Array(256).fill(0),arrived=Array(256).fill(false);
 let next=0,version=-1;
 const stats=[0,1].map(()=>({Issued:0,Pending:0,Applied:0,BankOnly:0,Stale:0,Censored:0}));
 const drain=()=>{while(next<256&&[2,3].includes(state[next])){for(const z of stats){z.Pending--;if(state[next]===3)z.Censored++;else if(Math.floor(next/32)===version)z.Applied++;else z.BankOnly++;}state[next]=4;next++;}};
 for(let clock=0;clock<288;clock++){
  for(let origin=0;origin<Math.min(clock,256);origin++){const s=r.Steps[origin];if(!arrived[origin]&&!s.Missing&&origin+s.Delay<=clock){assert(state[origin]===1,'expired arrival');arrived[origin]=true;state[origin]=2;drain();}}
  if(clock>=32){for(let j=0;j<Math.min(clock-31,256);j++)if(state[j]===1)state[j]=3;drain();}
  if(clock>=256)continue;
  if(clock%32===0){version=clock/32;const available=Array.from({length:16},(_,j)=>j-16);for(let j=0;j<clock;j++)if(arrived[j])available.push(j);const fit=r.Fits[version];assert(fit.Clock===clock&&eq(fit.Origins,[available.slice(-64),available.slice(-32)]),'fit evidence');}
  assert(state[clock]===0,'duplicate issue');state[clock]=1;for(const z of stats){z.Issued++;z.Pending++;assert(z.Pending<=64,'queue cap');}
  const s=r.Steps[clock];if(!s.Missing&&s.Delay===0){arrived[clock]=true;state[clock]=2;drain();}
  assert(eq(s.Stats,stats)&&eq(s.Selector,Array(2).fill(arrived.filter(Boolean).length)),`journal ${key}/${clock}`);
 }
 assert(next===256&&eq(stats,r.Final)&&r.Arrived===arrived.filter(Boolean).length&&eq(r.FinalSelector,[r.Arrived,r.Arrived]),'terminal accounting');
}
assert(pairs.size===576&&seeds.size===2880,'latent count');
for(const p of pairs.values()){assert(p.length===2&&eq(p[0].Teacher,p[1].Teacher),'teacher pair');for(let t=0;t<256;t++)for(const k of ['X','Y','Q'])assert(p[0].Steps[t][k]===p[1].Steps[t][k],'packet pair');}
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
function interval(xs){assert(xs.length===32,'paired n');const mu=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-mu)**2,0)/31/32);return {mean:mu,lower:mu-3.5*se,upper:mu+3.5*se};}
const gates=[],groups=[];
for(const ph of [0,1])for(const f of [0,1,2])for(const m of [0,1,2])for(const schedule of [0,1]){
 const rs=a.Records.filter(r=>r.Teacher.Spec.Phase===ph&&r.Teacher.Spec.Family===f&&r.Teacher.Spec.Mode===m&&r.Schedule===schedule);assert(rs.length===32,'cell size');
 for(const seg of [0,1]){
  const meta={phase:['design','confirmation'][ph],family:families[f],mode:modes[m],schedule,segment:seg?'terminal64':'all256'};
  const subset=r=>r.Steps.slice(seg?192:0);
  const metrics=[0,1,2].map(arm=>({brier:mean(rs.map(r=>r.Metrics[arm][seg].Brier)),accuracy:mean(rs.map(r=>r.Metrics[arm][seg].Accuracy)),logLoss:mean(rs.map(r=>r.LogLoss[arm][seg])),acquisitionReads:mean(rs.map(r=>mean(subset(r).map(s=>s.Cost[arm]))))}));
  groups.push({...meta,metrics,bayesBrierFloor:mean(rs.map(r=>mean(subset(r).map(s=>s.Q*(1-s.Q))))),bayesAccuracyCeiling:mean(rs.map(r=>mean(subset(r).map(s=>Math.max(s.Q,1-s.Q))))),fullAuditReadsPerTrajectory:mean(rs.map(r=>144+9*r.Arrived))});
  for(const candidate of [1,2]){
   for(const control of [0,1,2].filter(v=>v!==candidate)){const v=interval(rs.map(r=>r.Metrics[candidate][seg].Brier-r.Metrics[control][seg].Brier));gates.push({...meta,type:'nonharm',candidate,control,...v,pass:v.upper<=.01});}
   const controls=seg===1&&m!==0?[0,...(candidate===2&&schedule===1?[1]:[])]:[];
   for(const control of controls){const v=interval(rs.map(r=>r.Metrics[control][seg].Brier-r.Metrics[candidate][seg].Brier));gates.push({...meta,type:'gain',candidate,control,...v,pass:v.mean>=.005&&v.lower>0});}
  }
 }
}
assert(gates.length===348&&gates.filter(g=>g.type==='nonharm').length===288&&gates.filter(g=>g.type==='gain').length===60,'gate totals');
const candidates=[1,2].map(candidate=>{const gs=gates.filter(g=>g.candidate===candidate);assert(gs.length===(candidate===1?168:180),'candidate total');return {candidate,passed:gs.filter(g=>g.pass).length,total:gs.length,status:gs.every(g=>g.pass)?'PASS':'FAIL'};});
process.stdout.write(JSON.stringify({version:'transfer-v116',artifactSHA256:hash(raw),sourceHashes:Object.keys(a.Hashes).length,latentTrajectories:576,scheduleRuns:1152,steps:1152*256,uniqueEffectiveSeeds:seeds.size,passed:gates.filter(g=>g.pass).length,total:gates.length,status:gates.every(g=>g.pass)?'PASS':'FAIL',candidates,gates,groups},null,2)+'\n');
