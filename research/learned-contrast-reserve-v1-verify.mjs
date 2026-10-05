import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root=path.resolve(import.meta.dirname,'..');
const cases={
  early_bit2:{change:128,noise:.05,delay:0,known:true},
  late_bit2:{change:384,noise:.05,delay:0,known:true},
  noisy_bit2:{change:256,noise:.20,delay:0,known:true},
  delayed_bit2:{change:256,noise:.05,delay:32,known:true},
  skew_bit2:{change:256,noise:.05,delay:0,known:true},
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

const result={splits:{}};
for(const split of ['design','confirmation']){
  const file=path.join(root,`docs/experiments/mmm-learned-contrast-reserve-v1-${split}.jsonl.gz`);
  const reader=readline.createInterface({input:fs.createReadStream(file).pipe(zlib.createGunzip())});
  const byCase=new Map(Object.keys(cases).map(c=>[c,[]])),ids=new Set();
  let manifest,count=0;
  for await(const line of reader){
    const row=JSON.parse(line);
    if(!manifest){
      manifest=row;
      assert.equal(row.kind,'manifest');
      assert.equal(row.split,split);
      assert.equal(row.seedBase,split==='design'?2026102601:2026102602);
      assert.equal(row.fitOffset,split==='design'?1600:1700);
      assert.equal(row.trialsPerCase,256);
      assert.equal(row.budget,128);
      assert.deepEqual(row.stageBudgets,[80,48]);
      assert.ok(row.stateBytes>0&&row.stateBytes<4096);
      for(const [source,expected] of Object.entries(row.hashes)){
        const actual=crypto.createHash('sha256').update(fs.readFileSync(path.join(root,source))).digest('hex');
        assert.equal(actual,expected,`source changed: ${source}`);
      }
      continue;
    }
    assert.equal(row.Kind,'trial');
    assert.equal(row.Split,split);
    const c=cases[row.Scenario]; assert.ok(c);
    assert.ok(row.Fit>=0&&row.Fit<16&&row.Stream>=0&&row.Stream<16);
    const id=`${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!ids.has(id)); ids.add(id);
    assert.equal(row.Tape.length,512);
    assert.equal(row.Arms.length,6);
    const postStart=c.stable?256:c.change;
    const summary={Fit:row.Fit,Arms:[]};
    for(let clock=0;clock<512;clock++){
      const e=row.Tape[clock];
      assert.ok(Number.isInteger(e.X)&&e.X>=0&&e.X<512);
      near(e.PTrue,truth(row.Scenario,e.X,clock),`${id}/truth/${clock}`);
    }
    for(const [a,wrapped] of row.Arms.entries()){
      const arm=wrapped.Data;
      assert.equal(arm.Policy,['random','uncertainty','learned_disagreement'][a%3]);
      assert.equal(arm.Ticks.length,512);
      let full=0,post256=0,early256=0,fullExpected=0,postExpected=0,firstExpected=0,postRealized=0;
      let nominated=0,nominatedPost=0,correctPost=0,arrived=0,pending=0,stage1=0,postArrived=0;
      const selected=new Set(),delivered=new Set(),expected=[];
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
          if(clock<384)stage1++;
          if(clock>=256){nominatedPost++;if(c.known&&tick.Alternate===3)correctPost++;}
          if(!e.Missing&&clock+c.delay>=512)pending++;
        }
        for(const origin of tick.Delivered??[]){
          assert.ok(origin<=clock&&selected.has(origin),'future or unrequested label');
          assert.ok(!row.Tape[origin].Missing,'missing label delivered');
          assert.equal(origin+c.delay,clock,'wrong delivery clock');
          assert.ok(!delivered.has(origin),'duplicate label');
          delivered.add(origin);arrived++;
          if(origin>=384)postArrived++;
        }
      }
      for(const origin of selected){
        if(!row.Tape[origin].Missing&&origin+c.delay<512)assert.ok(delivered.has(origin));
      }
      assert.equal(nominated,128);
      if(a>=3)assert.equal(stage1,80);
      assert.equal(arm.Nominated,nominated);
      assert.equal(arm.NominatedPost,nominatedPost);
      assert.equal(arm.CorrectPost,correctPost);
      assert.equal(arm.Arrived,arrived);
      assert.equal(arm.Pending,pending);
      near(full,arm.FullBrier,`${id}/${a}/full`);
      near(post256,arm.PostBrier,`${id}/${a}/post256`);
      near(early256,arm.EarlyBrier,`${id}/${a}/early256`);
      near(fullExpected,wrapped.FullExpected,`${id}/${a}/fullExpected`);
      near(postExpected,wrapped.PostExpected,`${id}/${a}/postExpected`);
      near(firstExpected,wrapped.FirstExpected,`${id}/${a}/firstExpected`);
      near(postRealized,wrapped.PostRealized,`${id}/${a}/postRealized`);
      const rec=recover(expected,c);
      assert.equal(wrapped.RecoveryDelay,rec.delay);
      assert.equal(wrapped.MissedRecovery,rec.missed);
      summary.Arms.push({FullExpected:fullExpected,PostExpected:postExpected,
        FirstExpected:firstExpected,PostRealized:postRealized,
        RecoveryDelay:rec.delay,MissedRecovery:Number(rec.missed),Arrived:arrived,
        Stage1:stage1,PostArrived:postArrived});
    }
    byCase.get(row.Scenario).push(summary);count++;
  }
  assert.ok(manifest);
  assert.equal(count,1792);
  const splitReport={trajectories:count,cases:{},pass:true};
  let strongKnown=0;
  for(const [name,c] of Object.entries(cases)){
    const rows=byCase.get(name);assert.equal(rows.length,256);
    const mean=field=>Array.from({length:6},(_,a)=>avg(rows.map(r=>r.Arms[a][field])));
    const full=mean('FullExpected'),post=mean('PostExpected'),first=mean('FirstExpected');
    const realized=mean('PostRealized'),rec=mean('RecoveryDelay'),miss=mean('MissedRecovery');
    const arrived=mean('Arrived'),stage1=mean('Stage1'),postArrived=mean('PostArrived');
    const gain=(a,b,field)=>interval(rows,a,b,field);
    const checks={arrival:Math.abs(arrived[5]-arrived[3])<=2&&Math.abs(arrived[5]-arrived[4])<=2};
    if(c.known){
      checks.strong=[3,4].every(i=>{const g=gain(5,i,'PostExpected');return g.mean>=.01&&g.low>0;});
      if(checks.strong)strongKnown++;
    }
    if(name==='late_bit2'){
      const g=gain(5,2,'PostExpected');
      checks.late=rec[2]-rec[5]>=10&&miss[2]-miss[5]>=.15&&g.mean>=.01&&g.low>0;
    }
    if(['early_bit2','noisy_bit2','delayed_bit2','skew_bit2'].includes(name)){
      checks.protect=gain(5,2,'PostExpected').low>-.01&&first[5]-first[2]<=.005;
      if(name==='early_bit2'||name==='skew_bit2')
        checks.recovery=gain(5,2,'RecoveryDelay').low>-5;
    }
    if(c.stable)checks.stable=[2,3,4].every(i=>gain(5,i,'FullExpected').low>-.01);
    if(c.ood)checks.ood=[2,3,4].every(i=>gain(5,i,'PostExpected').low>-.01);
    const pass=Object.entries(checks).filter(([k])=>k!=='strong').every(([,v])=>v);
    splitReport.cases[name]={full,post,first,realized,recovery:rec,miss,arrived,stage1,postArrived,
      gainVsOld:gain(5,2,'PostExpected'),gainVsRandom:gain(5,3,'PostExpected'),
      gainVsUncertainty:gain(5,4,'PostExpected'),checks,pass};
    splitReport.pass&&=pass;
  }
  splitReport.strongKnown=strongKnown;
  splitReport.pass&&=strongKnown>=4;
  result.splits[split]=splitReport;
}
result.efficacyPass=result.splits.design.pass&&result.splits.confirmation.pass;
console.log(JSON.stringify(result,null,2));
