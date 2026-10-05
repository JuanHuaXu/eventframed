import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {prepare,bestResponse} from './joint-policy.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'risk-budget-evaluation.json'),d=JSON.parse(raw);
for(const[p,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const modelRaw=readFileSync(dir+'count-planning-exact.json');assert.equal(createHash('sha256').update(modelRaw).digest('hex'),d.inputSHA256);
const entropyControl=process.argv[2];assert(['entropy','tie_entropy'].includes(entropyControl));
const worlds=d.results.map(r=>({noise:r.noise,mask:r.mask})),graphs=prepare(JSON.parse(modelRaw),worlds),rows=[];
worlds.forEach((world,g)=>{for(const control of ['random',entropyControl]){
  rows.push({g,control,metric:'final',threshold:d.results[g].scores[control].finalBrier+.01});
  if(world.mask!==15)rows.push({g,control,metric:'area',threshold:d.results[g].scores[control].areaBrier});
}});
const logWeights=Array(rows.length).fill(0),mean=Array(rows.length).fill(0),rounds=[];
let bestLower=-Infinity,maxForwardError=0;
for(let t=0;t<128;t++){
  const max=Math.max(...logWeights),q=logWeights.map(v=>Math.exp(v-max)),sum=q.reduce((a,v)=>a+v,0);q.forEach((v,i)=>q[i]=v/sum);
  const fw=Array(worlds.length).fill(0),aw=Array(worlds.length).fill(0);
  rows.forEach((r,i)=>(r.metric==='final'?fw:aw)[r.g]+=q[i]);
  const response=bestResponse(graphs,fw,aw);maxForwardError=Math.max(maxForwardError,response.forwardError);
  for(const control of ['random',entropyControl]){
    const cost=worlds.reduce((v,_,g)=>v+fw[g]*d.results[g].scores[control].finalBrier+aw[g]*d.results[g].scores[control].areaBrier,0);
    assert(response.objective<=cost+1e-10);
  }
  const payoff=rows.map(r=>response[r.metric][r.g]-r.threshold);
  const lower=response.objective-rows.reduce((v,r,i)=>v+q[i]*r.threshold,0);bestLower=Math.max(bestLower,lower);
  payoff.forEach((v,i)=>{mean[i]+=(v-mean[i])/(t+1);logWeights[i]+=32*v;});
  const upper=Math.max(...mean);assert(bestLower<=upper+1e-10);
  rounds.push({round:t+1,weights:q,final:response.final,area:response.area,lower,bestLower,upper});
}
const violations=rows.map((r,i)=>({...r,...worlds[r.g],meanPayoff:mean[i],passes:r.metric==='final'?mean[i]<=0:mean[i]<0}));
const files=['JOINT_POLICY_PROTOCOL.md','joint-policy.mjs','joint-policy-two-control.mjs','global-policy.mjs','exact-regime.mjs'];
console.log(JSON.stringify({scope:'Two-control joint forecast/acquisition game on consumed finite worlds; numerical bounds, no assumed convergence or empirical validation',inputSHA256:createHash('sha256').update(raw).digest('hex'),hashes:Object.fromEntries(files.map(f=>[f,createHash('sha256').update(readFileSync(dir+f)).digest('hex')])),entropyControl,maxForwardError,rows,rounds,violations},null,2));
