// Preserve the completed CRLF preflight failure and reuse its verified archive.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
const hash=(b,a='sha256')=>crypto.createHash(a).update(b).digest('hex');
const from='research/public-task-pilot/nfcorpus-v1-acquire.mjs',old='research/public-task-pilot/nfcorpus-v1',to='research/public-task-pilot/nfcorpus-v2-acquire.mjs';
const original=fs.readFileSync(from,'utf8'),zip=fs.readFileSync(old+'/source.zip');assert.equal(hash(zip,'md5'),'a89dba18a62ef92f7d323ec890a0d38d');assert(!fs.existsSync(old+'/source.json'));assert.deepEqual(fs.readdirSync(old),['source.zip']);
fs.writeFileSync(old+'/failed-acquire.mjs',original,{flag:'wx',mode:0o600});
fs.writeFileSync(old+'/failure.json',JSON.stringify({time:new Date().toISOString(),terminalCode:1,reason:'Official test TSV uses CRLF; header parser retained terminal CR and rejected otherwise exact header.',observedActual:'query-id\tcorpus-id\tscore\r',expected:'query-id\tcorpus-id\tscore',scriptSHA256:hash(original),zipSHA256:hash(zip),md5:hash(zip,'md5'),documentsParsed:3633,queriesPredicted:0,outcomesEvaluated:0,scientificModelsChanged:false},null,2)+'\n',{flag:'wx',mode:0o600});
let next=original.replaceAll('nfcorpus-v1','nfcorpus-v2');
const download="execFileSync('/usr/bin/curl',['--fail','--location','--max-time','120','--proto','=https','--proto-redir','=https','--output',path.resolve(root+'/source.zip'),url],{stdio:['ignore','pipe','pipe']});";
assert.equal(next.split(download).length,2);next=next.replace(download,`fs.copyFileSync('${old}/source.zip',root+'/source.zip',fs.constants.COPYFILE_EXCL);`);
const header="const lines=testRaw.toString('utf8').trim().split('\\n');";assert.equal(next.split(header).length,2);next=next.replace(header,"const lines=testRaw.toString('utf8').trim().split(/\\r?\\n/);");
const stamp="url,md5:hash(zip,'md5'),";assert.equal(next.split(stamp).length,2);next=next.replace(stamp,`url,archiveReusedFrom:'${old}/source.zip',originalFailure:'${old}/failure.json',lineEndingRepairOnly:true,md5:hash(zip,'md5'),`);
fs.writeFileSync(to,next,{flag:'wx',mode:0o600});fs.writeFileSync('research/public-task-pilot/nfcorpus-v2-generation.json',JSON.stringify({time:new Date().toISOString(),from,to,fromSHA256:hash(original),toSHA256:hash(next),changes:['new owned root/script','reuse verified original archive, no second network acquisition','accept CRLF TSV line grammar','record failure/reuse lineage'],noCorpusOrQueryTextChange:true,predictions:0,outcomesEvaluated:0},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({to,originalFailurePreserved:true,zipMD5Verified:true}));
