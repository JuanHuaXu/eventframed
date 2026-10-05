// Consumed-artifact diagnosis, NOT independent confirmation. Frozen published
// forecasts are inputs. This tests no observer policy or production lifecycle.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const raw=fs.readFileSync(process.argv[2]),a=JSON.parse(raw);
assert.equal(a.Version,'soft-learners-v119');assert.equal(a.Records.length,2688);
const variants=[[0,1,2,3],[0,1,2,3,4,5],[0,1,2,3,8,9]];
const mean=xs=>xs.reduce((s,x)=>s+x,0)/xs.length;
const interval=xs=>{assert.equal(xs.length,32);const m=mean(xs),se=Math.sqrt(xs.reduce((s,x)=>s+(x-m)**2,0)/31/32);return {mean:m,lower:m-3.5*se,upper:m+3.5*se};};

function weightsBefore(steps,clock,arms){
  const prior=arms.map((_,j)=>j===0?.95:.05/(arms.length-1));
  let w=prior.slice();
  // Rebuild from the prior, admitting each eligible origin once. Unknown
  // labels have unit emission, with a transition after every issued frame.
  for(let j=0;j<clock;j++){
    const s=steps[j];
    if(!s.Missing&&j+s.Delay<=clock){
      let total=0;
      for(let k=0;k<arms.length;k++){const p=s.P[arms[k]];w[k]*=s.Y?p:1-p;total+=w[k];}
      assert(total>0);for(let k=0;k<w.length;k++)w[k]/=total;
    }
    for(let k=0;k<w.length;k++)w[k]=.999*w[k]+.001*prior[k];
  }
  return w;
}

const records=[];let forecasts=0,asOfChecks=0;
for(const r of a.Records){
  const scores=variants.map(()=>[0,0]);
  for(let clock=0;clock<256;clock++)for(let v=0;v<variants.length;v++){
    const arms=variants[v],w=weightsBefore(r.Steps,clock,arms),p=w.reduce((s,x,k)=>s+x*r.Steps[clock].P[arms[k]],0);
    assert(p>0&&p<1&&Math.abs(w.reduce((s,x)=>s+x,0)-1)<1e-12);
    const q=r.Steps[clock].Q,loss=(p-q)**2+q*(1-q);
    scores[v][0]+=loss/256;if(clock>=192)scores[v][1]+=loss/64;forecasts++;
  }
  // Poison current/future/unavailable outcomes, keeping the already-issued
  // forecast tape fixed. This does not test generation of that tape itself.
  if(r.Index===0)for(const clock of [0,32,160,255]){
    const changed=r.Steps.map((s,j)=>({...s,Y:(j>=clock||s.Missing||j+s.Delay>clock)?!s.Y:s.Y,Q:0}));
    for(const arms of variants){assert.deepEqual(weightsBefore(r.Steps,clock,arms),weightsBefore(changed,clock,arms));asOfChecks++;}
  }
  records.push({Phase:r.Phase,Case:r.Case,Index:r.Index,Schedule:r.Schedule,scores});
}
const groups=[];
for(const phase of [0,1])for(let scenario=0;scenario<21;scenario++)for(const schedule of [0,1])for(const segment of [0,1]){
  const rs=records.filter(r=>r.Phase===phase&&r.Case===scenario&&r.Schedule===schedule);
  groups.push({phase,scenario,schedule,segment,meanBrier:variants.map((_,v)=>mean(rs.map(r=>r.scores[v][segment]))),variationalGainVsFour:interval(rs.map(r=>r.scores[0][segment]-r.scores[2][segment])),variationalGainVsMAP:interval(rs.map(r=>r.scores[1][segment]-r.scores[2][segment]))});
}
assert.equal(forecasts,2688*256*3);assert.equal(asOfChecks,1008);
process.stdout.write(JSON.stringify({status:'CONSUMED_DIAGNOSTIC_ONLY',artifactSHA256:crypto.createHash('sha256').update(raw).digest('hex'),sourceSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),variants,prior:'generic64=.95; others share .05',transition:.001,forecasts,asOfChecks,groups},null,2)+'\n');
