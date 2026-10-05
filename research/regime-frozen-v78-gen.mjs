// Preserve V77's failed shortcut; mechanically fork before the conditional rescue.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const from='internal/researchregime',to='internal/researchregimefrozen',copies={};
assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go'))){
  const b=fs.readFileSync(from+'/'+n),s=b.toString().replaceAll('researchregime','researchregimefrozen').replaceAll('EVENTFRAME_REGIME_V77','EVENTFRAME_REGIME_V78');
  fs.writeFileSync(to+'/'+n,s,{flag:'wx',mode:0o600});
  copies[to+'/'+n]={source:from+'/'+n,sourceSHA256:hash(b),initialSHA256:hash(s)};
}
fs.writeFileSync('research/regime-frozen-v78-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'isolated fixed-support conditional rescue; original mathematical failures and timing preserved'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(copies).length}));
