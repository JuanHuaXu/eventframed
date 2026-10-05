import {test} from 'node:test';
import assert from 'node:assert/strict';
import {normalizeRecord} from './date-normalized-strict.mjs';
test('ISO token boundaries reject partial dates and timestamps',()=>{
  for(const s of ['2004-03-020','2004-03-02suffix','2004-03-02+01:00',
    '2004-03-02T07:17:00Z','2004-03-02 07:17:00','2004-03-02.1',
    '-2004-03-02','2004-03-02-other','2004-03-02/next']) assert.equal(normalizeRecord(s),'',s);
  for(const s of ['2004-03-02','Launched on 2004-03-02.','(2004-03-02)','date:2004-03-02'])
    assert.equal(normalizeRecord(s),'2 March 2004',s);
});
