import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),m=JSON.parse(fs.readFileSync(path.join(out,'manifest.json')));
const old=JSON.parse(fs.readFileSync(path.join(out,'../sharded-quality-v1/manifest.json')));
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
if(fs.existsSync(path.join(out,'service-manifest.json')))throw Error('preserve service preparation');
for(const[f,h]of Object.entries(old.sources.control))if(hash(fs.readFileSync(path.join(old.paths.control,f)))!==h)throw Error('service source drift '+f);
if(!JSON.parse(fs.readFileSync(path.join(out,'evaluation.json'))).qualityPass)throw Error('component quality prerequisite');
const paths={},sources={};
for(const arm of['control','candidate']){
 const dest=path.join(path.dirname(m.paths[arm]),'service-'+arm);fs.cpSync(old.paths.control,dest,{recursive:true});paths[arm]=dest;
 execFileSync('go',['mod','edit','-replace=github.com/xDarkicex/libravdb='+m.paths[arm]],{cwd:dest});
 sources[arm]=Object.fromEntries(Object.keys(old.sources.control).map(f=>[f,hash(fs.readFileSync(path.join(dest,f)))]));
}
fs.writeFileSync(path.join(out,'service-manifest.json'),JSON.stringify({paths,sources,serviceProtocolSHA256:hash(fs.readFileSync(path.join(out,'SERVICE_PROTOCOL.md'))),parentPinsSHA256:hash(fs.readFileSync(path.join(out,'SOURCE_PINS.json'))),testSHA256:sources.control['internal/service/sharded_quality_test.go'],bothUnsharded:true,productionTouched:false,privateDataUsed:false,wholeGoalValidation:false},null,2)+'\n');
console.log(JSON.stringify({prepared:true,paths}));
