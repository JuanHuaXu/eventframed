import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root=path.resolve(import.meta.dirname,'..');
const cases={
  early160:{change:160,noise:.05,delay:0,known:true},
  late360:{change:360,noise:.05,delay:0,known:true},
  late400:{change:400,noise:.05,delay:0,known:true},
  noisy256:{change:256,noise:.20,delay:0,known:true},
  delayed256:{change:256,noise:.05,delay:32,known:true},
  skew256:{change:256,noise:.05,delay:0,known:true,skew:true},
  stable20:{change:512,noise:.20,delay:0,stable:true},
  majority_ood:{change:256,noise:.05,delay:0,ood:true},
};
const avg=a=>a.reduce((s,x)=>s+x,0)/a.length;
const near=(a,b,label)=>assert.ok(Number.isFinite(a)&&Math.abs(a-b)<1e-9,`${label}: ${a} != ${b}`);
function interval(rows,candidate,control,field){
  const fit=Array.from({length:16},(_,f)=>avg(rows.filter(r=>r.Fit===f)
    .map(r=>r.Arms[control][field]-r.Arms[candidate][field])));
  const mean=avg(fit),variance=fit.reduce((s,x)=>s+(x-mean)**2,0)/15;
  return {mean,low:mean-3.5*Math.sqrt(variance/16),high:mean+3.5*Math.sqrt(variance/16)};
}
function truth(name,x,clock){
  const c=cases[name];
  let bit=Boolean(x&64)!==Boolean(x&128);
  bit=bit!==Boolean(x&256);
  if(clock>=c.change&&!c.stable)bit=c.ood?[1,2,4].filter(b=>x&b).length>=2:Boolean(x&4);
  return bit?1-c.noise:c.noise;
}
function altProbability(h,x){
  let bit;
  if(h<=9)bit=Boolean(x&(1<<(h-1)));
  else bit=(Boolean(x&1)!==Boolean(x&2))!==Boolean(x&4);
  return bit?.95:.05;
}
function recover(expected,c){
  if(!c.known||c.noise>=.20)return {delay:-1,missed:false};
  let consecutive=0;
  for(let clock=c.change+31;clock<512;clock++){
    const mean=avg(expected.slice(clock-31,clock+1));
    consecutive=mean<=.12?consecutive+1:0;
    if(consecutive===16)return {delay:clock-c.change,missed:false};
  }
  return {delay:512-c.change,missed:true};
}
function monitorStep(logs,origin,event,clock,alert){
  const ratios=Array.from({length:10},(_,j)=>{
    const p=altProbability(j+1,event.X),q=event.BaseP;
    return event.Y?Math.log(p)-Math.log(q):Math.log1p(-p)-Math.log1p(-q);
  });
  let maximum=-Infinity;
  for(let j=0;j<8;j++){
    if(origin>=j*64)for(let h=0;h<10;h++)logs[j][h]+=ratios[h];
    for(const value of logs[j])maximum=Math.max(maximum,value);
  }
  let total=0;
  for(const row of logs)for(const value of row)total+=Math.exp(value-maximum);
  assert.ok(Number.isFinite(total)&&total>0);
  if(alert<0&&maximum+Math.log(total/80)>=Math.log(100))return clock;
  return alert;
}

