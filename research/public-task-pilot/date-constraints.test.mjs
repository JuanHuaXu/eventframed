import {test} from 'node:test';
import assert from 'node:assert/strict';
import {constraints,recordDate,contradiction,partition} from './date-constraints.mjs';
test('calendar validation and ambiguous records',()=>{
  assert.equal(recordDate('29 February 1900'),null);
  assert.equal(recordDate('29 February 2000').key,20000229);
  assert.equal(recordDate('31 April 2016'),null);
  assert.equal(recordDate('1 November 2016 and 15 September 2016'),null);
  assert.equal(recordDate('no date'),null);
});
test('strict year comparisons',()=>{
  for(const [q,d,want]of [['before 2016','1 January 2016',true],['before 2016','31 December 2015',false],['after 2016','31 December 2016',true],['after 2016','1 January 2017',false]])
    assert.equal(contradiction(constraints(q),d),want);
});
test('exact and excluded dates are retained',()=>{
  const q='What status on 1 November 2016, rather than 15 September 2016?';
  assert.deepEqual(constraints(q),[{kind:'on',value:20161101},{kind:'exclude',value:20160915}]);
  assert(contradiction(constraints(q),'published 15 September 2016'));
  assert(!contradiction(constraints(q),'published 1 November 2016'));
});
test('unsupported and absent rules leave records alone',()=>{
  for(const q of ['before 2016 and after 2014','not before 2016','before 2016, rather than a draft?','on 31 April 2016','What happened?'])
    assert.deepEqual(constraints(q),[]);
  assert(!contradiction(constraints('before 2016'),'date unknown'));
});
test('partition is stable, preserves unknown, and flags all contradicted',()=>{
  const rows=[{id:1,text:'1 November 2016'},{id:2,text:'unknown'},{id:3,text:'28 October 2014'}];
  assert.deepEqual(partition('before 2016',rows).ordered.map(r=>r.id),[2,3,1]);
  assert.deepEqual(rows.map(r=>r.id),[1,2,3]);
  const all=partition('before 2010',[rows[0],rows[2]]);
  assert(all.allContradicted);assert.deepEqual(all.ordered.map(r=>r.id),[1,3]);
});
