import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {acquisitionBelief,acquisitionChoice,TYPES} from './interval-acquisition.mjs';
import {lookaheadChoice} from './interval-lookahead.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'interval-lookahead-experiment.json'),d=JSON.parse(raw);
for(const [p,h] of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
const replay=execFileSync(process.execPath,[dir+'interval-lookahead-experiment.mjs'],{maxBuffer:64*1024*1024});assert(raw.equals(replay));
let costChecks=0,scoreChecks=0,asofChoices=0,asofForecasts=0,copyChecks=0;
const loss=(p,y)=>p.reduce((s,v,i)=>s+(v-Number(i===y))**2,0);
for(const r of d.records)for(const [policy,arm] of Object.entries(r.arms)){
  const history=r.initial.map(s=>({action:s.action.slice(),outcome:s.outcome}));let area=0;
  assert.equal(arm.trace.length,6);
  for(const [step,s] of arm.trace.entries()){
    assert.equal(s.cost,2);assert.equal(s.credit,4+2*step);
    const [kind,t,slot]=s.action;assert.equal(kind,1);assert(TYPES.includes(t));
    assert.equal(slot,history.filter(h=>h.action[0]===1&&h.action[1]===t).length);
    if(r.mask&(1<<TYPES.indexOf(t))){assert.equal(s.outcome,history[TYPES.indexOf(t)].outcome);copyChecks++;}
    if(r.index<2){
      const b=acquisitionBelief(history);b.forecast.forEach((p,i)=>assert(Math.abs(p-s.forecast[i])<1e-12));asofForecasts++;
      if(policy!=='random'){assert.deepEqual(policy==='lookahead'?lookaheadChoice(history):acquisitionChoice(history,policy,()=>{throw Error('No random access');}),s.action);asofChoices++;}
    }
    area+=s.cost*loss(s.forecast,r.truth%4)/12;history.push({action:s.action.slice(),outcome:s.outcome});
  }
  assert.equal(arm.credits,4+arm.trace.reduce((s,t)=>s+t.cost,0));assert.equal(arm.credits,16);costChecks++;
  assert(Math.abs(area-arm.area_brier)<1e-12);assert(Math.abs(loss(arm.final,r.truth%4)-arm.final_brier)<1e-12);scoreChecks+=2;
}
const componentTests=JSON.parse(execFileSync(process.execPath,[dir+'test-interval-acquisition.mjs']));
const lookaheadTests=JSON.parse(execFileSync(process.execPath,[dir+'test-interval-lookahead.mjs']));
const out={scope:'Complete deterministic replay; all arm cost/score/copy checks; sampled prefixes reconstructed without hidden-world fields; independent quadrature component tests',inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-interval-lookahead.mjs')).digest('hex'),replay:true,costChecks,scoreChecks,asofChoices,asofForecasts,copyChecks,componentTests,lookaheadTests,passed:d.gates.filter(g=>g.passed).length,total:d.gates.length,failed:d.gates.filter(g=>!g.passed)};
writeFileSync(dir+'interval-lookahead-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({...out,failedCount:out.failed.length,failed:undefined},null,2));


