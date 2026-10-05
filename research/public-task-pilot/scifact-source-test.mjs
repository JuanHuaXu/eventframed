import assert from 'node:assert/strict';import{jsonLines,qrels,validate}from'./scifact-source.mjs';
assert.deepEqual(jsonLines(Buffer.from('{"a":1}\n{"a":2}\n')),[{a:1},{a:2}]);assert.throws(()=>jsonLines(Buffer.from('{bad}\n')));
const docs=[{_id:'d',title:'Title',text:'Abstract'}],queries=[{_id:'a',text:'Claim A'},{_id:'b',text:'Claim B'}];
const a=qrels(Buffer.from('query-id\tcorpus-id\tscore\na\td\t1\n')),b=qrels(Buffer.from('query-id\tcorpus-id\tscore\nb\td\t1\n'));assert.equal(validate(docs,queries,a,b).sharedEvidenceDocuments,1);
assert.throws(()=>qrels(Buffer.from('query-id\tcorpus-id\tscore\na\td\t1\na\td\t1\n')));assert.throws(()=>validate(docs,queries,a,a));assert.throws(()=>validate(docs,queries,a,[{query:'missing',document:'d',score:1}]));assert.throws(()=>validate(docs,queries,a,[{query:'b',document:'foreign',score:1}]));assert.throws(()=>validate(docs,queries,a,[{query:'b',document:'d',score:0}]));
console.log(JSON.stringify({positiveJSONAndQrelControls:2,negativeSchemaIdentitySplitControls:6,sharedDocumentOverlapExplicit:true,noBenchmarkResults:true}));
