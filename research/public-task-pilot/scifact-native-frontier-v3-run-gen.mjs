// Generate a fresh-copy full-prefix run, preserving the complete v2 runner.
import fs from 'node:fs';import path from 'node:path';import assert from 'node:assert/strict';import crypto from 'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const source='research/public-task-pilot/scifact-native-resume-v2-run.mjs',dest='research/public-task-pilot/scifact-native-frontier-v3-run.mjs';
const raw=fs.readFileSync(source,'utf8'),freeze=JSON.parse(fs.readFileSync('research/public-task-pilot/scifact-native-resume-v2/freeze.json'));
assert.equal(hash(raw),freeze.sources[path.resolve(source)].sha256);
const start=raw.indexOf("const failure=read(oldRoot+'/failure-audit.json');");
const end=raw.indexOf('fs.mkdirSync(root,{mode:0o700});',start);assert(start>0&&end>start);
const oldPrep=raw.slice(start,end);
const prep=`const failure=read(oldRoot+'/failure-audit.json');assert(failure.allCommandsTerminal);assert.equal(failure.fitQueriesCompleted,0);assert.equal(failure.fullCorpusImported,5183);\nconst receipts=fs.readFileSync(oldNative+'/trace.ndjson');assert.equal(hash(receipts),failure.traceSHA256);\nconst rows=receipts.toString().trim().split('\\n').map(JSON.parse);\nconst successful=rows.filter(r=>r.kind==='verify'||r.kind==='import');assert.equal(successful.length,5183);assert(successful.every((r,i)=>r.ok&&r.index===i));\nconst pool=read('research/public-task-pilot/scifact-pool-v1/pool.json');for(let i=0;i<successful.length;i++)assert.equal(successful[i].id,pool[i].candidate.ID);\n`;
let out=raw.replace(oldPrep,prep);
const pairs=[
 ["const root='research/public-task-pilot/scifact-native-resume-v2';","const root='research/public-task-pilot/scifact-native-frontier-v3';"],
 ["const oldRoot='research/public-task-pilot/scifact-native-fit-v1',oldNative='research/public-task-pilot/nativefit-v1';","const oldRoot='research/public-task-pilot/scifact-native-resume-v2',oldNative='research/public-task-pilot/nativeresume-v2';"],
 ["replaceAll('nativefit-v1','nativeresume-v2')","replaceAll('nativeresume-v2','nativefrontier-v3')"],
 ["replaceAll('nativeresume-v2','nativefit-v1')","replaceAll('nativefrontier-v3','nativeresume-v2')"],
 ['scifact-native-fit-v1-protocol.md','scifact-native-resume-v2-protocol.md'],
 ['scifact-native-resume-v2-protocol.md','scifact-native-frontier-v3-protocol.md'],
 ['research/public-task-pilot/scifact-native-resume-v2-run.mjs','research/public-task-pilot/scifact-native-frontier-v3-run.mjs'],
 ['research/public-task-pilot/scifact-native-resume-v2-gen.mjs','research/public-task-pilot/scifact-native-frontier-v3-gen.mjs'],
 ['research/public-task-pilot/scifact-native-resume-v2-generation.json','research/public-task-pilot/scifact-native-frontier-v3-generation.json'],
 ["'./internal/researchpublicresume','./cmd/research-public-native-resume'","'./internal/researchpublicresume','./internal/researchpublicrankguard','./cmd/research-public-native-frontier'"],
 ["'./cmd/research-public-native-resume'","'./cmd/research-public-native-frontier'"],
 ['claimedPrefix:4250','claimedPrefix:5183'],
];
// Protocol substitutions overlap intentionally; save exact intermediate states
// and verify each change can be inverted, rather than using a lossy inverse.
const transforms=[];for(const[a,b]of pairs){const before=out;out=out.replaceAll(a,b);assert.notEqual(out,before,a);transforms.push({a,b,before,after:out});}
// Fix the raw runner's protocol replacement direction explicitly: its source
// now names resume-v2, while its destination is the frontier-v3 contract.
const wrong="replaceAll('docs/experiments/scifact-native-frontier-v3-protocol.md','docs/experiments/scifact-native-frontier-v3-protocol.md')";
const right="replaceAll('docs/experiments/scifact-native-resume-v2-protocol.md','docs/experiments/scifact-native-frontier-v3-protocol.md')";
assert.equal(out.split(wrong).length,3);out=out.replace(wrong,right);
const wrongInverse="replaceAll('docs/experiments/scifact-native-frontier-v3-protocol.md','docs/experiments/scifact-native-frontier-v3-protocol.md')";
// The inverse line is the remaining same-name substitution after the first repair.
assert.equal(out.split(wrongInverse).length,2);out=out.replace(wrongInverse,"replaceAll('docs/experiments/scifact-native-frontier-v3-protocol.md','docs/experiments/scifact-native-resume-v2-protocol.md')");
// Every recorded transition derives from the exact frozen source; two explicit
// direction repairs above are checked separately. No claimed byte inverse here.
fs.writeFileSync(dest,out,{flag:'wx',mode:0o600});
fs.writeFileSync('research/public-task-pilot/scifact-native-frontier-v3-run-generation.json',JSON.stringify({source,dest,sourceSHA256:hash(raw),destSHA256:hash(out),
 transitions:transforms.map(t=>({from:hash(t.before),to:hash(t.after),search:t.a,replace:t.b})),
 prefixPreparation:'ALL5183 acknowledged or actual verified rows, no query-dependent selection',originalV2SourcePreserved:true,
 noCompleteInverseEqualityClaim:true},null,2)+'\n',{flag:'wx',mode:0o600});
console.log('frozen-source runner generation PASS; complete original corpus prefix');
