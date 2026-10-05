import fs from 'node:fs';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';

// Technical checks, not fresh outcome/cohort data. Exclusive evidence outputs.
const output=process.argv[2];if(!output)throw Error('output directory required');
fs.mkdirSync(output,{recursive:true});
const checks=[
 ['core-race',['test','-race','./internal/researchadmission','./internal/researchvalidity','-count=3','-v'],{}],
 ['lifecycle-race',['test','-race','./internal/store/libravdbstore','-run','^TestResearch(ScheduledWitness(Lifecycle|Cancellation)V24|JoinedWitness(Lifecycle|Cancellation|Batch|Gap)V25)$','-count=1','-v','-timeout=3m'],{EVENTFRAME_RUN_SCHEDULED_WITNESS_V24:'1',EVENTFRAME_RUN_JOINED_WITNESS_V25:'1'}],
 ['vet',['vet','./internal/researchadmission','./internal/researchvalidity','./internal/store/libravdbstore'],{}],
 ['lease-benchmark',['test','./internal/researchadmission','-run','^$','-bench','^BenchmarkLease$','-benchmem','-count=3'],{}],
];
const reports=[];let failed=false;
for(const [name,args,env] of checks){
 const file=`${output}/${name}.log`;if(fs.existsSync(file))throw Error('exclusive output exists '+file);
 const start=new Date().toISOString(),begin=performance.now();let code=0,log;
 try{log=execFileSync('go',args,{env:{...process.env,...env},encoding:'utf8',maxBuffer:8*1024*1024,timeout:200000});}
 catch(e){code=e.status??-1;log=(e.stdout??'')+'\n'+(e.stderr??'')+'\n'+e.message;failed=true;}
 fs.writeFileSync(file,log,{flag:'wx'});reports.push({name,args,env,start,end:new Date().toISOString(),wall_ms:performance.now()-begin,code,sha256:crypto.createHash('sha256').update(log).digest('hex')});
 console.log(name,code,log.slice(-2000));
}
fs.writeFileSync(`${output}/checks.json`,JSON.stringify(reports,null,2)+'\n',{flag:'wx'});process.exitCode=failed?1:0;
