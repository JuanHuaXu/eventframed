import fs from 'node:fs';
import crypto from 'node:crypto';
import cp from 'node:child_process';
import path from 'node:path';
const root='research/public-task-pilot/git-fusion-v1/';
const paths=new Set([root+'corpus.json',root+'queries.json',root+'PROTOCOL.md','go.mod','go.sum']);
const dirs=cp.execFileSync('go',['list','-deps','-f','{{if not .Standard}}{{.Dir}}{{end}}','./cmd/public-git-fusion'],{encoding:'utf8'}).trim().split('\n').filter(Boolean);
for(const dir of dirs){const rel=path.relative(process.cwd(),dir);if(rel.startsWith('..'))continue;for(const f of fs.readdirSync(dir)){if(f.endsWith('.go')&&(!f.endsWith('_test.go')||rel.startsWith('internal/researchfusion')||rel==='cmd/public-git-fusion'))paths.add(path.join(rel,f));}}
const hashes=ps=>Object.fromEntries([...ps].sort().map(p=>[p,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex')]));
fs.writeFileSync(root+'freeze.json',JSON.stringify(hashes(paths),null,2)+'\n',{flag:'wx',mode:0o600});
fs.writeFileSync(root+'audit-freeze.json',JSON.stringify(hashes([root+'freeze.json',root+'freeze.mjs',root+'facts.json',root+'oracle.json',root+'prepare.mjs',root+'score.mjs',root+'test-suite.mjs',root+'tests.json']),null,2)+'\n',{flag:'wx',mode:0o600});
console.log('frozen runtime inputs/sources',paths.size,'auditor/label hashes',8);
