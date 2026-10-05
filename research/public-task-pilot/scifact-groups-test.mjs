import assert from 'node:assert/strict';import{group}from'./scifact-groups.mjs';
const edge=(query,document)=>({query,document,score:1});
let r=group([edge('a','d'),edge('b','e'),edge('z','z')],[edge('c','d')]);assert.deepEqual(r.splits.confirmation,['c']);assert.deepEqual(r.splits.excludedOfficialTrain,['a']);assert.equal(r.components.length,3);
r=group([edge('a','d'),edge('b','d'),edge('b','e')],[edge('c','e')]);assert.equal(r.components.length,1);assert.deepEqual(r.splits.excludedOfficialTrain,['a','b']);assert.equal(r.units.fit.length+r.units.calibration.length,0);
const a=[edge('x','same'),edge('same','y')],b=[edge('confirm','d')];assert.equal(group(a,b).components.length,3);assert.deepEqual(group(a,b),group([...a].reverse(),b));assert.throws(()=>group([edge('q','d')],[edge('q','e')]));
console.log(JSON.stringify({positiveDirectTransitiveIdentityControls:3,orderInvariant:true,overlapNegative:true,noTopicIndependenceClaim:true}));
