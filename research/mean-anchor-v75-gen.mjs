// Mechanical isolation only. V74 and its independent reference remain frozen.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex'),copies={};
for(const [from,to]of [['internal/researchmeanjoint','internal/researchmeananchor'],['internal/researchmeanjointcheck','internal/researchmeananchorcheck']]){
 assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});
 for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go'))){
  const source=from+'/'+n,out=to+'/'+n,raw=fs.readFileSync(source);
  let s=raw.toString().replaceAll('researchmeanjointcheck','researchmeananchorcheck').replaceAll('researchmeanjoint"','researchmeananchor"').replaceAll('package researchmeanjoint','package researchmeananchor').replaceAll('EVENTFRAME_MEANJOINT_V74','EVENTFRAME_MEANANCHOR_V75');
  fs.writeFileSync(out,s,{flag:'wx',mode:0o600});copies[out]={source,sourceSHA256:sha(raw),initialSHA256:sha(s)};
 }
}
fs.writeFileSync('research/mean-anchor-v75-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'mechanical isolated fork; manual anchored-prefix changes follow; independent dense V74 reference is retained, not forked'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(copies).length,reference:'independent V74 dense reference unchanged'}));
