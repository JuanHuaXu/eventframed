import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const input=process.argv[2],raw=fs.readFileSync(input),a=JSON.parse(raw);
const assert=(x,m)=>{if(!x)throw Error(m)},eq=(x,y)=>JSON.stringify(x)===JSON.stringify(y);
const root=path.resolve(import.meta.dirname,'..');
assert(a.Version==='v111'&&a.SeedBase===2140111100&&a.PerCase===32,'metadata');
for(const [file,hash] of Object.entries(a.Hashes))assert(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,file))).digest('hex')===hash,`hash ${file}`);
const cases=['parity1','parity2','parity3','parity4','complement4','majority3','mux3','constant','null','dependent4','majority_to_parity','parity_to_majority'];
assert(a.Records.length===1536,'run count');
const seen=new Set(),pairs=new Map(),seeds=new Set();
const pop=x=>x.toString(2).replaceAll('0','').length;
for(const r of a.Records){
 const phase=['design','confirmation'].indexOf(r.Phase),c=cases.indexOf(r.Case);
 assert(phase>=0&&c>=0&&Number.isInteger(r.Index)&&r.Index>=0&&r.Index<32&&[0,1].includes(r.Schedule),'identity');
 const latent=[phase,c,r.Index].join('/'),key=latent+'/'+r.Schedule;assert(!seen.has(key),'duplicate');seen.add(key);
 if(!pairs.has(latent))pairs.set(latent,[]);pairs.get(latent)[r.Schedule]=r;
 if(r.Schedule===0)for(let role=0;role<5;role++){
  const seed=(a.SeedBase+phase*1e6+c*1e4+r.Index*10+role)%2147483647;assert(!seeds.has(seed),'seed');seeds.add(seed);
 }
 assert(r.Steps.length===256&&r.Fits.length===8,'counts');
 const b=Array.from({length:9},()=>[0,0]),ac=Array.from({length:9},()=>[0,0]),real=Array(9).fill(0);
 for(const [t,s] of r.Steps.entries()){
  assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512&&[.05,.5,.95].some(q=>Math.abs(q-s.Q)<1e-12)&&typeof s.Y==='boolean','target');
  assert(Number.isInteger(s.Delay)&&s.Delay>=0&&s.Delay<=31&&typeof s.Missing==='boolean','schedule');
  if(r.Schedule===0)assert(s.Delay===0&&!s.Missing,'immediate schedule');
  for(let arm=0;arm<9;arm++){
   const p=s.P[arm],mask=s.Mask[arm];assert(Number.isFinite(p)&&p>0&&p<1,'probability');
   assert(Number.isInteger(mask)&&mask>=1&&mask<512&&(mask&1)===1&&s.Cost[arm]===pop(mask)&&s.Cost[arm]<=6,'cost');
   for(let scope=0;scope<3;scope++)assert([0,1,3,7].includes((mask>>(3*scope))&7),'prefix');
   const loss=(p-s.Q)**2+s.Q*(1-s.Q),accuracy=p>=.5?s.Q:1-s.Q;
   b[arm][0]+=loss/256;ac[arm][0]+=accuracy/256;if(t>=128){b[arm][1]+=loss/128;ac[arm][1]+=accuracy/128;}
   real[arm]+=(p-Number(s.Y))**2/256;
  }
  if(r.Schedule===0)assert(s.P[1]===s.P[2]&&s.Mask[1]===s.Mask[2]&&s.P[2]===s.P[3]&&s.Mask[2]===s.Mask[3],'immediate variant parity');
 }
 for(let arm=0;arm<9;arm++){
  const expected=r.Steps.map(s=>-s.Q*Math.log(s.P[arm])-(1-s.Q)*Math.log1p(-s.P[arm]));
  for(let seg=0;seg<2;seg++){const xs=expected.slice(seg?128:0);assert(Math.abs(xs.reduce((a,b)=>a+b,0)/xs.length-r.LogLoss[arm][seg])<1e-12,'expected log score');}
  const lr=r.Steps.reduce((a,s)=>a-(s.Y?Math.log(s.P[arm]):Math.log1p(-s.P[arm])),0)/256;
  assert(Math.abs(lr-r.RealizedLog[arm])<1e-12,'realized log score');
  assert(Math.abs(real[arm]-r.Realized[arm])<1e-12,'realized reconstruction');
  for(let seg=0;seg<2;seg++)assert(Math.abs(b[arm][seg]-r.Metrics[arm][seg].Brier)<1e-12&&Math.abs(ac[arm][seg]-r.Metrics[arm][seg].Accuracy)<1e-12,'metric reconstruction');
 }
 // Independent origin-order simulation, not the Go ring or its counters.
 const state=Array(256).fill(0),arrived=Array(256).fill(false);
 let next=0,version=-1;
 const stats=[0,1,2,3,4,5,6,7].map(()=>({Issued:0,Pending:0,Applied:0,BankOnly:0,Stale:0,Censored:0}));
 const drain=()=>{
  while(next<256&&[2,3].includes(state[next])){
   for(let arm=0;arm<8;arm++){
    const z=stats[arm];z.Pending--;
    if(state[next]===3)z.Censored++;
    else if(Math.floor(next/32)===version)z.Applied++;
    else if(arm===0)z.Stale++;else z.BankOnly++;
   }
   state[next]=4;next++;
  }
 };
 for(let clock=0;clock<288;clock++){
  for(let origin=0;origin<Math.min(clock,256);origin++){
   const s=r.Steps[origin];
   if(!arrived[origin]&&!s.Missing&&origin+s.Delay<=clock){assert(state[origin]===1,'late expired label');arrived[origin]=true;state[origin]=2;drain();}
  }
  if(clock>=32){for(let i=0;i<Math.min(clock-31,256);i++)if(state[i]===1)state[i]=3;drain();}
  if(clock>=256)continue;
  if(clock%32===0){
   version=clock/32;
   const available=Array.from({length:16},(_,i)=>i-16);
   for(let i=0;i<clock;i++)if(arrived[i])available.push(i);
   const want=[available.slice(-64),available.slice(-32)],fit=r.Fits[version];
   assert(fit.Clock===clock&&eq(fit.Origins,want),'as-of fit origin mismatch');
  }
  assert(state[clock]===0,'duplicate issue');state[clock]=1;
  for(const z of stats){z.Issued++;z.Pending++;assert(z.Pending<=64,'capacity');}
  const s=r.Steps[clock];
  if(!s.Missing&&s.Delay===0){arrived[clock]=true;state[clock]=2;drain();}
  assert(eq([stats[0].Applied,stats[1].Applied+stats[1].BankOnly,...Array(6).fill(arrived.filter(Boolean).length)],s.Selector),'selector update clocks');
  assert(eq(stats,s.Stats),`per-step journal reconstruction ${key}/${clock}`);
 }
 assert(next===256&&eq(stats,r.Final)&&r.Arrived===arrived.filter(Boolean).length,'final reconstruction');
 assert(stats[1].Applied+stats[1].BankOnly===r.Arrived,'role retention');
 assert(eq([stats[0].Applied,stats[1].Applied+stats[1].BankOnly,...Array(6).fill(r.Arrived)],r.FinalSelector),'final selector clocks');
}
assert(pairs.size===768,'independent trajectory count');
for(const [key,p] of pairs){assert(p.length===2&&eq(p[0].Rules,p[1].Rules),'paired rules');for(let t=0;t<256;t++)for(const k of ['X','Q','Y'])assert(p[0].Steps[t][k]===p[1].Steps[t][k],`paired latent ${key}`);}
function interval(xs){const mean=xs.reduce((s,x)=>s+x,0)/xs.length,se=Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/(xs.length-1)/xs.length);return {mean,lower:mean-3.5*se,upper:mean+3.5*se};}
const gates=[],groups=[];
for(const phase of ['design','confirmation'])for(const c of cases)for(const schedule of [0,1]){
 const rs=a.Records.filter(r=>r.Phase===phase&&r.Case===c&&r.Schedule===schedule);assert(rs.length===32,'group size');
 const final=[0,1,2,3,4,5,6,7].map(arm=>Object.fromEntries(Object.keys(rs[0].Final[arm]).map(k=>[k,rs.reduce((s,r)=>s+r.Final[arm][k],0)/32])));
 for(let seg=0;seg<2;seg++){
  const metrics=Array.from({length:9},(_,arm)=>({logLoss:rs.reduce((s,r)=>s+r.LogLoss[arm][seg],0)/32,brier:rs.reduce((s,r)=>s+r.Metrics[arm][seg].Brier,0)/32,accuracy:rs.reduce((s,r)=>s+r.Metrics[arm][seg].Accuracy,0)/32,cost:rs.reduce((s,r)=>s+r.Steps.slice(seg?128:0).reduce((v,t)=>v+t.Cost[arm],0)/(seg?128:256),0)/32}));
  groups.push({phase,case:c,schedule,segment:seg?'late':'all',metrics,final});
  for(const candidate of [7,8]){
   for(const control of [0,1,2,3,4,5,6]){const v=interval(rs.map(r=>r.Metrics[candidate][seg].Brier-r.Metrics[control][seg].Brier));gates.push({type:'nonharm',candidate,phase,case:c,schedule,segment:seg,control,...v,pass:v.upper<=.01});}
   const generic=seg===0&&['parity3','parity4','complement4'].includes(c)||seg===1&&c.includes('_to_');
   const switching=schedule===1&&seg===1&&c.includes('_to_');
   const carry=schedule===1&&seg===1&&(c==='parity4'||c.includes('_to_'));
   for(const control of [...(generic?[0]:[]),...(carry?[1]:[]),...(switching?[2,3,4,5,6]:[])]){const v=interval(rs.map(r=>r.Metrics[control][seg].Brier-r.Metrics[candidate][seg].Brier));gates.push({type:'gain',candidate,phase,case:c,schedule,segment:seg,control,...v,pass:v.mean>=.005&&v.lower>0});}
  }
 }
}
assert(gates.length===1436,'gate count');
const candidates=[7,8].map(candidate=>{const gs=gates.filter(g=>g.candidate===candidate);assert(gs.length===718,'per-candidate gates');return{candidate,passed:gs.filter(g=>g.pass).length,total:gs.length,status:gs.every(g=>g.pass)?'PASS':'FAIL'};});
process.stdout.write(JSON.stringify({version:'v111',artifactSHA256:crypto.createHash('sha256').update(raw).digest('hex'),trajectories:768,scheduleRuns:1536,steps:1536*256,uniqueEffectiveSeeds:seeds.size,sourceHashes:Object.keys(a.Hashes).length,passed:gates.filter(g=>g.pass).length,total:gates.length,status:candidates.some(c=>c.status==='PASS')?'CANDIDATE_PASS':'ALL_CANDIDATES_FAIL',candidates,gates,groups},null,2)+'\n');
