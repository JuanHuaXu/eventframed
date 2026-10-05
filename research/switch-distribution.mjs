// Basic finite switch distribution, with delayed evidence replayed as-of clock.
// Q is deliberately not an input: this component receives issued probabilities.
export function switchWeights(rows, prior, switching = true) {
  const k = prior.length;
  if (!k || prior.some(x => !Number.isFinite(x) || x < 0) ||
      Math.abs(prior.reduce((a,b)=>a+b,0)-1)>1e-12) throw Error('invalid prior');
  let active = prior.map(x => switching ? x/2 : 0);
  let stopped = prior.map(x => switching ? x/2 : x);
  let logEvidence = 0;
  for (let j=0;j<rows.length;j++) {
    const {p,y} = rows[j];
    if (p.length!==k || p.some(x=>!Number.isFinite(x)||x<0||x>1) ||
        !(y===null || y===0 || y===1 || y===false || y===true)) throw Error('invalid row');
    if (y!==null) {
      let z=0;
      for(let i=0;i<k;i++) {
        const v=Math.max(1e-12,Math.min(1-1e-12,p[i]));
        const likelihood=y?v:1-v;
        active[i]*=likelihood;stopped[i]*=likelihood;z+=active[i]+stopped[i];
      }
      logEvidence+=Math.log(z);
      for(let i=0;i<k;i++){active[i]/=z;stopped[i]/=z;}
    }
    const hazard=1/(j+2),mass=active.reduce((a,b)=>a+b,0)*hazard;
    for(let i=0;i<k;i++) {
      active[i]=active[i]*(1-hazard)+mass*prior[i]/2;
      stopped[i]+=mass*prior[i]/2;
    }
  }
  return {weights:active.map((x,i)=>x+stopped[i]),logEvidence};
}
