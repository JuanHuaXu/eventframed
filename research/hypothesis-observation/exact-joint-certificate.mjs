import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {addTiePolicies} from './tie-policy.mjs';
import {pathWeights} from './exact-regime.mjs';
import {SCALE,DEN,enclosure,binaryFraction,exactJoint} from './exact-joint.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'joint-policy-two-tie_entropy.json'),d=JSON.parse(raw);
const modelRaw=readFileSync(dir+'count-planning-exact.json'),model=JSON.parse(modelRaw);
const frozen=JSON.parse(readFileSync(dir+'risk-budget-evaluation.json'));
assert.equal(createHash('sha256').update(modelRaw).digest('hex'),frozen.inputSHA256);
for(const[f,h]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(dir+f)).digest('hex'),h);
addTiePolicies(model);
const best=d.rounds.reduce((a,b)=>a.lower>b.lower?a:b),weights=best.weights.map(v=>BigInt(Math.floor(v*1e9))),W=weights.reduce((a,v)=>a+v,0n);assert(W>0n);
const metricWeights={final:Array(80).fill(0n),area:Array(80).fill(0n)},controlWeights={};
for(const p of ['random','tie_entropy'])controlWeights[p]={final:Array(80).fill(0n),area:Array(80).fill(0n)};
let finalWeight=0n;
d.rows.forEach((r,i)=>{metricWeights[r.metric][r.g]+=weights[i];controlWeights[r.control][r.metric][r.g]+=weights[i];if(r.metric==='final')finalWeight+=weights[i];});
let oracleLo=0n,oracleHi=0n,controlRawHi=0n,controlNormHi=0n,statesChecked=0;
for(const root of model.roots){
  const states=root.states,index=new Map(states.map((s,i)=>[s.counts.join(','),i])),level=states.map(s=>s.counts.reduce((a,v)=>a+v,0));
  const paths=pathWeights(root,['random','tie_entropy']),lower=Array(states.length).fill(0n),upper=Array(states.length).fill(0n);
  for(let i=0;i<states.length;i++){
    const s=states[i],terminal=level[i]===6,metric=terminal?'final':'area',factor=terminal?1n:6n,v=[0n,0n,0n,0n];
    const fractions=s.forecast.map(binaryFraction),bits=Math.max(...fractions.map(f=>f.bits)),P=1n<<BigInt(bits),p=fractions.map(f=>f.n<<BigInt(bits-f.bits)),sumP=p.reduce((a,x)=>a+x,0n);
    const loss=den=>Array.from({length:4},(_,y)=>p.reduce((a,x,z)=>a+(x-(z===y?den:0n))**2n,0n));
    const rawLoss=loss(P),normLoss=loss(sumP);let rawControl=0n,normControl=0n;
    for(let g=0;g<80;g++){
      const world=frozen.results[g],joint=exactJoint(root.pattern,s.counts,Math.round(world.noise*100),world.mask);
      for(let y=0;y<4;y++)v[y]+=metricWeights[metric][g]*joint[y];
      let pw=0n;
      for(const policy of ['random','tie_entropy']){
        const count=(paths[policy][level[i]].get(i)||0)*4096;assert(Number.isSafeInteger(count));
        pw+=BigInt(count)*controlWeights[policy][metric][g];
      }
      if(pw){rawControl+=pw*joint.reduce((a,j,y)=>a+j*rawLoss[y],0n);normControl+=pw*joint.reduce((a,j,y)=>a+j*normLoss[y],0n);}
    }
    controlRawHi+=enclosure(rawControl,DEN*W*4096n*factor*P*P)[1];
    controlNormHi+=enclosure(normControl,DEN*W*4096n*factor*sumP*sumP)[1];
    const m=v.reduce((a,x)=>a+x,0n);
    if(m)[lower[i],upper[i]]=enclosure(m*m-v.reduce((a,x)=>a+x*x,0n),m*DEN*W*factor);
    statesChecked++;
  }
  const order=states.map((_,i)=>i).sort((a,b)=>level[b]-level[a]);
  for(const i of order)if(level[i]<6){
    let lo=null,hi=null;
    for(let a=0;a<4;a++){
      let l=0n,u=0n;for(let y=0;y<2;y++){const c=states[i].counts.slice();c[2*a+y]++;const j=index.get(c.join(','));assert(j!==undefined);l+=lower[j];u+=upper[j];}
      if(lo===null||l<lo)lo=l;if(hi===null||u<hi)hi=u;
    }
    lower[i]+=lo;upper[i]+=hi;
  }
  const start=index.get('0,0,0,0,0,0,0,0');oracleLo+=lower[start];oracleHi+=upper[start];
}
const controlHi=controlRawHi>controlNormHi?controlRawHi:controlNormHi,allowanceHi=enclosure(finalWeight,100n*W)[1],bound=oracleLo-controlHi-allowanceHi;
const files=['EXACT_JOINT_PROTOCOL.md','exact-joint.mjs','exact-joint-certificate.mjs','tie-policy.mjs','exact-regime.mjs'];
console.log(JSON.stringify({scope:'Exact rational source model and integer outward bounds; both raw and simplex-normalized binary forecast controls; no proof-assistant verification',sourceSHA256:createHash('sha256').update(raw).digest('hex'),hashes:Object.fromEntries(files.map(f=>[f,createHash('sha256').update(readFileSync(dir+f)).digest('hex')])),witnessRound:best.round,weights:weights.map(String),weightSum:String(W),statesChecked,scale:String(SCALE),oracleLower:String(oracleLo),oracleUpper:String(oracleHi),controlRawUpper:String(controlRawHi),controlNormalizedUpper:String(controlNormHi),allowanceUpper:String(allowanceHi),lowerNumerator:String(bound),positive:bound>0n,approximateLower:Number(bound)/Number(SCALE)},null,2));
