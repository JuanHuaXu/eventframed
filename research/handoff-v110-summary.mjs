import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const bytes=fs.readFileSync(process.argv[2]??'docs/experiments/mmm-handoff-v110.json'),a=JSON.parse(bytes);
const sha=b=>crypto.createHash('sha256').update(b).digest('hex');
const near=(x,y)=>assert.ok(Number.isFinite(x)&&Number.isFinite(y)&&Math.abs(x-y)<1e-10,`${x} != ${y}`);
const risk=(p,q)=>(p-q)**2+q*(1-q);
const mix=(w,p)=>.5*w[0]+p.reduce((s,v,j)=>s+w[j+1]*v,0);
const weights=w=>{assert.equal(w.length,5);w.forEach(x=>assert.ok(Number.isFinite(x)&&x>=0&&x<=1));near(w.reduce((s,x)=>s+x,0),1);};
assert.equal(a.Version,'v110');assert.equal(a.ParentSHA256,sha(fs.readFileSync('docs/experiments/mmm-log-v109.json')));
for(const [p,h] of Object.entries(a.Hashes))assert.equal(sha(fs.readFileSync(p)),h,p);
assert.equal(a.Records.length,128);
const seen=new Set(),cells=new Map();let totalUpdates=0,crossVersion=0;
for(const {Record:r,Trace:tr} of a.Records){
 assert.ok(['design','confirmation'].includes(r.Phase)&&['majority_to_parity','parity_to_majority'].includes(r.Case));
 assert.ok(Number.isInteger(r.Index)&&r.Index>=0&&r.Index<32&&r.Schedule===1);
 const id=`${r.Phase}/${r.Case}/${r.Index}`;assert.ok(!seen.has(id));seen.add(id);
 assert.equal(r.Steps.length,256);assert.equal(tr.Profiles.length,8);assert.equal(tr.Arms.length,2);
 for(const [index,p]of tr.Profiles.entries()){
  assert.equal(p.Clock,32*index);assert.equal(p.Rows.length,512);assert.equal(p.Truth.length,512);
  const risks=[0,0,0,0],movement=[0,0,0,0];
  for(let x=0;x<512;x++){assert.equal(p.Rows[x].length,4);assert.ok(p.Truth[x]>=0&&p.Truth[x]<=1);
   for(let j=0;j<4;j++){assert.ok(p.Rows[x][j]>0&&p.Rows[x][j]<1);risks[j]+=risk(p.Rows[x][j],p.Truth[x])/512;if(index)movement[j]+=Math.abs(p.Rows[x][j]-tr.Profiles[index-1].Rows[x][j])/512;}
  }
  for(let j=0;j<4;j++){near(risks[j],p.Risk[j]);near(movement[j],p.Movement[j]);}
 }
 for(const [ai,arm]of tr.Arms.entries()){
  assert.equal(arm.Steps.length,256);assert.equal(arm.Updates.length,r.Arrived);
  const prior=ai===0?[0,.95,.05/3,.05/3,.05/3]:[.05,.95*.95,.95*.05/3,.95*.05/3,.95*.05/3];
  let state=[...prior],cursor=0;const delivered=new Set();
  const update=u=>{
   assert.ok(!delivered.has(u.Origin));delivered.add(u.Origin);
   const old=r.Steps[u.Origin];assert.ok(!old.Missing);assert.equal(u.Arrival,u.Origin+old.Delay);assert.equal(u.Clock,u.Arrival);assert.equal(u.Outcome,old.Y);
   assert.equal(u.IssueVersion,Math.floor(u.Origin/32)+1);
   const beforePublication=u.Clock%32===0&&u.Origin<u.Clock;
   assert.equal(u.UpdateVersion,Math.min(8,Math.floor((u.Clock-(beforePublication?1:0))/32)+1));
   weights(u.Before);weights(u.After);
   const like=[.5,...u.Raw].map(p=>u.Outcome?p:1-p),unnormalized=state.map((v,j)=>v*like[j]),sum=unnormalized.reduce((s,x)=>s+x,0);
   for(let j=0;j<4;j++)near(u.Raw[j],arm.Steps[u.Origin].Raw[j]);
   for(let j=0;j<5;j++){near(u.Before[j],state[j]);state[j]=.999*unnormalized[j]/sum+.001*prior[j];near(u.After[j],state[j]);}
   totalUpdates++;if(u.IssueVersion!==u.UpdateVersion)crossVersion++;
  };
  for(const [t,s]of arm.Steps.entries()){
   while(cursor<arm.Updates.length&&(arm.Updates[cursor].Clock<t||(arm.Updates[cursor].Clock===t&&arm.Updates[cursor].Origin<t)))update(arm.Updates[cursor++]);
   assert.equal(s.Origin,t);weights(s.Advice);weights(s.Routed);for(let j=0;j<5;j++)near(s.Advice[j],state[j]);
   const old=r.Steps[t],profile=tr.Profiles[Math.floor(t/32)];assert.equal(s.Mask,old.Mask[5+ai]);assert.equal(s.MatchedMask,old.Mask[0]);near(s.Served,old.P[5+ai]);
   const partial=mask=>{const raw=[0,0,0,0];let count=0;for(let x=0;x<512;x++)if((x&mask)===(old.X&mask)){count++;for(let j=0;j<4;j++)raw[j]+=profile.Rows[x][j];}return raw.map(x=>x/count);};
   const own=partial(s.Mask),matched=partial(s.MatchedMask);
   for(let j=0;j<4;j++){near(own[j],s.Raw[j]);near(matched[j],s.MatchedRaw[j]);near(profile.Rows[old.X][j],s.FullRaw[j]);}
   near(mix(s.Routed,s.Raw),s.Served);near(mix(s.Routed,s.MatchedRaw),s.Matched);near(mix(s.Routed,s.FullRaw),s.Full);near(matched[0],old.P[0]);
   const key=`${r.Phase}/${r.Case}/${Math.floor(t/32)}/${ai}`;
   if(!cells.has(key))cells.set(key,{phase:r.Phase,case:r.Case,start:32*Math.floor(t/32),arm:5+ai,n:0,own:0,matched:0,full:0,generic:0,genericOwn:0,raw:[0,0,0,0],profileRisk:[0,0,0,0],movement:[0,0,0,0],advice:[0,0,0,0,0],routed:[0,0,0,0,0]});
   const c=cells.get(key);c.n++;c.own+=risk(s.Served,old.Q)/1024;c.matched+=risk(s.Matched,old.Q)/1024;c.full+=risk(s.Full,old.Q)/1024;c.generic+=risk(old.P[0],old.Q)/1024;c.genericOwn+=risk(s.Raw[0],old.Q)/1024;
   for(let j=0;j<4;j++){c.raw[j]+=risk(s.Raw[j],old.Q)/1024;c.profileRisk[j]+=profile.Risk[j]/1024;c.movement[j]+=profile.Movement[j]/1024;}
   for(let j=0;j<5;j++){c.advice[j]+=s.Advice[j]/1024;c.routed[j]+=s.Routed[j]/1024;}
  }
  while(cursor<arm.Updates.length)update(arm.Updates[cursor++]);
  assert.equal(delivered.size,r.Arrived);
 }
}
assert.equal(cells.size,64);for(const c of cells.values()){assert.equal(c.n,1024);c.maskDifference=c.own-c.matched;c.sameMaskMixtureDifference=c.matched-c.generic;}
console.log(JSON.stringify({version:'v110',status:'CONSUMED_DIAGNOSTIC_NOT_A_RESCUE',artifactSHA256:sha(bytes),sourceHashes:Object.keys(a.Hashes).length,records:128,profileCount:1024,forecastRows:65536,totalUpdates,crossVersion,cells:[...cells.values()]},null,2));
