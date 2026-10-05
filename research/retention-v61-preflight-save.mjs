// Preserve the failed pre-freeze helper and compiler metadata before local repair.
import fs from 'node:fs';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
const root='research/retention-v61-preflight';
fs.mkdirSync(root,{mode:0o700,recursive:true});
// The initial preservation command hit EEXIST after lesson.md created this
// directory. Recover the exact first helper by inversing only the recorded patch.
let failed=fs.readFileSync('research/retention-v61-component-run.mjs','utf8');
function inverse(from,to) {if(!failed.includes(from))throw Error('inverse mismatch');failed=failed.replace(from,to);}
inverse('const closure = JSON.parse(raw), compiler = new Set(), generated = {};','const closure = JSON.parse(raw), compiler = new Set();');
inverse("  const absolute = path.resolve(p.Dir,n);\n  if (absolute.startsWith(process.cwd() + '/')) compiler.add(path.relative(process.cwd(),absolute));\n  else {\n   assert(p.ImportPath.endsWith('.test'), 'unexpected external compiler input');\n   generated[absolute] = await fileHash(absolute);\n  }", "  const absolute = path.resolve(p.Dir,n); assert(absolute.startsWith(process.cwd() + '/'));\n  compiler.add(path.relative(process.cwd(),absolute));");
inverse('generatedCompilerFiles:generated,','');
fs.writeFileSync(root+'/failed-run.mjs',failed,{flag:'wx',mode:0o600});
const raw=execFileSync('go',['run','./cmd/research-go-list-closure','./internal/researchretention'],{maxBuffer:32*1024*1024});
fs.writeFileSync(root+'/compiler-closure.json',raw,{flag:'wx',mode:0o600});
fs.writeFileSync(root+'/failure.json',JSON.stringify({stage:'before component root/freeze or any test/benchmark',
 error:"assert(absolute.startsWith(process.cwd() + '/'))",cause:'synthetic .test GoFiles points into Go build cache although package Dir is repository-local',
 failedSourceSHA256:crypto.createHash('sha256').update(fs.readFileSync(root+'/failed-run.mjs')).digest('hex'),
 sourceRecovery:'exact inverse of recorded helper patch, not a pre-repair copy',
 preservationAttemptFailure:'EEXIST: lesson.md had already created this directory; no failure artifact had been written',
 scientificContractChanged:false,repair:'record generated cache input separately; preserve entire closure and repo source copies'},null,2)+'\n',{flag:'wx',mode:0o600});
