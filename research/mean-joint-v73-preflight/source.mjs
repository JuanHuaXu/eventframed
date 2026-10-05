// Finite joint-law preflight only: no empirical outcomes or implementation claim.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/mean-joint-v73-preflight',source='research/mean-joint-v73-preflight.mjs';
const sha=p=>crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
assert(!fs.existsSync(root));
const parent='research/checkpoint-2026-10-04-dynvariance-v71/manifest.json';
assert.equal(sha(parent),'9d6cb6dcb7fc8281b407ebc9cb5035c27a2e8b29bcf9230013517e2da9a2b297');
const protectedFiles=JSON.parse(fs.readFileSync(parent)).trackedHashes;
for(const[p,h]of Object.entries(protectedFiles))assert.equal(sha(p),h,p);
fs.mkdirSync(root,{mode:0o700});fs.copyFileSync(source,root+'/source.mjs',fs.constants.COPYFILE_EXCL);
const freeze={time:new Date().toISOString(),source,sourceSHA256:sha(source),meanFamilySource:'internal/researchdispersion/model.go',meanFamilySourceSHA256:sha('internal/researchdispersion/model.go'),parent,protectedFiles,means:27,families:3,noiseStates:3,urnSize:20,hazards:[0,1/16,1],noiseModes:['common','independent'],scope:'complete finite maths over fixed public toy parameters; no consumed-world or reserved outcome collection'};
fs.writeFileSync(root+'/freeze.json',JSON.stringify(freeze,null,2)+'\n',{flag:'wx',mode:0o600});
let checks=0,maxDefect=0;
const close=(a,b,label)=>{assert(Number.isFinite(a)&&Number.isFinite(b),label);const d=Math.abs(a-b);maxDefect=Math.max(maxDefect,d);checks++;assert(d<2e-11,`${label}: ${a} vs ${b}`)};
const hp=[.8,.1,.1],eta=[0,.1,.2],mp=[.1,.8,...Array(25).fill(.004)];
function means(base){return base.map((b,i)=>{const out=[b,(.9*b-.05)/.8];for(const a of [.1,.3,.5,.7,.9])for(const c of [-.8,-.4,0,.4,.8])out.push(Math.max(.02,Math.min(.98,a+c*i/(base.length-1))));return out})}
function prior(p,a,dense){
 const out=Array(22).fill(0);if(a===0){out[0]=1;return out}
 const strength=a===1?1:2,spike=a===1?.8:0;out[0]=spike;
 let w;
 if(dense){w=[1];for(let n=0;n<20;n++){const v=Array(n+2).fill(0);for(let j=0;j<=n;j++){v[j+1]+=w[j]*(strength*p+j)/(strength+n);v[j]+=w[j]*(strength*(1-p)+n-j)/(strength+n)}w=v}}
 else {w=Array(21).fill(0);w[0]=1;for(let n=0;n<20;n++)w[0]*=(strength*(1-p)+n)/(strength+n);for(let j=0;j<20;j++)w[j+1]=w[j]*(20-j)/(j+1)*(strength*p+j)/(strength*(1-p)+19-j);const sum=w.reduce((s,x)=>s+x,0);w=w.map(x=>x/sum)}
 for(let j=0;j<21;j++)out[j+1]=(1-spike)*w[j];return out;
}
const rate=(p,z)=>z===0?p:(z-1)/20;
function first(p,e,y){const q=e+(1-2*e)*p;return y?q:1-q}
function pair(p,e,a,b){const u=a?1-e:e,v=b?1-e:e;return p*u*v+(1-p)*(1-u)*(1-v)}
// The independent likelihood explicitly sums the SAME latent Y for a pair.
function latent(p,e,a,b){let s=0;for(const y of [false,true]){let q=y?p:1-p;q*=y===a?1-e:e;if(b!==null)q*=y===b?1-e:e;s+=q}return s}
function filtered(p,a,h,values,pairValue,hazard){
 const pi=prior(p,a,false),e=eta[h];let state=[...pi],evidence=1;
 for(let n=0;n<values.length;n++){
  if(n>0){const sum=state.reduce((s,x)=>s+x,0);state=state.map((x,z)=>(1-hazard)*x+hazard*sum*pi[z])}
  state=state.map((x,z)=>x*(n===0&&pairValue!==null?pair(rate(p,z),e,values[n],pairValue):first(rate(p,z),e,values[n])));
  const sum=state.reduce((s,x)=>s+x,0);if(sum===0)return{z:0,q:0};evidence*=sum;state=state.map(x=>x/sum);
 }
 const sum=state.reduce((s,x)=>s+x,0);state=state.map((x,z)=>(1-hazard)*x+hazard*sum*pi[z]);const total=state.reduce((s,x)=>s+x,0);
 return{z:evidence,q:state.reduce((s,x,z)=>s+x*rate(p,z),0)/total};
}
function paths(p,a,h,values,pairValue,hazard){
 const pi=prior(p,a,true),e=eta[h];let z=0,numerator=0;
 for(let from=0;from<22;from++){
  if(pi[from]===0)continue;
  const w=pi[from]*latent(rate(p,from),e,values[0],pairValue);
  if(values.length===1){z+=w;numerator+=w*((1-hazard)*rate(p,from)+hazard*p);continue}
  for(let to=0;to<22;to++){
   const transition=hazard*pi[to]+(from===to?1-hazard:0);
   const x=w*transition*latent(rate(p,to),e,values[1],null);
   z+=x;numerator+=x*((1-hazard)*rate(p,to)+hazard*p);
  }
 }
 return{z,q:z?numerator/z:0};
}
function infer(base,values,pairValue,hazard,mode,dense){
 const mu=means(base),out={z:0,clean:[0,0],observed:[0,0],classes:Array(243).fill(0),family:[0,0,0],mean:Array(27).fill(0),meanFamily:Array(81).fill(0)};
 const evaluate=dense?paths:filtered;
 for(let t=0;t<27;t++)for(let a=0;a<3;a++){
  const rows=base.map((_,i)=>eta.map((_,h)=>evaluate(mu[i][t],a,h,values[i],i===0?pairValue:null,hazard)));
  if(!dense&&mode==='independent'){
   const localZ=rows.map(row=>row.reduce((s,x,h)=>s+hp[h]*x.z,0));
   const weight=mp[t]/3*localZ[0]*localZ[1];out.z+=weight;
   for(let i=0;i<2;i++){
    if(localZ[i]===0)continue;
    out.clean[i]+=weight*rows[i].reduce((s,x,h)=>s+hp[h]*x.z*x.q,0)/localZ[i];
    out.observed[i]+=weight*rows[i].reduce((s,x,h)=>s+hp[h]*x.z*(eta[h]+(1-2*eta[h])*x.q),0)/localZ[i];
   }
   for(let h=0;h<3;h++)out.classes[(t*3+a)*3+h]+=mp[t]/3*hp[h]*rows[0][h].z*localZ[1];
  }else{
   // Direct noise-index joint summation, separate from marginal-first inference.
   for(let h0=0;h0<3;h0++)for(let h1=0;h1<3;h1++){
    if(mode==='common'&&h0!==h1)continue;
    const weight=mp[t]/3*hp[h0]*(mode==='common'?1:hp[h1])*rows[0][h0].z*rows[1][h1].z;
    out.z+=weight;out.classes[(t*3+a)*3+h0]+=weight;
    for(let i=0;i<2;i++){const h=i===0?h0:h1;out.clean[i]+=weight*rows[i][h].q;out.observed[i]+=weight*(eta[h]+(1-2*eta[h])*rows[i][h].q)}
   }
  }
 }
 assert(out.z>0);
 out.classes=out.classes.map(x=>x/out.z);out.clean=out.clean.map(x=>x/out.z);out.observed=out.observed.map(x=>x/out.z);
 for(let k=0;k<243;k++){const t=Math.floor(k/9),a=Math.floor(k/3)%3;out.family[a]+=out.classes[k];out.mean[t]+=out.classes[k];out.meanFamily[3*t+a]+=out.classes[k]}
 return out;
}
const concentration=x=>x.reduce((s,p)=>s+p*p,0);
let histories=0,nonlocalWitness=0,initialMeanDifferences=[];
try{
 close(mp.reduce((s,x)=>s+x,0),1,'mean prior');
 for(const base of [[.25,.925],[.47,.7]]){
  const mu=means(base);
  for(let i=0;i<2;i++)for(let t=0;t<27;t++)for(let a=0;a<3;a++){
   const p=prior(mu[i][t],a,false),q=prior(mu[i][t],a,true);
   for(let z=0;z<22;z++)close(p[z],q[z],'urn ratios vs DP');
   close(p.reduce((s,x)=>s+x,0),1,'prior mass');close(p.reduce((s,x,z)=>s+x*rate(mu[i][t],z),0),mu[i][t],'conditional mean');
   const variance=p.reduce((s,x,z)=>s+x*(rate(mu[i][t],z)-mu[i][t])**2,0);
   const expected=a===0?0:(a===1?.2:1)*mu[i][t]*(1-mu[i][t])*(20+(a===1?1:2))/(20*((a===1?1:2)+1));close(variance,expected,'finite variance');
  }
  initialMeanDifferences.push(base.map((b,i)=>({baseline:b,jointPriorMean:mp.reduce((s,w,t)=>s+w*mu[i][t],0)})));
  for(const mode of ['common','independent'])for(const hazard of [0,1/16,1])for(let mask=0;mask<16;mask++){
   const values=[[Boolean(mask&1),Boolean(mask&2)],[Boolean(mask&4),Boolean(mask&8)]],results=[];
   for(const pairValue of [null,false,true]){
    const f=infer(base,values,pairValue,hazard,mode,false),d=infer(base,values,pairValue,hazard,mode,true);results.push(f);
    close(f.z,d.z,'joint evidence');for(const key of ['clean','observed','classes','family','mean','meanFamily'])for(let k=0;k<f[key].length;k++)close(f[key][k],d[key][k],'joint '+key);
    close(f.classes.reduce((s,x)=>s+x,0),1,'posterior mass');
   }
   const [before,no,yes]=results,p=yes.z/before.z;
   close((no.z+yes.z)/before.z,1,'pair partition');
   for(const key of ['clean','observed'])for(let i=0;i<2;i++)close(before[key][i],p*yes[key][i]+(1-p)*no[key][i],'all-target tower');
   for(const key of ['family','mean','meanFamily','classes']){
    const gain=p*concentration(yes[key])+(1-p)*concentration(no[key])-concentration(before[key]);
    const stable=p/(1-p)*yes[key].reduce((s,x,k)=>s+(x-before[key][k])**2,0);
    close(gain,stable,'actual class branches');assert(gain>-2e-11);
   }
   nonlocalWitness=Math.max(nonlocalWitness,Math.abs(yes.clean[1]-before.clean[1]),Math.abs(no.clean[1]-before.clean[1]));histories++;
  }
  for(const mode of ['common','independent'])for(let mask=0;mask<4;mask++){
   const values=[[Boolean(mask&1)],[Boolean(mask&2)]];
   for(const b of [null,false,true]){
    const f=infer(base,values,b,1/16,mode,false);
    for(const x of f.family)close(x,1/3,'one distinct outcome cannot identify dispersion');
   }
  }
 }
 assert(nonlocalWitness>1e-4);
 const payload={naiveOneConditionalTable200:27*3*3*22*200*8,compressedOneConditionalTable200:27*3*(1+22+21)*200*8,capBytes:8*1024*1024};
 assert(payload.naiveOneConditionalTable200>payload.capBytes);
 const result={checks,maxDefect,histories,nonlocalWitness,initialMeanDifferences,payload,meanFamily:27,hyperstates:243,urnSize:20,staticMeanField:true,naiveDenseLayoutFailsCap:true,compressedTableIsPayloadEstimateNotMeasuredAllocation:true,streamQualityRescueEstablished:false,performanceEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'};
 assert.equal(sha(source),freeze.sourceSHA256);for(const[p,h]of Object.entries(protectedFiles))assert.equal(sha(p),h,p);
 fs.writeFileSync(root+'/results.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});
 fs.writeFileSync(root+'/completed.json',JSON.stringify({exitCode:0,allJobsTerminal:true,sourceUnchanged:true,resultsSHA256:sha(root+'/results.json'),sourceSHA256:freeze.sourceSHA256},null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));
}catch(e){fs.writeFileSync(root+'/failure.json',JSON.stringify({exitCode:1,error:e.message,checks,maxDefect,sourceSHA256:freeze.sourceSHA256,allJobsTerminal:true,scientificGatesChanged:false},null,2)+'\n',{flag:'wx',mode:0o600});throw e}
