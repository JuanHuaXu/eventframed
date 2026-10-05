import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root=resolve(dirname(fileURLToPath(import.meta.url)),'..');
const load=name=>{
  const bytes=readFileSync(resolve(root,`docs/experiments/${name}`));
  return {hash:createHash('sha256').update(bytes).digest('hex'),rows:bytes.toString('utf8').trim().split('\n').map(JSON.parse)};
};
const ordinary=load('mmm-packet-innovation-v24-design.jsonl');
const raced=load('mmm-packet-innovation-v24-race-design.jsonl');
assert.equal(ordinary.rows.length,25);
assert.equal(raced.rows.length,25);
assert.deepEqual(raced.rows[0],ordinary.rows[0]);
for(const [name,hash] of Object.entries(raced.rows[0].hashes)) {
  const actual=createHash('sha256').update(readFileSync(resolve(root,name))).digest('hex');
  assert.equal(actual,hash,`source drift: ${name}`);
}
for(let i=1;i<25;i++) {
  const core=row=>{
    const copy={...row};
    for(const field of ['RecallNS','OutcomeNS','ScoreNS']) delete copy[field];
    return copy;
  };
  assert.deepEqual(core(raced.rows[i]),core(ordinary.rows[i]),`race world ${i} changed scientific output`);
}
const result={worlds:24,nomineeLaws:3600,outcomes:768,sourceHashes:13,designSHA256:ordinary.hash,raceSHA256:raced.hash,nonTimingFieldsIdentical:true};
const outIndex=process.argv.indexOf('--out');
if(outIndex>=0) writeFileSync(resolve(root,process.argv[outIndex+1]),JSON.stringify(result,null,2)+'\n',{flag:'wx'});
process.stdout.write(JSON.stringify(result,null,2)+'\n');
