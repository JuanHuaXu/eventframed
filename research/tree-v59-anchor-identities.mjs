// Prospective model identities only, no outcome access or empirical study.
import fs from 'node:fs';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const sigmoid=x=>1/(1+Math.exp(-x));
const offsets=[...Array(10)].map((_,i)=>-.6*(i+1)).concat([...Array(10)].map((_,i)=>.6*(i+1)));
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<1e-12,`${a} ${b}`);
function solve(b){let lo=-40,hi=40;for(let k=0;k<100;k++){const a=(lo+hi)/2,mean=offsets.reduce((s,d)=>s+sigmoid(a+d)/20,0);if(mean<b)lo=a;else hi=a}return (lo+hi)/2}
const cases=[];for(const b of[.001,.01,.025,.05,.1,.25,.3,.47,.5,.6,.7,.78,.8,.9,.925,.95,.975,.99,.999]) {
 const a=solve(b),rates=[b,...offsets.map(d=>sigmoid(a+d))],weights=[.8,...Array(20).fill(.01)];assert(rates.every(p=>p>0&&p<1));const initial=weights.reduce((s,w,j)=>s+w*rates[j],0);near(initial,b);assert.equal(rates[0],b);
 // Explicit one-latent-Y pair probabilities conserve the correct marginals.
 for(const eta of[0,.1,.2]) {const cells=[];for(let w1=0;w1<2;w1++)for(let w2=0;w2<2;w2++){let joint=0;for(let j=0;j<rates.length;j++)for(let y=0;y<2;y++){let p=weights[j]*(y?rates[j]:1-rates[j]);p*=w1===y?1-eta:eta;p*=w2===y?1-eta:eta;joint+=p}cells.push(joint)}near(cells.reduce((x,y)=>x+y,0),1);near(cells[2]+cells[3],eta+(1-2*eta)*b);near(cells[1]+cells[3],eta+(1-2*eta)*b);near(cells[1]+cells[2],2*eta*(1-eta))}
 cases.push({b,a,initial,identity:rates[0],minimum:Math.min(...rates),maximum:Math.max(...rates)})
}
// V58 small-tree prior loses context even before delayed evidence exists.
const base=[.25,.47,.78,.925],mean=base.reduce((s,b)=>s+b/4,0),v58=base.map((b,i)=>.5*mean+.25*(base[Math.floor(i/2)*2]+base[Math.floor(i/2)*2+1])/2+.25*b);
assert(v58.some((p,i)=>Math.abs(p-base[i])>.1));
const result={stage:'Finite mathematical preflight, not Go or empirical validation',cases:cases.length,pairedMassIdentities:cases.length*3,v58ZeroEvidenceCounterexample:{base,predictive:v58,squaredDeviation:v58.reduce((s,p,i)=>s+(p-base[i])**2/4,0)},checks:cases,offsets,wholeGoals:Array(7).fill('OPEN'),sourceSHA256:crypto.createHash('sha256').update(fs.readFileSync('research/tree-v59-anchor-identities.mjs')).digest('hex')};
fs.writeFileSync('research/tree-v59-anchor-identities.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({cases:result.cases,pairedIdentities:result.pairedMassIdentities,v58ZeroEvidence:result.v58ZeroEvidenceCounterexample,empiricalValidation:false},null,2));
