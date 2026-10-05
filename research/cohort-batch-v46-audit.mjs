// V46 extends only the explicitly frozen batch cap, not any scientific guard.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import vm from 'node:vm';
import assert from 'node:assert/strict';
import readline from 'node:readline';
import {readEnvelope} from './eager-load-v44-stream.mjs';

const hash = b => crypto.createHash('sha256').update(b).digest('hex');
export function checkBatchCap(row) {
  assert([4, 8].includes(row.BatchCap), 'declared batch cap');
  assert(Array.isArray(row.Batches) && row.Batches.length, 'batch coverage');
  for (const batch of row.Batches) assert(batch.IDs.length >= 1 && batch.IDs.length <= row.BatchCap, 'over-cap batch');
}
function replaceOne(source, from, to) {
  assert.equal(source.split(from).length, 2, 'unique sealed transformation: '+from);
  return source.replace(from, to);
}
const originalBatch = fs.readFileSync('research/batch-archive-v33-audit.mjs', 'utf8');
let batchSource = replaceOne(originalBatch, 'b.IDs.length<=4', 'b.IDs.length<=row.BatchCap');
batchSource = replaceOne(batchSource, '[1,2,3,4].map', '[1,2,3,4,5,6,7,8].map');
const readPaths = new Set();
const sealedFS = {...fs, readFileSync(p, ...rest) {
  if (String(p) === 'research/batch-archive-v33-audit.mjs') {
    readPaths.add(String(p)); assert(rest[0] === 'utf8'); return batchSource;
  }
  return fs.readFileSync(p, ...rest);
}};
const source = fs.readFileSync('research/eager-load-v44-audit.mjs', 'utf8');
const start = source.indexOf('const hash='), end = source.indexOf('// Run as a library');
assert(start >= 0 && end > start);
const core = source.slice(start, end).replaceAll('export function ', 'function ').replaceAll('export const ', 'const ');
const inherited = vm.runInNewContext(core+'\n({check, mutations, preflightFixture})', {fs:sealedFS, path, crypto, vm, process, Buffer, structuredClone}, {timeout:1000});
assert.equal(readPaths.size, 1, 'batch predicate must be on the actual inherited checker path');
export function check(row) {
  checkBatchCap(row);
  return {...inherited.check(row), batch_cap:row.BatchCap};
}
function corruptions(first) {
  const controls = [];
  for (const [name, mutate] of [...inherited.mutations,
    ['missing-cap', r => delete r.BatchCap], ['unsupported-cap', r => r.BatchCap=16]]) {
    const row = structuredClone(first); mutate(row);
    assert.throws(() => check(row), undefined, 'corruption escaped '+name); controls.push(name);
  }
  return controls;
}
export function boundControls() {
  for (const cap of [4,8]) {
    for (let n=1; n<=cap; n++) checkBatchCap({BatchCap:cap,Batches:[{IDs:Array(n).fill('unit-bound')} ]});
    assert.throws(()=>checkBatchCap({BatchCap:cap,Batches:[{IDs:Array(cap+1).fill('unit-bound')}]}));
    assert.throws(()=>checkBatchCap({BatchCap:cap,Batches:[{IDs:[]}]}));
  }
  for (const cap of [undefined,0,3,5,16,'8']) assert.throws(()=>checkBatchCap({BatchCap:cap,Batches:[{IDs:['unit-bound']}]}));
  return {positive:12,negative:10,scope:'bound-only fixtures, not load evidence'};
}
if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve('research/cohort-batch-v46-audit.mjs')) {
  if (process.argv[2] === '--self-test') {
    let consumed;
    for await (const line of readline.createInterface({input:fs.createReadStream('research/warm-search-v37/raw.ndjson'),crlfDelay:Infinity})) {
      const row=JSON.parse(line); if(row.WarmEnabled && row.Archived && !row.Visible) {consumed=row;break}
    }
    assert(consumed); const fixture=inherited.preflightFixture(consumed), controls={};
    for(const cap of [4,8]) {fixture.BatchCap=cap;check(fixture);controls[cap]=corruptions(fixture)}
    console.log(JSON.stringify({type:'consumed synthetic checker fixture ONLY',bounds:boundControls(),controls,adaptedBatchSHA256:hash(batchSource)},null,2));
  } else {
    const input=process.argv[2]; assert(input,'explicit raw input');
    const trials=[], factors=new Set(), controls={};
    const envelope=await readEnvelope(input,row=>{
      const factor=[row.Trial,row.Visible,row.Archived,row.EagerEnabled,row.BatchCap].join('/');
      assert(!factors.has(factor),'duplicate factor');factors.add(factor);trials.push(check(row));
      if(!controls[row.BatchCap] && row.EagerEnabled && row.Archived && !row.Visible) controls[row.BatchCap]=corruptions(row);
    },32);
    assert.equal(factors.size,32); assert(controls[4] && controls[8]);
    // Check the Cartesian product, not merely 32 distinct arbitrary factors.
    for(const repetition of [1,2]) for(const visible of [false,true]) for(const archived of [false,true]) for(const eager of [false,true]) for(const cap of [4,8]) assert(factors.has([repetition,visible,archived,eager,cap].join('/')),'missing declared cell');
    const freeze=fs.readFileSync(path.dirname(input)+'/freeze.json');assert.equal(envelope.header.FreezeSHA256,hash(freeze));
    console.log(JSON.stringify({time:new Date().toISOString(),rawSHA256:envelope.rawSHA256,adaptedBatchSHA256:hash(batchSource),controls,bounds:boundControls(),trials,allSevenWholeGoals:'OPEN'},null,2));
  }
}
