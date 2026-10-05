import assert from 'node:assert/strict';

const parity=Array.from({length:512},(_,x)=>x.toString(2).replaceAll('0','').length%2);
const prior=Array.from({length:512},(_,x)=>{const k=x.toString(2).replaceAll('0','').length;return (1/3)**k*(2/3)**(9-k);});

// Direct-probability reference: dense mask/value counts, no Go ternary lookup,
// log-sum recurrence, or fitter call. N<=64 keeps these likelihoods nonzero.
export function checkSegmentReference(r,fit,window,hazard=.01,mass=.95){
  const origins=Array.from({length:16},(_,j)=>j-16);
  for(let j=0;j<fit.Clock;j++){const s=r.Steps[j];if(!s.Missing&&j+s.Delay<=fit.Clock)origins.push(j);}
  const selected=origins.slice(-(window===0?64:32));
  assert.deepEqual(selected,fit.Origins[window]);
  const samples=selected.map(j=>j<0?r.Initial[j+16]:{Bits:r.Steps[j].X,Outcome:r.Steps[j].Y});
  const queries=r.Steps.slice(fit.Clock,fit.Clock+32).map(s=>s.X),n=samples.length,H=fit.Clock+16;
  const M=Array.from({length:n+1},()=>Array(n+1).fill(1));
  const tail=Array.from({length:n+1},()=>Array(queries.length).fill(.5));
  const counts=new Uint16Array(512*512),yes=new Uint16Array(512*512),agrees=new Uint16Array(512);
  for(let u=0;u<n;u++){
    counts.fill(0);yes.fill(0);agrees.fill(0);
    const g=Array(512).fill(1),b=Array(512).fill(1);
    for(let v=u;v<n;v++){
      const s=samples[v];let ge=0,be=0;
      for(let mask=0;mask<512;mask++){
        const idx=mask*512+(s.Bits&mask),y=Number(s.Outcome);
        g[mask]*=((y?yes[idx]:counts[idx]-yes[idx])+.5)/(counts[idx]+1);
        counts[idx]++;yes[idx]+=y;
        const match=Number(parity[s.Bits&mask]===y);
        b[mask]*=((match?agrees[mask]:v-u-agrees[mask])+.5)/(v-u+1);
        agrees[mask]+=match;ge+=prior[mask]*g[mask];be+=prior[mask]*b[mask];
      }
      M[u][v+1]=mass*ge+(1-mass)*be;
      assert(M[u][v+1]>0&&M[u][v+1]<=1+1e-12);
      if(v===n-1)for(let q=0;q<queries.length;q++){
        let numerator=0;
        for(let mask=0;mask<512;mask++){
          const idx=mask*512+(queries[q]&mask),gp=(yes[idx]+.5)/(counts[idx]+1),ap=(agrees[mask]+.5)/(n-u+1),bp=parity[queries[q]&mask]?ap:1-ap;
          numerator+=prior[mask]*(mass*g[mask]*gp+(1-mass)*b[mask]*bp);
        }
        tail[u][q]=numerator/M[u][n];
      }
    }
  }
  const before=Array.from({length:H+1},(_,j)=>selected.filter(t=>t<j-16).length),F=[1],last=Array(H).fill(0);
  for(let end=1;end<=H;end++){
    let total=0;
    for(let start=0;start<end;start++){
      const v=F[start]*(start===0?1:hazard)*(1-hazard)**(end-start-1)*M[before[start]][before[end]];
      total+=v;if(end===H)last[start]=v;
    }
    assert(total>0);F.push(total);
  }
  assert.equal(fit.SegmentStarts[window].length,H);
  let maxWeightError=0,maxForecastError=0,maxNoChangeError=0;
  for(let j=0;j<H;j++){last[j]/=F[H];maxWeightError=Math.max(maxWeightError,Math.abs(last[j]-fit.SegmentStarts[window][j]));}
  const evidenceError=Math.abs(Math.log(F[H])-fit.SegmentEvidence[window]);
  for(let q=0;q<queries.length;q++){
    const p=hazard*.5+(1-hazard)*last.reduce((s,w,j)=>s+w*tail[before[j]][q],0);
    maxForecastError=Math.max(maxForecastError,Math.abs(p-r.Steps[fit.Clock+q].P[10+window]));
    maxNoChangeError=Math.max(maxNoChangeError,Math.abs(tail[0][q]-r.Steps[fit.Clock+q].P[13+window]));
  }
  assert(evidenceError<1e-8&&maxWeightError<1e-8&&maxForecastError<1e-8&&maxNoChangeError<1e-8,'segment reference mismatch');
  return {evidenceError,maxWeightError,maxForecastError,maxNoChangeError,forecasts:queries.length};
}
