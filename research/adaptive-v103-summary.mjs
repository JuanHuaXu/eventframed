import fs from 'node:fs';
import crypto from 'node:crypto';
import path from 'node:path';

const input=process.argv[2];
const a=JSON.parse(fs.readFileSync(input,'utf8'));
const root=path.resolve(import.meta.dirname,'..');
const assert=(b,m)=>{if(!b)throw Error(m);};
assert(a.Version==='v103'&&a.SeedBase===2090110300&&a.PerCase===32,'metadata');
for(const [file,hash] of Object.entries(a.Hashes))
  assert(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,file))).digest('hex')===hash,`hash ${file}`);
const cases=['parity1','parity2','parity3','parity4','complement4','majority3','mux3','constant','null','dependent4','majority_to_parity','parity_to_majority'];
assert(a.Records.length===768,'trajectory count');
const seen=new Set(), seeds=new Set();
const pop=x=>x.toString(2).replaceAll('0','').length;
for(const r of a.Records){
  const phase=['design','confirmation'].indexOf(r.Phase),c=cases.indexOf(r.Case);
  assert(phase>=0&&c>=0&&Number.isInteger(r.Index)&&r.Index>=0&&r.Index<32,'record identity');
  const key=[phase,c,r.Index].join('/');assert(!seen.has(key),'duplicate');seen.add(key);
  for(let role=0;role<4;role++){
    const seed=(a.SeedBase+phase*1e6+c*1e4+r.Index*10+role)%2147483647;
    assert(!seeds.has(seed),'seed collision');seeds.add(seed);
  }
  assert(r.Steps.length===256,'step count');
  const b=Array.from({length:6},()=>[0,0]),acc=Array.from({length:6},()=>[0,0]),real=Array(6).fill(0);
  r.Steps.forEach((s,t)=>{
    assert(Number.isInteger(s.X)&&s.X>=0&&s.X<512&&[.05,.5,.95].some(q=>Math.abs(q-s.Q)<1e-12)&&typeof s.Y==='boolean','target');
    assert(s.Nodes>0&&s.Nodes<=19683&&s.Branches>=0,'planner work');
    for(let arm=0;arm<6;arm++){
      const p=s.P[arm],mask=s.Mask[arm],cost=s.Cost[arm];
      assert(Number.isFinite(p)&&p>=0&&p<=1,'probability');
      assert(Number.isInteger(mask)&&mask>=1&&mask<512&&(mask&1)===1&&cost===pop(mask)&&cost<=6,'cost/mask');
      for(let scope=0;scope<3;scope++)assert([0,1,3,7].includes((mask>>(3*scope))&7),'non-prefix observation');
      if(arm===1)assert(mask===s.Mask[0],'matched mask');
      if(arm===5)assert(mask===63,'fixed mask');
      const loss=(p-s.Q)**2+s.Q*(1-s.Q),ac=p>=.5?s.Q:1-s.Q;
      b[arm][0]+=loss/256;acc[arm][0]+=ac/256;
      if(t>=128){b[arm][1]+=loss/128;acc[arm][1]+=ac/128;}
      real[arm]+=(p-Number(s.Y))**2/256;
    }
  });
  for(let arm=0;arm<6;arm++){
    assert(Math.abs(real[arm]-r.Realized[arm])<1e-12,'realized reconstruction');
    for(let seg=0;seg<2;seg++){
      assert(Math.abs(b[arm][seg]-r.Metrics[arm][seg].Brier)<1e-12,'Brier reconstruction');
      assert(Math.abs(acc[arm][seg]-r.Metrics[arm][seg].Accuracy)<1e-12,'accuracy reconstruction');
    }
  }
}
function interval(xs){
  const n=xs.length,mean=xs.reduce((s,x)=>s+x,0)/n;
  const se=Math.sqrt(xs.reduce((s,x)=>s+(x-mean)**2,0)/(n-1)/n);
  return {mean,lower:mean-3.5*se,upper:mean+3.5*se};
}
const gates=[],groups=[];
for(const phase of ['design','confirmation'])for(const c of cases){
  const records=a.Records.filter(r=>r.Phase===phase&&r.Case===c);assert(records.length===32,'group size');
  for(let seg=0;seg<2;seg++){
    const metrics=Array.from({length:6},(_,arm)=>({
      brier:records.reduce((s,r)=>s+r.Metrics[arm][seg].Brier,0)/32,
      accuracy:records.reduce((s,r)=>s+r.Metrics[arm][seg].Accuracy,0)/32,
      cost:records.reduce((s,r)=>s+r.Steps.slice(seg?128:0).reduce((x,t)=>x+t.Cost[arm],0)/(seg?128:256),0)/32
    }));
    groups.push({phase,case:c,segment:seg?'late':'all',metrics});
    for(const control of [0,2]){
      const v=interval(records.map(r=>r.Metrics[3][seg].Brier-r.Metrics[control][seg].Brier));
      gates.push({type:'nonharm',phase,case:c,segment:seg,control,...v,pass:v.upper<=.01});
    }
    const genericGain=seg===0&&['parity3','parity4','complement4'].includes(c)||seg===1&&c.includes('_to_');
    const plannerGain=seg===1&&c==='parity4';
    for(const control of [...(genericGain?[0]:[]),...(plannerGain?[2]:[])]){
      const v=interval(records.map(r=>r.Metrics[control][seg].Brier-r.Metrics[3][seg].Brier));
      gates.push({type:'gain',phase,case:c,segment:seg,control,...v,pass:v.mean>=.005&&v.lower>0});
    }
  }
}
assert(gates.length===108,'gate count');
const output={version:'v103',artifactSHA256:crypto.createHash('sha256').update(fs.readFileSync(input)).digest('hex'),trajectories:a.Records.length,steps:a.Records.length*256,uniqueEffectiveSeeds:seeds.size,sourceHashes:Object.keys(a.Hashes).length,passed:gates.filter(g=>g.pass).length,total:gates.length,status:gates.every(g=>g.pass)?'PASS':'FAIL',gates,groups};
process.stdout.write(JSON.stringify(output,null,2)+'\n');
