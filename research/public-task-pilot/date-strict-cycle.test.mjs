import {test} from 'node:test';
import assert from 'node:assert/strict';
import {normalizeRecord,partition} from './date-normalized-strict.mjs';
import {recordDate,partition as originalPartition,contradiction,constraints} from './date-constraints.mjs';

test('full Gregorian cycle agrees with independent UTC calendar',()=>{
  let valid=0,invalid=0;
  const rules=[...constraints('before 1800'),...constraints('after 1800'),...constraints('on 29 February 1600')];
  for(let y=1600;y<2000;y++)for(let m=1;m<=12;m++)for(let d=1;d<=31;d++) {
    const iso=`${y}-${String(m).padStart(2,'0')}-${String(d).padStart(2,'0')}`;
    const utc=new Date(Date.UTC(y,m-1,d));
    const legal=utc.getUTCFullYear()===y&&utc.getUTCMonth()===m-1&&utc.getUTCDate()===d;
    const normalized=normalizeRecord(iso),parsed=recordDate(normalized);
    if(!legal){invalid++;assert.equal(parsed,null);continue;}
    valid++;assert.equal(parsed.key,y*10000+m*100+d);
    const english=utc.getUTCDate()+' '+utc.toLocaleString('en-US',{month:'long',timeZone:'UTC'})+' '+y;
    assert.deepEqual(parsed,recordDate(english));
    for(const r of rules)assert.equal(contradiction([r],normalized),contradiction([r],english));
  }
  assert.equal(valid,146097);assert.equal(invalid,2703);
});
test('ambiguous, mixed, invalid and timestamp dates remain unknown',()=>{
  for(const text of ['2004-03-02 and 2014-08-06','2 March 2004 and 2014-08-06',
    '31 April 2016 and 2004-03-02','2004-13-02','2004-00-02','2004-03-00',
    '1900-02-29','2004-03-02T07:17:00Z']) assert.equal(recordDate(normalizeRecord(text)),null,text);
});
test('written behavior and unsupported query behavior are unchanged',()=>{
  const rows=[{text:'2 March 2004'},{text:'6 August 2014'},{text:'no date'}];
  for(const q of ['before 2010','after 2010','on 2 March 2004','earlier than 2010','not before 2010'])
    assert.deepEqual(partition(q,rows),originalPartition(q,rows));
  const iso=rows.map(r=>({text:r.text.replace('2 March 2004','2004-03-02').replace('6 August 2014','2014-08-06')}));
  assert.deepEqual(partition('earlier than 2010',iso).ordered,iso.map(r=>({...r,contradiction:false})));
  assert.deepEqual(partition('before 2000',iso.slice(0,2)).ordered.map(r=>r.text),iso.slice(0,2).map(r=>r.text));
  assert(partition('before 2000',iso.slice(0,2)).allContradicted);
});
