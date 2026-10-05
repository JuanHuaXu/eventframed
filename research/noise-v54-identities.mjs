import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const close=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-14,`${a} ${b}`);
function joint(p,eta,u,v){let mass=0;for(const[y,prior]of[[true,p],[false,1-p]])mass+=prior*(u===y?1-eta:eta)*(v===y?1-eta:eta);return mass;}
function single(p,eta,u){return (u?p:1-p)*(1-eta)+(u?1-p:p)*eta;}
let parameterCases=0,stagedCases=0;
for(const p of[0,.1,.25,.5,.9,1])for(const eta of[0,.1,.2]){parameterCases++;let total=0,disagree=0;for(const u of[false,true]){let marginal=0;for(const v of[false,true]){const q=joint(p,eta,u,v);assert(q>=0&&q<=1);total+=q;marginal+=q;if(u!==v)disagree+=q;const first=single(p,eta,u);if(first>0){const conditional=q/first;close(first*conditional,q);stagedCases++}else close(q,0)}close(marginal,single(p,eta,u))}close(total,1);close(disagree,2*eta*(1-eta));close((1-Math.sqrt(1-2*disagree))/2,eta);}
for(const u of[false,true])close(single(.9,.1,u),single(.82,0,u));
close(joint(.9,.1,false,true)+joint(.9,.1,true,false),.18);
close(joint(.82,0,false,true)+joint(.82,0,true,false),0);
assert(Math.abs(joint(.9,.1,true,true)-single(.9,.1,true)**2)>.01,'must not resample latent outcome');
assert.equal(joint(.82,0,false,true),0,'impossible eta0 model must have zero mass');
const sources={};for(const p of['research/noise-v54-identities.mjs','research/noise-v54-direction.md'])sources[p]=hash(fs.readFileSync(p));
const result={parameterCases,stagedCases,jointNormalized:true,singleMarginalVerified:true,stagewiseFactorReplacementVerified:true,singleLabelNonidentifiabilityVerified:true,pairedDistinguishabilityUnderDeclaredIndependence:true,impossibleModelZeroMassVerified:true,proposalOnly:true,learnedImplementation:false,acquisitionOrPerformanceTest:false,wholeGoals:'OPEN',productionChanged:false,sources};
fs.writeFileSync('research/noise-v54-identities.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));
