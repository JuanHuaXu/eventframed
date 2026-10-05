// Mechanical fixture separation. The runner never opens facts.json or oracle.
import fs from 'node:fs';
import crypto from 'node:crypto';
const root='research/public-task-pilot/git-fusion-v1/';
const facts=JSON.parse(fs.readFileSync(root+'facts.json','utf8'));
if(facts.length!==48||new Set(facts.map(f=>f.c)).size!==16)throw Error('fixture shape');
const families=[...new Set(facts.map(f=>f.c))];
const corpus=[],queries=[],oracle=[];
facts.forEach((f,i)=>{
  if(!/^[\x20-\x7e]+$/.test(f.t+f.l+f.p)||!f.t.includes(f.a))throw Error('grounding/ASCII');
  const id=`record-${String(i+1).padStart(3,'0')}`;
  corpus.push({fixture_id:id,text:f.t});
  for(const [wording,question] of [['literal',f.l],['paraphrase',f.p]]){
    if(question.includes(f.a))throw Error(`answer token leakage ${id}`);
    const case_id=`question-${String(queries.length+1).padStart(3,'0')}`;
    queries.push({case_id,question});
    oracle.push({case_id,target:id,answer:f.a,family:f.c,split:families.indexOf(f.c)<8?'design':'confirmation',wording,source:`https://git-scm.com/docs/git-${f.c}`});
  }
});
for(const [family,question] of [
  ['reset','Which git reset option encrypts the staging area with a password?'],
  ['merge','What compression ratio does git merge guarantee for image files?'],
  ['fetch','What network bandwidth does git fetch guarantee?'],
  ['worktree','Which git worktree command guarantees a fixed number of CPU cores?']]){
  const case_id=`question-${String(queries.length+1).padStart(3,'0')}`;
  queries.push({case_id,question});oracle.push({case_id,target:null,answer:null,family,split:families.indexOf(family)<8?'design':'confirmation',wording:'absent',source:null});
}
// Deterministic insertion shuffle independent of labels and query wording.
let state=32452843;
for(let i=corpus.length-1;i>0;i--){state=(Math.imul(state,1664525)+1013904223)>>>0;const j=state%(i+1);[corpus[i],corpus[j]]=[corpus[j],corpus[i]];}
for(const [name,data] of [['corpus.json',corpus],['queries.json',queries],['oracle.json',oracle]]){
  fs.writeFileSync(root+name,JSON.stringify(data,null,2)+'\n',{flag:'wx',mode:0o600});
  console.log(name,crypto.createHash('sha256').update(fs.readFileSync(root+name)).digest('hex'));
}
