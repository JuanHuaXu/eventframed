// Finite joint-model identities, not an implemented observer or an experiment.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const near=(a,b)=>assert(Number.isFinite(a)&&Number.isFinite(b)&&Math.abs(a-b)<2e-12,`${a} ${b}`);
const loss=(q,y)=>(q-y)**2;
let cases=0;const rows=[];
for(const eta of[0,.1,.2])for(const rates of[[.2,.8],[.3,.6],[.4,.4]])for(const first of[false,true]) {
 // Enumerate parameter, old clean outcome, second measurement and NEW clean
 // outcome; W1/W2 share old Y, while the future Y is independent given p.
 const joint=Array.from({length:2},()=>[0,0]);let normalizer=0;
 const coarse=Array.from({length:2},()=>[0]);
 for(let h=0;h<2;h++)for(let old=0;old<2;old++)for(let second=0;second<2;second++)for(let future=0;future<2;future++) {
  const p=rates[h],mass=.5*(old?p:1-p)*(Boolean(old)===first?1-eta:eta)*(second===old?1-eta:eta)*(future?p:1-p);
  joint[second][future]+=mass;coarse[second][0]+=mass;normalizer+=mass;
 }
 for(const row of joint)for(let j=0;j<row.length;j++)row[j]/=normalizer;
 const probability=joint.map(row=>row[0]+row[1]),mean=joint[0][1]+joint[1][1],conditional=joint.map((row,j)=>probability[j]===0?null:row[1]/probability[j]);
 let riskBefore=0,riskAfter=0,projectionGain=0;
 for(let second=0;second<2;second++){if(probability[second]===0)continue;for(let future=0;future<2;future++){riskBefore+=joint[second][future]*loss(mean,future);riskAfter+=joint[second][future]*loss(conditional[second],future);}projectionGain+=probability[second]*(conditional[second]-mean)**2;}
 near(mean,conditional.reduce((s,q,j)=>s+(q??0)*probability[j],0));near(riskBefore-riskAfter,projectionGain);assert(projectionGain>=0);
 // The coarse hypothesis class is already certain in every branch: its
 // concentration gain is zero, despite possible predictive improvement.
 let concentration=-1;for(let j=0;j<2;j++){if(probability[j]===0)continue;const posterior=coarse[j][0]/normalizer/probability[j];near(posterior,1);concentration+=probability[j]*posterior**2;}near(concentration,0);
 if(eta===0||rates[0]===rates[1])near(projectionGain,0);
 rows.push({eta,rates,first,mean,probability,conditional,riskBefore,riskAfter,projectionGain,coarseConcentrationGain:concentration});cases++;
}
const nonvacuous=rows.filter(x=>x.projectionGain>1e-6);assert(nonvacuous.length===8);
const src='research/paired-v57-predictive-value-identities.mjs',sha256=crypto.createHash('sha256').update(fs.readFileSync(src)).digest('hex');
const result={study:'paired-v57-predictive-value-identities',cases,nonvacuousCollapsedClassCases:nonvacuous.length,rows,sources:{[src]:sha256},proposalOnly:true,implementedObserver:false,performanceMeasured:false,controlledCohortEvaluated:false,wholeGoals:'OPEN',goal:'ACTIVE'};
fs.writeFileSync('research/paired-v57-predictive-value-identities.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({cases,nonvacuousCollapsedClassCases:nonvacuous.length,example:nonvacuous[0],proposalOnly:true}));