const result={splits:{}};
for(const split of ['design','confirmation']){
  const file=path.join(root,`docs/experiments/mmm-observation-responsive-v1-${split}.jsonl.gz`);
  const reader=readline.createInterface({input:fs.createReadStream(file).pipe(zlib.createGunzip())});
  const byCase=new Map(Object.keys(cases).map(c=>[c,[]])),ids=new Set();
  let manifest,count=0;
  for await(const line of reader){
    const row=JSON.parse(line);
    if(!manifest){
      manifest=row;
      assert.equal(row.kind,'manifest');
      assert.equal(row.split,split);
      assert.equal(row.seedBase,split==='design'?2026102701:2026102702);
      assert.equal(row.fitOffset,split==='design'?1800:1900);
      assert.equal(row.trialsPerCase,256);
      assert.equal(row.budget,128);
      assert.equal(row.monitorStarts,8);
      assert.equal(row.alternatives,10);
      assert.equal(row.threshold,100);
      assert.equal(row.baseRate,.20);
      assert.equal(row.burstRate,.60);
      assert.equal(row.burstClocks,64);
      assert.ok(row.stateBytes>0&&row.stateBytes<4096);
      for(const [source,expected] of Object.entries(row.hashes)){
        const actual=crypto.createHash('sha256').update(fs.readFileSync(path.join(root,source))).digest('hex');
        assert.equal(actual,expected,`source changed: ${source}`);
      }
      continue;
    }
    assert.equal(row.Kind,'trial');
    assert.equal(row.Split,split);
    const c=cases[row.Scenario];assert.ok(c);
    assert.ok(row.Fit>=0&&row.Fit<16&&row.Stream>=0&&row.Stream<16);
    const id=`${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!ids.has(id));ids.add(id);
    assert.equal(row.Tape.length,512);
    assert.equal(row.Arms.length,6);
    const postStart=c.stable?256:c.change;
    const baseByContext=new Map();
    const summary={Fit:row.Fit,Arms:[],SkewRate:0};
    for(let clock=0;clock<512;clock++){
      const e=row.Tape[clock];
      assert.ok(Number.isInteger(e.X)&&e.X>=0&&e.X<512);
      assert.ok(Number.isFinite(e.BaseP)&&e.BaseP>0&&e.BaseP<1);
      if(baseByContext.has(e.X))near(e.BaseP,baseByContext.get(e.X),`${id}/base/${e.X}`);
      else baseByContext.set(e.X,e.BaseP);
      if(c.skew)summary.SkewRate+=Number(Boolean(e.X&4))/512;
      near(e.PTrue,truth(row.Scenario,e.X,clock),`${id}/truth/${clock}`);
    }
    for(const [a,wrapped] of row.Arms.entries()){
      const arm=wrapped.Data.Data;
      assert.equal(arm.Policy,['random','uncertainty','learned_disagreement'][a%3]);
      assert.equal(arm.Ticks.length,512);
      let full=0,post256=0,early256=0,fullExpected=0,postExpected=0,firstExpected=0,postRealized=0;
      let nominated=0,nominatedPost=0,correctPost=0,arrived=0,pending=0;
      let preRequests=0,postRequests=0,postArrived=0,alert=-1;
      const selected=new Set(),delivered=new Set(),expected=[];
      const logs=Array.from({length:8},()=>Array(10).fill(0));
      for(let clock=0;clock<512;clock++){
        const e=row.Tape[clock],tick=arm.Ticks[clock];
        assert.ok(Number.isFinite(tick.P)&&tick.P>0&&tick.P<1);
        assert.ok(tick.Alternate>=1&&tick.Alternate<=10);
        const score=(tick.P-Number(e.Y))**2;
        const proper=e.PTrue*(1-e.PTrue)+(tick.P-e.PTrue)**2;
        expected.push(proper);
        full+=score/512;fullExpected+=proper/512;
        if(clock>=256){post256+=score/256;if(clock<320)early256+=score/64;}
        if(clock>=postStart){
          postExpected+=proper/(512-postStart);
          postRealized+=score/(512-postStart);
          if(clock<postStart+64)firstExpected+=proper/64;
        }
        if(tick.Selected){
          selected.add(clock);nominated++;
          if(clock<c.change)preRequests++;else postRequests++;
          if(clock>=256){nominatedPost++;if(c.known&&tick.Alternate===3)correctPost++;}
          if(!e.Missing&&clock+c.delay>=512)pending++;
        }
        for(const origin of tick.Delivered??[]){
          assert.ok(origin<=clock&&selected.has(origin),'future or unrequested label');
          assert.ok(!row.Tape[origin].Missing,'missing label delivered');
          assert.equal(origin+c.delay,clock,'wrong delivery clock');
          assert.ok(!delivered.has(origin),'duplicate label');
          delivered.add(origin);arrived++;
          if(origin>=c.change)postArrived++;
          if(a>=3)alert=monitorStep(logs,origin,row.Tape[origin],clock,alert);
        }
      }
      for(const origin of selected){
        if(!row.Tape[origin].Missing&&origin+c.delay<512)assert.ok(delivered.has(origin));
      }
      assert.equal(nominated,128);
      assert.equal(arm.Nominated,nominated);
      assert.equal(arm.NominatedPost,nominatedPost);
      assert.equal(arm.CorrectPost,correctPost);
      assert.equal(arm.Arrived,arrived);
      assert.equal(arm.Pending,pending);
      assert.equal(wrapped.AlertAt,alert);
      assert.equal(wrapped.MonitorUpdates,a>=3?arrived:0);
      near(full,arm.FullBrier,`${id}/${a}/full`);
      near(post256,arm.PostBrier,`${id}/${a}/post256`);
      near(early256,arm.EarlyBrier,`${id}/${a}/early256`);
      near(fullExpected,wrapped.Data.FullExpected,`${id}/${a}/fullExpected`);
      near(postExpected,wrapped.Data.PostExpected,`${id}/${a}/postExpected`);
      near(firstExpected,wrapped.Data.FirstExpected,`${id}/${a}/firstExpected`);
      near(postRealized,wrapped.Data.PostRealized,`${id}/${a}/postRealized`);
      const rec=recover(expected,c);
      assert.equal(wrapped.Data.RecoveryDelay,rec.delay);
      assert.equal(wrapped.Data.MissedRecovery,rec.missed);
      summary.Arms.push({FullExpected:fullExpected,PostExpected:postExpected,
        FirstExpected:firstExpected,PostRealized:postRealized,
        RecoveryDelay:rec.delay,MissedRecovery:Number(rec.missed),Arrived:arrived,
        AlertAt:alert,PreRequests:preRequests,PostRequests:postRequests,PostArrived:postArrived});
    }
    byCase.get(row.Scenario).push(summary);count++;
  }
  assert.ok(manifest);
  assert.equal(count,2048);
  const splitReport={trajectories:count,cases:{},pass:true};
  let strongKnown=0;
  for(const [name,c] of Object.entries(cases)){
    const rows=byCase.get(name);assert.equal(rows.length,256);
    const mean=field=>Array.from({length:6},(_,a)=>avg(rows.map(r=>r.Arms[a][field])));
    const full=mean('FullExpected'),post=mean('PostExpected'),first=mean('FirstExpected');
    const realized=mean('PostRealized'),rec=mean('RecoveryDelay'),miss=mean('MissedRecovery');
    const arrived=mean('Arrived'),preReq=mean('PreRequests'),postReq=mean('PostRequests');
    const postArrived=mean('PostArrived');
    const alerts=Array.from({length:6},(_,a)=>rows.filter(r=>r.Arms[a].AlertAt>=0).length/256);
    const preAlerts=Array.from({length:6},(_,a)=>rows.filter(r=>r.Arms[a].AlertAt>=0&&r.Arms[a].AlertAt<c.change).length/256);
    const detected=Array.from({length:6},(_,a)=>rows.filter(r=>r.Arms[a].AlertAt>=c.change).length/256);
    const alertDelays=Array.from({length:6},(_,a)=>{
      const d=rows.map(r=>r.Arms[a].AlertAt).filter(x=>x>=c.change).map(x=>x-c.change).sort((x,y)=>x-y);
      return d.length?d[Math.floor(d.length/2)]:-1;
    });
    const gain=(a,b,field)=>interval(rows,a,b,field);
    const checks={arrival:Math.abs(arrived[5]-arrived[3])<=2&&Math.abs(arrived[5]-arrived[4])<=2};
    if(c.known){
      checks.strong=[3,4].every(i=>{const g=gain(5,i,'PostExpected');return g.mean>=.01&&g.low>0;});
      if(checks.strong)strongKnown++;
    }
    if(name==='late360'||name==='late400'){
      const g=gain(5,2,'PostExpected');
      checks.late=g.mean>=.01&&g.low>0&&rec[2]-rec[5]>=10&&miss[2]-miss[5]>=.10;
    }
    if(['early160','noisy256','delayed256','skew256'].includes(name)){
      checks.protect=gain(5,2,'PostExpected').low>-.01&&first[5]-first[2]<=.005;
      if(name==='early160'||name==='skew256')checks.recovery=gain(5,2,'RecoveryDelay').low>-5;
    }
    if(c.stable)checks.stable=alerts[5]<=.05&&[2,3,4].every(i=>gain(5,i,'FullExpected').low>-.01);
    if(c.ood)checks.ood=[2,3,4].every(i=>gain(5,i,'PostExpected').low>-.01);
    if(c.skew)checks.skew=avg(rows.map(r=>r.SkewRate))>=.08&&avg(rows.map(r=>r.SkewRate))<=.12;
    const pass=Object.entries(checks).filter(([k])=>k!=='strong').every(([,v])=>v);
    splitReport.cases[name]={full,post,first,realized,recovery:rec,miss,arrived,
      preReq,postReq,postArrived,alerts,preAlerts,detected,alertDelays,
      gainVsOld:gain(5,2,'PostExpected'),gainVsRandom:gain(5,3,'PostExpected'),
      gainVsUncertainty:gain(5,4,'PostExpected'),checks,pass};
    splitReport.pass&&=pass;
  }
  splitReport.strongKnown=strongKnown;
  splitReport.pass&&=strongKnown>=5;
  result.splits[split]=splitReport;
}
result.efficacyPass=result.splits.design.pass&&result.splits.confirmation.pass;
console.log(JSON.stringify(result,null,2));
