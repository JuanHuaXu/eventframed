import assert from 'node:assert/strict';

// Center at the neutral half mixture. Only arrived evidence supplies either
// sufficient statistic; pending samples are not zero-outcome observations.
export function directMixture(rows,{window=Infinity,currentCovariance=false}={}){
  assert(window===Infinity||Number.isInteger(window)&&window>0);
  return rows.map((r,t)=>{
    const admitted=[];
    for(let j=0;j<t;j++)if(!rows[j].missing&&j+rows[j].delay<=t)admitted.push(j);
    const origins=window===Infinity?admitted:admitted.slice(-window);
    let gram=1,cross=0;
    for(const j of origins){const s=rows[j],d=s.c-s.b,m=(s.b+s.c)/2;gram+=d*d;cross+=d*(Number(s.y)-m);}
    const d=r.c-r.b;if(currentCovariance)gram+=d*d;
    const w=Math.max(0,Math.min(1,.5+cross/gram));
    return{w,gram,cross,origins};
  });
}

export function testDirectMixture(){
  let checks=0;
  const rows=Array.from({length:96},(_,i)=>({b:.1+(i%7)*.08,c:.9-(i%5)*.1,y:i%3===0,delay:i%11,missing:i%7===0}));
  for(const window of [Infinity,64])for(const currentCovariance of [false,true]){
    const out=directMixture(rows,{window,currentCovariance});
    const poison=rows.map((r,i)=>({...r,y:i>=40||r.missing||i+r.delay>40?!r.y:r.y}));
    assert.deepEqual(out.slice(0,41),directMixture(poison,{window,currentCovariance}).slice(0,41));
    for(let t=0;t<rows.length;t++){
      const f=out[t];assert(f.w>=0&&f.w<=1);assert(f.origins.every(j=>j<t&&!rows[j].missing&&j+rows[j].delay<=t));assert(f.origins.length<=window);
      const available=Array.from({length:t},(_,j)=>j).filter(j=>!rows[j].missing&&j+rows[j].delay<=t);
      assert.deepEqual(f.origins,window===Infinity?available:available.slice(-window));
      const objective=w=>{
        let s=(w-.5)**2;
        for(const j of f.origins){const r=rows[j];s+=(r.b+w*(r.c-r.b)-Number(r.y))**2;}
        if(currentCovariance)s+=(rows[t].c-rows[t].b)**2*(w-.5)**2;
        return s;
      };
      for(let k=0;k<=100;k++){assert(objective(f.w)<=objective(k/100)+1e-12);checks++;}
    }
  }
  const pending=rows.map(r=>({...r,missing:true,y:NaN}));assert(directMixture(pending).every(f=>f.w===.5));
  const same=rows.map(r=>({...r,c:r.b}));assert(directMixture(same).every(f=>f.w===.5));
  const simple=[{b:0,c:1,y:true,delay:0,missing:false},{b:0,c:1,y:false,delay:0,missing:false}];
  assert.equal(directMixture(simple)[1].w,.75);
  assert.equal(directMixture(simple,{currentCovariance:true})[1].w,2/3);
  console.log(`direct mixture objective/as-of checks PASS: ${checks}`);
}
