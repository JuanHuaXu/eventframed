import assert from 'node:assert/strict';
import {createReadStream, readFileSync} from 'node:fs';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';

// Consumed-data diagnostic, not confirmation. One frozen equal-prior comparison;
// do not multiply Bayes factors from overlapping windows. Each fit starts from
// the same model prior and conditions once on exactly its eligible labels.
const path=process.argv[2];
assert(path,'artifact path required');
const digest=createHash('sha256'),stream=createReadStream(path);
stream.on('data',b=>digest.update(b));
const lines=createInterface({input:stream,crlfDelay:Infinity});
const parity=Array.from({length:512},(_,m)=>m.toString(2).replaceAll('0','').length%2);
const prior=Array.from({length:512},(_,m)=>{const n=m.toString(2).replaceAll('0','').length;return (1/3)**n*(2/3)**(9-n);});
const n=new Uint16Array(512*512),y=new Uint16Array(n.length),agree=new Uint16Array(512);
const g=new Float64Array(512),b=new Float64Array(512);
function evidence(samples,query){
  n.fill(0);y.fill(0);agree.fill(0);g.fill(1);b.fill(1);
  for(let j=0;j<samples.length;j++){
    const s=samples[j],out=Number(s.Outcome);
    for(let m=0;m<512;m++){
      const k=m*512+(s.Bits&m),match=Number(parity[s.Bits&m]===out);
      g[m]*=((out?y[k]:n[k]-y[k])+.5)/(n[k]+1);
      b[m]*=((match?agree[m]:j-agree[m])+.5)/(j+1);
      n[k]++;y[k]+=out;agree[m]+=match;
    }
  }
  let z=0,num=0;
  for(let m=0;m<512;m++){
    const k=m*512+(query&m),ap=(agree[m]+.5)/(samples.length+1);
    z+=prior[m]*(.95*g[m]+.05*b[m]);
    num+=prior[m]*(.95*g[m]*(y[k]+.5)/(n[k]+1)+.05*b[m]*(parity[query&m]?ap:1-ap));
  }
  assert(z>0&&Number.isFinite(z));
  return {logZ:Math.log(z),p:num/z};
}
assert(Math.abs(evidence([],0).logZ)<1e-14);
assert(Math.abs(evidence([],0).p-.5)<1e-14);
assert(Math.abs(evidence([{Bits:0,Outcome:true}],0).logZ-Math.log(.5))<1e-14);
const mean=a=>a.reduce((s,x)=>s+x,0)/a.length;
function interval(a){const m=mean(a),se=Math.sqrt(a.reduce((s,x)=>s+(x-m)**2,0)/(a.length*(a.length-1)));return {mean:m,lower:m-3.5*se,upper:m+3.5*se};}
let header,fitChecks=0,maxNoChangeError=0;
const records=[];
for await(const line of lines){
  if(!header){header=JSON.parse(line);assert.equal(header.Version,'soft-learners-v120');continue;}
  const r=JSON.parse(line),weights=[[],[]],scores=Array.from({length:2},()=>[0,0]);
  assert.equal(r.Fits.length,8);assert.equal(r.Steps.length,256);
  for(let f=0;f<8;f++){
    const fit=r.Fits[f];assert.equal(fit.Clock,f*32);
    const available=Array.from({length:16},(_,i)=>i-16);
    for(let j=0;j<fit.Clock;j++)if(!r.Steps[j].Missing&&j+r.Steps[j].Delay<=fit.Clock)available.push(j);
    for(let w=0;w<2;w++){
      const origins=available.slice(-(w===0?64:32));assert.deepEqual(origins,fit.Origins[w]);
      const samples=origins.map(j=>j<0?r.Initial[j+16]:{Bits:r.Steps[j].X,Outcome:r.Steps[j].Y});
      const e=evidence(samples,r.Steps[fit.Clock].X);
      maxNoChangeError=Math.max(maxNoChangeError,Math.abs(e.p-r.Steps[fit.Clock].P[13+w]));
      assert(maxNoChangeError<1e-8,'no-change evidence/predictive mismatch');
      const d=e.logZ-fit.SegmentEvidence[w],weight=d>=0?Math.exp(-d)/(1+Math.exp(-d)):1/(1+Math.exp(d));
      assert(weight>=0&&weight<=1);weights[w].push(weight);fitChecks++;
      for(let t=fit.Clock;t<fit.Clock+32;t++){
        const s=r.Steps[t],p=weight*s.P[10+w]+(1-weight)*s.P[13+w];
        const loss=(p-s.Q)**2+s.Q*(1-s.Q);
        scores[w][0]+=loss/256;if(t>=192)scores[w][1]+=loss/64;
      }
    }
  }
  records.push({phase:r.Phase,case:r.Case,index:r.Index,schedule:r.Schedule,weights,scores,control:r.Metrics.map(m=>m.map(x=>x.Brier))});
}
assert.equal(records.length,2688);assert.equal(fitChecks,43008);
const gates=[],groups=[];
for(let ph=0;ph<2;ph++)for(let c=0;c<21;c++)for(let sch=0;sch<2;sch++){
  const rs=records.filter(r=>r.phase===ph&&r.case===c&&r.schedule===sch);
  assert.equal(rs.length,32);assert.deepEqual(rs.map(r=>r.index),Array.from({length:32},(_,i)=>i));
  for(let w=0;w<2;w++)for(let seg=0;seg<2;seg++){
    const meta={phase:ph,case:c,schedule:sch,window:w,segment:seg};
    groups.push({...meta,brier:mean(rs.map(r=>r.scores[w][seg])),parentBrier:mean(rs.map(r=>r.control[10+w][seg])),noChangeBrier:mean(rs.map(r=>r.control[13+w][seg])),segmentWeightByFit:Array.from({length:8},(_,f)=>mean(rs.map(r=>r.weights[w][f])))});
    for(const control of [2*w,2*w+1,12,13+w]){
      const v=interval(rs.map(r=>r.scores[w][seg]-r.control[control][seg]));
      gates.push({...meta,type:'nonharm',control,...v,pass:v.upper<=.01});
    }
    if(seg===1&&((c<9&&c%3!==0)||c>=19))for(const control of [2*w,12,13+w]){
      const v=interval(rs.map(r=>r.control[control][seg]-r.scores[w][seg]));
      gates.push({...meta,type:'gain',control,...v,pass:v.mean>=.005&&v.lower>0});
    }
  }
}
assert.equal(gates.length,1536);
const candidates=[0,1].map(window=>({window,types:Object.fromEntries(['nonharm','gain'].map(type=>{const gs=gates.filter(g=>g.window===window&&g.type===type);return [type,{passed:gs.filter(g=>g.pass).length,total:gs.length}];}))}));
process.stdout.write(JSON.stringify({scope:'Consumed v120 equal-prior static-atom diagnostic; no fresh confirmation',artifactSHA256:digest.digest('hex'),scriptSHA256:createHash('sha256').update(readFileSync(new URL(import.meta.url))).digest('hex'),records:records.length,fitChecks,maxNoChangeError,candidates,gates,groups},null,2)+'\n');
