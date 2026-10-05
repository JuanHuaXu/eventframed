import {recordDate,constraints,contradiction} from './date-constraints.mjs';
const months=['January','February','March','April','May','June','July','August','September','October','November','December'];
const written=/\b\d{1,2} (?:January|February|March|April|May|June|July|August|September|October|November|December) \d{4}\b/gi;

// Unknown input becomes empty evidence, never a fabricated or partially parsed date.
export function normalizeRecord(text) {
  const iso=[...text.matchAll(/\b\d{4}-\d{2}-\d{2}(?:[Tt][^\s,;!?]*)?/g)];
  if(iso.length===0) return text;
  if(iso.length!==1||[...text.matchAll(written)].length!==0) return '';
  const m=iso[0][0].match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if(!m) return '';
  const month=Number(m[2]);
  if(month<1||month>12) return '';
  const converted=`${Number(m[3])} ${months[month-1]} ${m[1]}`;
  if(!recordDate(converted)) return '';
  return converted;
}

export function partition(original,records) {
  const rules=constraints(original);
  const annotated=records.map(r=>({...r,contradiction:contradiction(rules,normalizeRecord(r.text))}));
  const allContradicted=annotated.length>0&&annotated.every(r=>r.contradiction);
  return {rules,allContradicted,ordered:allContradicted?annotated:
    [...annotated.filter(r=>!r.contradiction),...annotated.filter(r=>r.contradiction)]};
}
