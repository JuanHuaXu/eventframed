// Expected Brier Bayes-risk reduction on declared virtual predictive probes.
// Probe outcomes are conditionally independent of the queried old label given
// the current expert. These are not observations replayed as new evidence.
export function predictiveQueryValue(base,information,probes){
 if(!probes.length)throw Error('empty probes');
 let value=0;
 for(const p of probes){
  if(p.length!==base.length||p.some(x=>!Number.isFinite(x)||x<0||x>1))throw Error('invalid probe');
  const mean=base.reduce((v,w,k)=>v+w*p[k],0);
  const conditional=information.conditional.map(w=>w.reduce((v,x,k)=>v+x*p[k],0));
  const variance=conditional.reduce((v,x,y)=>v+information.probabilities[y]*(x-mean)**2,0);
  const reduction=mean*(1-mean)-conditional.reduce((v,x,y)=>v+information.probabilities[y]*x*(1-x),0);
  if(Math.abs(variance-reduction)>1e-9)throw Error('risk identity');
  value+=variance/probes.length;
 }
 return value;
}
