import assert from 'node:assert/strict';

export function weightedRisk(losses,p){
  assert(Array.isArray(losses)&&losses.length===2&&losses.every(x=>Number.isFinite(x)&&x>=0&&x<=1));
  assert(Number.isFinite(p)&&p>=0&&p<=1);
  return (1-p)*losses[0]+p*losses[1];
}

// This selector is an oracle diagnostic: losses are not observable at decision
// time. Its only answer-weight input is probability, never a hidden teacher.
export function minimumMassRisk(actions){
  assert(Array.isArray(actions)&&actions.length>0);
  assert(new Set(actions.map(a=>a.origin)).size===actions.length);
  let chosen;
  for(const a of actions){
    assert(Number.isInteger(a.origin)&&a.origin>=-1);
    const risk=weightedRisk(a.losses,a.probability);
    if(!chosen||risk<chosen.risk||(risk===chosen.risk&&a.origin<chosen.origin))chosen={origin:a.origin,risk};
  }
  return chosen;
}
