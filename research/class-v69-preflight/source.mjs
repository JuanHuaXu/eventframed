// Finite algebra only. No corpus, outcomes, runtime model or production access.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/class-v69-preflight',source='research/class-v69-preflight.mjs';assert(!fs.existsSync(root));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),raw=fs.readFileSync(source);
fs.mkdirSync(root,{mode:0o700});fs.writeFileSync(root+'/source.mjs',raw,{flag:'wx',mode:0o600});
const save=(p,x)=>fs.writeFileSync(root+'/'+p,JSON.stringify(x,null,2)+'\n',{flag:'wx',mode:0o600});
save('freeze.json',{time:new Date().toISOString(),source,sourceSHA256:hash(raw),scope:'finite class-concentration algebra; NOT stream experiment,certification or performance',empiricalLabelsOpened:false});
let checks=0,maxDefect=0;
function near(a,b){const d=Math.abs(a-b);maxDefect=Math.max(maxDefect,d);checks++;assert(d<2e-12,`${a} != ${b}`)}
function compressed(rows){const classes=new Map();let p=0;
 for(const x of rows){assert(x.mass>=0&&x.yes>=0&&x.yes<=1);const c=classes.get(x.cls)??[0,0];c[0]+=x.mass;c[1]+=x.mass*x.yes;classes.set(x.cls,c);p+=x.mass*x.yes}
 let out=0;for(const[m,y]of classes.values()){out-=m*m;if(p>0)out+=y*y/p;if(p<1)out+=(m-y)*(m-y)/(1-p)}return out;
}
// Independent original joint P(latent,W) construction and branch normalization.
function enumerated(rows){let before=0,after=0;const prior={};for(const x of rows)prior[x.cls]=(prior[x.cls]??0)+x.mass;for(const m of Object.values(prior))before+=m*m;
 for(const w of[false,true]){let total=0;const posterior={};for(const x of rows){const m=x.mass*(w?x.yes:1-x.yes);posterior[x.cls]=(posterior[x.cls]??0)+m;total+=m}if(total>0)for(const m of Object.values(posterior))after+=total*(m/total)**2}return after-before;
}
function latent(rows){return compressed(rows.map((x,i)=>({...x,cls:i})))}
function refine(rows,count){return rows.flatMap((x,i)=>Array.from({length:count(i)},()=>({...x,mass:x.mass/count(i)})))}
const x=[.9,.9,.1,.1].map((yes,i)=>({mass:.25,yes,cls:Math.floor(i/2)}));
const y=[1,0,.5,.5].map((yes,i)=>({mass:.25,yes,cls:Math.floor(i/2)}));
const rx=refine(x,i=>i<2?1:20),ry=refine(y,i=>i<2?1:20);
const counterexample={before:{latentX:latent(x),latentY:latent(y),classX:compressed(x),classY:compressed(y)},after:{latentX:latent(rx),latentY:latent(ry),classX:compressed(rx),classY:compressed(ry)},refinement:'split each B latent atom into20 indistinguishable copies; preserve P(class,W)'};
near(counterexample.before.latentX,.16);near(counterexample.before.latentY,.125);near(counterexample.after.latentX,.084);near(counterexample.after.latentY,.125);near(counterexample.before.classX,.32);near(counterexample.before.classY,0);near(counterexample.after.classX,.32);near(counterexample.after.classY,0);
assert(counterexample.before.latentX>counterexample.before.latentY&&counterexample.after.latentX<counterexample.after.latentY);
let state=19790321;function random(){state=(Math.imul(state,1664525)+1013904223)>>>0;return state/4294967296}
for(let n=0;n<256;n++){
 const rows=Array.from({length:4+n%29},(_,i)=>({mass:.001+random(),yes:random(),cls:i%(2+n%7)})),total=rows.reduce((s,x)=>s+x.mass,0);for(const x of rows)x.mass/=total;
 const v=compressed(rows);near(v,enumerated(rows));assert(v>=-2e-12);near(v,compressed(refine(rows,i=>1+i%5)));
 const joint=new Map();for(const z of rows)joint.set(z.cls,(joint.get(z.cls)??0)+z.mass);assert(v<=1-[...joint.values()].reduce((s,m)=>s+m*m,0)+2e-12);
 for(const p of[0,.3,1])near(compressed(rows.map(x=>({...x,yes:p}))),0);
}
near(compressed([{mass:1,yes:.6,cls:0}]),0);
assert.equal(hash(fs.readFileSync(source)),hash(raw));
save('results.json',{checks,maxDefect,counterexample,randomJointCases:256,latentOrderingNotRefinementInvariant:true,classGainRefinementInvariant:true,notEC2Implementation:true,adaptiveSubmodularityEstablished:false,qualityRescueEstablished:false,equalTotalCostEstablished:false,goals:Array(7).fill('OPEN'),goal:'ACTIVE'});
save('completed.json',{exitCode:0,sourceUnchanged:true,allJobsTerminal:true,resultsSHA256:hash(fs.readFileSync(root+'/results.json'))});console.log(JSON.stringify({checks,maxDefect,counterexample,componentOnly:true},null,2));
