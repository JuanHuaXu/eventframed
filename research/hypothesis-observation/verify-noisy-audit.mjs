import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {auditBelief,auditLikelihood} from './noisy-audit.mjs';
import {canonicalHistory} from './count-planning.mjs';
import {acquisitionBelief} from './interval-acquisition.mjs';
import {actualJoint} from './exact-regime.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'noisy-audit-evaluation.json'),d=JSON.parse(raw);
for(const [f,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(dir+f)).digest('hex'),h);
const nodes=[-.9324695142031521,-.6612093864662645,-.2386191860831969,.2386191860831969,.6612093864662645,.9324695142031521];
const weights=[.1713244923791704,.3607615730481386,.4679139345726910,.4679139345726910,.3607615730481386,.1713244923791704];
for(let k=0;k<=11;k++)assert(Math.abs(nodes.reduce((v,z,i)=>v+weights[i]/2*((z+1)/2)**k,0)-1/(k+1))<1e-14);
let posteriorChecks=0,quadratureChecks=0,maxError=0;
for(const root of d.roots)for(const s of root.states.filter((_,i)=>i%127===0)){
  const base=acquisitionBelief(canonicalHistory(root.pattern,s.counts)),a=auditBelief(root.pattern,s.counts,0),b=auditBelief(root.pattern,s.counts,1),neutral=auditBelief(root.pattern,s.counts,root.audit,.5);
  const joint=[0,0,0,0];
  for(let q=0;q<6;q++)for(let mask=0;mask<16;mask++){
    const prior=.5/16+(mask===0||mask===15?.25:0),factor=weights[q]/2*prior*auditLikelihood(mask,root.audit);
    const mass=actualJoint(root.pattern,s.counts,.2+.1*nodes[q],mask);
    mass.forEach((v,y)=>joint[y]+=factor*v);
  }
  const total=joint.reduce((v,x)=>v+x,0);
  for(let y=0;y<4;y++){
    assert(Math.abs(a.mass*a.forecast[y]+b.mass*b.forecast[y]-base.mass*base.forecast[y])<1e-12);
    assert(Math.abs(neutral.forecast[y]-base.forecast[y])<1e-12);posteriorChecks+=2;
    const error=Math.abs(joint[y]/total-s.forecast[y]);assert(error<1e-12);maxError=Math.max(maxError,error);quadratureChecks++;
  }
}
let backwardChecks=0,maxScoreError=0;
for(const [noise,mask]of [[.1,0],[.2,1],[.25,2],[.3,6],[.3,15],[.2,10]]){
  const row=d.results.find(r=>r.noise===noise&&r.mask===mask),totals=Object.fromEntries(['target','entropy','random'].map(p=>[p,{finalBrier:0,areaBrier:0}]));
  for(const root of d.roots){
    const states=new Map(root.states.map(s=>[s.counts.join(','),s])),cache=new Map();
    const world=key=>{
      if(!cache.has(key)){
        const s=states.get(key),joint=actualJoint(root.pattern,s.counts,noise,mask),mass=joint.reduce((a,v)=>a+v,0);
        const loss=mass?(mass*(1+s.forecast.reduce((a,p)=>a+p*p,0))-2*joint.reduce((a,v,y)=>a+v*s.forecast[y],0))/mass:0;
        cache.set(key,{s,mass,loss});
      }return cache.get(key);
    };
    for(const policy of Object.keys(totals)){
      const memo=new Map();
      function solve(key){
        if(memo.has(key))return memo.get(key);
        const w=world(key),step=w.s.counts.reduce((a,v)=>a+v,0);assert(w.mass>0);
        if(step===5)return {final:w.loss,area:0};
        const out={final:0,area:w.loss/6};
        for(let a=0;a<4;a++){
          const pa=policy==='random'?.25:Number(w.s.actions[policy]===a);if(!pa)continue;
          for(let y=0;y<2;y++){const c=w.s.counts.slice();c[2*a+y]++;const nextKey=c.join(','),child=world(nextKey);if(!child.mass)continue;
            const next=solve(nextKey),p=pa*child.mass/w.mass;out.final+=p*next.final;out.area+=p*next.area;
          }
        }memo.set(key,out);return out;
      }
      const key='0,0,0,0,0,0,0,0',mass=world(key).mass*auditLikelihood(mask,root.audit),out=solve(key);
      const initial=actualJoint(root.pattern,Array(8).fill(0),noise,mask),pre=initial.reduce((a,v,y)=>a+v*root.preAudit.reduce((b,p,z)=>b+(p-Number(y===z))**2,0),0)*auditLikelihood(mask,root.audit)/6;
      totals[policy].finalBrier+=mass*out.final;totals[policy].areaBrier+=mass*out.area+pre;
    }
  }
  for(const [p,out]of Object.entries(totals))for(const k of Object.keys(out)){const error=Math.abs(out[k]-row.scores[p][k]);assert(error<1e-11);maxScoreError=Math.max(maxScoreError,error);backwardChecks++;}
}
assert(raw.equals(execFileSync(process.execPath,[dir+'noisy-audit-evaluation.mjs'],{maxBuffer:64*1024*1024})));
const out={scope:'Audit marginalization/neutral channel, independent quadrature posterior, separate backward scores and full replay; no empirical audit validation',posteriorChecks,quadratureChecks,maxError,backwardChecks,maxScoreError,replay:true,inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-noisy-audit.mjs')).digest('hex')};
writeFileSync(dir+'noisy-audit-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
