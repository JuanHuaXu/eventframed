import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url));
const inherited=JSON.parse(fs.readFileSync(path.join(out,'../sharded-quality-scale-v2/manifest.json')));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'manifest.json')))throw Error('preserve existing preparation');
for(const [f,h]of Object.entries(inherited.sources.fork))if(hash(fs.readFileSync(path.join(inherited.paths.fork,f)))!==h)throw Error('source drift '+f);
const root=fs.mkdtempSync(path.join(os.tmpdir(),'eventframe-index-overlay-'));
const paths={};
for(const arm of ['control','candidate']){
  paths[arm]=path.join(root,arm);
  fs.cpSync(inherited.paths.fork,paths[arm],{recursive:true});
}
fs.writeFileSync(path.join(out,'manifest.json'),JSON.stringify({paths,inheritedSources:inherited.sources.fork,inheritedRoot:inherited.paths.fork,protocolSHA256:hash(fs.readFileSync(path.join(out,'PROTOCOL.md'))),preparationSHA256:hash(fs.readFileSync(fileURLToPath(import.meta.url))),go:execFileSync('go',['version']).toString().trim(),workers:10,productionTouched:false,privateDataUsed:false,wholeGoalValidation:false},null,2)+'\n');
console.log(JSON.stringify({prepared:true,root,paths}));
