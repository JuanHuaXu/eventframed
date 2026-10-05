import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';import path from 'node:path';import{fileURLToPath}from'node:url';
export const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const contract='public-source-document-pool-v1';
export function audit(docs,entries){
 assert.equal(entries.length,docs.length);let bytes=0,metaBytes=0,spans=0,maxBytes=0;
 const seen=new Set();for(let i=0;i<docs.length;i++){
  const d=docs[i],e=entries[i];assert.deepEqual(Object.keys(e).sort(),['source_id','source_sha256','available_at','span_ids','candidate'].sort());
  assert.equal(e.source_id,d.id);assert.equal(e.source_sha256,d.source_sha256);assert(!seen.has(e.source_id));seen.add(e.source_id);
  assert.deepEqual(e.span_ids,d.spans.map(s=>s.event.id));assert.equal(e.available_at,'2026-10-04T00:00:00Z');
  const c=e.candidate;assert.deepEqual(Object.keys(c).sort(),['ID','Text','Score','Metadata'].sort());assert.equal(c.Score,0);
  assert.equal(c.ID,'public-document-'+hash(contract+'\0'+e.span_ids[0]+'\0'+d.source_sha256));
  const expected=d.spans.map(s=>'representation: eventframe-5w1h-v1\nwhat: '+s.event.what.value.trim().replace(/ +/g,' ')+'\nwhere: '+s.event.where.value).join('\n\n');
  assert.equal(c.Text,expected);assert(Buffer.byteLength(c.Text)<=16384);const m=JSON.parse(Buffer.from(c.Metadata,'base64'));
  assert.deepEqual(m,{collection:'public-scientific',ts:Date.parse(e.available_at),authored:false,access_count:0,authority:0,salience:0,
   source_document_id:d.id,source_sha256:d.source_sha256,available_at:e.available_at,frame_contract:'public-source-assertion-span-v1',pool_contract:contract,span_ids:e.span_ids});
  bytes+=Buffer.byteLength(c.Text);metaBytes+=Buffer.from(c.Metadata,'base64').length;spans+=e.span_ids.length;maxBytes=Math.max(maxBytes,Buffer.byteLength(c.Text));
 }
 return{documents:entries.length,linkedSpans:spans,textBytes:bytes,metadataBytes:metaBytes,maxTextBytes:maxBytes,
  dimension768DocumentCachePayload:bytes+entries.length*(2+768*4),noDuplicateDocumentEvidence:true,labelsNotRead:true};
}
function corruptions(docs,entries){
 const mutations=[e=>e.pop(),e=>e.push(e[0]),e=>e[0].source_id='other',e=>e[0].source_sha256='bad',e=>e[0].span_ids.pop(),
  e=>e[0].available_at='2025-01-01T00:00:00Z',e=>e[0].candidate.ID='unknown',e=>e[0].candidate.Text+=' leaked',
  e=>e[0].candidate.Score=1,e=>e[0].candidate.Metadata=Buffer.from('{}').toString('base64'),e=>e[0].candidate.label='SUPPORT',
  e=>e[0].span_ids.reverse(),e=>e[0].candidate.Text=e[0].candidate.Text.split('\n\n')[0],e=>e[0].source_id=e[1].source_id];
 const d=docs.slice(0,2),e=entries.slice(0,2);audit(d,e);for(const m of mutations){const copy=structuredClone(e);m(copy);assert.throws(()=>audit(d,copy));}return mutations.length;
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 const[frames,pooled,out]=process.argv.slice(2),a=fs.readFileSync(frames),b=fs.readFileSync(pooled);
 const docs=JSON.parse(a),entries=JSON.parse(b),inventory=audit(docs,entries);assert.equal(inventory.documents,5183);assert.equal(inventory.linkedSpans,10869);
 const report={time:new Date().toISOString(),framesSHA256:hash(a),poolSHA256:hash(b),inventory,corruptionRejections:corruptions(docs,entries),allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
 fs.writeFileSync(out,JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(report,null,2));
}
