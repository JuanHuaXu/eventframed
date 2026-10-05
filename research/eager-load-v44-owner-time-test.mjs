import assert from 'node:assert/strict';
import {checkOwnerTime, checkReuseTime} from './eager-load-v44-audit.mjs';
const t = n => '2026-10-04T00:00:00.' + String(n).padStart(3, '0') + 'Z';
const trace = {Owner: true, Begin: t(1), OwnerAt: t(3), End: t(4)};
checkReuseTime(trace, t(0));
checkReuseTime(trace, t(2)); // Core published by another owner while this caller waited.
checkReuseTime(trace, t(3));
checkOwnerTime({Owner: false, Begin: t(1), End: t(4), OwnerAt: '0001-01-01T00:00:00Z'});
for (const bad of [{...trace, OwnerAt: t(0)}, {...trace, OwnerAt: t(5)}, {...trace, OwnerAt: undefined},
  {...trace, Owner: false}]) assert.throws(() => checkOwnerTime(bad));
assert.throws(() => checkReuseTime(trace, t(4)), 'future core must fail');
console.log(JSON.stringify({positive: 4, negative: 5, ownerWaitPublicationAccepted: true, futureCoreRejected: true}));
