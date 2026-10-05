// Mechanical derivation of the existing exclusive frozen runner.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const source='research/joint-v66-run.mjs',raw=fs.readFileSync(source);let s=raw.toString();
s=s.replaceAll('joint-v66','shared-v68').replaceAll('researchjointsequence','researchsharedsequence').replaceAll('researchjointsequenceref','researchsharedsequenceref').replaceAll('EVENTFRAME_JOINT_V66','EVENTFRAME_SHARED_V68').replaceAll('TestJointV66','TestSharedV68');
s=s.replace('research/checkpoint-2026-10-04-retention-v63-v65/manifest.json','research/checkpoint-2026-10-04-joint-v66-v68/manifest.json').replace('8b26825dd2a90f5fc2ca50b3f4f50d9969c7ddb5b141c6ffc3091e8ff1e3a932','aaa2f3b819d752aab0c96ca71d9d3a7916ea0ea36785f63060a4a7681bcf96e0');
const a=s.indexOf('const extras='),b=s.indexOf('const files=',a);assert(a>0&&b>a);
s=s.slice(0,a)+`const extras=['go.mod','go.sum','research/shared-v68-run.mjs','research/shared-v68-run-gen.mjs','research/shared-v68-run-generation.json','research/shared-v68-readback.mjs','research/shared-v68-fixture-gen.mjs','research/shared-v68-fixture-generation.json','research/shared-v68-boundary-preflight.mjs','research/shared-v68-boundary-preflight/command.json','research/shared-v68-boundary-preflight/command.log','docs/experiments/mmm-shared-v68-protocol.md'];\n`+s.slice(b);
const d=s.indexOf('const data={};'),e=s.indexOf('fs.mkdirSync(root',d);assert(d>0&&e>d);
s=s.slice(0,d)+`const data={};for(const p of ['research/tree-v60-diagnostic/diagnostic.jsonl','research/joint-v66-diagnostic/diagnostic.jsonl','research/joint-v66-diagnostic/freeze.json'])data[p]=await digest(p);\n`+s.slice(e);
s=s.replace('full eight-arm consumed-cohort joint sequence diagnostic','full fourteen-arm consumed-cohort shared-context diagnostic').replaceAll('60*60*1000','90*60*1000').replace("'-timeout=30m'","'-timeout=60m'").replace("'-timeout=50m'","'-timeout=80m'");
assert(s.includes('researchsharedsequenceref'));assert(s.includes('TestSharedV68Experiment'));assert(!s.includes('EVENTFRAME_JOINT_V66'));
const out='research/shared-v68-run.mjs',hash=b=>crypto.createHash('sha256').update(b).digest('hex');fs.writeFileSync(out,s,{flag:'wx',mode:0o600});
fs.writeFileSync('research/shared-v68-run-generation.json',JSON.stringify({source,sourceSHA256:hash(raw),output:out,outputSHA256:hash(s),changes:'package, experiment/env names, chained parent,14-arm protocol and frozen support inputs only'},null,2)+'\n',{flag:'wx',mode:0o600});
