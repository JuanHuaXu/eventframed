import fs from 'node:fs';import crypto from 'node:crypto';import path from 'node:path';
const root=path.resolve(import.meta.dirname,'..'),raw=fs.readFileSync(process.argv[2]),a=JSON.parse(raw);
const assert=(x,m)=>{if(!x)throw Error(m)},eq=(x,y)=>JSON.stringify(x)===JSON.stringify(y);
const parentRaw=fs.readFileSync(path.join(root,'docs/experiments/mmm-delayed-v104.json')),parent=JSON.parse(parentRaw);
assert(a.Version==='v105'&&a.ParentSHA256===crypto.createHash('sha256').update(parentRaw).digest('hex'),'parent hash');
for(const [file,hash] of Object.entries({...parent.Hashes,...a.Hashes}))assert(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,file))).digest('hex')===hash,`source ${file}`);
assert(a.Records.length===128,'record count');
const key=r=>[r.Phase,r.Case,r.Index,r.Schedule].join('/'),index=new Map(parent.Records.map(r=>[key(r),r])),seen=new Set();
const prior=[.95,.05/3,.05/3,.05/3],total=prior.reduce((s,v)=>s+v,0);for(let i=0;i<4;i++)prior[i]/=total;
function risk(profile,w){let v=profile.C;for(let i=0;i<4;i++){v-=2*w[i]*profile.B[i];for(let j=0;j<4;j++)v+=w[i]*w[j]*profile.A[i][j];}return v;}
for(const d of a.Records){
 const r=d.Record,k=key(r);assert(!seen.has(k)&&eq(r,index.get(k)),'parent replay');seen.add(k);
 assert(r.Schedule===1&&r.Case.includes('_to_')&&d.Updates.length===r.Arrived&&d.Profiles.length===8,'selection');
 let previousOrigin=-1,previousWeights=prior;
 for(const u of d.Updates){
  assert(u.Origin>previousOrigin&&u.Origin<256&&u.Clock>=u.Origin&&u.Clock<=287,'update order');previousOrigin=u.Origin;
  const s=r.Steps[u.Origin];assert(!s.Missing&&u.Outcome===s.Y&&u.Arrival===u.Origin+s.Delay&&u.Wait===u.Clock-u.Arrival&&u.Wait>=0,'arrival');
  assert(u.IssueVersion===1+Math.floor(u.Origin/32)&&u.BankOnly===(u.IssueVersion!==u.UpdateVersion),'version');
  const losses=u.Raw.map(p=>(p-Number(u.Outcome))**2),unnorm=u.Before.map((w,i)=>w*Math.exp(-.5*losses[i])),sum=unnorm.reduce((s,x)=>s+x,0);
  for(let j=0;j<4;j++){
   assert(Math.abs(u.Before[j]-previousWeights[j])<1e-12,'weight continuity');
   assert(Math.abs(u.After[j]-(.999*unnorm[j]/sum+.001*prior[j]))<1e-12,'loss/share reconstruction');
  }
  previousWeights=u.After;
  assert(u.RiskMeasured===(u.Clock<256),'risk coverage');
  if(u.RiskMeasured){assert(u.Profile===Math.floor(u.Clock/32),'risk clock');const p=d.Profiles[u.Profile];assert(Math.abs(risk(p,u.Before)-u.RiskBefore)<1e-12&&Math.abs(risk(p,u.After)-u.RiskAfter)<1e-12,'risk reconstruction');}
 }
}
const groups=[];
for(const phase of ['design','confirmation'])for(const c of ['majority_to_parity','parity_to_majority']){
 const rows=a.Records.filter(d=>d.Record.Phase===phase&&d.Record.Case===c);assert(rows.length===32,'group count');
 for(const originGroup of ['pre_change','post_change'])for(const bankOnly of [false,true]){
  const us=rows.flatMap(d=>d.Updates).filter(u=>u.RiskMeasured&&u.Clock>=128&&(u.Origin<128)===(originGroup==='pre_change')&&u.BankOnly===bankOnly);
  const mean=f=>us.length?us.reduce((s,u)=>s+f(u),0)/us.length:null;
  groups.push({phase,case:c,originGroup,bankOnly,updates:us.length,meanAge:mean(u=>u.Clock-u.Origin),meanArrivalDelay:mean(u=>u.Arrival-u.Origin),meanAdditionalWait:mean(u=>u.Wait),meanRawRiskChange:mean(u=>u.RiskAfter-u.RiskBefore),sumRawRiskChange:us.reduce((s,u)=>s+u.RiskAfter-u.RiskBefore,0),harmful:us.filter(u=>u.RiskAfter-u.RiskBefore>1e-12).length});
 }
}
process.stdout.write(JSON.stringify({version:'v105',type:'consumed diagnostic, not confirmation',parentSHA256:a.ParentSHA256,artifactSHA256:crypto.createHash('sha256').update(raw).digest('hex'),records:128,updates:a.Records.reduce((s,d)=>s+d.Updates.length,0),sourceHashes:Object.keys({...parent.Hashes,...a.Hashes}).length,groups},null,2)+'\n');
