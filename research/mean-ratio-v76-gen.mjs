// New isolated fork; no edits to frozen V75/V74 candidate or independent reference.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex'),copies={};
for(const[from,to]of [['internal/researchmeananchor','internal/researchmeanratio'],['internal/researchmeananchorcheck','internal/researchmeanratiocheck']]){
 assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go'))){
  const source=from+'/'+n,out=to+'/'+n,raw=fs.readFileSync(source),s=raw.toString().replaceAll('researchmeananchor','researchmeanratio').replaceAll('EVENTFRAME_MEANANCHOR_V75','EVENTFRAME_MEANRATIO_V76');
  fs.writeFileSync(out,s,{flag:'wx',mode:0o600});copies[out]={source,sourceSHA256:sha(raw),initialSHA256:sha(s)};
 }
}
fs.writeFileSync('research/mean-ratio-v76-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'mechanical fork followed by separately audited at-anchor same-outcome conditional-factor replacement; full model retained and dense V74 reference unchanged'},null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify({copies:Object.keys(copies).length}));
