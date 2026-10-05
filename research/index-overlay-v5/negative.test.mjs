import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import test from 'node:test';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {verify} from './verify.mjs';
const root=path.dirname(fileURLToPath(import.meta.url));
test('actual archive/report validates before mutations',()=>assert.equal(verify(root).verified,true));
for(const kind of['fabricated-mean','hidden-failure','omitted-cell','changed-protocol'])test('reject '+kind,()=>{
 const copy=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-overlay-negative-'));fs.cpSync(root,copy,{recursive:true});
 function mutate(f,fn){const p=path.join(copy,f),x=JSON.parse(fs.readFileSync(p));fn(x);fs.writeFileSync(p,JSON.stringify(x));}
 if(kind==='fabricated-mean')mutate('evaluation.json',x=>{x.cells[0].meanTieRecall+=.01;});
 if(kind==='hidden-failure')mutate('evaluation.json',x=>{x.failed.push({status:0,label:'invented'});});
 if(kind==='omitted-cell')mutate('cost-results.json',x=>{x.results.pop();});
 if(kind==='changed-protocol')fs.appendFileSync(path.join(copy,'PROTOCOL.md'),'\nchanged gate\n');
 assert.throws(()=>verify(copy,{sources:false}),{name:'AssertionError'});
});
