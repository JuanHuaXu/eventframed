import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),old=path.join(out,'../index-overlay-v1');
const pins=JSON.parse(fs.readFileSync(path.join(old,'SOURCE_PINS.json'))),manifest=JSON.parse(fs.readFileSync(path.join(old,'manifest.json')));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'manifest.json')))throw Error('preserve preparation');
for(const[f,h]of Object.entries(pins.files))if(hash(fs.readFileSync(f))!==h)throw Error('V1 source drift '+f);
const root=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-index-overlay-v2-')),paths={};
for(const arm of['control','candidate']){paths[arm]=path.join(root,arm);fs.cpSync(manifest.paths[arm],paths[arm],{recursive:true});const f=path.join(paths[arm],'libravdb/research_key_capacity_test.go');fs.copyFileSync(path.join(old,'key_capacity_test.go.txt'),f);execFileSync('gofmt',['-w',f]);}
fs.cpSync(path.join(path.dirname(manifest.paths.control),'lexer'),path.join(root,'lexer'),{recursive:true});
for(const f of['command.mjs','run.mjs','pin.mjs','database_test.go.txt','preflight_test.go.txt'])fs.copyFileSync(path.join(old,f),path.join(out,f));
fs.writeFileSync(path.join(out,'manifest.json'),JSON.stringify({paths,inheritedSources:manifest.inheritedSources,inheritedRoot:manifest.inheritedRoot,protocolSHA256:hash(fs.readFileSync(path.join(out,'PROTOCOL.md'))),preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),parentPinsSHA256:hash(fs.readFileSync(path.join(old,'SOURCE_PINS.json'))),workers:10,productionTouched:false,privateDataUsed:false,wholeGoalValidation:false},null,2)+'\n');
console.log(JSON.stringify({prepared:true,paths}));
