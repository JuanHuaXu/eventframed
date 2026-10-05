import assert from 'node:assert/strict';

// The saved "actual" branch may be label zero or one. Weight by its label,
// not by its position in the artifact, to integrate out the purchased answer.
export function expectedQueryValue(actualY,q,actualLoss,flippedLoss){
  assert.equal(typeof actualY,'boolean');
  for(const x of [q,actualLoss,flippedLoss])assert(Number.isFinite(x)&&x>=0&&x<=1);
  const probabilityActual=actualY?q:1-q;
  return probabilityActual*actualLoss+(1-probabilityActual)*flippedLoss;
}
