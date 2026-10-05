import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),old=path.join(out,'../index-overlay-v3');
const m=JSON.parse(fs.readFileSync(path.join(old,'manifest.json'))),pins=JSON.parse(fs.readFileSync(path.join(old,'SOURCE_PINS.json'))),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'manifest.json')))throw Error('preserve preparation');
for(const[f,h]of Object.entries(pins.files))if(hash(fs.readFileSync(f))!==h)throw Error('V3 source drift '+f);
const root=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-index-overlay-v4-')),paths={};
for(const arm of['control','candidate']){paths[arm]=path.join(root,arm);fs.cpSync(m.paths[arm],paths[arm],{recursive:true});}
fs.cpSync(path.join(path.dirname(m.paths.control),'lexer'),path.join(root,'lexer'),{recursive:true});
for(const f of['command.mjs','run.mjs','evaluate.mjs','evaluate.test.mjs','independent.mjs','verify.mjs','negative.test.mjs','audit-tests.mjs','database_test.go.txt','preflight_test.go.txt','key_capacity_test.go.txt','profile_test.go.txt','ownership_test.go.txt','SERVICE_PROTOCOL.md','service-prepare.mjs','service-run.mjs','derive-service-evaluators.mjs'])fs.copyFileSync(path.join(old,f),path.join(out,f));
fs.writeFileSync(path.join(out,'manifest.json'),JSON.stringify({paths,parentPinsSHA256:hash(fs.readFileSync(path.join(old,'SOURCE_PINS.json'))),protocolSHA256:hash(fs.readFileSync(path.join(out,'PROTOCOL.md'))),preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),workers:10,productionTouched:false,privateDataUsed:false,wholeGoalValidation:false},null,2)+'\n');console.log(JSON.stringify({prepared:true,paths}));
