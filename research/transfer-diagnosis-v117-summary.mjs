import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const assert=(x,m)=>{if(!x)throw Error(m)},eq=(x,y)=>JSON.stringify(x)===JSON.stringify(y);
const root=path.resolve(import.meta.dirname,'..'),raw=fs.readFileSync(process.argv[2]),a=JSON.parse(raw);
const parentRaw=fs.readFileSync(path.join(root,'docs/experiments/mmm-transfer-v116.json')),parent=JSON.parse(parentRaw);
assert(a.Version==='transfer-diagnosis-v117'&&a.ParentSHA256==='cad9f6f36879280044f0148895910f9fc10e3d0df902da95cc8276980f079d1e'&&hash(parentRaw)===a.ParentSHA256,'parent');
assert(a.Records.length===1152&&parent.Records.length===1152&&Object.keys(a.Hashes).length===157,'counts');
for(const [name,h] of Object.entries(a.Hashes))assert(hash(fs.readFileSync(path.join(root,name)))===h,`source ${name}`);
for(const [name,h] of Object.entries(parent.Hashes))assert(a.Hashes[name]===h,'parent source manifest');
const segments=[['all256',0,256],['terminal64',192,256],...Array.from({length:8},(_,i)=>['block'+(i*32),i*32,i*32+32])];
const loss=(p,q)=>(p-q)**2+q*(1-q);
const hull=(ps,q)=>loss(Math.max(Math.min(...ps,.5),Math.min(q,Math.max(...ps,.5))),q);
const records=[];let checks=0,identities=0;
for(let n=0;n<a.Records.length;n++){
 const r=a.Records[n],p=parent.Records[n];
 assert(eq(r.Spec,p.Teacher.Spec)&&r.Schedule===p.Schedule&&r.Profiles.length===8&&r.Steps.length===256,'record identity');
 const aggregate=segments.map(()=>({}));
 for(let block=0;block<8;block++){
  const profile=r.Profiles[block];assert(profile.Clock===block*32&&profile.Full.length===4,'profile');
  for(const table of profile.Full)assert(table.length===512&&table.every(v=>Number.isFinite(v)&&v>0&&v<1),'full table');
  const cache=new Map();
  const partial=(model,mask,values)=>{
   const key=model+'/'+mask+'/'+values;
   if(cache.has(key))return cache.get(key);
   let sum=0,count=0;for(let x=0;x<512;x++)if((x&mask)===values){sum+=profile.Full[model][x];count++;}
   assert(count>0,'empty completions');const q=sum/count;cache.set(key,q);return q;
  };
  for(let t=block*32;t<block*32+32;t++){
   const s=p.Steps[t],d=r.Steps[t],q=s.Q,v={floor:q*(1-q)};
   assert(d.Raw.length===3&&d.Generic===s.P[0],'generic identity');
   const full=[];
   for(let k=0;k<4;k++){
    full.push(profile.Full[k][s.X]);
    v['full'+k]=loss(full[k],q);
    v['fixed'+k]=loss(partial(k,63,s.X&63),q);
   }
   v.hullFull=hull(full,q);v.fitLoss=v.full0-v.floor;
   for(let arm=0;arm<3;arm++){
    assert(d.Raw[arm].length===4,'raw models');
    for(let k=0;k<4;k++){
     const expected=partial(k,s.Mask[arm],s.X&s.Mask[arm]);
     assert(Math.abs(expected-d.Raw[arm][k])<1e-12,'independent partial probability');checks++;
     v['view'+arm+'model'+k]=loss(expected,q);
    }
    v['served'+arm]=loss(s.P[arm],q);
    v['route'+arm]=v['served'+arm]-v['view'+arm+'model0'];
    v['viewLoss'+arm]=v['view'+arm+'model0']-v.full0;
    v['excess'+arm]=v['served'+arm]-v.floor;
    v['hull'+arm]=hull(d.Raw[arm],q);
    v['delta'+arm]=v['served'+arm]-loss(s.P[0],q);
    v['obsDiff'+arm]=v['view'+arm+'model0']-loss(s.P[0],q);
    assert(Math.abs(v['route'+arm]+v['viewLoss'+arm]+v.fitLoss-v['excess'+arm])<1e-12,'floor decomposition');
    assert(Math.abs(v['route'+arm]+v['obsDiff'+arm]-v['delta'+arm])<1e-12,'matched decomposition');identities+=2;
   }
   assert(Math.abs(d.Raw[0][0]-s.P[0])<1e-12,'generic own-view');
   for(let seg=0;seg<segments.length;seg++){
    const [,lo,hi]=segments[seg];if(t<lo||t>=hi)continue;
    for(const [key,value] of Object.entries(v))aggregate[seg][key]=(aggregate[seg][key]??0)+value/(hi-lo);
   }
  }
 }
 records.push({spec:r.Spec,schedule:r.Schedule,aggregate});
}
assert(checks===1152*256*12&&identities===1152*256*6,'coverage');
function interval(xs){assert(xs.length===32,'cell size');const mean=xs.reduce((s,x)=>s+x,0)/32,se=Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/31/32);return {mean,lower:mean-3.5*se,upper:mean+3.5*se};}
const groups=[];
for(const phase of [0,1])for(const family of [0,1,2])for(const mode of [0,1,2])for(const schedule of [0,1]){
 const rs=records.filter(r=>r.spec.Phase===phase&&r.spec.Family===family&&r.spec.Mode===mode&&r.schedule===schedule);
 assert(rs.length===32,'group size');
 for(let seg=0;seg<segments.length;seg++){
  const metrics=Object.fromEntries(Object.keys(rs[0].aggregate[seg]).map(k=>[k,interval(rs.map(r=>r.aggregate[seg][k]))]));
  groups.push({phase:['design','confirmation'][phase],family:['additive','hierarchy','local_table'][family],mode:['stationary','abrupt','gradual'][mode],schedule,segment:segments[seg][0],metrics});
 }
}
process.stdout.write(JSON.stringify({version:a.Version,artifactSHA256:hash(raw),parentSHA256:a.ParentSHA256,sourceHashes:Object.keys(a.Hashes).length,scheduleRuns:a.Records.length,profiles:1152*8,steps:1152*256,independentPartialChecks:checks,decompositionIdentities:identities,status:'DIAGNOSIS_NOT_VALIDATION',groups},null,2)+'\n');
