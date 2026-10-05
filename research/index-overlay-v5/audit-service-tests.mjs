import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const root=path.dirname(fileURLToPath(import.meta.url)),dest=path.join(root,'service-evaluator-audit.json');
if(fs.existsSync(dest))throw Error('preserve audit');
const args=['--test',path.join(root,'service-evaluate.test.mjs'),path.join(root,'service-negative.test.mjs')],r=spawnSync(process.execPath,args,{encoding:'utf8'}),data=r.stdout+r.stderr;
fs.writeFileSync(path.join(root,'service-evaluator-audit.txt'),data);fs.writeFileSync(dest,JSON.stringify({args,status:r.status,signal:r.signal,error:r.error?.message??null,sha256:crypto.createHash('sha256').update(data).digest('hex'),bytes:Buffer.byteLength(data),wholeGoalValidation:false},null,2)+'\n');console.log(JSON.stringify({verified:r.status===0,status:r.status}));process.exit(r.status??1);
