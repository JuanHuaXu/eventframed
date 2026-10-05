// Small finite shared-latent proposal, NOT a quality/performance experiment.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/joint-shared-v68-preflight';assert(!fs.existsSync(root));
const self='research/joint-shared-v68-preflight.mjs',source=fs.readFileSync(self),sha=b=>crypto.createHash('sha256').update(b).digest('hex');
const base=[.25,.47,.7,.925],eta=[0,.1,.2],noiseWeight=[.8,.1,.1],hazard=1/16;
const sigmoid=x=>1/(1+Math.exp(-x)),grid=[-4,0,4],gridWeight=[.3,.4,.3],familyWeight=[.08,.06,.06];
const load=(family,b)=>family===0?1:family===1?2*b-1:2*(2*b-1)**2-1;
const atoms=[base.slice()],weights=[.8];
for(let f=0;f<3;f++){
 const intercept=base.map(b=>{let lo=-64,hi=64;for(let n=0;n<100;n++){const c=(lo+hi)/2,mean=grid.reduce((s,x,j)=>s+gridWeight[j]*sigmoid(c+load(f,b)*x),0);if(mean<b)lo=c;else hi=c}return(lo+hi)/2});
 for(let j=0;j<3;j++){atoms.push(base.map((b,i)=>sigmoid(intercept[i]+load(f,b)*grid[j])));weights.push(familyWeight[f]*gridWeight[j])}
}
const states=[];for(let h=0;h<3;h++)for(let z=0;z<atoms.length;z++)states.push({h,z,p:noiseWeight[h]*weights[z]});
const normalize=p=>{const sum=p.reduce((a,b)=>a+b,0);assert(sum>0&&Number.isFinite(sum));return p.map(x=>x/sum)};
let maxDifference=0,checks=0;
const near=(a,b)=>{assert(Number.isFinite(a)&&Number.isFinite(b));maxDifference=Math.max(maxDifference,Math.abs(a-b));checks++;assert(Math.abs(a-b)<2e-12,`${a} != ${b}`)};
near(weights.reduce((a,b)=>a+b,0),1);
for(let i=0;i<base.length;i++)near(atoms.reduce((s,a,z)=>s+weights[z]*a[i],0),base[i]);
const bit=(x,i)=>(x>>i)&1;
const first=(state,i,w)=>{const q=eta[state.h]+(1-2*eta[state.h])*atoms[state.z][i];return w?q:1-q};
const pair=(state,i,a,b)=>{const q=atoms[state.z][i],e=eta[state.h];return q*(a?1-e:e)*(b?1-e:e)+(1-q)*(a?e:1-e)*(b?e:1-e)};
const efficient=(w,member,second)=>normalize(states.map(s=>{let p=s.p;for(let i=0;i<4;i++)p*=i===member?pair(s,i,bit(w,i),second):first(s,i,bit(w,i));return p}));
const ordinary=w=>normalize(states.map(s=>s.p*base.reduce((p,_,i)=>p*first(s,i,bit(w,i)),1)));
// Independent joint enumeration sums all 16 ORIGINAL Y vectors. The optional
// W2 uses the same Y coordinate, never a newly sampled outcome.
const enumerate=(w,member,second)=>{
 const mass=states.map(s=>{let sum=0;for(let y=0;y<16;y++){let p=s.p;for(let i=0;i<4;i++){const yi=bit(y,i),q=atoms[s.z][i],e=eta[s.h];p*=yi?q:1-q;p*=yi===bit(w,i)?1-e:e;if(i===member)p*=yi===second?1-e:e}sum+=p}return sum});
 return{evidence:mass.reduce((a,b)=>a+b,0),posterior:normalize(mass)};
};
const future=p=>base.map((b,i)=>hazard*b+(1-hazard)*states.reduce((s,x,k)=>s+p[k]*atoms[x.z][i],0));
const denseFuture=p=>{
 const next=states.map(()=>0);
 for(let from=0;from<states.length;from++)for(let to=0;to<states.length;to++)if(states[from].h===states[to].h)next[to]+=p[from]*((from===to?1-hazard:0)+hazard*weights[states[to].z]);
 return base.map((_,i)=>states.reduce((s,x,k)=>s+next[k]*atoms[x.z][i],0));
};
fs.mkdirSync(root,{mode:0o700});fs.writeFileSync(root+'/source.mjs',source,{flag:'wx',mode:0o600});
const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('freeze.json',{source:self,sourceSHA256:sha(source),base,eta,noiseWeight,hazard,grid,gridWeight,familyWeight,atoms,weights,states:states.length,stage:'finite toy mathematical preflight only; no consumed/fresh/private labels or production',goals:Array(7).fill('OPEN')});
const begin=performance.now(),cases=[];
try{
 for(let w=0;w<16;w++){
  const before=ordinary(w),enumBase=enumerate(w,-1,0),q0=future(before),dense0=denseFuture(enumBase.posterior);
  for(let s=0;s<states.length;s++)near(before[s],enumBase.posterior[s]);for(let i=0;i<4;i++)near(q0[i],dense0[i]);
  for(let member=0;member<4;member++){
   const yes=efficient(w,member,1),no=efficient(w,member,0),ey=enumerate(w,member,1),en=enumerate(w,member,0);
   let p=0;for(let s=0;s<states.length;s++)p+=before[s]*pair(states[s],member,bit(w,member),1)/first(states[s],member,bit(w,member));
   near(p,ey.evidence/enumBase.evidence);near(1-p,en.evidence/enumBase.evidence);
   for(let s=0;s<states.length;s++){near(yes[s],ey.posterior[s]);near(no[s],en.posterior[s])}
   const q1=future(yes),qn=future(no),d1=denseFuture(ey.posterior),dn=denseFuture(en.posterior);let value=0,local=0;
   for(let i=0;i<4;i++){near(q1[i],d1[i]);near(qn[i],dn[i]);near(p*q1[i]+(1-p)*qn[i],q0[i]);const v=(p*(q1[i]-q0[i])**2+(1-p)*(qn[i]-q0[i])**2)/4;value+=v;if(i===member)local=v}
   cases.push({firstPattern:w,member,secondProbability:p,globalPredictionValue:value,localOnlyShortcut:local,omittedNonlocalValue:value-local});
  }
 }
 const maxNonlocal=Math.max(...cases.map(x=>x.omittedNonlocalValue));assert(maxNonlocal>1e-5,'vacuous sharing');
 const result={stage:'mathematical component only, NOT integrated controller or empirical rescue',checks,maxDifference,contexts:16,queries:64,latentStates:states.length,originalYPathsPerState:16,maxNonlocalValue:maxNonlocal,cases,baselineMeanPreserved:true,towerIdentityPassed:true,efficientVersusEnumerationPassed:true,allSevenGoalsValidated:false,goals:Array(7).fill('OPEN')};
 save('results.json',result);save('completed.json',{exitCode:0,sourceSHA256:sha(source),wallMS:performance.now()-begin,allJobsTerminal:true,goal:'ACTIVE',goals:Array(7).fill('OPEN')});console.log(JSON.stringify({...result,cases:undefined},null,2));
}catch(e){save('failure.json',{error:e.message,checks,maxDifference,sourceSHA256:sha(source),allJobsTerminal:true,goal:'ACTIVE'});throw e}
