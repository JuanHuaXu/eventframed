import assert from 'node:assert/strict';
const pc=Array.from({length:512},(_,x)=>x.toString(2).replaceAll('0','').length);
const prior=pc.map(k=>(1/3)**k*(2/3)**(9-k));
const half=[0],fact=[0];for(let n=1;n<=64;n++){half[n]=half[n-1]+Math.log(n-.5);fact[n]=fact[n-1]+Math.log(n);}
const beta=Array.from({length:65},(_,n)=>Array.from({length:n+1},(_,y)=>half[y]+half[n-y]-fact[n]));

// Direct integrated likelihood and predictive numerator. At <=64 labels,
// products stay representable; assert denominators rather than clipping them.
export function regimeReference(samples,queries){
 assert(samples.length<=64);for(const [x,y] of samples)assert(Number.isInteger(x)&&x>=0&&x<512&&typeof y==='boolean');for(const x of queries)assert(Number.isInteger(x)&&x>=0&&x<512);
 if(!samples.length)return{genericLog:0,booleanLog:0,logEvidence:0,genericMass:.95,predictions:queries.map(()=>.5)};
 let zg=0,zb=0;const nums=queries.map(()=>0),count=new Uint8Array(512),yes=new Uint8Array(512);
 for(let mask=0;mask<512;mask++){
  const touched=[];let agree=0;
  for(const [x,y] of samples){const key=x&mask;if(count[key]===0)touched.push(key);count[key]++;if(y)yes[key]++;if(Boolean(pc[key]%2)===y)agree++;}
  let lg=0;for(const key of touched)lg+=beta[count[key]][yes[key]];
  const g=prior[mask]*Math.exp(lg),b=prior[mask]*Math.exp(beta[samples.length][agree]);zg+=g;zb+=b;
  queries.forEach((x,i)=>{const key=x&mask,pg=(yes[key]+.5)/(count[key]+1),agreement=(agree+.5)/(samples.length+1),pb=pc[key]%2?agreement:1-agreement;nums[i]+=.95*g*pg+.05*b*pb;});
  for(const key of touched){count[key]=0;yes[key]=0;}
 }
 const z=.95*zg+.05*zb;assert(z>0&&Number.isFinite(z));return{genericLog:Math.log(zg),booleanLog:Math.log(zb),logEvidence:Math.log(z),genericMass:.95*zg/z,predictions:nums.map(n=>n/z)};
}
