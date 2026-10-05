// Mechanical clones with exact declared replacements; original studies untouched.
import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),records=[];
const write=(p,b,from)=>{assert(!fs.existsSync(p));fs.mkdirSync(p.substring(0,p.lastIndexOf('/')),{recursive:true});fs.writeFileSync(p,b,{flag:'wx',mode:0o600});records.push({path:p,sha256:hash(b),from,fromSHA256:hash(fs.readFileSync(from))});};
function replace(s,from,to,n=1){assert.equal(s.split(from).length-1,n,from);return s.replaceAll(from,to);}
const original='cmd/research-public-sourceonly-pairrank/main.go';let command=fs.readFileSync(original,'utf8');
command=replace(command,'r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"','idf "github.com/JuanHuaXu/eventframed/internal/researchpublicidf"\n\tr "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"');
command=replace(command,'r.NewSources(ctx, in.Sources, in.AsOf, in.Epoch)','idf.NewSources(ctx, in.Sources, in.AsOf, in.Epoch)');
command=replace(command,'"contract": r.Contract','"contract": r.Contract, "featureContract": idf.Contract',2);
const formatted=execFileSync('gofmt',[],{input:command,encoding:'utf8'});
write('cmd/research-public-idf-pairrank/main.go',formatted,original);
const parent='research/public-task-pilot/scifact-sourceonly-pairrank-v1-run.mjs';let runner=fs.readFileSync(parent,'utf8');
runner=runner.replaceAll('scifact-sourceonly-pairrank-v1','scifact-idf-pairrank-v1').replaceAll('research-public-sourceonly-pairrank','research-public-idf-pairrank');
runner=replace(runner,'scifact-sourceonly-v1-gen.mjs','scifact-idf-pairrank-v1-gen.mjs');runner=replace(runner,'scifact-sourceonly-v1-generation.json','scifact-idf-pairrank-v1-generation.json');runner=replace(runner,'docs/experiments/scifact-sourceonly-v1-protocol.md','docs/experiments/scifact-idf-pairrank-v1-protocol.md');
// Explicit root includes IDF tests in both source closure and actual race runs.
runner=runner.replaceAll("'./internal/researchpublicrankmask'","'./internal/researchpublicrankmask','./internal/researchpublicidf'");
runner=replace(runner,"const inputs={};",`paths.add('research/public-task-pilot/scifact-sourceonly-pairrank-v1-audit.mjs');\nconst inputs={};`);
runner=replace(runner,"root+'/fit-labels.json']","root+'/fit-labels.json','research/public-task-pilot/scifact-sourceonly-pairrank-v1/predictions.json','research/public-task-pilot/scifact-sourceonly-pairrank-v1/manifest.json','research/public-task-pilot/scifact-sourceonly-pairrank-v1/audit-results.json']");
write('research/public-task-pilot/scifact-idf-pairrank-v1-run.mjs',runner,parent);
const parentAudit='research/public-task-pilot/scifact-sourceonly-pairrank-v1-audit.mjs';let audit=fs.readFileSync(parentAudit,'utf8').replaceAll('scifact-sourceonly-pairrank-v1','scifact-idf-pairrank-v1');
const begin=audit.indexOf('function feature(q,id,n){'),end=audit.indexOf('\nconst family=',begin);assert(begin>0&&end>begin);
const feature=`const corpusDF=new Map();for(const d of doc.values())for(const t of new Set([...d.ts,...d.bs]))corpusDF.set(t,(corpusDF.get(t)??0)+1);
const termWeight=t=>Math.log1p((5183-(corpusDF.get(t)??0)+.5)/((corpusDF.get(t)??0)+.5));
function feature(q,id,n){const d=doc.get(id),o=token(q),u=[...new Set(o)].sort();let den=0,title=0,body=0,num=0,numHit=0,bigDen=0,bigHit=0;
 for(const t of u){const w=termWeight(t);den+=w;if(d.ts.has(t))title+=w;if(d.bs.has(t))body+=w;if(/\\p{Nd}/u.test(t)){num+=w;if(d.bs.has(t))numHit+=w;}}
 for(let i=0;i+1<o.length;i++){const w=(termWeight(o[i])+termWeight(o[i+1]))/2;bigDen+=w;if(d.title.some((t,j)=>j+1<d.title.length&&t===o[i]&&d.title[j+1]===o[i+1]))bigHit+=w;}
 return[title/den,body/den,bigDen?bigHit/bigDen:0,num?numHit/num:0,Math.min(1,Math.log1p(d.title.length)/Math.log1p(8192)),Math.min(1,Math.log1p(d.body.length)/Math.log1p(8192)),0,0];}
`;
audit=audit.slice(0,begin)+feature+audit.slice(end);
audit=replace(audit,'const input=read(root+',`const rawRoot='research/public-task-pilot/scifact-sourceonly-pairrank-v1',rawAudit=read(rawRoot+'/audit-results.json'),rawManifest=read(rawRoot+'/manifest.json'),rawPred=read(rawRoot+'/predictions.json');assert.equal(await sha(rawRoot+'/manifest.json'),rawAudit.manifestSHA256);assert.equal(await sha(rawRoot+'/predictions.json'),rawManifest.artifacts['predictions.json']);assert.equal(rawAudit.calibrationPredictions,0);\nconst input=read(root+`);
audit=replace(audit,'assert.equal(features.labelsRead,false);',`assert.equal(features.labelsRead,false);assert.equal(features.featureContract,'public-source-idf-features-v1');assert.equal(pred.featureContract,'public-source-idf-features-v1');`);
audit=replace(audit,"'TestSourceOnlyInvalidMaskedCueAndCancel'];","'TestSourceOnlyInvalidMaskedCueAndCancel','TestIDFUnionAndAbsentTerms','TestIDFSourceEpochFutureAndCancellation','TestIDFImmutableConcurrentAndNativeInvariant'];");
audit=replace(audit,'raceRoots:8','raceRoots:11');
const marker='const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);';
const compare=`for(let i=0;i<351;i++){assert.equal(rawPred.predictions[i].id,pred.predictions[i].id);assert.deepEqual(rawPred.predictions[i].baseline,pred.predictions[i].baseline);}
const rawDetails=new Map(rawAudit.details.map(d=>[d.id,d]));const rawUnits=new Map(rawAudit.unitRows.map(d=>[d.family,d]));
const vsRaw=Object.fromEntries(Object.keys(summary.baseline).map(k=>{const delta=unitRows.map(u=>u.learned[k]-rawUnits.get(u.family).learned[k]);return[k,{mean:mean(delta),better:delta.filter(v=>v>1e-14).length,worse:delta.filter(v=>v< -1e-14).length,tied:delta.filter(v=>Math.abs(v)<=1e-14).length}];}));
const meanScreenPassed=['recall10','ndcg10'].every(k=>summary.learned[k]>summary.baseline[k]&&summary.learned[k]>rawAudit.summary.learned[k]&&unitSummary.learned[k]>unitSummary.baseline[k]&&unitSummary.learned[k]>rawAudit.unitSummary.learned[k]);
`;
audit=replace(audit,marker,compare+marker);
audit=replace(audit,'summary,unitSummary,paired,cost,',`summary,unitSummary,paired,vsRaw,unweightedSummary:rawAudit.summary.learned,unweightedUnitSummary:rawAudit.unitSummary.learned,meanScreenPassed,dfTerms:corpusDF.size,featureContract:'public-source-idf-features-v1',cost,`);
write('research/public-task-pilot/scifact-idf-pairrank-v1-audit.mjs',audit,parentAudit);
fs.writeFileSync('research/public-task-pilot/scifact-idf-pairrank-v1-generation.json',JSON.stringify({time:new Date().toISOString(),records,originalInputsFrontierOptimizerUnchanged:true,changes:['source-IDF import/NewSources','featureContract markers','new owned paths and provenance','IDF test roots','independent weighted feature reconstruction','immutable unweighted comparator and conjunctive mean screen']},null,2)+'\n',{flag:'wx',mode:0o600});
console.log(JSON.stringify({generated:records.map(r=>r.path)}));
