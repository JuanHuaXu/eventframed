import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url));
const m=JSON.parse(fs.readFileSync(path.join(out,'manifest.json')));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const pins=JSON.parse(fs.readFileSync(path.join(out,'SOURCE_PINS.json')));
for(const [file,h]of Object.entries(pins.files))if(hash(fs.readFileSync(file))!==h)throw Error('prospective source drift '+file);
if(fs.existsSync(path.join(out,'cost-results.json')))throw Error('preserve existing run');
const results=[];
for(let repeat=0;repeat<3;repeat++)for(const dimension of[32,128])for(const initial of[256,1024]){
 for(const arm of(repeat%2?['candidate','control']:['control','candidate'])){
  const label=`cost-${repeat}-${dimension}-${initial}-${arm}`;
  const child=spawnSync(process.execPath,[path.join(out,'command.mjs'),arm,label,'test','./libravdb','-run','^TestResearchOverlayCostV1$','-count=1','-v','-timeout=5m'],{env:{...process.env,RESEARCH_OVERLAY_DIM:String(dimension),RESEARCH_OVERLAY_SIZE:String(initial)},stdio:['ignore','inherit','inherit']});
  const receipt=JSON.parse(fs.readFileSync(path.join(out,label+'.json')));results.push({repeat,dimension,initial,...receipt});
  fs.writeFileSync(path.join(out,'cost-results.json'),JSON.stringify({results,sourcePinsSHA256:hash(fs.readFileSync(path.join(out,'SOURCE_PINS.json'))),protocolSHA256:hash(fs.readFileSync(path.join(out,'PROTOCOL.md')))},null,2)+'\n');
  if(child.status!==0)process.exit(child.status??1);
 }
}
console.log(JSON.stringify({completed:true,commands:results.length,productionTouched:false,wholeGoalValidation:false}));
