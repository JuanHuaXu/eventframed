import assert from 'node:assert/strict';
export const SCALE=1n<<80n;
export const DEN=8n*100n**10n;
export function enclosure(n,d){
  assert(n>=0n&&d>0n);const scaled=n*SCALE;return [scaled/d,(scaled+d-1n)/d];
}
export function binaryFraction(x){
  assert(Number.isFinite(x)&&x>=0);
  if(x===0)return {n:0n,bits:0};
  const b=new DataView(new ArrayBuffer(8));b.setFloat64(0,x);const v=b.getBigUint64(0),raw=Number((v>>52n)&2047n);
  const n=(v&((1n<<52n)-1n))+(raw?1n<<52n:0n),e=(raw?raw-1023:1-1023)-52;
  return e>=0?{n:n<<BigInt(e),bits:0}:{n,bits:-e};
}
const powers=Array.from({length:101},(_,p)=>Array.from({length:7},(_,k)=>BigInt(p)**BigInt(k)));
export function exactJoint(pattern,counts,noise,mask){
  assert(Number.isInteger(noise)&&noise>0&&noise<50);
  let degree=4;
  for(let j=0;j<4;j++){
    if(mask&(1<<j)){if(counts[2*j+1-((pattern>>j)&1)]>0)return [0n,0n,0n,0n];}
    else degree+=counts[2*j]+counts[2*j+1];
  }
  const joint=[0n,0n,0n,0n],pad=100n**BigInt(10-degree);
  for(let h=0;h<8;h++){
    let numerator=1n;
    for(let j=0;j<4;j++){
      const truth=j===3?((h&1)^((h>>2)&1)):((h>>j)&1),error=j===2?1:noise,p=truth?100-error:error;
      numerator*=BigInt((pattern>>j)&1?p:100-p);
      if(!(mask&(1<<j)))numerator*=powers[100-p][counts[2*j]]*powers[p][counts[2*j+1]];
    }
    joint[h%4]+=numerator*pad;
  }
  return joint;
}
