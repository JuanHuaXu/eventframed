// Fork the tested computation; the protected support intentionally changes law.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const from='internal/researchregimeincremental',to='internal/researchregimeprotected',hash=b=>crypto.createHash('sha256').update(b).digest('hex'),copies={};
assert(!fs.existsSync(to));fs.mkdirSync(to,{mode:0o700});
for(const n of fs.readdirSync(from).filter(n=>n.endsWith('.go'))) {
  const b=fs.readFileSync(from+'/'+n),s=b.toString().replaceAll('researchregimeincremental','researchregimeprotected').replaceAll('EVENTFRAME_REGIME_V80','EVENTFRAME_REGIME_V81');
  fs.writeFileSync(to+'/'+n,s,{flag:'wx',mode:0o600});copies[to+'/'+n]={source:from+'/'+n,sourceSHA256:hash(b),initialSHA256:hash(s)};
}
fs.writeFileSync('research/regime-protected-v81-generation.json',JSON.stringify({time:new Date().toISOString(),copies,scope:'protected-support law experiment, not equivalent pruning or inherited theoretical guarantee'},null,2)+'\n',{flag:'wx',mode:0o600});
let runner=fs.readFileSync('research/regime-incremental-v80-run.mjs','utf8')
  .replaceAll('regime-incremental-v80','regime-protected-v81').replaceAll('researchregimeincremental','researchregimeprotected').replaceAll('EVENTFRAME_REGIME_V80','EVENTFRAME_REGIME_V81')
  .replaceAll('checkpoint-2026-10-05-regime-ledger-v79','checkpoint-2026-10-05-regime-incremental-v80')
  .replaceAll('c5582c55cf000fbfa2d34d91c1da54f7a6e06fdd0493f2113f403b162b05c839','d2ae1995bfaa0caff9a2ddee5864847606fd532ec1ae7cd27e16d5a50fcd965e');
runner=runner.replaceAll("'./internal/researchregimeledger','./internal/researchregimeprotected'","'./internal/researchregimeincremental','./internal/researchregimeprotected'");
fs.writeFileSync('research/regime-protected-v81-run.mjs',runner,{flag:'wx',mode:0o600});
console.log(JSON.stringify({copies:Object.keys(copies).length}));
