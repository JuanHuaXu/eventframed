import fs from 'node:fs';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const out=path.dirname(fileURLToPath(import.meta.url)),repo=path.resolve(out,'../..'),destination=process.argv[2];
if(!destination)throw Error('explicit clean checkpoint checkout required');
const status=execFileSync('git',['status','--porcelain'],{cwd:destination,encoding:'utf8'});if(status.trim())throw Error('destination is not clean');
const check=JSON.parse(fs.readFileSync(path.join(out,'CHECKPOINT_VERIFICATION.json')));if(!check.verified||check.wholeGoalValidation)throw Error('missing honest verified checkpoint');
for(let n=1;n<=6;n++) {
 const source=path.join(repo,'research/index-overlay-v'+n),dest=path.join(destination,'research/index-overlay-v'+n);if(fs.existsSync(dest))throw Error('preserve staged directory');
 const omitted=new Set(fs.existsSync(path.join(source,'LOCAL_ONLY.json'))?JSON.parse(fs.readFileSync(path.join(source,'LOCAL_ONLY.json'))).omittedFiles.map(x=>x.file):[]);
 fs.cpSync(source,dest,{recursive:true,filter:p=>!omitted.has(path.relative(source,p))});
}
fs.copyFileSync(path.join(repo,'research-direction.md'),path.join(destination,'research-direction.md'));
console.log(JSON.stringify({staged:true,scope:'six index-overlay research directories and research-direction.md only',rawCopiesRemainLocal:true,automaticPush:false}));
