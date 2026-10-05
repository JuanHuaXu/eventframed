const months = new Map(['january','february','march','april','may','june','july','august','september','october','november','december'].map((m,i)=>[m,i+1]));
const datePattern = '\\b(\\d{1,2}) (January|February|March|April|May|June|July|August|September|October|November|December) (\\d{4})\\b';

function calendar(match) {
  const d=Number(match[1]),m=months.get(match[2].toLowerCase()),y=Number(match[3]);
  const leap=y%4===0&&(y%100!==0||y%400===0);
  const days=[31,leap?29:28,31,30,31,30,31,31,30,31,30,31];
  if(y<1000||d<1||d>days[m-1]) return null;
  return {year:y,key:y*10000+m*100+d};
}

export function recordDate(text) {
  const dates=[...text.matchAll(new RegExp(datePattern,'gi'))];
  return dates.length===1?calendar(dates[0]):null;
}

export function constraints(original) {
  const pieces=original.split(', rather than ');
  if(pieces.length>2) return [];
  const positive=pieces[0];
  const out=[];
  const bounds=[...positive.matchAll(/\b(before|after) (\d{4})\b/gi)];
  const exact=[...positive.matchAll(new RegExp('\\bon '+datePattern,'gi'))];
  // This grammar cannot resolve multiple target anchors or temporal negation.
  if(bounds.length+exact.length>1||/\b(not|never|unless|except)\b/i.test(original)) return [];
  if(bounds.length) out.push({kind:bounds[0][1].toLowerCase(),value:Number(bounds[0][2])});
  if(exact.length) {
    const date=calendar(exact[0]); if(!date) return [];
    out.push({kind:'on',value:date.key});
  }
  if(pieces.length===2) {
    const suffix=pieces[1].trim().replace(/\?$/,'');
    const m=suffix.match(new RegExp('^'+datePattern+'$','i'));
    if(!m) return [];
    const date=calendar(m); if(!date) return [];
    out.push({kind:'exclude',value:date.key});
  }
  return out;
}

export function contradiction(rules,text) {
  const date=recordDate(text);
  if(!date||rules.length===0) return false;
  return rules.some(r=>r.kind==='before'?date.year>=r.value:
    r.kind==='after'?date.year<=r.value:r.kind==='on'?date.key!==r.value:
    r.kind==='exclude'?date.key===r.value:false);
}

// Keep exclusion and anchor information from the original query, not its focus rewrite.
export function partition(original,records) {
  const rules=constraints(original);
  const annotated=records.map(r=>({...r,contradiction:contradiction(rules,r.text)}));
  const allContradicted=annotated.length>0&&annotated.every(r=>r.contradiction);
  const ordered=allContradicted?annotated:[...annotated.filter(r=>!r.contradiction),...annotated.filter(r=>r.contradiction)];
  return {rules,allContradicted,ordered};
}
