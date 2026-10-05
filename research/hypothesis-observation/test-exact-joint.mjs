import assert from 'node:assert/strict';
import {SCALE,DEN,enclosure,binaryFraction,exactJoint} from './exact-joint.mjs';
let likelihoodChecks=0,enclosureChecks=0;
// Independent16-latent repeated-product fractions, without power padding until end.
for(let root=0;root<16;root++)for(const counts of [Array(8).fill(0),[1,1,0,1,0,0,0,0],[1,0,0,1,1,0,0,3]])for(const noise of [10,15,20,25,30])for(let mask=0;mask<16;mask++){
  const observed=exactJoint(root,counts,noise,mask),reference=[0n,0n,0n,0n];
  for(let h=0;h<16;h++){
    let n=1n,den=16n;
    for(let j=0;j<4;j++){
      const truth=j===3?((h%2)^(Math.floor(h/4)%2)):Math.floor(h/2**j)%2,p=truth?100-(j===2?1:noise):(j===2?1:noise),r=(root>>j)&1;
      n*=BigInt(r?p:100-p);den*=100n;
      for(let y=0;y<2;y++)for(let k=0;k<counts[2*j+y];k++){
        if(mask&(1<<j)){if(y!==r)n=0n;}
        else{n*=BigInt(y?p:100-p);den*=100n;}
      }
    }
    // DEN/den can be fractional by2 because latent bit3 is duplicated.
    assert((n*DEN*2n)%den===0n);reference[h%4]+=n*DEN*2n/den;
  }
  observed.forEach((v,y)=>{assert.equal(v*2n,reference[y]);likelihoodChecks++;});
}
for(let n=0n;n<100n;n++)for(let d=1n;d<33n;d++){
  const [lo,hi]=enclosure(n,d);assert(lo*d<=n*SCALE&&hi*d>=n*SCALE&&hi-lo<=1n);enclosureChecks++;
}
for(const x of [0,.01,.1,.25,.5,1,2,Number.MIN_VALUE]){
  const f=binaryFraction(x);
  if(x===Number.MIN_VALUE){assert.equal(f.n,1n);assert.equal(f.bits,1074);}
  else assert.equal(Number(f.n)*2**(-f.bits),x);
}
assert.throws(()=>enclosure(-1n,2n));assert.throws(()=>enclosure(1n,0n));
assert.throws(()=>binaryFraction(-1));assert.throws(()=>binaryFraction(NaN));
console.log(JSON.stringify({likelihoodChecks,enclosureChecks,floatFixtures:8,invalidFixtures:4}));
