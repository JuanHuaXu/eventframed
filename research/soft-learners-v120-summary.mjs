import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import readline from 'node:readline';
import {checkSegmentReference} from './segment-v120-reference.mjs';
const root=path.resolve(import.meta.dirname,'..');
const assert=(x,m)=>{if(!x)throw Error(m)},eq=(x,y)=>JSON.stringify(x)===JSON.stringify(y),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const stream=fs.createReadStream(process.argv[2]),digest=crypto.createHash('sha256');
stream.on('data',chunk=>digest.update(chunk));
const lines=readline.createInterface({input:stream,crlfDelay:Infinity})[Symbol.asyncIterator]();
const first=await lines.next();assert(!first.done,'header');
const a=JSON.parse(first.value);a.Records=[];
assert(a.Version==='soft-learners-v120'&&a.TransferBase===2200112200&&a.BooleanBase===2204112300&&a.Workers===4&&a.Hazard===.01&&a.GenericMass===.95,'contract');
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
// Independent Simpson rule on a wider interval, checked against the dense
// component reference; neither calls the production quadrature nor uses its weights.
const normal=Array.from({length:513},(_,j)=>{const z=-12+j*24/512,w=(j===0||j===512)?1:(j%2?4:2);return {z,w:w*Math.exp(-z*z/2)/Math.sqrt(2*Math.PI)*24/512/3};});
function integral(m,v){const sd=Math.sqrt(v);let p=0;for(const {z,w}of normal)p+=w*(1+Math.tanh((m+sd*z)/2))/2;return p;}
const seen=new Set(),pairs=new Map(),seeds=new Set();let ridgeChecks=0,fitChecks=0,variationalChecks=0,variationalFitChecks=0,markovChecks=0,segmentWeightChecks=0,segmentReferenceFits=0,segmentReferenceForecasts=0;const segmentErrors={evidenceError:0,maxWeightError:0,maxForecastError:0,maxNoChangeError:0};
for await(const line of lines){
 const r=JSON.parse(line);
 const {Phase:ph,Case:c,Index:i,Schedule:sch}=r;
 assert([0,1].includes(ph)&&Number.isInteger(c)&&c>=0&&c<21&&Number.isInteger(i)&&i>=0&&i<32&&[0,1].includes(sch),'identity');
 const latent=[ph,c,i].join('/'),key=latent+'/'+sch;assert(!seen.has(key),'duplicate');seen.add(key);if(!pairs.has(latent))pairs.set(latent,[]);pairs.get(latent)[sch]=hash(JSON.stringify([r.Initial,r.Teacher,r.Rules,r.Steps.map(s=>[s.X,s.Y,s.Q])]));
 const base=c<9?a.TransferBase+ph*1e6+Math.floor(c/3)*1e5+(c%3)*1e4+i*10:a.BooleanBase+ph*1e6+(c-9)*1e4+i*10;
 for(let role=0;role<5;role++){assert(r.Seeds[role]===base+role,'seed');if(sch===0){const s=(base+role)%2147483647;assert(!seeds.has(s),'seed collision');seeds.add(s);}}
 if(c<9){const s=r.Teacher.Spec;assert(s.SeedBase===a.TransferBase&&s.Phase===ph&&s.Index===i&&s.Family===Math.floor(c/3)&&s.Mode===c%3&&eq(r.Seeds,r.Teacher.Seeds),'teacher identity');}else assert(r.Teacher===null,'unexpected teacher');
 assert(r.Initial.length===16&&r.Steps.length===256&&r.Fits.length===8,'record sizes');
 for(const s of r.Initial)assert(Number.isInteger(s.Bits)&&s.Bits>=0&&s.Bits<512&&typeof s.Outcome==='boolean','initial packet');
 for(const [fitIndex,fit] of r.Fits.entries()){
  assert(fit.Clock===fitIndex*32,'fit ordering');
  assert(fit.Clock>=0&&fit.Clock<256&&fit.Clock%32===0,'fit clock');
  const available=Array.from({length:16},(_,j)=>j-16);for(let j=0;j<fit.Clock;j++){const s=r.Steps[j];if(!s.Missing&&j+s.Delay<=fit.Clock)available.push(j);}
  assert(eq(fit.Origins,[available.slice(-64),available.slice(-32)]),'as-of fits');
  for(let window=0;window<2;window++){
   const b=fit.RidgeBeta[window],stats=fit.RidgeStats[window];assert(b.length===10&&b.every(Number.isFinite),'coefficients');
   let f=dot(b,b)/2;const g=b.slice();
   for(const origin of fit.Origins[window]){const s=origin<0?r.Initial[origin+16]:{Bits:r.Steps[origin].X,Outcome:r.Steps[origin].Y},x=features(s.Bits),z=dot(b,x),v=s.Outcome?-z:z;f+=Math.max(v,0)+Math.log1p(Math.exp(-Math.abs(v)));const residual=s.Outcome?-sigmoid(-z):sigmoid(z);for(let j=0;j<10;j++)g[j]+=residual*x[j];}
   const norm=Math.max(...g.map(Math.abs));assert(norm<=1.0001e-8&&Math.abs(norm-stats.GradientInf)<1e-11&&Math.abs(f-stats.Objective)<1e-10,'ridge stationarity/objective');
   assert(stats.Iterations>=0&&stats.Iterations<=32&&stats.Evaluations>=1&&stats.Evaluations<=769,'solver cap');fitChecks++;
   const mu=fit.VariationalMean[window],cov=fit.VariationalCov[window];
   assert(mu.length===10&&mu.every(Number.isFinite)&&cov.length===10&&cov.every(row=>row.length===10&&row.every(Number.isFinite)),'variational moments');
   assert(Number.isInteger(fit.VariationalIterations[window])&&fit.VariationalIterations[window]>=1&&fit.VariationalIterations[window]<=1024&&Number.isFinite(fit.VariationalResidual[window])&&fit.VariationalResidual[window]>=0&&fit.VariationalResidual[window]<=1.0001e-10,'variational convergence');
   const A=Array.from({length:10},(_,j)=>Array.from({length:10},(_,k)=>Number(j===k))),bb=Array(10).fill(0);let bound=0;
   for(const origin of fit.Origins[window]){
    const packet=origin<0?r.Initial[origin+16]:{Bits:r.Steps[origin].X,Outcome:r.Steps[origin].Y},x=features(packet.Bits),m=dot(mu,x),v=dot(x,cov.map(row=>dot(row,x)));
    assert(v>0&&v<=10+1e-10,'variational projected variance');
    const xi=Math.sqrt(v+m*m),lambda=Math.tanh(xi/2)/(4*xi);
    bound+=-Math.log1p(Math.exp(-xi))-xi/2+lambda*xi*xi;
    for(let j=0;j<10;j++){bb[j]+=(Number(packet.Outcome)-.5)*x[j];for(let k=0;k<10;k++)A[j][k]+=2*lambda*x[j]*x[k];}
   }
   const triangular=A.map(row=>row.slice());
   for(let j=0;j<10;j++){
    assert(triangular[j][j]>0,'precision SPD');bound-=Math.log(triangular[j][j])/2;
    for(let k=j+1;k<10;k++){const factor=triangular[k][j]/triangular[j][j];for(let h=j+1;h<10;h++)triangular[k][h]-=factor*triangular[j][h];}
    assert(Math.abs(dot(A[j],mu)-bb[j])<1e-7,'variational mean fixed point');
    for(let k=0;k<10;k++){assert(Math.abs(cov[j][k]-cov[k][j])<1e-12,'covariance symmetry');assert(Math.abs(dot(A[j],cov.map(row=>row[k]))-Number(j===k))<1e-7,'variational covariance inverse');}
   }
   bound+=dot(bb,mu)/2;
   assert(Number.isFinite(fit.VariationalBound[window])&&Math.abs(bound-fit.VariationalBound[window])<1e-8,'variational ELBO');variationalFitChecks++;
   const weights=fit.SegmentStarts[window];
   assert(weights.length===fit.Clock+16&&weights.every(x=>Number.isFinite(x)&&x>=0&&x<=1)&&Math.abs(weights.reduce((s,x)=>s+x,0)-1)<1e-10,'segment weights');
   assert(Number.isFinite(fit.SegmentEvidence[window])&&fit.SegmentEvidence[window]<=1e-10,'segment evidence');
   segmentWeightChecks++;
   if(i===0&&[0,128,224].includes(fit.Clock)){
    const e=checkSegmentReference(r,fit,window,a.Hazard,a.GenericMass);segmentReferenceFits++;segmentReferenceForecasts+=e.forecasts;
    for(const key of Object.keys(segmentErrors))segmentErrors[key]=Math.max(segmentErrors[key],e[key]);
   }
  }
 }
 const b=Array.from({length:15},()=>[0,0]),acc=Array.from({length:15},()=>[0,0]),ll=Array.from({length:15},()=>[0,0]),real=Array(15).fill(0),rl=Array(15).fill(0);
 for(const [t,s]of r.Steps.entries()){
  assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512&&typeof s.Y==='boolean'&&Math.abs(s.Q-truth(r,s.X,t))<1e-13,'truth');
  assert(Number.isInteger(s.Delay)&&s.Delay>=0&&s.Delay<=31&&typeof s.Missing==='boolean','schedule');if(sch===0)assert(s.Delay===0&&!s.Missing,'immediate');
  if(c===18)assert(((s.X>>8)&1)===((s.X>>1)&1),'dependent input');
  assert(s.P.length===15,'policy count');
  const priorW=[.95,.05/3,.05/3,.05/3];let w=priorW.slice();
  for(let j=0;j<t;j++){
   const old=r.Steps[j];
   if(!old.Missing&&j+old.Delay<=t){let total=0;for(let k=0;k<4;k++){w[k]*=old.Y?old.P[k]:1-old.P[k];total+=w[k];}assert(total>0,'Markov normalization');for(let k=0;k<4;k++)w[k]/=total;}
   for(let k=0;k<4;k++)w[k]=.999*w[k]+.001*priorW[k];
  }
  assert(Math.abs(s.P[12]-w.reduce((sum,v,k)=>sum+v*s.P[k],0))<1e-10,'Markov issued forecast');markovChecks++;
  for(let arm=0;arm<15;arm++){
   const p=s.P[arm];assert(Number.isFinite(p)&&p>0&&p<1,'forecast');
   if(arm===4||arm===5){const beta=r.Fits[Math.floor(t/32)].RidgeBeta[arm-4],expected=Math.max(1e-12,Math.min(1-1e-12,sigmoid(dot(beta,features(s.X)))));assert(Math.abs(p-expected)<1e-12,'ridge issued forecast');ridgeChecks++;}
   if(arm===8||arm===9){
    const fit=r.Fits[Math.floor(t/32)],x=features(s.X),mu=fit.VariationalMean[arm-8],cov=fit.VariationalCov[arm-8],m=dot(mu,x),v=dot(x,cov.map(row=>dot(row,x)));
    assert(v>=0&&v<=10+1e-10,'issued variance');
    const expected=Math.max(1e-12,Math.min(1-1e-12,integral(m,v)));
    assert(Math.abs(p-expected)<1e-10,'variational issued forecast');variationalChecks++;
   }
   const loss=(p-s.Q)**2+s.Q*(1-s.Q),ac=p>=.5?s.Q:1-s.Q,log=-s.Q*Math.log(p)-(1-s.Q)*Math.log1p(-p);
   for(const seg of t>=192?[0,1]:[0]){const n=seg?64:256;b[arm][seg]+=loss/n;acc[arm][seg]+=ac/n;ll[arm][seg]+=log/n;}
   real[arm]+=(p-Number(s.Y))**2/256;rl[arm]-=(s.Y?Math.log(p):Math.log1p(-p))/256;
  }
 }
 for(let arm=0;arm<15;arm++){assert(Math.abs(real[arm]-r.Realized[arm])<1e-12&&Math.abs(rl[arm]-r.RealizedLog[arm])<1e-12,'realized scores');for(let seg=0;seg<2;seg++)assert(Math.abs(b[arm][seg]-r.Metrics[arm][seg].Brier)<1e-12&&Math.abs(acc[arm][seg]-r.Metrics[arm][seg].Accuracy)<1e-12&&Math.abs(ll[arm][seg]-r.LogLoss[arm][seg])<1e-12,'expected scores');}
 assert(r.Arrived===r.Steps.filter(s=>!s.Missing).length,'final feedback');
 a.Records.push({Phase:ph,Case:c,Index:i,Schedule:sch,Metrics:r.Metrics,LogLoss:r.LogLoss,Steps:r.Steps.map(s=>({Q:s.Q}))});
}
assert(a.Records.length===2688&&markovChecks===2688*256&&segmentWeightChecks===2688*8*2&&segmentReferenceFits===504&&segmentReferenceForecasts===16128,'new coverage');
assert(pairs.size===1344&&seeds.size===6720&&ridgeChecks===2688*256*2&&fitChecks===2688*8*2&&variationalChecks===ridgeChecks&&variationalFitChecks===fitChecks,'coverage');
for(const p of pairs.values())assert(p[0]===p[1],'paired generator');
const gates=[],groups=[];
for(const ph of [0,1])for(let c=0;c<21;c++)for(const sch of [0,1]){
 const rs=a.Records.filter(r=>r.Phase===ph&&r.Case===c&&r.Schedule===sch);assert(rs.length===32,'cell count');
 for(const seg of [0,1]){
  const meta={phase:['design','confirmation'][ph],case:cases[c],schedule:sch,segment:seg?'terminal64':'all256'};
  const metrics=Array.from({length:15},(_,arm)=>({brier:mean(rs.map(r=>r.Metrics[arm][seg].Brier)),accuracy:mean(rs.map(r=>r.Metrics[arm][seg].Accuracy)),logLoss:mean(rs.map(r=>r.LogLoss[arm][seg]))}));
  groups.push({...meta,metrics,bayesBrierFloor:mean(rs.map(r=>mean(r.Steps.slice(seg?192:0).map(s=>s.Q*(1-s.Q))))),bayesAccuracyCeiling:mean(rs.map(r=>mean(r.Steps.slice(seg?192:0).map(s=>Math.max(s.Q,1-s.Q)))))});
  for(const candidate of [4,5,6,7,8,9]){
   const control=(candidate%2)*2,targetBase=(candidate===6||candidate===7)?3:0;
   for(const ctrl of [control,control+1]){const v=interval(rs.map(r=>r.Metrics[candidate][seg].Brier-r.Metrics[ctrl][seg].Brier));gates.push({...meta,type:'nonharm',candidate,control:ctrl,...v,pass:v.upper<=.01});}
   const changed=(c<9&&c%3!==0)||c>=19;
   if(candidate>=8){
    const ctrl=4+candidate%2,v=interval(rs.map(r=>r.Metrics[candidate][seg].Brier-r.Metrics[ctrl][seg].Brier));
    gates.push({...meta,type:'map_nonharm',candidate,control:ctrl,...v,pass:v.upper<=.01});
    if((seg===1&&changed)||(seg===0&&(c===0||c===17))){
     const g=interval(rs.map(r=>r.Metrics[ctrl][seg].Brier-r.Metrics[candidate][seg].Brier));
     gates.push({...meta,type:'map_gain',candidate,control:ctrl,...g,pass:g.mean>=.005&&g.lower>0});
    }
   }
   if((seg===1&&changed)||(seg===0&&c===targetBase)){
    const v=interval(rs.map(r=>r.Metrics[control][seg].Brier-r.Metrics[candidate][seg].Brier));
    gates.push({...meta,type:'gain',candidate,control,structuralTarget:c>=targetBase&&c<=targetBase+2,...v,pass:v.mean>=.005&&v.lower>0});
   }
  }
 }
}

