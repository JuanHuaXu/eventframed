import assert from 'node:assert/strict';
import {expectedQueryValue as value} from './regime-query-counterfactual-value.mjs';
const close=(x,y)=>assert(Math.abs(x-y)<1e-14);
close(value(true,.8,.1,.5),.18);close(value(false,.8,.1,.5),.42);
for(const q of [0,.05,.3,.5,.8,.95,1])for(const l0 of [0,.1,.5,1])for(const l1 of [0,.2,.6,1]){
  close(value(true,q,l1,l0),(1-q)*l0+q*l1);
  close(value(false,q,l0,l1),(1-q)*l0+q*l1);
  close(value(true,q,l0,l0),l0);
}
for(const args of [[0,.5,.1,.2],[true,NaN,.1,.2],[true,-.1,.1,.2],[true,1.1,.1,.2],[true,.5,Infinity,.2],[true,.5,.1,-.2]])assert.throws(()=>value(...args));
console.log('PASS: 112 outcome pairs, label-swap/degenerate/constant invariants and six invalid-input cases');
