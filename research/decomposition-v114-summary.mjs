import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const root=path.resolve(import.meta.dirname,'..');
const bytes=fs.readFileSync(process.argv[2]),a=JSON.parse(bytes);
const parentBytes=fs.readFileSync(path.join(root,'docs/experiments/mmm-hazard-v113.json'));
const parent=JSON.parse(parentBytes);
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const assert=(x,m)=>{if(!x)throw Error(m)};
const close=(x,y,m)=>assert(Math.abs(x-y)<1e-11,m);
assert(a.Version==='v114'&&a.ParentSHA256===hash(parentBytes),'parent');
for(const [file,digest] of Object.entries(a.Hashes))assert(hash(fs.readFileSync(path.join(root,file)))===digest,`source ${file}`);
const id=r=>[r.Phase,r.Case,r.Index].join('/');
const eligible=parent.Records.filter(r=>r.Schedule===1&&['parity4','majority_to_parity','parity_to_majority'].includes(r.Case));
const parents=new Map(eligible.map(r=>[id(r),r])),seen=new Set(),blocks=[];
assert(parents.size===192&&a.Records.length===192,'coverage');
const risk=(p,q)=>(p-q)**2+q*(1-q);
let partialChecks=0;
for(const r of a.Records){
 const key=id(r),p=parents.get(key);assert(p&&!seen.has(key),'identity');seen.add(key);
 assert(r.Profiles.length===8&&r.Steps.length===256,'shape');
 for(let block=0;block<8;block++){
  const profile=r.Profiles[block];assert(profile.Clock===block*32&&profile.Full.length===4,'profile');
  for(const row of profile.Full)assert(row.length===512&&row.every(v=>Number.isFinite(v)&&v>0&&v<1),'full predictions');
  const cache=new Map();
  const partial=(mask,x)=>{
   const values=x&mask,k=mask+'/'+values;if(cache.has(k))return cache.get(k);
   const sums=[0,0,0,0];let n=0;
   for(let z=0;z<512;z++)if((z&mask)===values){n++;for(let j=0;j<4;j++)sums[j]+=profile.Full[j][z];}
   assert(n>0,'completion support');const out=sums.map(v=>v/n);cache.set(k,out);return out;
  };
  for(const [j,arm] of [5,7,8].entries()){
   const totals={};
   const add=(k,v)=>{assert(Number.isFinite(v),k);totals[k]=(totals[k]??0)+v/32;};
   for(let t=block*32;t<(block+1)*32;t++){
    const s=p.Steps[t],d=r.Steps[t],ps=d.Raw[j],want=partial(s.Mask[arm],s.X);
    for(let k=0;k<4;k++){close(ps[k],want[k],'partial reconstruction');partialChecks++;}
    close(d.Generic,s.P[0],'generic parent');close(d.Generic,partial(s.Mask[0],s.X)[0],'generic completion');
    const served=risk(s.P[arm],s.Q),generic=risk(d.Generic,s.Q),matched=risk(ps[0],s.Q);
    const lo=Math.min(.5,...ps),hi=Math.max(.5,...ps),oracle=Math.max(lo,Math.min(hi,s.Q));
    assert(s.P[arm]>=lo-1e-11&&s.P[arm]<=hi+1e-11,'served outside raw/neutral hull');
    const envelope=risk(oracle,s.Q),noise=s.Q*(1-s.Q);
    assert(envelope<=served+1e-11&&envelope>=noise-1e-11,'envelope bound');
    add('served',served);add('genericOwn',generic);add('genericMatched',matched);
    add('totalDifference',served-generic);add('modelMixtureDifference',served-matched);add('observationDifference',matched-generic);
    add('envelope',envelope);add('noise',noise);add('mixtureHeadroom',served-envelope);add('envelopeExcess',envelope-noise);
    add('cost',s.Cost[arm]);add('genericCost',s.Cost[0]);
    for(let k=0;k<4;k++){add('raw'+k,risk(ps[k],s.Q));add('full'+k,risk(profile.Full[k][s.X],s.Q));}
   }
   close(totals.totalDifference,totals.modelMixtureDifference+totals.observationDifference,'additive decomposition');
   close(totals.served,totals.noise+totals.envelopeExcess+totals.mixtureHeadroom,'headroom decomposition');
   blocks.push({phase:r.Phase,case:r.Case,index:r.Index,arm,clock:block*32,...totals});
  }
 }
}
function interval(xs){const mean=xs.reduce((s,v)=>s+v,0)/xs.length,se=Math.sqrt(xs.reduce((s,v)=>s+(v-mean)**2,0)/(xs.length-1)/xs.length);return{mean,lower:mean-3.5*se,upper:mean+3.5*se};}
const groups=[];
for(const phase of ['design','confirmation'])for(const c of ['parity4','majority_to_parity','parity_to_majority'])for(const arm of [5,7,8])for(let clock=0;clock<256;clock+=32){
 const rows=blocks.filter(r=>r.phase===phase&&r.case===c&&r.arm===arm&&r.clock===clock);assert(rows.length===32,'group size');
 const metrics={};for(const k of Object.keys(rows[0]).filter(k=>!['phase','case','index','arm','clock'].includes(k)))metrics[k]=interval(rows.map(r=>r[k]));
 groups.push({phase,case:c,arm,clock,metrics});
}
process.stdout.write(JSON.stringify({version:'v114',status:'DIAGNOSTIC_NOT_VALIDATION',artifactSHA256:hash(bytes),parentSHA256:a.ParentSHA256,sourceHashes:Object.keys(a.Hashes).length,trajectories:192,profiles:192*8,stepRows:192*256,partialChecks,groups},null,2)+'\n');
