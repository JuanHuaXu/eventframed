import assert from 'node:assert/strict';import{ceiling}from'./metrology-headroom.mjs';
const fixture=()=>Array.from({length:36},(_,i)=>({id:String(i),cluster:String(Math.floor(i/6)),wording:i%6<3?'literal':'paraphrase',top1:1,nominated:1,retained:1,survival:1}));
let p=fixture();assert.equal(ceiling(p).best,null);assert.equal(ceiling(p).oracleGatePossible,false);
p=fixture();p[3].top1=0;p[9].top1=0;let c=ceiling(p);assert.equal(c.maximumNetGains,2);assert.equal(c.best.net,2);assert(c.best.lower<0);assert.equal(c.oracleGatePossible,false);
p=fixture();p[3].top1=0;p[9].top1=0;p[15].top1=0;assert.equal(ceiling(p).oracleGatePossible,true);
p=fixture();p[3].top1=p[4].top1=p[5].top1=0;assert.equal(ceiling(p).oracleGatePossible,false);
p=fixture();p[0].retained=0;assert.throws(()=>ceiling(p));p=fixture();p[0].id=p[1].id;assert.throws(()=>ceiling(p));
console.log(JSON.stringify({positiveFeasibilityControl:1,impossibleCeilingControls:3,invalidInputControls:2,oracleDoesNotClaimImplementability:true}));
