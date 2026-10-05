import assert from 'node:assert/strict';
import {fitDependence} from './query-dependence.mjs';
export function observedInputs(origins,source){
  assert(Array.isArray(origins)&&origins.length<=63&&new Set(origins).size===origins.length);
  const values=new Set();
  for(const j of origins){
    assert(Number.isInteger(j)&&j>=-16&&j<160);
    let x;
    if(j<0)x=source.Initial[j+16].Bits;
    else {const s=source.Steps[j];assert(!s.Missing&&Number.isInteger(s.Delay)&&s.Delay>=0&&j+s.Delay<=160);x=s.X;}
    assert(Number.isInteger(x)&&x>=0&&x<512);values.add(x);
  }
  return values;
}
export function dependenceCell(p,x,observed){
  assert(Number.isFinite(p)&&p>0&&p<1&&Number.isInteger(x)&&x>=0&&x<512&&observed instanceof Set);
  return 2*Number(4*p*(1-p)>=.5)+Number(observed.has(x));
}
export function fitLocalDependence(rows){
  assert(Array.isArray(rows));
  const cells=Array.from({length:4},()=>[]);
  for(const r of rows){assert(Number.isInteger(r.cell)&&r.cell>=0&&r.cell<4);cells[r.cell].push({base:r.base,delta:r.delta,y:r.y,weight:r.weight});}
  return cells.map(rs=>({rows:rs.length,weight:rs.reduce((s,r)=>s+r.weight,0),...(rs.length?fitDependence(rs):{lambda:1,d0:null,d1:null})}));
}
