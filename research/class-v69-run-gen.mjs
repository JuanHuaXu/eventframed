// Reuse the serial frozen-run protocol mechanically, not its collected results.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const source='research/shared-v68-run.mjs',raw=fs.readFileSync(source);let s=raw.toString();
s=s.replaceAll('shared-v68','class-v69').replaceAll('researchsharedsequenceref','researchclasssequenceref').replaceAll('researchsharedsequence','researchclasssequence').replaceAll('EVENTFRAME_SHARED_V68','EVENTFRAME_CLASS_V69').replaceAll('TestSharedV68','TestClassV69');
s=s.replace('research/checkpoint-2026-10-04-joint-v66-v68/manifest.json','research/checkpoint-2026-10-04-shared-v68-class-v69/manifest.json').replace('aaa2f3b819d752aab0c96ca71d9d3a7916ea0ea36785f63060a4a7681bcf96e0','409e55ab5208c107077a30957fd33720d7bbb3b6260ade01ee45e34a6bdd9d53');
const a=s.indexOf('const extras='),b=s.indexOf('const files=',a);assert(a>0&&b>a);
s=s.slice(0,a)+`const extras=['go.mod','go.sum','research/class-v69-run.mjs','research/class-v69-run-gen.mjs','research/class-v69-run-generation.json','research/class-v69-readback.mjs','research/class-v69-gen.mjs','research/class-v69-generation.json','research/class-v69-direction.md','research/class-v69-preflight.mjs','research/class-v69-preflight/results.json','docs/experiments/mmm-class-v69-protocol.md'];\n`+s.slice(b);
const d=s.indexOf('const data={};'),e=s.indexOf('fs.mkdirSync(root',d);assert(d>0&&e>d);
s=s.slice(0,d)+`const data={};for(const p of ['research/shared-v68-diagnostic/diagnostic.jsonl','research/shared-v68-diagnostic/freeze.json','research/shared-v68-diagnostic/readback.json'])data[p]=await digest(p);\n`+s.slice(e);
s=s.replace('full fourteen-arm consumed-cohort shared-context diagnostic','full ten-arm consumed-cohort class-observation diagnostic');
const out='research/class-v69-run.mjs',hash=b=>crypto.createHash('sha256').update(b).digest('hex');fs.writeFileSync(out,s,{flag:'wx',mode:0o600});
fs.writeFileSync('research/class-v69-run-generation.json',JSON.stringify({source,sourceSHA256:hash(raw),output:out,outputSHA256:hash(s),changes:'new packages/env/test names,chained parent,support list and10-arm declaration only'},null,2)+'\n',{flag:'wx',mode:0o600});
