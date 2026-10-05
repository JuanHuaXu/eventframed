import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const bytes=readFileSync(process.argv[2]),data=JSON.parse(bytes);
for(const [p,h]of Object.entries(data.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
function table(noise){const errors=[noise,noise,.01,.01,Math.min(.4,2*noise),.05,.5,noise];return Array.from({length:8},(_,t)=>Array.from({length:16},(_,h)=>{const b=Array.from({length:4},(_,i)=>(h>>i)&1),v=[...b,b[0]^b[1],b[2]^b[3],0,b[0]^b[2]];return v[t]?1-errors[t]:errors[t];}));}
function component(history,likelihood,freshPrior){
  return Array.from({length:16},(_,h)=>{
    let weight=1/16;
    for(let t=0;t<8;t++){
      const q=likelihood[t][h],reports=history.filter(s=>s.action[1]===t);let local=0;
      for(let ordinary=0;ordinary<2;ordinary++)for(let fresh=0;fresh<2;fresh++)for(let root=0;root<2;root++){
        let p=.5*(fresh?freshPrior:1-freshPrior)*(root?q:1-q);
        for(const s of reports){const [kind,,slot]=s.action,independent=kind?fresh:ordinary&&slot>0,emission=independent?q:root;p*=s.outcome?emission:1-emission;}
        local+=p;
      }
      weight*=local;
    }
    return weight;
  });
}
function expected(history,likelihood){
  const ws=[0,1,.5].map(p=>component(history,likelihood,p)),priors=[.25,.25,.5];
  const masses=ws.map((w,i)=>w.reduce((s,x)=>s+x,0)*priors[i]),total=masses.reduce((s,x)=>s+x,0);
  const localTotal=ws[2].reduce((s,x)=>s+x,0),local=Array(4).fill(0),mix=Array(4).fill(0);
  for(let h=0;h<16;h++){local[h%4]+=ws[2][h]/localTotal;mix[h%4]+=ws.reduce((s,w,i)=>s+priors[i]*w[h],0)/total;}
  const hard=Array(4).fill(0),hardZ=ws[1].reduce((s,x)=>s+x,0);ws[1].forEach((w,h)=>hard[h%4]+=w/hardZ);
  return {ps:[hard,local,mix],components:masses.map(m=>m/total),evidence:Math.log(total)};
}

const likelihood=table(.2),names=['certain','local','hierarchical'],cases=['genuine','counterfeit','even_counterfeit','odd_counterfeit'];
const results=Object.fromEntries(cases.map(c=>[c,{mass:0,oracle_brier:0,models:Object.fromEntries(names.map(n=>[n,{brier:0,confident_wrong:0}]))}]));
for(let bits=0;bits<1024;bits++){
  const history=data.schedule.map((action,i)=>({action,outcome:Boolean((bits>>(9-i))&1)})),predictions=expected(history,likelihood).ps;
  for(const c of cases){
    const weights=Array.from({length:16},(_,h)=>{
      const roots=new Map();let w=1/16;
      for(const s of history){
        const [kind,t]=s.action,q=likelihood[t][h];
        if(kind===0){roots.set(t,s.outcome);w*=s.outcome?q:1-q;}
        else {const copied=c==='counterfeit'||c==='even_counterfeit'&&t%2===0||c==='odd_counterfeit'&&t%2===1;w*=copied?Number(s.outcome===roots.get(t)):s.outcome?q:1-q;}
      }
      return w;
    });
    const mass=weights.reduce((s,x)=>s+x,0),out=results[c];out.mass+=mass;if(mass===0)continue;
    const classMass=Array(4).fill(0);weights.forEach((w,h)=>classMass[h%4]+=w);
    out.oracle_brier+=mass-classMass.reduce((s,x)=>s+x*x/mass,0);
    predictions.forEach((p,i)=>{
      const chosen=p.indexOf(Math.max(...p)),norm=p.reduce((s,x)=>s+x*x,0);
      out.models[names[i]].brier+=weights.reduce((s,w,h)=>s+w*(1+norm-2*p[h%4]),0);
      if(Math.max(...p)>=.9)out.models[names[i]].confident_wrong+=mass-classMass[chosen];
    });
  }
}
let maxError=0,checks=0;
for(const c of cases){
  for(const key of ['mass','oracle_brier']){const e=Math.abs(results[c][key]-data.results[c][key]);assert(e<1e-11);maxError=Math.max(maxError,e);checks++;}
  for(const n of names)for(const key of ['brier','confident_wrong']){const e=Math.abs(results[c].models[n][key]-data.results[c].models[n][key]);assert(e<1e-11);maxError=Math.max(maxError,e);checks++;}
}
console.log(JSON.stringify({scope:'Independent batch likelihood and squared-loss expansion; classification ties not independently checked',results,checks,maxError,artifactSHA256:createHash('sha256').update(bytes).digest('hex'),sourceSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex')},null,2));

