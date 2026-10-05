import fs from 'node:fs';
import path from 'node:path';
import {spawnSync} from 'node:child_process';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
const root=path.dirname(fileURLToPath(import.meta.url));
if(fs.existsSync(path.join(root,'evaluator-audit.json')))throw Error('preserve audit output');
const args=['--test',path.join(root,'evaluate.test.mjs'),path.join(root,'negative.test.mjs')],r=spawnSync(process.execPath,args,{encoding:'utf8'}),text=r.stdout+r.stderr;
fs.writeFileSync(path.join(root,'evaluator-audit.txt'),text);fs.writeFileSync(path.join(root,'evaluator-audit.json'),JSON.stringify({args,status:r.status,signal:r.signal,error:r.error?.message??null,sha256:crypto.createHash('sha256').update(text).digest('hex'),bytes:Buffer.byteLength(text),wholeGoalValidation:false},null,2)+'\n');console.log(JSON.stringify({status:r.status,verified:r.status===0}));process.exit(r.status??1);
