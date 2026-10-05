import test from 'node:test';
import assert from 'node:assert/strict';
import {summarize} from './load-evaluate.mjs';
import {summarize as qualitySummary} from './quality-core.mjs';

function row(){const t={functional:true,frontier:50,completed:64,replay_completed:64,ledger_rows:128,capture_bytes:1000,init_bytes:4000,init_ns:1000,replay_ns:1000,overlaps:64};
 for(const[k,n]of[['recall',64],['feedback',64],['live_age',64],['capture',256]])t[k+'_ns']=Array(n).fill(10);
 t.recall_start_ns=Array.from({length:64},(_,i)=>i*20);t.recall_end_ns=t.recall_start_ns.map(x=>x+10);
 t.capture_start_ns=Array.from({length:256},(_,i)=>i*20);t.capture_end_ns=t.capture_start_ns.map(x=>x+10);
 t.candidate_counts=Array(64).fill(50);t.packed_counts=Array(64).fill(10);
 return{arm:'both',frontier:50,pair:0,wallMS:100,status:0,trace:t};}
test('complete raw trace accepted with explicit finite timing',()=>{const x=summarize(row());assert.equal(x.metrics.recall.p99_ms,.00001);assert.equal(x.absolute,true);});
test('truncated capture array is rejected',()=>{const x=row();x.trace.capture_ns.pop();assert.throws(()=>summarize(x));});
test('frontier truncation is rejected',()=>{const x=row();x.trace.candidate_counts[0]=10;assert.throws(()=>summarize(x));});
test('false publication completion is rejected',()=>{const x=row();x.trace.completed=63;assert.throws(()=>summarize(x));});
test('fabricated overlap is rejected',()=>{const x=row();x.trace.overlaps=65;assert.throws(()=>summarize(x));});
test('inconsistent monotonic intervals are rejected',()=>{const x=row();x.trace.recall_end_ns[0]++;assert.throws(()=>summarize(x));});
test('p99 includes last compaction outlier',()=>{const x=row();x.trace.recall_ns[63]=100000000;x.trace.recall_end_ns[63]=x.trace.recall_start_ns[63]+100000000;x.trace.overlaps=256;const y=summarize(x);assert.equal(y.absolute,false);assert.equal(y.metrics.recall.p99_ms,100);});
test('future nominated identity is rejected before score interpretation',()=>{assert.throws(()=>qualitySummary({k:1,nomination:[{id:'future-0000'},{id:'past-0000'},{id:'past-0001'}],exact_top:[{id:'past-0000'},{id:'past-0001'},{id:'past-0002'}]}));});
