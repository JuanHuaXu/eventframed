import assert from 'node:assert/strict';
const floating=new Set(['Issued','ObservedIssued','ExpertIssued','NoiseIssued','NoiseSnapshots','Template','Weights','Forecast','Values','IssuedBrier','IssuedPriority','Recovery','Brier','PriorityBrier','PacketUsefulness','PacketBias','Observed','Uncertainty','Information','EdgeCut']);
export function compare(a,b,tolerance=2e-10){
 const r={scalars:0,maxDefect:0,differences:0,first:null,examples:[]};
 function fail(path,x,y,reason){r.differences++;const z={path:path.join('.'),actual:x,reference:y,reason};r.first??=z;if(r.examples.length<8)r.examples.push(z)}
 function visit(x,y,path=[]){
  if(typeof x==='number'||typeof y==='number'){
   if(typeof x!=='number'||typeof y!=='number'||!Number.isFinite(x)||!Number.isFinite(y)){fail(path,x,y,'type/nonfinite');return}
   r.scalars++;const d=Math.abs(x-y);r.maxDefect=Math.max(r.maxDefect,d);
   if(path.some(p=>floating.has(p))){if(d>tolerance)fail(path,x,y,'float tolerance')}else if(x!==y)fail(path,x,y,'discrete number');return;
  }
  if(x===null||y===null||typeof x!=='object'||typeof y!=='object'){if(x!==y)fail(path,x,y,'discrete/type');return}
  if(Array.isArray(x)!==Array.isArray(y)){fail(path,x,y,'array/object');return}
  if(Array.isArray(x)){if(x.length!==y.length){fail(path,x.length,y.length,'array length');return}for(let i=0;i<x.length;i++)visit(x[i],y[i],[...path,String(i)]);return}
  const k=Object.keys(x).sort(),l=Object.keys(y).sort();if(JSON.stringify(k)!==JSON.stringify(l)){fail(path,k,l,'object keys');return}for(const p of k)visit(x[p],y[p],[...path,p]);
 }
 visit(a,b);return r;
}
export function selftest(){
 const a={Issued:[.5],Decisions:[{At:2,Members:[4],Options:[{Observed:.4}]}],PeakPending:1},copy=()=>structuredClone(a);let n=0;
 assert.equal(compare(a,a).differences,0);n++;
 let b=copy();b.Issued[0]+=1e-11;assert.equal(compare(b,a).differences,0);n++;
 for(const change of [b=>b.Issued[0]+=.1,b=>b.Issued[0]=NaN,b=>b.Issued[0]=Infinity,b=>b.Decisions[0].Members[0]=5,b=>b.Decisions[0].At+=1e-11,b=>delete b.PeakPending,b=>b.Issued.push(.5),b=>b.Decisions[0].Options[0].Observed=-.1]){b=copy();change(b);assert(compare(b,a).differences>0);n++}
 assert(compare({Issued:[NaN]},{Issued:[NaN]}).differences>0);n++;
 return n;
}
