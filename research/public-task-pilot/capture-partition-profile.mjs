import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {spawnSync} from 'node:child_process';
const input='research/public-task-pilot/partition-profile-results.json';
const d=JSON.parse(fs.readFileSync(input));assert.equal(d.Arms.length,2);
for(const [p,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),h,p);
const out='research/public-task-pilot/partition-profile-analysis';fs.mkdirSync(out);
const hashes={};const summaries=[];
for(const a of d.Arms){
 assert([3200,6400].includes(a.N)&&a.Repeat===0&&a.GapMS===5);
 assert.equal(a.Reads.length,1024);assert.equal(a.Writes.length,512);
 assert.deepEqual(JSON.parse(fs.readFileSync(`${input}.arm-n${a.N}-gap5-r0.json`)),a);
 const dir=`${input}.stores/n${a.N}-gap5-r0`,cpu=`${dir}/cpu.pprof`,before=`${dir}/alloc-before.pprof`,after=`${dir}/alloc-after.pprof`;
 for(const p of [cpu,before,after])hashes[p]=crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
 const commands={tags:['-tags',cpu],compact:['-top','-nodecount=25','-tagfocus=phase=compact',cpu],read:['-top','-nodecount=20','-tagfocus=phase=read',cpu],write:['-top','-nodecount=20','-tagfocus=phase=write',cpu],alloc:['-top','-nodecount=25','-sample_index=alloc_space','-base',before,after]};
 for(const [name,args]of Object.entries(commands)){
  const r=spawnSync('go',['tool','pprof',...args],{encoding:'utf8'});assert.equal(r.status,0,r.stderr);
  fs.writeFileSync(`${out}/n${a.N}-${name}.txt`,r.stdout,{flag:'wx'});
 }
 const b=a.MemBefore,c=a.MemAfter;
 const pauses=[];for(let i=b.NumGC;i<c.NumGC;i++)pauses.push(c.PauseNs[i%256]);
 summaries.push({n:a.N,allocationMiB:(c.TotalAlloc-b.TotalAlloc)/2**20,mallocs:c.Mallocs-b.Mallocs,gcCycles:c.NumGC-b.NumGC,gcPauseMS:(c.PauseTotalNs-b.PauseTotalNs)/1e6,maxGCPauseMS:Math.max(0,...pauses)/1e6,builds:a.Builds.length,ack:a.Acknowledged,reopened:a.Reopened});
}
fs.writeFileSync(`${out}/manifest.json`,JSON.stringify({input,hashes,summaries},null,2),{flag:'wx'});
console.log(JSON.stringify(summaries,null,2));
