import {switchWeights} from './switch-distribution.mjs';
const entropy=w=>-w.reduce((v,p)=>v+(p>0?p*Math.log(p):0),0);
// Integrate the old label's two outcomes and the intervening switching states.
// Rows contain only known labels; the nominated origin must still be unknown.
export function currentStateInformation(rows,prior,origin,base=switchWeights(rows,prior)) {
  if(!Number.isInteger(origin)||origin<0||origin>=rows.length||rows[origin].y!==null)throw Error('origin not unknown');
  const conditional=[0,1].map(y=>switchWeights(rows.map((r,j)=>j===origin?{...r,y}:r),prior));
  const probabilities=conditional.map(c=>Math.exp(c.logEvidence-base.logEvidence));
  if(Math.abs(probabilities[0]+probabilities[1]-1)>1e-9)throw Error('outcome mass');
  for(let k=0;k<prior.length;k++)if(Math.abs(conditional.reduce((v,c,y)=>v+probabilities[y]*c.weights[k],0)-base.weights[k])>1e-9)throw Error('marginal mismatch');
  const information=entropy(base.weights)-conditional.reduce((v,c,y)=>v+probabilities[y]*entropy(c.weights),0);
  if(information < -1e-9||!Number.isFinite(information))throw Error('invalid information');
  return {information:Math.max(0,information),probabilities,conditional:conditional.map(c=>c.weights)};
}
