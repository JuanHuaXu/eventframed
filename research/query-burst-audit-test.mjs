import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import {createInterface} from 'node:readline';
import {auditBurstRecord} from './query-burst-audit.mjs';

// Hand-built constant-forecast tapes isolate clock accounting from Go fitting.
for(const missing of [false,true]) {
 const raw={Phase:0,Case:0,Index:0,Schedule:Number(missing),Steps:Array.from({length:256},()=>({Missing:missing,Delay:0,Y:false,Q:.5,P:Array(15).fill(.5)}))};
 const r={Phase:0,Case:0,Index:0,Schedule:Number(missing),Results:[]};
 for(let arm=0;arm<7;arm++) {
  const result={Predictions:Array.from({length:256},()=>Array(5).fill(.5)),Queries:[],Fits:[],Arrivals:[],EmptyQueryClocks:[]};
  const paid=new Map();
  for(let t=0;t<256;t++) {
   const eligible=[];
   for(let j=-16;j<t;j++)if(j<0||!missing||(paid.has(j)&&paid.get(j)<=t))eligible.push(j);
   if(t%32===0)result.Fits.push({Clock:t,Origins:[eligible.slice(-64),eligible.slice(-32)]});
   if(arm===0)continue;
   if(!missing)result.Arrivals.push({Clock:t,Origin:t,Paid:false,Surprise:false,Mixer:true});
   for(const [j,arrival] of paid)if(arrival===t)result.Arrivals.push({Clock:t,Origin:j,Paid:true,Surprise:false,Mixer:j>=t-32});
   const pool=[];for(let j=Math.max(0,t-32);j<t;j++)if(missing&&!paid.has(j))pool.push(j);
   const periodic=t>0&&t%8===0;
   // With no surprising observations, forced clocks are explicitly known.
   const forced=(t>=29&&t<=31)||(t>=60&&t<=223&&t%32>=28)||(t>=245&&t<=248);
   if(pool.length===0&&(arm>=4||periodic))result.EmptyQueryClocks.push(t);
   if(pool.length>0&&(arm<4?periodic:forced)) {
    let j=pool[0];
    if(arm===1||arm===4){const hash=crypto.createHash('sha256').update(`0:0:0:${Number(missing)}:${t}`).digest();j=pool[Number((BigInt(hash.readUInt32BE(0))*BigInt(pool.length))>>32n)];}
    paid.set(j,t+1);result.Queries.push({Clock:t,Origin:j,Reveal:t+1});
   }
  }
  r.Results.push(result);
 }
 const checked=auditBurstRecord(r,raw);
 for(const values of checked.brier)for(const pair of values)assert.deepEqual(pair,[.25,.25]);
 for(let a=1;a<7;a++)assert.equal(checked.diagnostics[a].queries,missing?31:0);
 const bad=structuredClone(r);bad.Results[6].Fits[1].Origins[0]=[255];assert.throws(()=>auditBurstRecord(bad,raw));
 if(missing) {
  const late=structuredClone(r);late.Results[6].Queries[0].Reveal++;assert.throws(()=>auditBurstRecord(late,raw));
  const wrong=structuredClone(r);wrong.Results[4].Queries[0].Origin=255;assert.throws(()=>auditBurstRecord(wrong,raw));
 }
}
console.log('PASS: immediate and all-missing handcrafted tapes, exact .25 Brier, seven arms, mutation rejection');

if(process.argv[2]) {
 const reader=createInterface({input:fs.createReadStream(process.argv[2]),crlfDelay:Infinity});
 let checked=0;
 for await(const line of reader) {
  const raw=JSON.parse(line);if(!raw.Steps||raw.Schedule!==0||raw.Index!==0)continue;
  const r={Phase:raw.Phase,Case:raw.Case,Index:raw.Index,Schedule:raw.Schedule,Results:[]};
  const predictions=raw.Steps.map(s=>[0,1,2,3,12].map(a=>s.P[a]));
  const arrivals=raw.Steps.map((s,t)=>({Clock:t,Origin:t,Paid:false,Surprise:(s.Y?predictions[t][4]:1-predictions[t][4])<.2,Mixer:true}));
  for(let arm=0;arm<7;arm++)r.Results.push({Predictions:predictions,Fits:raw.Fits,Queries:[],Arrivals:arm?arrivals:[],EmptyQueryClocks:arm===0?[]:Array.from({length:256},(_,t)=>t).filter(t=>arm>=4||(t>0&&t%8===0))});
  auditBurstRecord(r,raw);checked++;
  if(checked===3)break;
 }
 reader.close();assert.equal(checked,3);
 console.log('PASS: direct-probability mixer replay against three original Go immediate-feedback trajectories');
}
