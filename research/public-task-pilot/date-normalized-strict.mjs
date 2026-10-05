import {normalizeRecord as normalize} from './date-normalized.mjs';
import {constraints,contradiction} from './date-constraints.mjs';

export function normalizeRecord(text) {
  for(const m of text.matchAll(/\b\d{4}-\d{2}-\d{2}/g)) {
    const before=text[m.index-1]??'',tail=text.slice(m.index+m[0].length),after=tail[0]??'';
    if(before&&!/[\s(:"']/.test(before)) return '';
    if(after&&!/[\s.,;!?)"']/.test(after)) return '';
    if(/^\.\d|^\s+\d{2}:\d{2}/.test(tail)) return '';
  }
  return normalize(text);
}

export function partition(original,records) {
  const rules=constraints(original);
  const annotated=records.map(r=>({...r,contradiction:contradiction(rules,normalizeRecord(r.text))}));
  const allContradicted=annotated.length>0&&annotated.every(r=>r.contradiction);
  return {rules,allContradicted,ordered:allContradicted?annotated:
    [...annotated.filter(r=>!r.contradiction),...annotated.filter(r=>r.contradiction)]};
}
