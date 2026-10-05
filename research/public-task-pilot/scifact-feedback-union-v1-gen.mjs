// Mechanical isolated command/runner/auditor clone; no original source edits.
import fs from'node:fs';import assert from'node:assert/strict';import crypto from'node:crypto';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex'),receipt={changes:'Nomination union<=400, original-query scoring before cap, fixed original models/feedback; explicit repeated scans',clones:[]};
function clone(src,dst,edits){let b=fs.readFileSync(src,'utf8'),n=b;for(const[old,replacement]of edits){assert(n.includes(old),old);n=n.replaceAll(old,replacement);}fs.mkdirSync(dst.substring(0,dst.lastIndexOf('/')),{recursive:true,mode:0o700});fs.writeFileSync(dst,n,{flag:'wx',mode:0o600});receipt.clones.push({source:src,sourceSHA256:hash(b),target:dst,generatedBeforeGofmtSHA256:hash(n)});}
clone('cmd/research-public-feedback/main.go','cmd/research-public-feedback-union/main.go',[
 ['r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"','r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"\n u "github.com/JuanHuaXu/eventframed/internal/researchpublicfrontierunion"'],
 ['\t\tbr, e := rows(ctx, s, q, in.Epoch, base)','union, e := u.OriginalScores(ctx,x,q.Text,base,expanded)\n if e!=nil{return e}\n unionNS:=time.Since(t).Nanoseconds()\n t=time.Now()\n br, e := rows(ctx, s, q, in.Epoch, base)'],
 ['rows(ctx, s, q, in.Epoch, expanded)','rows(ctx, s, q, in.Epoch, union)'],
 ['rank(ctx, r.Model{}, er)','u.Rank(ctx, r.Model{}, er,200)'],['rank(ctx, model[q.Fold], er)','u.Rank(ctx, model[q.Fold], er,200)'],
 ['"expandedRows": er','"unionRows": er, "unionNominees": union'],['"feedback": ex, "feedbackIDF": ec','"union": ex, "unionIDF": ec'],['"expandedNS": expandedNS','"expandedNS": expandedNS, "unionNS": unionNS'],['"kind": "start",','"kind": "start", "unionRescue":true, "maxUnion":400,']]);
clone('cmd/research-public-feedback/main_test.go','cmd/research-public-feedback-union/main_test.go',[]);
clone('research/public-task-pilot/scifact-feedback-v1-run.mjs','research/public-task-pilot/scifact-feedback-union-v1-run.mjs',[
 ['scifact-feedback-v1','scifact-feedback-union-v1'],['./cmd/research-public-feedback','./cmd/research-public-feedback-union'],['\'./internal/researchpublicfeedback\',','\'./internal/researchpublicfeedback\',\'./internal/researchpublicfrontierunion\','],
 ['const files=new Set([','const files=new Set([\'research/public-task-pilot/scifact-feedback-union-v1-gen.mjs\','],
 ['prior+\'/predictions.json\',','prior+\'/predictions.json\',\'research/public-task-pilot/scifact-feedback-v1/manifest.json\',\'research/public-task-pilot/scifact-feedback-v1/audit-results.json\',\'research/public-task-pilot/scifact-feedback-v1/predictions.ndjson\',']]);
clone('research/public-task-pilot/scifact-feedback-v1-audit.mjs','research/public-task-pilot/scifact-feedback-union-v1-audit.mjs',[
 ['scifact-feedback-v1','scifact-feedback-union-v1'],
 ['assert.equal(start.kind,\'start\');','assert.equal(start.kind,\'start\');assert.equal(start.unionRescue,true);assert.equal(start.maxUnion,400);'],
 ['function scores(terms){','function scores(terms,cap=200){'],['.slice(0,200);}', '.slice(0,cap);}'],
 ['compare(p.expanded,ex);','compare(p.expanded,ex);const allOriginal=new Map(scores(new Map([...new Set(token(q.text))].map(t=>[t,1])),5183).map(h=>[h.id,h.score]));const unionIDs=[...b.map(h=>h.id),...ex.filter(h=>!b.some(v=>v.id===h.id)).map(h=>h.id)],union=unionIDs.map(id=>({id,score:allOriginal.get(id)??0}));assert(union.length<=400);compare(p.unionNominees,union);'],
 ["[ex,'expandedRows']","[union,'unionRows']"],['rows[j].base,hits[j].score/hits[0].score','rows[j].base,hits[j].score/Math.max(...hits.map(h=>h.score))'],
 ["['feedback',p.expandedRows,Array(8).fill(0)],['feedbackIDF',p.expandedRows,models[q.fold].weights]","['union',p.unionRows,Array(8).fill(0)],['unionIDF',p.unionRows,models[q.fold].weights]"],
 ['compare(p[arm],sort(v));','compare(p[arm],sort(v).slice(0,200));'],
 ["'expandedNS','featureNS'","'expandedNS','unionNS','featureNS'"],['expandedRows','unionRows'],['feedbackIDF','unionIDF'],["'feedback','unionIDF'","'union','unionIDF'"],
 ['p.expandedNS+p.featureNS','p.expandedNS+p.unionNS+p.featureNS'],
 ['expandedFullSearchNS:dist(pred.map(p=>p.expandedNS)),','expandedFullSearchNS:dist(pred.map(p=>p.expandedNS)),unionOriginalScansNS:dist(pred.map(p=>p.unionNS)),'],
 ['pred.forEach((p,i)=>check(i,p));','pred.forEach((p,i)=>check(i,p));for(const p of pred){assert.deepEqual(p.union,p.plain);}'],
 ['nomination:{meanAdded:','nomination:{maxUnion:Math.max(...pred.map(p=>p.unionNominees.length)),meanUnion:mean(pred.map(p=>p.unionNominees.length)),allUnionScoredBeforeCap:true,meanAdded:']]);
fs.writeFileSync('research/public-task-pilot/scifact-feedback-union-v1-generation.json',JSON.stringify(receipt,null,2)+'\n',{flag:'wx',mode:0o600});
