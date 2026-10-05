import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
const root=path.dirname(fileURLToPath(import.meta.url)),source=path.join(root,'../sharded-quality-v1/evaluate.test.mjs');
const before=fs.readFileSync(source),after=before.toString().replace('"./evaluate.mjs"','"./service-evaluate.mjs"'),dest=path.join(root,'service-evaluate.test.mjs');
if(fs.existsSync(dest))throw Error('preserve test derivation');fs.writeFileSync(dest,after);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
fs.writeFileSync(path.join(root,'SERVICE_TEST_DERIVATION.json'),JSON.stringify({source:'../sharded-quality-v1/evaluate.test.mjs',beforeSHA256:hash(before),afterSHA256:hash(Buffer.from(after)),change:'only imported evaluator basename; nine arithmetic/leakage/equality tests unchanged'},null,2)+'\n');
