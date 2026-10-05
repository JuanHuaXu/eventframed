import assert from 'node:assert/strict';
export function binaryEntropy(p){
  assert(Number.isFinite(p)&&p>=0&&p<=1);
  return p===0||p===1?0:-p*Math.log(p)-(1-p)*Math.log1p(-p);
}
export function predictiveInformation(base,conditional,mass,weights){
  assert(base.length>0&&base.length<=512&&weights.length===base.length&&conditional.length===2&&conditional.every(v=>v.length===base.length));
  assert(mass.length===2&&mass.every(p=>Number.isFinite(p)&&p>=0&&p<=1)&&Math.abs(mass[0]+mass[1]-1)<1e-10);
  assert(weights.every(w=>Number.isInteger(w)&&w>=0));const total=weights.reduce((s,w)=>s+w,0);assert(total>0);
  let gain=0;
  for(let x=0;x<base.length;x++){
    const p=base[x],a=conditional[0][x],b=conditional[1][x],h=binaryEntropy(p),ha=binaryEntropy(a),hb=binaryEntropy(b);
    assert(Math.abs(mass[0]*a+mass[1]*b-p)<1e-10);
    gain+=(h-mass[0]*ha-mass[1]*hb)*weights[x]/total;
  }
  assert(Number.isFinite(gain)&&gain>=-1e-10&&gain<=binaryEntropy(mass[1])+1e-10);return gain;
}
