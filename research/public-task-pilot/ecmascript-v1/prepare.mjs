import fs from 'node:fs';import path from 'node:path';import vm from 'node:vm';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {fileURLToPath} from 'node:url';import {cases,groups} from './cases.mjs';
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
export function observe(program){
 let value;try{value=vm.runInNewContext(program,Object.create(null),{timeout:100,contextCodeGeneration:{strings:false,wasm:false}})}catch(e){return 'error:'+e.name}
 if(value===undefined)return 'undefined';if(value===null)return 'null';
 if(typeof value==='number'){if(Object.is(value,-0))return 'number:-0';return 'number:'+String(value)}
 if(typeof value==='string'||typeof value==='boolean')return typeof value+':'+String(value);
 throw Error('non-scalar observable');
}
export function prepare(){
 const corpus=[],queries=[],oracle=[];
 for(const c of cases){const actual=observe(c.program);assert.equal(actual,c.expected,c.group+'/'+c.label);
  const id=hash('ecma-corpus:'+c.group+'/'+c.label).slice(0,24);
  corpus.push({id,frame:{who:'ECMAScript runtime',what:c.group+': '+c.label,when:'specified ECMAScript 2025 behavior',where:'JavaScript language',why:'resolve program behavior',how:c.program+' returns '+actual},metadata:{source:c.source,program:c.program}});
  for(const[wording,text]of [['literal','In JavaScript, what value or exception does this expression produce? '+c.program],['paraphrase',c.paraphrase]]){
   const qid=hash('ecma-question:'+wording+':'+text).slice(0,24);queries.push({id:qid,split:c.split,wording,text});oracle.push({question_id:qid,target_id:id,cluster:c.group,split:c.split,observable:actual,program:c.program});
  }
 }
 assert.equal(new Set(corpus.map(x=>x.id)).size,54);assert.equal(new Set(queries.map(x=>x.id)).size,108);
 return {corpus,queries,oracle,groups:groups.length};
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 const dir=path.resolve('research/public-task-pilot/ecmascript-v1/prepared');fs.mkdirSync(dir,{recursive:true});
 const sources={};for(const p of ['cases.mjs','prepare.mjs','prepare.test.mjs','STATUS.md'])sources[p]=hash(fs.readFileSync(new URL(p,import.meta.url)));
 fs.writeFileSync(path.join(dir,'freeze.json'),JSON.stringify({time:new Date().toISOString(),sources,node:process.version,v8:process.versions.v8,retrieval_runs:0,whole_goal_complete:false},null,2)+'\n',{flag:'wx',mode:0o600});
 const result=prepare(),artifacts={};for(const name of ['corpus','queries','oracle']){const data=JSON.stringify(result[name],null,2)+'\n';fs.writeFileSync(path.join(dir,name+'.json'),data,{flag:'wx',mode:0o600});artifacts[name]=hash(data)}
 const report={time:new Date().toISOString(),corpus:result.corpus.length,questions:result.queries.length,clusters:result.groups,split_clusters:{fit:6,design:6,confirmation:6},split_questions:{fit:36,design:36,confirmation:36},verified_programs:54,artifacts,retrieval_runs:0,agent_runs:0,whole_goal_complete:false};fs.writeFileSync(path.join(dir,'prepared.json'),JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(report,null,2));
}
