import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const d=JSON.parse(fs.readFileSync(new URL('./nobel-v1/bounded-semantic-results.json',import.meta.url)));
for(const [f,h]of Object.entries(d.Hashes))assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.resolve(import.meta.dirname,'../..',f))).digest('hex'),h);
assert.equal(d.Results.length,42);const seen=new Set();let maxDelta=0,maxLawError=0;
for(const r of d.Results){const key=[r.Arm,r.Case].join(':');assert(!seen.has(key));seen.add(key);assert.equal(r.Candidates.length,19);assert.equal(new Set(r.Candidates.map(c=>c.Fixture)).size,19);}
const summaries=[];
for(const arm of ['baseline','sparse','bounded']){
 const rows=d.Results.filter(r=>r.Arm===arm);
 summaries.push({arm,literalTop1:rows.filter(r=>r.Case.endsWith('-literal')&&r.SupportRank===1).length,paraphraseTop1:rows.filter(r=>r.Case.endsWith('-paraphrase')&&r.SupportRank===1).length,positiveRetained:rows.filter(r=>r.SupportRank>0).length});
 if(arm==='baseline')continue;
 for(const r of rows){assert.equal(r.Frontier,19);const base=d.Results.find(b=>b.Arm==='baseline'&&b.Case===r.Case);
  for(const c of r.Candidates){const b=base.Candidates.find(b=>b.Fixture===c.Fixture);assert(b);
   for(const k of Object.keys(b.Law)){maxLawError=Math.max(maxLawError,Math.abs(b.Law[k]-c.Law[k]));assert(Math.abs(b.Law[k]-c.Law[k])<=1e-12);}
   if(arm==='bounded'){maxDelta=Math.max(maxDelta,Math.abs(b.Score-c.Score));assert(Math.abs(b.Score-c.Score)<=.005+1e-12);}
  }
 }
}
const b=summaries[0],c=summaries[2];
console.log(JSON.stringify({summaries,maxDelta,maxLawError,passed:c.positiveRetained===12&&c.literalTop1>=b.literalTop1&&c.paraphraseTop1>=b.paraphraseTop1&&c.literalTop1+c.paraphraseTop1>b.literalTop1+b.paraphraseTop1,scope:'consumed public tasks, rank-only correction, no generation'},null,2));