assert(gates.length===2648,'retained gates');
for(const ph of [0,1])for(let c=0;c<21;c++)for(const sch of [0,1]){
 const rs=a.Records.filter(r=>r.Phase===ph&&r.Case===c&&r.Schedule===sch);
 for(const seg of [0,1])for(const candidate of [10,11]){
  const meta={phase:['design','confirmation'][ph],case:cases[c],schedule:sch,segment:seg?'terminal64':'all256'};
  const generic=2*(candidate-10),noChange=13+candidate-10;
  for(const ctrl of [generic,generic+1,12,noChange]){
   const v=interval(rs.map(r=>r.Metrics[candidate][seg].Brier-r.Metrics[ctrl][seg].Brier));
   gates.push({...meta,type:'segment_nonharm',candidate,control:ctrl,...v,pass:v.upper<=.01});
  }
  const changed=(c<9&&c%3!==0)||c>=19;
  if(seg===1&&changed)for(const ctrl of [generic,12,noChange]){
   const v=interval(rs.map(r=>r.Metrics[ctrl][seg].Brier-r.Metrics[candidate][seg].Brier));
   gates.push({...meta,type:'segment_gain',candidate,control:ctrl,...v,pass:v.mean>=.005&&v.lower>0});
  }
 }
}
assert(gates.length===4184&&gates.filter(g=>g.type.endsWith('nonharm')).length===3696&&gates.filter(g=>g.type.endsWith('gain')).length===488,'total gates');
const candidates=[4,5,6,7,8,9,10,11].map(candidate=>{
 const gs=gates.filter(g=>g.candidate===candidate),expected=candidate>=10?768:(candidate>=8?580:372);
 assert(gs.length===expected,'candidate scope');
 const byType={};for(const type of new Set(gs.map(g=>g.type))){const sub=gs.filter(g=>g.type===type);byType[type]={passed:sub.filter(g=>g.pass).length,total:sub.length};}
 return {candidate,passed:gs.filter(g=>g.pass).length,total:expected,status:gs.every(g=>g.pass)?'PASS':'FAIL',byType};
});
process.stdout.write(JSON.stringify({version:a.Version,artifactSHA256:digest.digest('hex'),sourceHashes:Object.keys(a.Hashes).length,latentTrajectories:1344,scheduleRuns:2688,steps:2688*256,uniqueEffectiveSeeds:seeds.size,ridgeForecastChecks:ridgeChecks,ridgeFitChecks:fitChecks,variationalForecastChecks:variationalChecks,variationalFitChecks,markovForecastChecks:markovChecks,segmentWeightChecks,segmentReferenceFits,segmentReferenceForecasts,noChangeReferenceForecasts:segmentReferenceForecasts,segmentErrors,passed:gates.filter(g=>g.pass).length,total:gates.length,status:candidates.filter(c=>c.candidate>=10).some(c=>c.status==='PASS')?'SEGMENT_CANDIDATE_PASS':'SEGMENT_CANDIDATES_FAIL',candidates,gates,groups},null,2)+'\n');

