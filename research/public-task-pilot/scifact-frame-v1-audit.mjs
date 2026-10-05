// Independent source-span reconstruction: no candidate imports or Go calls.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';

export const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const contract='public-source-assertion-span-v1';
const imported='2026-10-04T00:00:00Z';
// Unicode White_Space, matching the published normalization contract. JS \s
// differs at NEL and BOM, so its default regex is not an independent oracle.
const white=/[\u0009-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+/gu;
const canon=s=>s.split(white).filter(Boolean).join(' ');
// Go escapes these two scalar values even with HTML escaping disabled.
const encoded=value=>JSON.stringify(value).replaceAll('\u2028','\\u2028').replaceAll('\u2029','\\u2029');
const identity=value=>hash(encoded(value));
const keys=(x,want)=>assert.deepEqual(Object.keys(x).sort(),want.slice().sort());
const field=(value,evidence)=>({value,source:'observed',confidence:1,evidence});

export function audit(records,docs){
  assert.equal(docs.length,records.length);const seen=new Set();
  let frames=0,canonicalBytes=0,repeatedRawBytes=0,docCachePayload=0,maxFrames=0,maxField=0;
  for(let i=0;i<records.length;i++){
    const r=records[i],d=docs[i];keys(d,['id','source_sha256','canonical_title','canonical_abstract','spans']);
    assert.equal(d.id,r.id);assert(!seen.has(d.id));seen.add(d.id);
    const sha=identity([r.id,r.title,r.text]);assert.equal(d.source_sha256,sha);
    assert.equal(d.canonical_title,canon(r.title));assert.equal(d.canonical_abstract,canon(r.text));
    assert(d.spans.length>=2&&d.spans.length<=64);maxFrames=Math.max(maxFrames,d.spans.length);
    let index=0;
    for(const[section,source]of[['title',d.canonical_title],['abstract',d.canonical_abstract]]){
      const b=Buffer.from(source);canonicalBytes+=b.length;let offset=0;
      while(offset<b.length){
        const s=d.spans[index];assert(s,'missing source span');keys(s,['section','start','end','event']);
        assert.equal(s.section,section);assert.equal(s.start,offset);assert(Number.isSafeInteger(s.end)&&s.end>offset&&s.end<=b.length&&s.end-offset<=2048);
        // Invariant reconstruction plus independent maximal-prefix boundary:
        // UTF-8 continuation byte membership is tested directly, not via Go.
        let end=Math.min(b.length,offset+2048);
        while(end<b.length&&(b[end]&0xc0)===0x80)end--;
        if(end<b.length){const gap=b.lastIndexOf(0x20,end-1);if(gap>=offset)end=gap+1;}
        assert.equal(s.end,end,'changed span boundary');
        const value=b.subarray(offset,end).toString('utf8');assert.equal(Buffer.byteLength(value),end-offset);assert(value.trim());
        const e=s.event;
        keys(e,['id','tenant_id','session_id','sequence','kind','content','occurred_at','observed_at','available_at',
          'who','what','where','when','why','how','priority','provenance','attributes','embedding_model']);
        assert.equal(e.id,'public-span-'+identity([contract,'research-scifact','public-import',imported,r.id,sha,section,offset,end]));
        assert.equal(e.tenant_id,'research-scifact');assert.equal(e.session_id,'public-import');assert.equal(e.sequence,index+1);
        assert.equal(e.kind,'public_source_assertion');assert.equal(e.content,r.title+'\n\n'+r.text);
        for(const k of['occurred_at','observed_at','available_at'])assert.equal(e[k],imported);
        for(const k of['who','when','why','how'])assert.deepEqual(e[k],{value:'',source:'',confidence:0});
        assert.deepEqual(e.what,field(value,`canonical-${section}[bytes:${offset}:${end}]`));
        assert.deepEqual(e.where,field('public scientific '+section,'source section, not physical location'));
        assert.equal(e.priority,.5);assert.equal(e.embedding_model,'');
        assert.deepEqual(e.provenance,{producer:contract});
        assert.deepEqual(e.attributes,{source_document_id:r.id,source_sha256:sha,source_section:section,semantic_extractor:contract,
          clock_semantics:'import-not-publication',confidence_semantics:'transcription-not-truth'});
        const frame='representation: eventframe-5w1h-v1\nwhat: '+canon(value)+'\nwhere: public scientific '+section;
        docCachePayload+=2+Buffer.byteLength(frame)+768*4;
        repeatedRawBytes+=Buffer.byteLength(e.content);maxField=Math.max(maxField,Buffer.byteLength(value));frames++;index++;offset=end;
      }
    }
    assert.equal(index,d.spans.length,'extra source spans');
  }
  const queryEntries=351+180+300;
  const conservativeQueryPayload=queryEntries*(2+16384+768*4);
  return{documents:records.length,frames,maxFrames,maxField,canonicalBytes,repeatedRawBytes,docCachePayload,
    queryEntries,totalCacheEntries:frames+queryEntries,conservativeQueryPayload,
    conservativeCachePayload:docCachePayload+conservativeQueryPayload,
    existingCacheEntryCap:12000,existingCachePayloadCap:64<<20,
    cacheEntriesFit:frames+queryEntries<=12000,cachePayloadUpperEnvelopeFits:docCachePayload+conservativeQueryPayload<=64<<20,
    confidenceMeansTranscriptionNotTruth:true,unknownFieldsRemainUnknown:true,labelsNotRead:true,noQualityClaim:true};
}

export function corruptions(records,docs){
  const changes=[
    d=>d.pop(),d=>d.push(d[0]),d=>d[0].id='other',d=>d[0].source_sha256='0'.repeat(64),
    d=>d[0].canonical_abstract+=' hidden',d=>d[0].spans[0].start=1,d=>d[0].spans[0].end--,
    d=>d[0].spans[0].event.what.value+=' injected',d=>d[0].spans[0].event.content='replaced',
    d=>d[0].spans[0].event.who={value:'invented',source:'inferred',confidence:1},
    d=>d[0].spans[0].event.available_at='2025-01-01T00:00:00Z',
    d=>d[0].spans[0].event.attributes.label='SUPPORT',d=>d[0].spans[0].event.id='foreign',
    d=>d[0].spans[0].event.sequence=99,d=>d[0].spans[0].event.embedding_model='unfrozen',
    d=>d[0].spans[0].event.priority=1,d=>d[0].spans.reverse(),
    d=>d[0].spans[0].event.what.confidence=.5,d=>d[0].spans[0].event.why={value:'because invented',source:'inferred',confidence:1},
    d=>d[0].spans[0].event.tags=['fit'],
  ];
  // Full baseline reconstruction executes first; fixture mutations then retain
  // the same independent checks without cloning tens of MiB for each mutation.
  const r=records.slice(0,2),small=docs.slice(0,2);audit(r,small);
  for(const change of changes){const altered=structuredClone(small);change(altered);assert.throws(()=>audit(r,altered));}
  return changes.length;
}

if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
  const [corpus,out,destination]=process.argv.slice(2);assert(corpus&&out&&destination);
  const raw=fs.readFileSync(corpus),bytes=fs.readFileSync(out);
  const records=JSON.parse(raw),docs=JSON.parse(bytes);const inventory=audit(records,docs);
  assert.equal(inventory.documents,5183);
  const report={time:new Date().toISOString(),corpusSHA256:hash(raw),framesSHA256:hash(bytes),outputBytes:bytes.length,
    inventory,corruptionRejections:corruptions(records,docs),allSevenWholeGoals:'OPEN',goal:'ACTIVE'};
  fs.writeFileSync(destination,JSON.stringify(report,null,2)+'\n',{flag:'wx',mode:0o600});
  console.log(JSON.stringify(report,null,2));
}
