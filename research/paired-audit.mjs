import assert from 'node:assert/strict';

// Declare inclusion and prediction before drawing the audit coin. A skipped
// episode supplies no outcome, even though its DR increment may be nonzero.
export function preparePairedAudit(q,prediction){
  assert(Number.isFinite(q)&&q>=.5&&q<=1);
  assert(Number.isFinite(prediction)&&prediction>=-1&&prediction<=1);
  const lower=prediction+(-1-prediction)/q,upper=prediction+(1-prediction)/q;
  return Object.freeze({lower,upper,observe(included,difference){
    assert(typeof included==='boolean');
    if(!included){assert(q<1&&difference===undefined,'skipped audit supplied an outcome');return prediction;}
    assert(Number.isFinite(difference)&&difference>=-1&&difference<=1);
    return prediction+(difference-prediction)/q;
  }});
}

export const identicalAudit=Object.freeze({lower:0,upper:0,observe(included,value){assert(included===false&&value===undefined);return 0;}});
