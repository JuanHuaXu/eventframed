import test from 'node:test';
import assert from 'node:assert/strict';
import {quality} from './evaluate.mjs';
function row(){const e=Array.from({length:50},(_,i)=>({id:`e-${String(i).padStart(5,'0')}`,exact:i===0?.9:.5,native:i===0?.9:.5}));return {records:51,actual:structuredClone(e),exact_top:e};}
test('boundary tie cannot replace strictly better neighbor',()=>{const r=row();r.actual[0]={id:'e-00050',exact:.5,native:.5};assert.equal(quality(r).tieRecall,49/50);});
test('exact geometry passes',()=>assert.equal(quality(row()).tieRecall,1));
test('future nominee rejected',()=>{const r=row();r.actual[0].id='e-00051';assert.throws(()=>quality(r));});
test('duplicate nominee rejected',()=>{const r=row();r.actual[0].id=r.actual[1].id;assert.throws(()=>quality(r));});
test('nonfinite native score rejected',()=>{const r=row();r.actual[0].native=NaN;assert.throws(()=>quality(r));});
