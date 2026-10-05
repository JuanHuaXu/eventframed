// Research identity checks only: no cohort, learner adoption or timing claim.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const atoms=[.15,.5,.85],prior=[.2,.3,.5];
function emission(p,eta,x){if(!x.known)return 1;const a=x.first?1-eta:eta;if(!x.paired)return p*a+(1-p)*(1-a);const b=x.second?1-eta:eta;return p*a*b+(1-p)*(1-a)*(1-b);}
function normalize(a){const s=a.reduce((x,y)=>x+y,0);assert(Number.isFinite(s)&&s>=0);return s===0?null:a.map(x=>x/s);}
function fullPath(rows,eta,hazard,target,second){let denominator=0,numerator=0;function visit(t,last,weight){if(t===rows.length){denominator+=weight;return;}for(let z=0;z<atoms.length;z++){const transition=t===0?prior[z]:hazard*prior[z]+(z===last?1-hazard:0),x=rows[t];const next=weight*transition*emission(atoms[z],eta,x);if(t===target){const updated={...x,paired:true,second};const ratio=emission(atoms[z],eta,updated)/emission(atoms[z],eta,x);walk(t+1,z,next*ratio);}visit(t+1,z,next);}}function walk(t,last,weight){if(t===rows.length){numerator+=weight;return;}for(let z=0;z<atoms.length;z++)walk(t+1,z,weight*(hazard*prior[z]+(z===last?1-hazard:0))*emission(atoms[z],eta,rows[t]));}visit(0,-1,1);return {denominator,numerator};}
function smoothing(rows,eta,hazard,target,second){let row=prior.slice(),alpha;for(let t=0;t<=target;t++){if(t>0)row=row.map((x,z)=>(1-hazard)*x+hazard*prior[z]);row=normalize(row.map((x,z)=>x*emission(atoms[z],eta,rows[t])));if(!row)return null;}alpha=row;let beta=atoms.map(()=>1);for(let t=rows.length-1;t>target;t--){const g=beta.map((b,z)=>b*emission(atoms[z],eta,rows[t])),reset=g.reduce((s,x,z)=>s+prior[z]*x,0);beta=normalize(g.map(x=>(1-hazard)*x+hazard*reset));if(!beta)return null;}const gamma=normalize(alpha.map((x,z)=>x*beta[z]));if(!gamma)return null;return gamma.reduce((s,x,z)=>s+x*emission(atoms[z],eta,{...rows[target],paired:true,second})/emission(atoms[z],eta,rows[target]),0);}
const histories=[
 [{known:true,first:true},{known:false,first:false},{known:true,first:false},{known:true,first:true}],
 [{known:true,first:false},{known:true,first:true,paired:true,second:true},{known:false,first:true},{known:true,first:false}],
 [{known:true,first:true,paired:true,second:false},{known:false,first:false},{known:true,first:true},{known:true,first:false}],
 [{known:false,first:false},{known:true,first:false},{known:true,first:true,paired:true,second:false},{known:false,first:true},{known:true,first:true}],
];
let cases=0,zeroSupport=0,futureForks=0;
for(const eta of[0,.1,.2])for(const hazard of[0,.125,.8])for(const rows of histories)for(let target=0;target<rows.length;target++){if(!rows[target].known||rows[target].paired)continue;let qsum=0;for(const second of[false,true]){const ref=fullPath(rows,eta,hazard,target,second),q=smoothing(rows,eta,hazard,target,second);cases++;if(ref.denominator===0){assert.equal(q,null);zeroSupport++;continue;}assert(q!==null&&Math.abs(q-ref.numerator/ref.denominator)<2e-12);qsum+=q;}if(qsum>0)assert(Math.abs(qsum-1)<2e-12);const fork=rows.map(x=>x.known?{...x}:{...x,first:!x.first,second:!x.second});assert.equal(smoothing(rows,eta,hazard,target,true),smoothing(fork,eta,hazard,target,true));futureForks++;}
const source='research/paired-v55-smoothing-identities.mjs',sha256=crypto.createHash('sha256').update(fs.readFileSync(source)).digest('hex');
const result={study:'paired-v55-smoothing-identities',cases,zeroSupportChecks:zeroSupport,unknownEvidenceForks:futureForks,atoms:3,histories:4,hazards:[0,.125,.8],noise:[0,.1,.2],sources:{[source]:sha256},independentExhaustivePaths:true,proposalOnly:true,implementedLearner:false,performanceMeasured:false,wholeGoals:'OPEN'};
fs.writeFileSync('research/paired-v55-smoothing-identities.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));
