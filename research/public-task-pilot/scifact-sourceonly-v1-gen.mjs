import fs from'node:fs';import assert from'node:assert/strict';import crypto from'node:crypto';import{execFileSync}from'node:child_process';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),records=[];
function put(path,s,from){fs.mkdirSync(path.slice(0,path.lastIndexOf('/')),{recursive:true,mode:0o700});fs.writeFileSync(path,s,{flag:'wx',mode:0o600});records.push({path,sha256:hash(s),from});}
for(const kind of ['pairrank','lambdarank']){
 const old='scifact-'+kind+'-v1',name='scifact-sourceonly-'+kind+'-v1',cmd='cmd/research-public-'+kind+'/main.go',nextCmd='cmd/research-public-sourceonly-'+kind+'/main.go',raw=fs.readFileSync(cmd,'utf8');
 const marker='if _, e := (r.Model{}).Rank(ctx, c.Rows); e != nil {';assert.equal(raw.split(marker).length,2);
 let source=raw.replace('r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"','r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"\nmask "github.com/JuanHuaXu/eventframed/internal/researchpublicrankmask"').replace(marker,'c.Rows, e = mask.SourceOnly(ctx, c.Rows)\nif e != nil { return e }\n'+marker);
 source=execFileSync('gofmt',[],{input:source,encoding:'utf8'});put(nextCmd,source,{path:cmd,sha256:hash(raw),change:'Validate/copy/zero only6,7 before fitting and scoring; post-gofmt hash'});
 const rp='research/public-task-pilot/'+old+'-run.mjs',r=fs.readFileSync(rp,'utf8');let runner=r.replaceAll(old,name).replaceAll('cmd/research-public-'+kind,'cmd/research-public-sourceonly-'+kind);
 runner=runner.replaceAll("'./internal/researchpublicpairrank'","'./internal/researchpublicpairrank','./internal/researchpublicrankmask'");
 // Both commands share this predeclared protocol, not an arm-specific rewrite.
 runner=runner.replace('docs/experiments/'+name+'-protocol.md','docs/experiments/scifact-sourceonly-v1-protocol.md');
 runner=runner.replace("'go.mod','go.sum'","'go.mod','go.sum','research/public-task-pilot/scifact-sourceonly-v1-gen.mjs','research/public-task-pilot/scifact-sourceonly-v1-generation.json'");
 put('research/public-task-pilot/'+name+'-run.mjs',runner,{path:rp,sha256:hash(r),change:'New owned root/mask tests and complete closure'});
 const ap='research/public-task-pilot/'+old+(kind==='pairrank'?'-audit-v2.mjs':'-audit.mjs'),a=fs.readFileSync(ap,'utf8');let audit=a.replaceAll(old,name).replace(name+'-audit-v2.mjs',name+'-audit.mjs');
 const fm='const f=feature(q.text,r.id,r.nativeRank);';assert.equal(audit.split(fm).length,2);audit=audit.replace(fm,fm+'f[6]=0;f[7]=0;');
 audit=audit.replace("'TestPairRankConcurrent'","'TestPairRankConcurrent','TestSourceOnlyExactOwnedCopy','TestSourceOnlyNativeInvariance','TestSourceOnlyInvalidMaskedCueAndCancel'");audit=audit.replace('raceRoots:'+(kind==='pairrank'?5:7),'raceRoots:'+(kind==='pairrank'?8:10));
 // The Lambda evaluator's scientific byte-identity reference is the original
 // unmasked study, not a nonexistent source-only counterpart.
 audit=audit.replaceAll("research/public-task-pilot/scifact-sourceonly-pairrank-v1/input.json","research/public-task-pilot/scifact-pairrank-v1/input.json").replaceAll("research/public-task-pilot/scifact-sourceonly-pairrank-v1/fit-labels.json","research/public-task-pilot/scifact-pairrank-v1/fit-labels.json");
 audit=audit.replace('const input=read(root+',"assert.equal(await sha(root+'/input.json'),await sha('research/public-task-pilot/"+old+"/input.json'));assert.equal(await sha(root+'/fit-labels.json'),await sha('research/public-task-pilot/"+old+"/fit-labels.json'));\nconst input=read(root+");
 put('research/public-task-pilot/'+name+'-audit.mjs',audit,{path:ap,sha256:hash(a),change:'Independent raw feature reconstruction then exact mask, full6model refit and original input byte identity'});
}
fs.writeFileSync('research/public-task-pilot/scifact-sourceonly-v1-generation.json',JSON.stringify({generatorSHA256:hash(fs.readFileSync('research/public-task-pilot/scifact-sourceonly-v1-gen.mjs')),records},null,2)+'\n',{flag:'wx',mode:0o600});console.log(records.map(r=>r.path));
