import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const raw=fs.readFileSync(process.argv[2]),a=JSON.parse(raw),root=path.resolve(import.meta.dirname,'..');
const assert=(x,m)=>{if(!x)throw Error(m)},eq=(x,y)=>JSON.stringify(x)===JSON.stringify(y),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
assert(a.Version==='soft-learners-v118'&&a.TransferBase===2184111800&&a.BooleanBase===2188111900&&a.Records.length===2688,'contract');
for(const [f,h] of Object.entries(a.Hashes))assert(hash(fs.readFileSync(path.join(root,f)))===h,`source ${f}`);
const cases=['additive_stationary','additive_abrupt','additive_gradual','hierarchy_stationary','hierarchy_abrupt','hierarchy_gradual','local_table_stationary','local_table_abrupt','local_table_gradual','parity1','parity2','parity3','parity4','complement4','majority3','mux3','constant','null','dependent4','majority_to_parity','parity_to_majority'];
const pop=x=>x.toString(2).replaceAll('0','').length;
function endpoint(p,f,x){
 const bit=c=>(x>>c)&1;
 if(f===0){let z=p.Bias;for(let j=0;j<5;j++)z+=p.Coefficients[j]*(bit(p.Coordinates[j])?1:-1);return .5+.4*Math.tanh(z/2);}
 if(f===1){const a=bit(p.Branches[0]),b=bit(p.Branches[1+a]),c=bit(p.Branches[3+2*a+b]);return p.Probabilities[4*a+2*b+c];}
 let k=0;for(let j=0;j<4;j++)k+=bit(p.Coordinates[j])*(2**j);return p.Probabilities[k];
}
function truth(r,x,t){
 if(r.Case<9){const d=r.Teacher;let w=0;if(d.Spec.Mode!==0&&t>=d.Change)w=d.Spec.Mode===1?1:Math.min(1,(t-d.Change+1)/32);return (1-w)*endpoint(d.Models[0],d.Spec.Family,x)+w*endpoint(d.Models[1],d.Spec.Family,x);}
 let name=cases[r.Case],rule=r.Rules[0];
 if(r.Case>=19){name=r.Case===19?'majority3':'parity4';if(t>=128){rule=r.Rules[1];name=r.Case===19?'parity4':'majority3';}}
 if(name==='null')return .5;
 let yes=pop(x&rule)%2===1;
 if(name==='constant')yes=false;
 if(name==='complement4')yes=!yes;
 if(name==='majority3')yes=pop(x&rule)>=2;
 if(name==='mux3'){const bs=[];for(let j=0;j<9;j++)if(rule&(1<<j))bs.push(Boolean(x&(1<<j)));assert(bs.length===3,'mux arity');yes=bs[0]?bs[1]:bs[2];}
 return yes?.95:.05;
}
const sigmoid=z=>z>=0?1/(1+Math.exp(-z)):Math.exp(z)/(1+Math.exp(z));
const features=x=>[1,...Array.from({length:9},(_,j)=>((x>>j)&1)?1:-1)];
const dot=(x,y)=>x.reduce((s,v,j)=>s+v*y[j],0);
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
function interval(xs){assert(xs.length===32,'paired n');const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
const seen=new Set(),pairs=new Map(),seeds=new Set();let ridgeChecks=0,fitChecks=0;
for(const r of a.Records){
 const {Phase:ph,Case:c,Index:i,Schedule:sch}=r;
 assert([0,1].includes(ph)&&Number.isInteger(c)&&c>=0&&c<21&&Number.isInteger(i)&&i>=0&&i<32&&[0,1].includes(sch),'identity');
 const latent=[ph,c,i].join('/'),key=latent+'/'+sch;assert(!seen.has(key),'duplicate');seen.add(key);if(!pairs.has(latent))pairs.set(latent,[]);pairs.get(latent)[sch]=r;
 const base=c<9?a.TransferBase+ph*1e6+Math.floor(c/3)*1e5+(c%3)*1e4+i*10:a.BooleanBase+ph*1e6+(c-9)*1e4+i*10;
 for(let role=0;role<5;role++){assert(r.Seeds[role]===base+role,'seed');if(sch===0){const s=(base+role)%2147483647;assert(!seeds.has(s),'seed collision');seeds.add(s);}}
 if(c<9){const s=r.Teacher.Spec;assert(s.SeedBase===a.TransferBase&&s.Phase===ph&&s.Index===i&&s.Family===Math.floor(c/3)&&s.Mode===c%3&&eq(r.Seeds,r.Teacher.Seeds),'teacher identity');}else assert(r.Teacher===null,'unexpected teacher');
 assert(r.Initial.length===16&&r.Steps.length===256&&r.Fits.length===8,'record sizes');
 for(const s of r.Initial)assert(Number.isInteger(s.Bits)&&s.Bits>=0&&s.Bits<512&&typeof s.Outcome==='boolean','initial packet');
 for(const fit of r.Fits){
  assert(fit.Clock>=0&&fit.Clock<256&&fit.Clock%32===0,'fit clock');
  const available=Array.from({length:16},(_,j)=>j-16);for(let j=0;j<fit.Clock;j++){const s=r.Steps[j];if(!s.Missing&&j+s.Delay<=fit.Clock)available.push(j);}
  assert(eq(fit.Origins,[available.slice(-64),available.slice(-32)]),'as-of fits');
  for(let window=0;window<2;window++){
   const b=fit.RidgeBeta[window],stats=fit.RidgeStats[window];assert(b.length===10&&b.every(Number.isFinite),'coefficients');
   let f=dot(b,b)/2;const g=b.slice();
   for(const origin of fit.Origins[window]){const s=origin<0?r.Initial[origin+16]:{Bits:r.Steps[origin].X,Outcome:r.Steps[origin].Y},x=features(s.Bits),z=dot(b,x),v=s.Outcome?-z:z;f+=Math.max(v,0)+Math.log1p(Math.exp(-Math.abs(v)));const residual=s.Outcome?-sigmoid(-z):sigmoid(z);for(let j=0;j<10;j++)g[j]+=residual*x[j];}
   const norm=Math.max(...g.map(Math.abs));assert(norm<=1.0001e-8&&Math.abs(norm-stats.GradientInf)<1e-11&&Math.abs(f-stats.Objective)<1e-10,'ridge stationarity/objective');
   assert(stats.Iterations>=0&&stats.Iterations<=32&&stats.Evaluations>=1&&stats.Evaluations<=769,'solver cap');fitChecks++;
  }
 }
 const b=Array.from({length:8},()=>[0,0]),acc=Array.from({length:8},()=>[0,0]),ll=Array.from({length:8},()=>[0,0]),real=Array(8).fill(0),rl=Array(8).fill(0);
 for(const [t,s]of r.Steps.entries()){
  assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512&&typeof s.Y==='boolean'&&Math.abs(s.Q-truth(r,s.X,t))<1e-13,'truth');
  assert(Number.isInteger(s.Delay)&&s.Delay>=0&&s.Delay<=31&&typeof s.Missing==='boolean','schedule');if(sch===0)assert(s.Delay===0&&!s.Missing,'immediate');
  if(c===18)assert(((s.X>>8)&1)===((s.X>>1)&1),'dependent input');
  assert(s.P.length===8,'policy count');
  for(let arm=0;arm<8;arm++){
   const p=s.P[arm];assert(Number.isFinite(p)&&p>0&&p<1,'forecast');
   if(arm===4||arm===5){const beta=r.Fits[Math.floor(t/32)].RidgeBeta[arm-4],expected=Math.max(1e-12,Math.min(1-1e-12,sigmoid(dot(beta,features(s.X)))));assert(Math.abs(p-expected)<1e-12,'ridge issued forecast');ridgeChecks++;}
   const loss=(p-s.Q)**2+s.Q*(1-s.Q),ac=p>=.5?s.Q:1-s.Q,log=-s.Q*Math.log(p)-(1-s.Q)*Math.log1p(-p);
   for(const seg of t>=192?[0,1]:[0]){const n=seg?64:256;b[arm][seg]+=loss/n;acc[arm][seg]+=ac/n;ll[arm][seg]+=log/n;}
   real[arm]+=(p-Number(s.Y))**2/256;rl[arm]-=(s.Y?Math.log(p):Math.log1p(-p))/256;
  }
 }
 for(let arm=0;arm<8;arm++){assert(Math.abs(real[arm]-r.Realized[arm])<1e-12&&Math.abs(rl[arm]-r.RealizedLog[arm])<1e-12,'realized scores');for(let seg=0;seg<2;seg++)assert(Math.abs(b[arm][seg]-r.Metrics[arm][seg].Brier)<1e-12&&Math.abs(acc[arm][seg]-r.Metrics[arm][seg].Accuracy)<1e-12&&Math.abs(ll[arm][seg]-r.LogLoss[arm][seg])<1e-12,'expected scores');}
 assert(r.Arrived===r.Steps.filter(s=>!s.Missing).length,'final feedback');
}
assert(pairs.size===1344&&seeds.size===6720&&ridgeChecks===2688*256*2&&fitChecks===2688*8*2,'coverage');
for(const p of pairs.values()){assert(eq(p[0].Initial,p[1].Initial)&&eq(p[0].Teacher,p[1].Teacher)&&eq(p[0].Rules,p[1].Rules),'paired generator');for(let t=0;t<256;t++)for(const key of ['X','Y','Q'])assert(p[0].Steps[t][key]===p[1].Steps[t][key],'paired packets');}
const gates=[],groups=[];
for(const ph of [0,1])for(let c=0;c<21;c++)for(const sch of [0,1]){
 const rs=a.Records.filter(r=>r.Phase===ph&&r.Case===c&&r.Schedule===sch);assert(rs.length===32,'cell count');
 for(const seg of [0,1]){
  const meta={phase:['design','confirmation'][ph],case:cases[c],schedule:sch,segment:seg?'terminal64':'all256'};
  const metrics=Array.from({length:8},(_,arm)=>({brier:mean(rs.map(r=>r.Metrics[arm][seg].Brier)),accuracy:mean(rs.map(r=>r.Metrics[arm][seg].Accuracy)),logLoss:mean(rs.map(r=>r.LogLoss[arm][seg]))}));
  groups.push({...meta,metrics,bayesBrierFloor:mean(rs.map(r=>mean(r.Steps.slice(seg?192:0).map(s=>s.Q*(1-s.Q))))),bayesAccuracyCeiling:mean(rs.map(r=>mean(r.Steps.slice(seg?192:0).map(s=>Math.max(s.Q,1-s.Q)))))});
  for(const candidate of [4,5,6,7]){
   const control=(candidate%2)*2,targetBase=candidate<6?0:3;
   for(const ctrl of [control,control+1]){const v=interval(rs.map(r=>r.Metrics[candidate][seg].Brier-r.Metrics[ctrl][seg].Brier));gates.push({...meta,type:'nonharm',candidate,control:ctrl,...v,pass:v.upper<=.01});}
   const changed=(c<9&&c%3!==0)||c>=19;
   if((seg===1&&changed)||(seg===0&&c===targetBase)){
    const v=interval(rs.map(r=>r.Metrics[control][seg].Brier-r.Metrics[candidate][seg].Brier));
    gates.push({...meta,type:'gain',candidate,control,structuralTarget:c>=targetBase&&c<=targetBase+2,...v,pass:v.mean>=.005&&v.lower>0});
   }
  }
 }
}
assert(gates.length===1488&&gates.filter(g=>g.type==='nonharm').length===1344&&gates.filter(g=>g.type==='gain').length===144,'gate count');
const candidates=[4,5,6,7].map(candidate=>{const gs=gates.filter(g=>g.candidate===candidate),target=gs.filter(g=>g.structuralTarget);assert(gs.length===372&&target.length===12,'per-candidate scope');return {candidate,passed:gs.filter(g=>g.pass).length,total:372,status:gs.every(g=>g.pass)?'PASS':'FAIL',structuralTargetPassed:target.filter(g=>g.pass).length,structuralTargetTotal:12};});
process.stdout.write(JSON.stringify({version:a.Version,artifactSHA256:hash(raw),sourceHashes:Object.keys(a.Hashes).length,latentTrajectories:1344,scheduleRuns:2688,steps:2688*256,uniqueEffectiveSeeds:seeds.size,ridgeForecastChecks:ridgeChecks,ridgeFitChecks:fitChecks,passed:gates.filter(g=>g.pass).length,total:gates.length,status:candidates.some(c=>c.status==='PASS')?'STANDALONE_CANDIDATE_PASS':'ALL_STANDALONE_CANDIDATES_FAIL',candidates,gates,groups},null,2)+'\n');
