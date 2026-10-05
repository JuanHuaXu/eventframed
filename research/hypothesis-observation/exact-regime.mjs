import assert from 'node:assert/strict';
const types=[0,1,2,7];
// Joint mass for ONE ordered report history represented by sufficient counts.
// Path multiplicities belong to policy propagation, not this likelihood.
export function actualJoint(pattern,counts,noise,mask){
  assert(noise>0&&noise<.5&&Number.isInteger(mask)&&mask>=0&&mask<16);
  const joint=[0,0,0,0];
  for(let j=0;j<4;j++)if(mask&(1<<j)){
    const root=(pattern>>j)&1;if(counts[2*j+1-root]>0)return joint;
  }
  for(let h=0;h<16;h++){
    let w=1/16;
    for(let j=0;j<4;j++){
      const t=types[j],truth=t===7?((h&1)^((h>>2)&1)):((h>>t)&1),error=t===2?.01:noise;
      const q=truth?1-error:error,root=(pattern>>j)&1;
      w*=root?q:1-q;
      if(!(mask&(1<<j)))w*=(1-q)**counts[2*j]*q**counts[2*j+1];
    }
    joint[h%4]+=w;
  }
  return joint;
}
export function pathWeights(root,policies){
  const index=new Map(root.states.map((s,i)=>[s.counts.join(','),i]));
  const output={};
  for(const policy of policies){
    const layers=[new Map([[index.get('0,0,0,0,0,0,0,0'),1]])];
    for(let step=0;step<6;step++){
      const next=new Map();
      for(const[i,w]of layers[step]){
        const s=root.states[i];
        for(let a=0;a<4;a++){
          const aw=policy==='random'?.25:Number(s.actions[policy]===a);if(!aw)continue;
          for(let y=0;y<2;y++){const c=s.counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);next.set(j,(next.get(j)||0)+w*aw);}
        }
      }
      layers.push(next);
    }
    output[policy]=layers;
  }
  return output;
}

