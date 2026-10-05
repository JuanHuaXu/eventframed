import fs from 'node:fs';
import assert from 'node:assert/strict';
const [dir,output]=process.argv.slice(2);assert(dir&&output);
const results=[];
for(const [name,file]of [['staticGlobal','mmm-spike-continuous-v1-results.json'],['shareGlobal','mmm-spike-fixed-share-v1-results.json'],['staticLocal','mmm-spike-static-local-v1-results.json'],['shareLocal','mmm-spike-fixed-share-local-v1-results.json']]){
  const a=JSON.parse(fs.readFileSync(`${dir}/${file}`,'utf8'));assert.equal(a.results.length,672);
  const blocks=[];
  for(let block=0;block<8;block++){
    const groups=new Map();let mean=0,harms=0,maxHarm=-Infinity;
    for(const r of a.results){const [phase,scenario,index,schedule]=r.key.split(':').map(Number),key=[phase,scenario,schedule].join(':'),delta=r.blocks[block].expected[0]-r.blocks[block].expected[2];mean+=delta/672;harms+=Number(delta>.01);maxHarm=Math.max(maxHarm,delta);if(!groups.has(key))groups.set(key,new Map());assert(!groups.get(key).has(index));groups.get(key).set(index,delta);}
    assert.equal(groups.size,84);const scenarios=[];for(const [key,rs]of groups){assert.equal(rs.size,8);scenarios.push({key,meanDelta:[...rs.values()].reduce((a,b)=>a+b,0)/8});}
    blocks.push({start:block*32,meanDelta:mean,recordHarmsAbovePoint01:harms,maxRecordHarm:maxHarm,scenarioMeansAbovePoint01:scenarios.filter(r=>r.meanDelta>.01),scenarios});
  }
  results.push({name,blocks});
}
fs.writeFileSync(output,JSON.stringify({results,limitations:'Descriptive all-window screen on consumed data. No post-selection or simultaneous confidence claim.'},null,2)+'\n',{flag:'wx'});
for(const r of results)console.log(JSON.stringify({name:r.name,blocks:r.blocks.map(({scenarios,...r})=>r)},null,2));
