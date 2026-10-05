// Mechanical fork of lifecycle/tests; arithmetic is supplied by a separate patch.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const from='internal/researchregimeprotected',to='internal/researchregimelog',hash=b=>crypto.createHash('sha256').update(b).digest('hex'),copies={};
assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});
for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go')&&!['model.go','resume.go'].includes(n))){
 const b=fs.readFileSync(from+'/'+n),s=b.toString().replaceAll('researchregimeprotected','researchregimelog').replaceAll('EVENTFRAME_REGIME_V81','EVENTFRAME_REGIME_V82');
 fs.writeFileSync(to+'/'+n,s,{flag:'wx',mode:0o600});copies[to+'/'+n]={source:from+'/'+n,sourceSHA256:hash(b),initialSHA256:hash(s)};
}
fs.writeFileSync('research/regime-log-v82-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'isolated persistent log-weight/log-odds repair, not quality or equivalent trace claim'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(copies).length}));
