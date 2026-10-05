// Mechanical isolated fork before the exact incremental computation changes.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const from='internal/researchregimeledger',to='internal/researchregimeincremental';
assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),copies={};
for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go'))) {
  const b=fs.readFileSync(from+'/'+n),s=b.toString().replaceAll('researchregimeledger','researchregimeincremental').replaceAll('EVENTFRAME_REGIME_V79','EVENTFRAME_REGIME_V80');
  fs.writeFileSync(to+'/'+n,s,{flag:'wx',mode:0o600});
  copies[to+'/'+n]={source:from+'/'+n,sourceSHA256:hash(b),initialSHA256:hash(s)};
}
fs.writeFileSync('research/regime-incremental-v80-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'equivalent incremental computation; original frozen law and limitations retained'},null,2)+'\n',{flag:'wx',mode:0o600});
const runner=fs.readFileSync('research/regime-ledger-v79-run.mjs','utf8')
  .replaceAll('regime-ledger-v79','regime-incremental-v80').replaceAll('researchregimeledger','researchregimeincremental').replaceAll('EVENTFRAME_REGIME_V79','EVENTFRAME_REGIME_V80')
  .replaceAll('checkpoint-2026-10-05-regime-v77-frozen-v78','checkpoint-2026-10-05-regime-ledger-v79')
  .replaceAll('5aa836d99096fe6c843edc2a8d855e261d3e0a0a526b2d1d7c4d4bc75cd05249','c5582c55cf000fbfa2d34d91c1da54f7a6e06fdd0493f2113f403b162b05c839');
fs.writeFileSync('research/regime-incremental-v80-run.mjs',runner,{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(copies).length,runner:'research/regime-incremental-v80-run.mjs'}));
