// Mechanical isolated fork; do not alter either preserved model's source.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const from='internal/researchregimefrozen',to='internal/researchregimeledger';
assert(!fs.existsSync(to));
fs.mkdirSync(to,{mode:0o700});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const copies={};
for(const n of ['model.go','oracle_test.go']) {
  const b=fs.readFileSync(from+'/'+n);
  let s=b.toString().replaceAll('researchregimefrozen','researchregimeledger');
  if(n==='model.go') s=s.slice(0,s.indexOf('// Journal is a single-owner'));
  fs.writeFileSync(to+'/'+n,s,{flag:'wx',mode:0o600});
  copies[to+'/'+n]={source:from+'/'+n,sourceSHA256:hash(b),initialSHA256:hash(s)};
}
fs.writeFileSync('research/regime-ledger-v79-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'isolated canonical support/law ledger; no quality or runtime adoption'},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(copies).length}));
