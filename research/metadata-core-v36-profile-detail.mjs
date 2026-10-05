// Supplement the original selector without rewriting its preserved outputs.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import{execFileSync}from'node:child_process';
const dir=path.resolve('research/metadata-core-v36'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const profile=JSON.parse(fs.readFileSync(path.join(dir,'profile.json')));if(profile.code!==0)throw Error('profile not terminal');
for(const kind of ['binary','cpu','mutex','block']){const p=profile.files[kind];if(hash(fs.readFileSync(path.join(dir,p.path)))!==p.sha256)throw Error('changed profile '+kind)}
const reports=[];for(const[name,kind,options]of[
 ['recall-cpu','cpu',['-top','-cum','-nodecount=40','-focus=recallOnce']],
 ['posterior-cpu','cpu',['-top','-cum','-nodecount=40','-focus=GetBayesianPosterior']],
 ['archive-cpu','cpu',['-top','-cum','-nodecount=40','-focus=applyV34']],
 ['recall-block','block',['-top','-cum','-nodecount=40','-focus=recallOnce']]
]){
 const args=['tool','pprof',...options,path.join(dir,profile.files.binary.path),path.join(dir,profile.files[kind].path)];
 const text=execFileSync('go',args,{encoding:'utf8'});
 if(!text.includes('Showing')||text.includes('no samples'))throw Error('empty selector '+name);
 fs.writeFileSync(path.join(dir,`profile-${name}.txt`),text,{flag:'wx'});reports.push({name,args,sha256:hash(text)});
}
const r={type:'posthoc-profile-detail-not-confirmation',time:new Date().toISOString(),selector_repair:'original Service.Recall/coreStoreV36.Search pattern did not match parenthesized Go receiver notation; original summary retained, explicit recallOnce selector added',source_profile_sha256:hash(fs.readFileSync(path.join(dir,'profile.json'))),script_sha256:hash(fs.readFileSync('research/metadata-core-v36-profile-detail.mjs')),reports,scope:'16 diagnostic arms aggregate, includes fixture setup; cumulative CPU overlaps; block/mutex aggregate waiter delay is not per-request time'};
fs.writeFileSync(path.join(dir,'profile-detail.json'),JSON.stringify(r,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(r,null,2));
